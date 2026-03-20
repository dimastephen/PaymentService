package storage

import (
	"context"
)

type outboxRepository struct {
	db TXDB
}

func NewOutboxRepository(db TXDB) *outboxRepository {
	return &outboxRepository{
		db: db,
	}
}

func (o *outboxRepository) Add(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	query := "INSERT INTO outbox (topic,key,value,headers) VALUES ($1,$2,$3,$4)"
	_, err := o.db.Exec(ctx, query, topic, key, value, headers)
	if err != nil {
		return err
	}
	return nil
}

func (o *outboxRepository) GetUnsent(ctx context.Context, limit int) ([]OutboxMessage, error) {
	query := "SELECT id,topic,key,value,headers,sent,created_at FROM outbox WHERE sent=false ORDER BY id LIMIT $1"
	rows, err := o.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []OutboxMessage
	for rows.Next() {
		message := &OutboxMessage{}
		err = rows.Scan(&message.Id, &message.Topic, &message.Key, &message.Value, &message.Headers, &message.Sent, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}
	return messages, nil
}

func (o *outboxRepository) MarkSent(ctx context.Context, id int64) error {
	query := "UPDATE outbox SET sent=true WHERE id=$1"
	_, err := o.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
