package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/payment-service/shared/domain"
	"github.com/payment-service/shared/kafka"
	pb "github.com/payment-service/shared/proto/gen/proto/payment"
	"github.com/payment-service/shared/psp"
	"github.com/payment-service/shared/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/proto"
)

var tracer = otel.Tracer("worker")

type Worker interface {
	Handle(ctx context.Context, key []byte, value []byte) error
}

type WorkerHandler struct {
	psp         psp.Provider
	store       *storage.Storage
	eventRepo   storage.PaymentsEventsRepository
	commandRepo storage.ProcessedCommandsRepository
	paymentRepo storage.PaymentRepository
	producer    kafka.Producer
	eventTopic  string
}

func NewWorkerHandler(psp psp.Provider, eventRepo storage.PaymentsEventsRepository, commandRepo storage.ProcessedCommandsRepository, paymentRepo storage.PaymentRepository, producer kafka.Producer, store *storage.Storage, eventTopic string) *WorkerHandler {
	return &WorkerHandler{psp, store, eventRepo, commandRepo, paymentRepo, producer, eventTopic}
}

func (w *WorkerHandler) Handle(ctx context.Context, key []byte, value []byte) error {
	start := time.Now()
	ctx, span := tracer.Start(ctx, "Handle")
	defer span.End()
	slog.Info("Command received: ", "idempotency_key", string(key))
	cmd, err := w.deserialize(ctx, value)
	if err != nil {
		slog.Error("Error deserializing command: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	processed, err := w.checkIdempotencyKey(ctx, key)
	if err != nil {
		slog.Error("Error checking idempotency key: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if processed {
		slog.Info("Idempotency key already processed")
		return fmt.Errorf("idempotency Key is already processed")
	}
	payment, err := w.createPayment(ctx, cmd)
	if err != nil {
		slog.Error("Error creating payment: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	response, err := w.chargePSP(ctx, payment)
	if err != nil {
		slog.Error("Error charging payment: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	payment, err = w.checkStatus(ctx, response, payment)
	if err != nil {
		slog.Error("Error checking status: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	event, err := w.serializeEvent(ctx, payment)
	if err != nil {
		slog.Error("Error serializing event: ", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	err = w.store.WithTransaction(ctx, func(txdb storage.TXDB) error {
		commandRepo := storage.NewCommandRepository(txdb)
		eventRepo := storage.NewEventRepository(txdb)
		paymentRepo := storage.NewPaymentRepository(txdb)
		err = paymentRepo.Create(ctx, payment)
		slog.Debug("Starting payment creation")
		if err != nil {
			return err
		}
		err = eventRepo.Add(ctx, payment.Id, string(payment.Status), event)
		if err != nil {
			return err
		}
		err = commandRepo.Save(ctx, payment.IdempotencyKey, payment.Id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	err = w.producer.SendMessage(ctx, w.eventTopic, []byte(payment.IdempotencyKey), event)
	if err != nil {
		slog.Error("Error publishing event: ", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	slog.Info("payment processed", "payment_id", payment.Id, "status", payment.Status)
	commandDuration.Observe(time.Since(start).Seconds())
	commandsProcessed.WithLabelValues(string(payment.Status)).Inc()
	return err
}

func (w *WorkerHandler) checkIdempotencyKey(ctx context.Context, key []byte) (bool, error) {
	ctx, span := tracer.Start(ctx, "checkIdempotencyKey")
	defer span.End()
	ok, err := w.commandRepo.IsProcessed(ctx, string(key))
	if err != nil {
		return false, err
	}
	if ok {
		return true, err
	}
	return false, err
}

func (w *WorkerHandler) deserialize(ctx context.Context, value []byte) (*domain.CreatePayment, error) {
	ctx, span := tracer.Start(ctx, "Deserialize")
	defer span.End()
	cmd := &pb.CreatePaymentCommand{}
	err := proto.Unmarshal(value, cmd)
	if err != nil {
		return nil, err
	}
	return &domain.CreatePayment{
		IdempotencyKey: cmd.IdempotencyKey,
		Amount:         cmd.Amount,
		Currency:       domain.Currency(cmd.Currency),
		MerchantID:     cmd.MerchantId,
	}, nil
}

func (w *WorkerHandler) createPayment(ctx context.Context, cmd *domain.CreatePayment) (*domain.Payment, error) {
	ctx, span := tracer.Start(ctx, "createPayment")
	defer span.End()
	payment, err := domain.NewPayment(cmd)
	if err != nil {
		return nil, err
	}
	err = payment.TransitionTo(domain.PaymentStatusProcessing)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (w *WorkerHandler) chargePSP(ctx context.Context, payment *domain.Payment) (*psp.TransactionResponse, error) {
	ctx, span := tracer.Start(ctx, "chargePSP")
	defer span.End()
	request := &psp.TransactionRequest{Amount: payment.Amount, CurrencyCode: string(payment.Currency)}
	slog.Info("Charging payment request", "request", request)
	response, err := w.psp.Charge(ctx, request)
	if err != nil {
		return nil, err
	}
	slog.Info("Payment response", "response", response.Status)
	return response, nil
}

func (w *WorkerHandler) checkStatus(ctx context.Context, response *psp.TransactionResponse, payment *domain.Payment) (*domain.Payment, error) {
	ctx, span := tracer.Start(ctx, "checkStatus")
	defer span.End()
	switch response.Status {
	case psp.StatusFailed:
		err := payment.TransitionTo(domain.PaymentStatusFailed)
		if err != nil {
			return nil, err
		}
		payment.PSPTransactionID = response.PspTransactionID
		payment.ErrorMessage = response.ErrorMessage
		return payment, nil

	case psp.StatusSuccess:
		err := payment.TransitionTo(domain.PaymentStatusCompleted)
		if err != nil {
			return nil, err
		}
		payment.PSPTransactionID = response.PspTransactionID
		return payment, nil

	default:
		return nil, fmt.Errorf("failed to map status")
	}
}

func (w *WorkerHandler) serializeEvent(ctx context.Context, payment *domain.Payment) ([]byte, error) {
	ctx, span := tracer.Start(ctx, "serializeEvent")
	defer span.End()
	switch payment.Status {
	case domain.PaymentStatusCompleted:
		event := &pb.PaymentCompletedEvent{
			PaymentId:        payment.Id.String(),
			PspTransactionId: payment.PSPTransactionID,
			OccurredAt:       payment.UpdatedAt.Unix(),
		}
		return proto.Marshal(event)

	case domain.PaymentStatusFailed:
		event := &pb.PaymentFailedEvent{
			PaymentId:    payment.Id.String(),
			ErrorMessage: payment.ErrorMessage,
			OccurredAt:   payment.UpdatedAt.Unix(),
		}
		return proto.Marshal(event)

	default:
		return nil, fmt.Errorf("invalid payment status")
	}

}
