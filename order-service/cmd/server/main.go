package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	orderGrpc "order-service/internal/adapters/grpc"
	"order-service/internal/adapters/kafka"
	"order-service/internal/adapters/models"
	"order-service/internal/adapters/repository"
	"order-service/internal/clients"
	"order-service/internal/domain/interfaces"
	"order-service/internal/usecases"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	loadEnv()

	models.LoadSchema("configs/order_event_schema.json")
	schemaStr := models.OrderEventSchema

	logger := createLogger()
	defer logger.Sync()

	orderRepo := buildRepository(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userSvcConn, err := grpc.DialContext(ctx,
		getEnv("USER_SERVICE_ADDRESS", "localhost:50051"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Fatal("failed to dial user service", zap.Error(err))
	}
	realUserClient := clients.NewGRPCUserServiceClient(userSvcConn)

	invAddress := getEnv("INVENTORY_SERVICE_ADDRESS", "inventory-service:30051")
	if err := waitForInventoryService(invAddress, 60*time.Second); err != nil {
		logger.Fatal("inventory service not reachable", zap.Error(err))
	}

	invConn, err := grpc.DialContext(ctx,
		invAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Fatal("failed to dial inventory service", zap.Error(err))
	}
	realInventoryClient := clients.NewGRPCInventoryServiceClient(invConn)

	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	orderTopic := getEnv("KAFKA_ORDER_TOPIC", "order-events")
	schemaRegistryURL := getEnv("SCHEMA_REGISTRY_URL", "http://localhost:8081")
	subject := "order-events-value"

	orderEventProducer, err := kafka.NewOrderEventProducer(kafkaBrokers, orderTopic, schemaRegistryURL, subject, schemaStr)
	if err != nil {
		logger.Fatal("failed to create order event producer", zap.Error(err))
	}
	defer orderEventProducer.Close()

	orderUseCase := usecases.NewOrderUsecase(orderRepo, realUserClient, realInventoryClient, orderEventProducer, logger)
	startGRPC(logger, orderUseCase)
}

func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: no .env file found, proceeding...")
	}
}

func createLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

func buildRepository(logger *zap.Logger) interfaces.IOrderRepository {
	repoType := getEnv("REPO_TYPE", "gorm")
	switch repoType {
	case "gorm":
		return buildGormRepo(logger)
	default:
		return buildGormRepo(logger)
	}
}

func buildGormRepo(logger *zap.Logger) interfaces.IOrderRepository {
	dbDriver := getEnv("DB_DRIVER", "postgres")
	db, err := connectGorm(dbDriver, logger)
	if err != nil {
		logger.Fatal("Failed to connect GORM DB", zap.Error(err))
	}
	return repository.NewGormOrderRepo(db, logger)
}

func connectGorm(driver string, logger *zap.Logger) (*gorm.DB, error) {
	switch driver {
	case "mysql":
		return connectMySQL(logger)
	default:
		return connectPostgres(logger)
	}
}

func connectPostgres(logger *zap.Logger) (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "devuser")
	pass := getEnv("DB_PASS", "devpass")
	dbname := getEnv("DB_NAME", "order_service")
	dbschema := getEnv("DB_SCHEMA", "public")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		host, port, user, pass, dbname, sslmode, dbschema)
	logger.Info("Connecting to Postgres", zap.String("dsn", dsn))

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func connectMySQL(logger *zap.Logger) (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASS", "")
	dbname := getEnv("DB_NAME", "order_service")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)
	logger.Info("Connecting to MySQL", zap.String("dsn", dsn))

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func startGRPC(logger *zap.Logger, uc interfaces.IOrderUseCase) {
	port := getEnv("GRPC_PORT", "60051")
	if err := orderGrpc.StartGRPCServer(port, uc, logger); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func waitForInventoryService(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("inventory service %s not reachable after %s", address, timeout)
}
