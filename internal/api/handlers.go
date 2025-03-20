package api

import (
	"context"
	"log"
	"net/http"

	"github.com/KarmaBeLike/message-processor/internal/kafka"
	"github.com/KarmaBeLike/message-processor/internal/models"
	"github.com/KarmaBeLike/message-processor/internal/repository"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	Repo     *repository.MessageRepository
	Producer *kafka.KafkaProducer
}

func NewMessageHandler(repo *repository.MessageRepository /*, producer *kafka.KafkaProducer*/) *MessageHandler {
	return &MessageHandler{
		Repo: repo,
		// Producer: producer,
	}
}

func (h *MessageHandler) PostMessage(c *gin.Context) {
	log.Println("PostMessage")
	var message models.Message
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	message.Processed = false
	err := h.Repo.SaveMessage(context.Background(), &message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = h.Producer.ProduceMessage(context.Background(), []byte(message.Content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *MessageHandler) GetStatistics(c *gin.Context) {
	count, err := h.Repo.GetProcessedMessagesCount(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"processed_messages_count": count})
}
