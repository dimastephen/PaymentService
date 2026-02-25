package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *eventRepository {
	return &eventRepository{pool: pool}
}

func (e *eventRepository) Add(ctx context.Context, paymentId uuid.UUID, event string, payload []byte) error {
	query := "INSERT INTO payment_events (payment_id, event_type, payload) VALUES ($1, $2, $3)"
	_, err := e.pool.Exec(ctx, query, paymentId, event, payload)
	return err
}
