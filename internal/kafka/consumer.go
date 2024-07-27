package kafka

import (
	"context"
	"log"

	"github.com/KarmaBeLike/message-processor/internal/repository"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Reader *kafka.Reader
	Repo   *repository.MessageRepository
}

func NewKafkaConsumer(brokers []string, topic string, groupID string, repo *repository.MessageRepository) *KafkaConsumer {
	return &KafkaConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		Repo: repo,
	}
}

func (c *KafkaConsumer) StartConsuming(ctx context.Context) {
	for {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("could not read message: %v", err)
			continue
		}
		log.Printf("received: %s", string(msg.Value))

		// Пометить сообщение как обработанное
		if err := c.Repo.MarkMessageAsProcessed(ctx, string(msg.Value)); err != nil {
			log.Printf("could not mark message as processed: %v", err)
		}
	}
}
