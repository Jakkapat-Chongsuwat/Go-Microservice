package repository

import (
	"context"
	"errors"
	"user-service/internal/domain"
	"user-service/internal/usecases"

	"go.uber.org/zap"
)

type InMemoryUserRepository struct {
	store  map[string]*domain.User
	logger *zap.Logger
}

func NewInMemoryUserRepo(logger *zap.Logger) *InMemoryUserRepository {
	return &InMemoryUserRepository{
		store:  make(map[string]*domain.User),
		logger: logger,
	}
}

var _ usecases.UserRepository = (*InMemoryUserRepository)(nil)

func (r *InMemoryUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.store[user.ID] = user
	r.logger.Info("user saved", zap.String("id", user.ID), zap.String("username", user.Username), zap.String("email", user.Email))
	return nil
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user, found := r.store[id]
	if !found {
		r.logger.Warn("user not found", zap.String("id", id))
		return nil, errors.New("user not found")
	}
	r.logger.Debug("user retrieved", zap.String("id", id))
	return user, nil
}

func (r *InMemoryUserRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(r.store))
	for _, user := range r.store {
		users = append(users, user)
	}
	r.logger.Debug("all users retrieved", zap.Int("count", len(users)))
	return users, nil
}
