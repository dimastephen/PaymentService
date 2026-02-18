package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentCreated struct {
	Id             uuid.UUID
	Amount         int64
	Currency       Currency
	IdempotencyKey string
	MerchantID     string
	OccurredAt     time.Time
}

type PaymentFailed struct {
	Id           uuid.UUID
	ErrorMessage string
	OccurredAt   time.Time
}

type PaymentCancelled struct {
	Id           uuid.UUID
	ErrorMessage string
	OccurredAt   time.Time
}

type PaymentCompleted struct {
	Id               uuid.UUID
	PSPTransactionID string
	OccurredAt       time.Time
}
