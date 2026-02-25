package storage

import (
	"context"

	"github.com/google/uuid"
)

type commandRepository struct {
	db TXDB
}

func NewCommandRepository(db TXDB) *commandRepository {
	return &commandRepository{
		db: db,
	}
}

func (c *commandRepository) IsProcessed(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM processed_commands WHERE idempotency_key = $1)"
	err := c.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *commandRepository) Save(ctx context.Context, idempotencyKey string, paymentId uuid.UUID) error {
	query := "INSERT INTO processed_commands(idempotency_key, payment_id) VALUES ($1, $2)"
	_, err := c.db.Exec(ctx, query, idempotencyKey, paymentId)
	return err
}
