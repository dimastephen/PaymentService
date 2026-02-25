package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/payment-service/shared/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	Update(ctx context.Context, payment *domain.Payment) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
}

type PaymentsEventsRepository interface {
	Add(ctx context.Context, paymentId uuid.UUID, event string, payload []byte) error
}

type ProcessedCommandsRepository interface {
	IsProcessed(ctx context.Context, id string) (bool, error)
	Save(ctx context.Context, idempotencyKey string, paymentId uuid.UUID) error
}
