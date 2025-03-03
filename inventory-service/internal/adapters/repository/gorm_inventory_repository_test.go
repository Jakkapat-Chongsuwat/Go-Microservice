package repository

import (
	"context"
	"fmt"
	"inventory-service/internal/domain"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormInventoryRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:15-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testDb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=test password=test dbname=testDb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(60 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=testDb sslmode=disable", host, mappedPort.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Product{})
	require.NoError(t, err)

	repo := NewGormInventoryRepo(db, logger)

	t.Run("CreateAndGetProduct_Success", func(t *testing.T) {
		product := domain.NewProduct("Test Product", 50, 19.99)
		created, err := repo.CreateProduct(ctx, product)
		require.NoError(t, err)
		require.NotEmpty(t, created.ID)

		fetched, err := repo.GetProduct(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, fetched.ID)
		require.Equal(t, created.Name, fetched.Name)
	})

	t.Run("UpdateProduct_Success", func(t *testing.T) {
		product := domain.NewProduct("Update Test", 100, 29.99)
		created, err := repo.CreateProduct(ctx, product)
		require.NoError(t, err)

		err = created.AdjustStock(-20)
		require.NoError(t, err)
		updated, err := repo.UpdateProduct(ctx, created)
		require.NoError(t, err)
		require.Equal(t, 80, updated.Quantity)
	})

	t.Run("ListProducts_Success", func(t *testing.T) {
		p1 := domain.NewProduct("List Product 1", 10, 9.99)
		p2 := domain.NewProduct("List Product 2", 20, 14.99)
		_, err := repo.CreateProduct(ctx, p1)
		require.NoError(t, err)
		_, err = repo.CreateProduct(ctx, p2)
		require.NoError(t, err)

		products, err := repo.ListProducts(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(products), 2)
	})

	t.Run("GetProduct_NotFound", func(t *testing.T) {
		nonExistingID := uuid.NewString()
		fetched, err := repo.GetProduct(ctx, nonExistingID)
		require.Error(t, err)
		require.Nil(t, fetched)
	})

	t.Run("ListProducts_Empty", func(t *testing.T) {
		err = db.Exec("DELETE FROM products").Error
		require.NoError(t, err)

		products, err := repo.ListProducts(ctx)
		require.NoError(t, err)
		require.Len(t, products, 0)
	})
}
