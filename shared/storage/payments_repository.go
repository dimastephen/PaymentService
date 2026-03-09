package storage

import (
	"context"

	"github.com/payment-service/shared/domain"
)

type paymentRepository struct {
	db TXDB
}

func NewPaymentRepository(db TXDB) *paymentRepository {
	return &paymentRepository{
		db: db,
	}
}

func (p *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	model := fromDomain(payment)
	query := "INSERT INTO payments(id,status,amount,currency,merchant,idempotency_key,psp_transaction_id,error_message,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)"
	_, err := p.db.Exec(ctx, query, model.Id, model.Status, model.Amount, model.Currency, model.MerchantId, model.IdempotencyKey, model.PspTransactionId, model.Error, model.CreatedAt, model.UpdatedAt)
	return err
}

func (p *paymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	model := fromDomain(payment)
	query := "UPDATE payments SET status=$1, psp_transaction_id=$2, error_message=$3, updated_at=$4 WHERE id = $5"
	_, err := p.db.Exec(ctx, query, model.Status, model.PspTransactionId, model.Error, model.UpdatedAt, model.Id)
	return err
}

func (p *paymentRepository) Get(ctx context.Context, idempotencyKey string) (*domain.Payment, error) {
	model := &PaymentsModel{}
	query := "SELECT id,status,amount,currency,merchant,idempotency_key,psp_transaction_id,created_at,updated_at,error_message FROM payments WHERE idempotency_key = $1"
	err := p.db.QueryRow(ctx, query, idempotencyKey).Scan(&model.Id, &model.Status, &model.Amount, &model.Currency, &model.MerchantId, &model.IdempotencyKey, &model.PspTransactionId, &model.CreatedAt, &model.UpdatedAt, &model.Error)
	if err != nil {
		return nil, err
	}
	payment := toDomain(model)
	return payment, err
}
