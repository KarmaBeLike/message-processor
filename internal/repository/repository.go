package repository

import (
	"context"

	"github.com/KarmaBeLike/message-processor/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

type MessageRepository struct {
	DB *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) SaveMessage(ctx context.Context, message *models.Message) error {
	_, err := r.DB.Exec(ctx, "INSERT INTO messages (content, processed) VALUES ($1, $2)", message.Content, message.Processed)
	return err
}

func (r *MessageRepository) GetProcessedMessagesCount(ctx context.Context) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE processed=true").Scan(&count)
	return count, err
}

func (r *MessageRepository) MarkMessageAsProcessed(ctx context.Context, content string) error {
	_, err := r.DB.Exec(ctx, "UPDATE messages SET processed=true WHERE content=$1", content)
	return err
}
