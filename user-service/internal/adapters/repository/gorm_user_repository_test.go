package repository

import (
	"context"
	"fmt"
	"testing"
	"time"
	"user-service/internal/adapters/models"
	"user-service/internal/domain"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormUserRepository(t *testing.T) {
	t.Log("➡️  Starting Test: GormUserRepository")

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
		}).WithStartupTimeout(60 * time.Second)}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err, "Failed to start postgres container")

	defer postgresC.Terminate(ctx)

	host, err := postgresC.Host(ctx)
	require.NoError(t, err, "Failed to get postgres container host")
	mappedPort, err := postgresC.MappedPort(ctx, "5432")
	require.NoError(t, err, "Failed to get postgres container mapped port")
	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=testDb sslmode=disable", host, mappedPort.Port())

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to postgres")

	err = db.AutoMigrate(&models.GormDBUser{})
	require.NoError(t, err, "Failed to migrate users table")

	logger, _ := zap.NewDevelopment()
	repo := NewGormUserRepo(db, logger)

	t.Run("SaveAndFindByID_Success", func(t *testing.T) {
		t.Log("➡️  Starting Test: SaveAndFindByID_Success")

		user := &domain.User{
			ID:       "1",
			Username: "Alice",
			Email:    "alice@example.com",
		}

		err := repo.Save(ctx, user)
		require.NoError(t, err, "Failed to save user")

		fetched, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err, "Failed to fetch user by ID")
		require.Equal(t, user.ID, fetched.ID)
		require.Equal(t, user.Username, fetched.Username)
		require.Equal(t, user.Email, fetched.Email)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		fetched, err := repo.FindByID(ctx, "no-id")
		require.Error(t, err, "Expected error for user not found")
		require.Nil(t, fetched, "Expected nil user for user not found")
	})

	t.Run("UpsertOverwriteExisting", func(t *testing.T) {
		user1 := &domain.User{
			ID:       "1",
			Username: "Oldname",
			Email:    "old@example.com",
		}

		user2 := &domain.User{
			ID:       "1",
			Username: "Newname",
			Email:    "new@example.com",
		}

		err := repo.Save(ctx, user1)
		require.NoError(t, err, "Failed to save user")

		err = repo.Save(ctx, user2)
		require.NoError(t, err, "Failed to save user")

		fetched, err := repo.FindByID(ctx, user2.ID)
		require.NoError(t, err, "Failed to fetch user by ID")
		require.Equal(t, user2.ID, fetched.ID)
		require.Equal(t, user2.Username, fetched.Username)
		require.Equal(t, user2.Email, fetched.Email)
	})
}
