package main

import (
	"fmt"
	"log"
	"order-service/internal/adapters/grpc"
	"order-service/internal/adapters/repository"
	"order-service/internal/usecases"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	loadEnv()

	logger := createLogger()
	defer logger.Sync()

	orderRepo := buildRepository(logger)

	orderUseCase := usecases.NewOrderUsecase(orderRepo, nil, nil, logger)

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

func buildRepository(logger *zap.Logger) usecases.OrderRepository {
	repoType := getEnv("REPO_TYPE", "gorm")
	switch repoType {
	case "gorm":
		return buildGormRepo(logger)
	default:
		return buildGormRepo(logger)
	}
}

func buildGormRepo(logger *zap.Logger) usecases.OrderRepository {
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
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbname, sslmode)
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

func startGRPC(logger *zap.Logger, uc usecases.OrderUseCase) {
	if err := grpc.StartGRPCServer("60051", uc, logger); err != nil {
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
