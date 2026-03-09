package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Producer interface {
	SendMessage(ctx context.Context, topic string, key []byte, value []byte) error
	Close() error
}

type SyncProducer struct {
	producer sarama.SyncProducer
}

func NewSyncProducer(brokers []string) (*SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(brokers, NewProducerConfig())
	if err != nil {
		return nil, err
	}
	return &SyncProducer{
		producer: producer,
	}, nil
}

func (p *SyncProducer) SendMessage(ctx context.Context, topic string, key []byte, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	headers := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, headers)
	for k, v := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
	}
	_, _, err := p.producer.SendMessage(msg)

	return err
}

func (p *SyncProducer) Close() error {
	return p.producer.Close()
}
