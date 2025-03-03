package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"notification-service/internal/adapters/kafka"
	"notification-service/internal/usecases"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	loadEnv()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "notification-consumer-group"
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "notifications"
	}

	schemaRegistryURL := os.Getenv("SCHEMA_REGISTRY_URL")
	if schemaRegistryURL == "" {
		schemaRegistryURL = "http://localhost:8081"
	}

	notificationUseCase := usecases.NewNotificationUseCase(logger)

	consumerGroup, err := kafka.NewKafkaConsumerGroup(brokers, groupID, topic, schemaRegistryURL, notificationUseCase, logger)
	if err != nil {
		logger.Fatal("failed to create Kafka consumer group", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := consumerGroup.Start(ctx); err != nil {
		logger.Fatal("error during consuming", zap.Error(err))
	}

	logger.Info("Notification service consumer shut down")
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}
