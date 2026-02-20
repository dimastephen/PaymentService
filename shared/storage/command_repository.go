package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type commandRepository struct {
	pool *pgxpool.Pool
}

func NewCommandRepository(pool *pgxpool.Pool) *commandRepository {
	return &commandRepository{
		pool: pool,
	}
}

func (c *commandRepository) IsProcessed(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM processed_commands WHERE idempotency_key = $1)"
	err := c.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *commandRepository) Save(ctx context.Context, idempotencyKey string, paymentId uuid.UUID) error {
	query := "INSERT INTO processed_commands(idempotency_key, payment_id) VALUES ($1, $2)"
	_, err := c.pool.Exec(ctx, query, idempotencyKey, paymentId)
	return err
}
