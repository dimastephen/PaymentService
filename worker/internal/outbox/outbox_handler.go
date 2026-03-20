package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/payment-service/shared/kafka"
	"github.com/payment-service/shared/storage"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("outbox")

type OutboxHandler struct {
	producer   kafka.Producer
	outboxRepo storage.OutboxRepository
}

func NewOutboxHandler(producer kafka.Producer, outboxRepo storage.OutboxRepository) *OutboxHandler {
	return &OutboxHandler{
		producer:   producer,
		outboxRepo: outboxRepo,
	}
}

func (o *OutboxHandler) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			ctx, span := tracer.Start(ctx, "OutboxHandler")
			messages, err := o.outboxRepo.GetUnsent(ctx, 50)
			if err != nil {
				slog.Error("Failed to get unsent outboxes", "err", err)
				continue
			}
			for _, message := range messages {
				start := time.Now()
				err = o.producer.SendMessage(ctx, message.Topic, message.Key, message.Value, message.Headers)
				if err != nil {
					slog.Error("Failed to send message", "err", err, "topic", message.Topic, "key", message.Key, "id", message.Id)
					continue
				}
				err = o.outboxRepo.MarkSent(ctx, message.Id)
				if err != nil {
					slog.Error("Failed to mark sent", "err", err, "topic", message.Topic, "key", message.Key, "id", message.Id)
					continue
				}
				eventsProcessed.WithLabelValues(message.Headers["event_type"]).Inc()
				eventsDuration.Observe(time.Since(start).Seconds())
			}
			span.End()
		}
	}
}
