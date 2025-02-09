// user-service\cmd\server\main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"user-service/internal/adapters/grpc"
	"user-service/internal/adapters/repository"
	"user-service/internal/usecases"

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

	repo := buildRepository(logger)
	userUsecase := usecases.NewUserUseCase(repo, logger)

	demoUseCase(logger, userUsecase)
	startGRPC(logger, userUsecase)
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

func buildRepository(logger *zap.Logger) usecases.UserRepository {
	repoType := getEnv("REPO_TYPE", "memory")
	switch repoType {
	case "gorm":
		return buildGormRepo(logger)
	default:
		logger.Info("Using In-Memory Repository (default)")
		return repository.NewInMemoryUserRepo(logger)
	}
}

func buildGormRepo(logger *zap.Logger) usecases.UserRepository {
	dbDriver := getEnv("DB_DRIVER", "postgres")
	db, err := connectGorm(dbDriver, logger)
	if err != nil {
		logger.Fatal("Failed to connect GORM DB", zap.Error(err))
	}
	return repository.NewGormUserRepo(db, logger)
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
	port := getEnv("DB_PORT", "5555")
	user := getEnv("DB_USER", "devuser")
	pass := getEnv("DB_PASS", "devpass")
	dbname := getEnv("DB_NAME", "user_service")
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
	dbname := getEnv("DB_NAME", "user_service")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)
	logger.Info("Connecting to MySQL", zap.String("dsn", dsn))

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func demoUseCase(logger *zap.Logger, u usecases.UserUseCase) {
	fmt.Println("Running demo use case")
	ctx := context.Background()

	if user, err := u.CreateUser(ctx, "1", "Alice", "alice@example.com"); err == nil {
		logger.Info("Created user1", zap.Any("user", user))
	}
	if user, err := u.CreateUser(ctx, "2", "Bob", "bob@example.com"); err == nil {
		logger.Info("Created user2", zap.Any("user", user))
	}
	if user, err := u.CreateUser(ctx, "3", "Charlie", "charlie@example.com"); err == nil {
		logger.Info("Created user3", zap.Any("user", user))
	}

	ids := []string{"1", "2", "3"}
	users, err := u.GetUserInParallel(ctx, ids)
	if err != nil {
		logger.Error("Failed parallel retrieval", zap.Strings("ids", ids), zap.Error(err))
		return
	}
	fmt.Printf("Got users in parallel:\n")
	for _, usr := range users {
		logger.Info("User in parallel", zap.Any("user", usr))
	}
}

func startGRPC(logger *zap.Logger, u usecases.UserUseCase) {
	if err := grpc.StartGRPCServer("50051", u, logger); err != nil {
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
