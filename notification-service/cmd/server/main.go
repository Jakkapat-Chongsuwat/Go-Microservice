package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"notification-service/internal/adapters/kafka"
	ws "notification-service/internal/adapters/websocket"
	"notification-service/internal/usecases"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type kafkaConfig struct {
	brokers []string
	groupID string
	topic   string
}

func main() {
	loadEnv()
	logger := initLogger()
	defer logger.Sync()

	hub := ws.NewHub()
	startWebSocketServer(logger, hub)

	kafkaCfg := getKafkaConfig()
	schemaRegistryURL := getSchemaRegistryURL()

	notificationUseCase := usecases.NewNotificationUseCase(logger, hub)

	consumerGroup := createKafkaConsumerGroup(kafkaCfg, schemaRegistryURL, notificationUseCase, logger)
	runConsumerGroup(consumerGroup, logger)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

func initLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	return logger
}

func getKafkaConfig() kafkaConfig {
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

	return kafkaConfig{brokers: brokers, groupID: groupID, topic: topic}
}

func getSchemaRegistryURL() string {
	url := os.Getenv("SCHEMA_REGISTRY_URL")
	if url == "" {
		url = "http://localhost:8081"
	}
	return url
}

func getWsPort() string {
	port := os.Getenv("WS_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func startWebSocketServer(logger *zap.Logger, hub *ws.Hub) {
	wsPort := getWsPort()
	go func() {
		http.HandleFunc("/ws", wsHandler(hub))
		logger.Info("WebSocket server listening", zap.String("port", wsPort))
		if err := http.ListenAndServe(":"+wsPort, nil); err != nil {
			logger.Fatal("WebSocket server failed", zap.Error(err))
		}
	}()
}

func createKafkaConsumerGroup(cfg kafkaConfig, schemaRegistryURL string, useCase usecases.NotificationUseCase, logger *zap.Logger) *kafka.KafkaConsumerGroup {
	consumerGroup, err := kafka.NewKafkaConsumerGroup(cfg.brokers, cfg.groupID, cfg.topic, schemaRegistryURL, useCase, logger)
	if err != nil {
		logger.Fatal("failed to create Kafka consumer group", zap.Error(err))
	}
	return consumerGroup
}

func runConsumerGroup(consumerGroup *kafka.KafkaConsumerGroup, logger *zap.Logger) {
	ctx := context.Background()
	if err := consumerGroup.Start(ctx); err != nil {
		logger.Fatal("error during consuming", zap.Error(err))
	}
	logger.Info("Notification service consumer shut down")
}

func wsHandler(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("failed to upgrade connection:", err)
			return
		}
		hub.Add(conn)
		defer hub.Remove(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}
