package main

import (
	"fmt"
	"log"
	"os"
	"time"
	fiber_http "user-service/internal/adapters/fiber"
	"user-service/internal/adapters/grpc"
	"user-service/internal/adapters/repository"
	"user-service/internal/usecases"

	"github.com/gofiber/fiber/v2"
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

	go startGRPC(logger, userUsecase)

	startHTTP(logger, userUsecase)
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
	dbschema := getEnv("DB_SCHEMA", "public")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		host, port, user, pass, dbname, sslmode, dbschema)
	logger.Info("Connecting to Postgres", zap.String("dsn", dsn))

	var db *gorm.DB
	var err error
	maxRetries := 15

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		logger.Warn("Postgres not ready, retrying...", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to Postgres after %d attempts: %w", maxRetries, err)
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

func startGRPC(logger *zap.Logger, u usecases.UserUseCase) {
	port := getEnv("GRPC_PORT", "50051")
	if err := grpc.StartGRPCServer(port, u, logger); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}
}

func startHTTP(logger *zap.Logger, u usecases.UserUseCase) {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("User Service is running")
	})

	fiber_http.RegisterUserRoutes(app, fiber_http.NewUserHttpHandler(u, logger))

	port := getEnv("HTTP_PORT", "8080")
	logger.Info("Starting HTTP server on port", zap.String("port", port))

	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
