package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/KarmaBeLike/message-processor/config"
	"github.com/KarmaBeLike/message-processor/internal/api"
	"github.com/KarmaBeLike/message-processor/internal/kafka"

	"github.com/KarmaBeLike/message-processor/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Println("config", cfg)

	dbpool, err := pgxpool.Connect(context.Background(),
		fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
			cfg.DBConfig.DBUser, cfg.DBConfig.DBPassword, cfg.DBConfig.DBHost, cfg.DBConfig.DBPort, cfg.DBConfig.DBName))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbpool.Close()

	messageRepo := repository.NewMessageRepository(dbpool)
	kafkaProducer := kafka.NewKafkaProducer([]string{cfg.KafkaConfig.KafkaHost + ":" + cfg.KafkaConfig.KafkaPort}, cfg.KafkaConfig.Topic)
	kafkaConsumer := kafka.NewKafkaConsumer([]string{cfg.KafkaConfig.KafkaHost + ":" + cfg.KafkaConfig.KafkaPort}, cfg.KafkaConfig.Topic, cfg.KafkaConfig.Group, messageRepo)

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
		if err := srv.ListenAndServe(); err != nil /*&& err != http.ErrServerClosed*/ {
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
