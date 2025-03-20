package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/KarmaBeLike/message-processor/config"
	"github.com/KarmaBeLike/message-processor/internal/api"
	"github.com/KarmaBeLike/message-processor/internal/database"
	"github.com/KarmaBeLike/message-processor/internal/kafka"

	"github.com/KarmaBeLike/message-processor/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Println("config", cfg)

	dbpool, err := database.ConnectDB(cfg)
	if err != nil {
		slog.Error("failed to connect to db", slog.Any("error", err))
		return
	}
	defer dbpool.Close()

	if err := database.RunMigrations(dbpool); err != nil {
		slog.Error("error running migrations", slog.Any("error", err))
		return
	}

	messageRepo := repository.NewMessageRepository(dbpool)
	kafkaProducer := kafka.NewKafkaProducer([]string{cfg.KafkaHost + ":" + cfg.KafkaPort}, cfg.Topic)
	kafkaConsumer := kafka.NewKafkaConsumer([]string{cfg.KafkaHost + ":" + cfg.KafkaPort}, cfg.Topic, cfg.Group, messageRepo)

	router := gin.Default()

	messageHandler := api.NewMessageHandler(messageRepo, kafkaProducer)

	router.POST("/messages", messageHandler.PostMessage)
	router.GET("/statistics", messageHandler.GetStatistics)

	go kafkaConsumer.StartConsuming(context.Background())

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server started")

	// Создаем канал для получения сигналов прерывания
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Блокируем выполнение до получения сигнала
	<-quit
	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для завершения сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершаем работу сервера
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}
