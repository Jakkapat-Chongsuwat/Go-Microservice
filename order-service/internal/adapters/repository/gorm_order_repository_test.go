package repository

import (
	"context"
	"fmt"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormOrderRepository(t *testing.T) {
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

	logger, _ := zap.NewDevelopment()
	repo := NewGormOrderRepo(db, logger)

	t.Run("CreateAndGetOrder_Success", func(t *testing.T) {
		order := &domain.Order{
			ID:     "order-1",
			UserID: "user-123",
			Status: domain.OrderStatusCreated,
			Items: []*domain.OrderItem{
				{
					ProductID: "product-abc",
					Quantity:  2,
				},
			},
		}

		created, err := repo.CreateOrder(ctx, order)
		require.NoError(t, err)
		require.Equal(t, order.ID, created.ID)
		require.Equal(t, order.UserID, created.UserID)
		require.Equal(t, order.Status, created.Status)

		fetched, err := repo.GetOrder(ctx, order.ID)
		require.NoError(t, err)
		require.Equal(t, order.ID, fetched.ID)
		require.Equal(t, order.UserID, fetched.UserID)
		require.Equal(t, order.Status, fetched.Status)
		require.Len(t, fetched.Items, 1)
		require.Equal(t, order.Items[0].ProductID, fetched.Items[0].ProductID)
		require.Equal(t, order.Items[0].Quantity, fetched.Items[0].Quantity)
	})

	t.Run("GetOrder_NotFound", func(t *testing.T) {
		fetched, err := repo.GetOrder(ctx, "unknown-order-id")
		require.Error(t, err)
		require.Nil(t, fetched)
	})

	t.Run("GetOrdersByUserID_Success", func(t *testing.T) {
		orders := []*domain.Order{
			{
				ID:     "order-2",
				UserID: "user-xyz",
				Status: domain.OrderStatusCreated,
				Items: []*domain.OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  1,
					},
				},
			},
			{
				ID:     "order-3",
				UserID: "user-xyz",
				Status: domain.OrderStatusPaid,
				Items: []*domain.OrderItem{
					{
						ProductID: "prod-2",
						Quantity:  4,
					},
				},
			},
		}

		for _, o := range orders {
			_, err := repo.CreateOrder(ctx, o)
			require.NoError(t, err)
		}

		_, err := repo.CreateOrder(ctx, &domain.Order{
			ID:     "order-4",
			UserID: "user-other",
			Status: domain.OrderStatusCreated,
			Items: []*domain.OrderItem{
				{
					ProductID: "prod-9",
					Quantity:  10,
				},
			},
		})
		require.NoError(t, err)

		fetched, err := repo.GetOrdersByUserID(ctx, "user-xyz")
		require.NoError(t, err)
		require.Len(t, fetched, 2)
		ids := []string{fetched[0].ID, fetched[1].ID}
		require.ElementsMatch(t, []string{"order-2", "order-3"}, ids)
	})
}
