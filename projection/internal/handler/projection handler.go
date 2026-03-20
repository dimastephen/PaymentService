package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/payment-service/shared/domain"
	pb "github.com/payment-service/shared/proto/gen/proto/payment"
	"github.com/payment-service/shared/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/proto"
)

var tracer = otel.Tracer("projection")

type ProjectionHandler struct {
	paymentRepo storage.PaymentRepository
}

func NewProjectionHandler(paymentRepo storage.PaymentRepository) *ProjectionHandler {
	return &ProjectionHandler{
		paymentRepo: paymentRepo,
	}
}

func (p *ProjectionHandler) Handle(ctx context.Context, key []byte, value []byte, headers map[string]string) error {
	start := time.Now()
	ctx, span := tracer.Start(ctx, "Projection")
	defer span.End()
	eventType, ok := headers["event_type"]
	if !ok {
		slog.Error("event_type header not found")
		span.RecordError(fmt.Errorf("event_type header not found"))
		span.SetStatus(codes.Error, "ProjectionHandler")
		return errors.New("header not found")
	}
	slog.Info("Event received: ", "event_type", eventType)

	switch eventType {
	case "payment_created":
		payment, err := p.deserializeCreation(ctx, value)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		err = p.paymentRepo.Create(ctx, payment)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		eventsDuration.Observe(time.Since(start).Seconds())
		eventsProcessed.WithLabelValues(eventType).Inc()
		return nil

	default:
		payment, err := p.deserializeStatus(ctx, value, eventType)
		if err != nil {
			slog.Error("error deserializing payment status", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		err = p.paymentRepo.Update(ctx, payment)
		if err != nil {
			slog.Error("error updating payment", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		eventsDuration.Observe(time.Since(start).Seconds())
		eventsProcessed.WithLabelValues(eventType).Inc()
		return nil
	}
}

func (p *ProjectionHandler) deserializeCreation(ctx context.Context, value []byte) (*domain.Payment, error) {
	ctx, span := tracer.Start(ctx, "deserialize event")
	defer span.End()
	paymentCreated := &pb.PaymentCreatedEvent{}
	err := proto.Unmarshal(value, paymentCreated)
	if err != nil {
		return nil, err
	}
	payment := &domain.Payment{
		Id:             uuid.MustParse(paymentCreated.PaymentId),
		Amount:         paymentCreated.Amount,
		Status:         domain.PaymentStatusProcessing,
		Currency:       domain.Currency(paymentCreated.Currency),
		IdempotencyKey: paymentCreated.IdempotencyKey,
		MerchantID:     paymentCreated.MerchantId,
		CreatedAt:      time.Unix(paymentCreated.OccuredAt, 0),
	}
	val := &domain.CreatePayment{
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		IdempotencyKey: payment.IdempotencyKey,
		MerchantID:     payment.MerchantID,
	}
	_, err = domain.NewPayment(val)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (p *ProjectionHandler) deserializeStatus(ctx context.Context, value []byte, status string) (*domain.Payment, error) {
	ctx, span := tracer.Start(ctx, "Create")
	defer span.End()
	switch status {
	case string(domain.PaymentStatusCompleted):
		completed := &pb.PaymentCompletedEvent{}
		err := proto.Unmarshal(value, completed)
		if err != nil {
			return nil, err
		}
		payment := &domain.Payment{
			Id:               uuid.MustParse(completed.PaymentId),
			PSPTransactionID: completed.PspTransactionId,
			UpdatedAt:        time.Unix(completed.OccurredAt, 0),
			Status:           domain.PaymentStatusCompleted,
		}
		return payment, nil

	case string(domain.PaymentStatusFailed):
		failed := &pb.PaymentFailedEvent{}
		err := proto.Unmarshal(value, failed)
		if err != nil {
			return nil, err
		}
		payment := &domain.Payment{
			Id:           uuid.MustParse(failed.PaymentId),
			ErrorMessage: failed.ErrorMessage,
			UpdatedAt:    time.Unix(failed.OccurredAt, 0),
			Status:       domain.PaymentStatusFailed,
		}
		return payment, nil
	default:
		return nil, fmt.Errorf("unknown status: %s", status)
	}
}
