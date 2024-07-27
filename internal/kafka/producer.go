package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer(addr []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(addr...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) ProduceMessage(ctx context.Context, message []byte) error {
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
	if err != nil {
		log.Printf("could not write message %v: %v", message, err)
	}
	return err
}
