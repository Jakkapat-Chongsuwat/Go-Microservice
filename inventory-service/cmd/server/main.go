package main

import (
	"fmt"
	"log"
	"os"
	"time"

	fiber_http "inventory-service/internal/adapters/fiber"
	inventoryGrpc "inventory-service/internal/adapters/grpc"
	"inventory-service/internal/adapters/repository"
	"inventory-service/internal/usecases"

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

	inventoryRepo := buildRepository(logger)
	inventoryUseCase := usecases.NewInventoryUsecase(inventoryRepo, logger)

	go startGRPC(logger, inventoryUseCase)
	startHTTP(logger, inventoryUseCase)
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

func buildRepository(logger *zap.Logger) usecases.InventoryRepository {
	repoType := getEnv("REPO_TYPE", "gorm")
	switch repoType {
	case "gorm":
		return buildGormRepo(logger)
	default:
		return buildGormRepo(logger)
	}
}

func buildGormRepo(logger *zap.Logger) usecases.InventoryRepository {
	dbDriver := getEnv("DB_DRIVER", "postgres")
	db, err := connectGorm(dbDriver, logger)
	if err != nil {
		logger.Fatal("Failed to connect GORM DB", zap.Error(err))
	}
	return repository.NewGormInventoryRepo(db, logger)
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
	dbname := getEnv("DB_NAME", "inventory_service")
	dbschema := getEnv("DB_SCHEMA", "public")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		host, port, user, pass, dbname, sslmode, dbschema)
	logger.Info("Connecting to Postgres", zap.String("dsn", dsn))

	var db *gorm.DB
	var err error
	for i := 0; i < 15; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		logger.Warn("Postgres not ready yet, retrying...", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to Postgres after multiple attempts: %w", err)
}

func connectMySQL(logger *zap.Logger) (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASS", "")
	dbname := getEnv("DB_NAME", "inventory_service")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)
	logger.Info("Connecting to MySQL", zap.String("dsn", dsn))

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func startGRPC(logger *zap.Logger, uc usecases.InventoryUseCase) {
	port := getEnv("GRPC_PORT", "30051")
	if err := inventoryGrpc.StartGRPCServer(port, uc, logger); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}
}

func startHTTP(logger *zap.Logger, uc usecases.InventoryUseCase) {
	app := fiber.New()
	handler := fiber_http.NewInventoryHTTPHandler(uc, logger)
	fiber_http.RegisterInventoryRoutes(app, handler)

	port := getEnv("HTTP_PORT", "30052")
	logger.Info("Starting Inventory HTTP server", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
