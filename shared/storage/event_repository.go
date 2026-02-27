package storage

import (
	"context"

	"github.com/google/uuid"
)

type eventRepository struct {
	db TXDB
}

func NewEventRepository(db TXDB) *eventRepository {
	return &eventRepository{db: db}
}

func (e *eventRepository) Add(ctx context.Context, paymentId uuid.UUID, event string, payload []byte) error {
	query := "INSERT INTO payment_events (payment_id, event_type, payload) VALUES ($1, $2, $3)"
	_, err := e.db.Exec(ctx, query, paymentId, event, payload)
	return err
}
