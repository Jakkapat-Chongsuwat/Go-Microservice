package repository

import (
	"context"
	"fmt"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
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

func TestGormOrderRepository(t *testing.T) {
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

	err = db.AutoMigrate(&models.GormDBOrder{}, &models.GormDBOrderItem{})
	require.NoError(t, err)

	repo := NewGormOrderRepo(db, logger)

	t.Run("CreateAndGetOrderWithItems_Success", func(t *testing.T) {
		order := domain.NewOrder("user-123")
		order.ID = uuid.NewString()
		item1 := domain.NewOrderItem("product-abc", 2)
		item1.ID = uuid.NewString()
		item2 := domain.NewOrderItem("product-def", 3)
		item2.ID = uuid.NewString()
		order.Items = []*domain.OrderItem{item1, item2}

		logger.Info("Creating order", zap.Any("order", order))
		created, err := repo.CreateOrderWithItems(ctx, order, order.Items)
		require.NoError(t, err)
		require.NotEmpty(t, created.ID)

		logger.Info("Fetching order", zap.String("orderID", created.ID))
		fetched, err := repo.GetOrder(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, fetched.ID)
		require.Len(t, fetched.Items, len(order.Items))
	})

	t.Run("GetOrder_NotFound", func(t *testing.T) {
		invalidUUID := uuid.NewString()
		fetched, err := repo.GetOrder(ctx, invalidUUID)
		require.Error(t, err)
		require.Nil(t, fetched)
	})

	t.Run("GetOrdersByUserID_Success", func(t *testing.T) {
		order2 := domain.NewOrder("user-xyz")
		order2.ID = uuid.NewString()
		order2.Items = []*domain.OrderItem{
			domain.NewOrderItem("prod-1", 1),
		}
		order2.Items[0].ID = uuid.NewString()

		order3 := domain.NewOrder("user-xyz")
		order3.ID = uuid.NewString()
		order3.Status = domain.OrderStatusPaid
		order3.Items = []*domain.OrderItem{
			domain.NewOrderItem("prod-2", 4),
		}
		order3.Items[0].ID = uuid.NewString()

		for _, o := range []*domain.Order{order2, order3} {
			_, err := repo.CreateOrderWithItems(ctx, o, o.Items)
			require.NoError(t, err)
		}

		fetched, err := repo.GetOrdersByUserID(ctx, "user-xyz")
		require.NoError(t, err)
		require.Len(t, fetched, 2)

		ids := []string{fetched[0].ID, fetched[1].ID}
		require.ElementsMatch(t, []string{order2.ID, order3.ID}, ids)
	})
}
