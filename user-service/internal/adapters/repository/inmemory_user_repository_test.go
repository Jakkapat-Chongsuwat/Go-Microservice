package repository

import (
	"context"
	"testing"
	"user-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInMemoryUserRepository(t *testing.T) {
	t.Run("SaveAndFindSuccess", func(t *testing.T) {
		t.Log("➡️  Starting Test: SaveAndFindSuccess")
		logger, _ := zap.NewDevelopment()
		repo := NewInMemoryUserRepo(logger)
		ctx := context.Background()
		user := &domain.User{
			ID:       "123",
			Username: "alice",
			Email:    "alice@example.com",
		}
		t.Log("Saving user...")
		errSave := repo.Save(ctx, user)
		fetchedUser, errFind := repo.FindByID(ctx, "123")
		assert.NoError(t, errSave)
		assert.NoError(t, errFind)
		assert.Equal(t, user, fetchedUser)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Log("➡️  Starting Test: UserNotFound")
		logger, _ := zap.NewDevelopment()
		repo := NewInMemoryUserRepo(logger)
		ctx := context.Background()
		t.Log("Finding user...")
		fetchedUser, err := repo.FindByID(ctx, "non-existent-id")
		assert.Nil(t, fetchedUser)
		assert.Error(t, err)
	})

	t.Run("OverwriteExistingUser", func(t *testing.T) {
		t.Log("➡️  Starting Test: OverwriteExistingUser")
		logger, _ := zap.NewDevelopment()
		repo := NewInMemoryUserRepo(logger)
		ctx := context.Background()
		user1 := &domain.User{
			ID:       "123",
			Username: "alice",
			Email:    "alice2@example.com",
		}
		user2 := &domain.User{
			ID:       "123",
			Username: "alice",
			Email:    "alice2@example.com",
		}
		_ = repo.Save(ctx, user1)
		_ = repo.Save(ctx, user2)
		fetchedUser, err := repo.FindByID(ctx, "123")
		assert.NoError(t, err)
		assert.Equal(t, user2, fetchedUser)
	})
}
