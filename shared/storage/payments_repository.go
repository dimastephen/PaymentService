package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/payment-service/shared/domain"
)

type paymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *paymentRepository {
	return &paymentRepository{
		pool: pool,
	}
}

func (p *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	model := fromDomain(payment)
	query := "INSERT INTO payments(id,status,amount,currency,merchant,idempotency_key) VALUES ($1,$2,$3,$4,$5,$6)"
	_, err := p.pool.Exec(ctx, query, model.Id, model.Status, model.Amount, model.Currency, model.MerchantId, model.IdempotencyKey)
	return err
}

func (p *paymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	model := fromDomain(payment)
	query := "UPDATE payments SET status=$1, psp_transaction_id=$2, error_message=$3, updated_at=$4 WHERE id = $5"
	_, err := p.pool.Exec(ctx, query, model.Status, model.PspTransactionId, model.Error, model.UpdatedAt, model.Id)
	return err
}

func (p *paymentRepository) Get(ctx context.Context, paymentId uuid.UUID) (*domain.Payment, error) {
	model := &PaymentsModel{}
	query := "SELECT id,status,amount,currency,merchant_id,idempotency_key,psp_transaction_id,created_at,updated_at,error_message FROM payments WHERE id = $1"
	err := p.pool.QueryRow(ctx, query, paymentId).Scan(&model.Id, &model.Status, &model.Amount, &model.Currency, &model.MerchantId, &model.IdempotencyKey, &model.PspTransactionId, &model.CreatedAt, &model.UpdatedAt, &model.Error)
	if err != nil {
		return nil, err
	}
	payment := toDomain(model)
	return payment, err
}
