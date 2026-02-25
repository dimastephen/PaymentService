package storage

import (
	"time"

	"github.com/google/uuid"
)

type PaymentsModel struct {
	Id               uuid.UUID `db:"id"`
	Status           string    `db:"status"`
	Amount           int64     `db:"amount"`
	Currency         string    `db:"currency"`
	MerchantId       string    `db:"merchant_id"`
	IdempotencyKey   string    `db:"idempotency_key"`
	PspTransactionId string    `db:"psp_transaction_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Error            string    `db:"error_message"`
}
