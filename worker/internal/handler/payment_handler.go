package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/payment-service/shared/domain"
	"github.com/payment-service/shared/kafka"
	"github.com/payment-service/shared/psp"
	"github.com/payment-service/shared/storage"
)

type Worker interface {
	Handle(ctx context.Context, key []byte, value []byte) error
}

type WorkerHandler struct {
	psp         psp.Provider
	store       storage.Storage
	eventRepo   storage.PaymentsEventsRepository
	commandRepo storage.ProcessedCommandsRepository
	paymentRepo storage.PaymentRepository
	producer    kafka.Producer
}

func NewWorkerHandler(psp psp.Provider, eventRepo storage.PaymentsEventsRepository, commandRepo storage.ProcessedCommandsRepository, paymentRepo storage.PaymentRepository, producer kafka.Producer, store storage.Storage) *WorkerHandler {
	return &WorkerHandler{psp, store, eventRepo, commandRepo, paymentRepo, producer}
}

func (w *WorkerHandler) Handle(ctx context.Context, key []byte, value []byte) error {
	cmd, err := w.deserialize(ctx, value)
	if err != nil {
		return err
	}
	processed, err := w.checkIdempotencyKey(ctx, key)
	if err != nil {
		return err
	}
	if processed {
		return fmt.Errorf("idempotency Key is already processed")
	}
	payment, err := w.createPayment(ctx, cmd)
	if err != nil {
		return err
	}
	response, err := w.chargePSP(ctx, payment)
	if err != nil {
		return err
	}
	payment, err = w.checkStatus(ctx, response, payment)
	if err != nil {
		return err
	}
	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(payment)
	if err != nil {
		return err
	}
	err = w.store.WithTransaction(ctx, func(txdb storage.TXDB) error {
		commandRepo := storage.NewCommandRepository(txdb)
		eventRepo := storage.NewEventRepository(txdb)
		paymentRepo := storage.NewPaymentRepository(txdb)
		err = commandRepo.Save(ctx, payment.IdempotencyKey, payment.Id)
		if err != nil {
			return err
		}

		err = eventRepo.Add(ctx, payment.Id, string(payment.Status), buff.Bytes())
		if err != nil {
			return err
		}
		err = paymentRepo.Create(ctx, payment)
		if err != nil {
			return err
		}
		return nil

	})
	err = w.producer.SendMessage(ctx, "payment-events", []byte(payment.IdempotencyKey), buff.Bytes())
	if err != nil {
		return err
	}
	return err
}

func (w *WorkerHandler) checkIdempotencyKey(ctx context.Context, key []byte) (bool, error) {
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
	cmd := &domain.CreatePayment{}
	decoder := json.NewDecoder(bytes.NewBuffer(value))
	if err := decoder.Decode(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

func (w *WorkerHandler) createPayment(ctx context.Context, cmd *domain.CreatePayment) (*domain.Payment, error) {
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
	request := &psp.TransactionRequest{Amount: payment.Amount, CurrencyCode: string(payment.Currency)}
	response, err := w.psp.Charge(ctx, request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (w *WorkerHandler) checkStatus(ctx context.Context, response *psp.TransactionResponse, payment *domain.Payment) (*domain.Payment, error) {
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
