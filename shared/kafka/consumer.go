package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type Handler func(ctx context.Context, key []byte, value []byte) error

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type consumer struct {
	brokers       []string
	group         string
	topic         string
	fn            Handler
	consumerGroup sarama.ConsumerGroup
}

func NewConsumer(brokers []string, group string, topic string, fn Handler) *consumer {
	return &consumer{
		brokers: brokers,
		group:   group,
		topic:   topic,
		fn:      fn,
	}
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ch := claim.Messages()
	for v := range ch {
		err := c.fn(session.Context(), v.Key, v.Value)
		if err != nil {
			fmt.Printf("Error on kafka message: %s\n", err)
			continue
		}
		session.MarkMessage(v, "")
	}
	return nil
}

func (c *consumer) Start(ctx context.Context) error {
	group, err := sarama.NewConsumerGroup(c.brokers, c.group, NewConsumerConfig())
	if err != nil {
		return err
	}
	c.consumerGroup = group
	for {
		err = group.Consume(ctx, []string{c.topic}, c)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *consumer) Close() error {
	return c.consumerGroup.Close()
}
