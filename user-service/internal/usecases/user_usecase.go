// user-service\internal\usecases\user_usecase.go

package usecases

import (
	"context"
	"fmt"
	"sync"
	"user-service/internal/domain"

	"go.uber.org/zap"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
}

type UserUseCase interface {
	CreateUser(ctx context.Context, username, email string) (*domain.User, error)
	GetUserInParallel(ctx context.Context, userIDs []string) ([]*domain.User, error)
	GetUsersWithConcurrencyLimit(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error)
	GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error)
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
}

type UserUseCaseImpl struct {
	userRepo UserRepository
	logger   *zap.Logger
}

func NewUserUseCase(userRepo UserRepository, logger *zap.Logger) UserUseCase {
	return &UserUseCaseImpl{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *UserUseCaseImpl) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
	u.logger.Info("CreateUser called", zap.String("username", username), zap.String("email", email))
	if email == "" {
		u.logger.Error("invalid email", zap.String("username", username))
		return nil, domain.ErrInvalidEmail
	}

	user := domain.NewUser(username, email)

	err := u.userRepo.Save(ctx, user)
	if err != nil {
		u.logger.Error("failed to save user", zap.String("id", user.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	u.logger.Info("user created", zap.String("id", user.ID))
	return user, nil
}

// for testing, batch query can be more efficient
func (u *UserUseCaseImpl) GetUserInParallel(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	u.logger.Info("GetUserInParallel called", zap.Int("count", len(userIDs)))
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  []*domain.User
		firstErr error
	)
	wg.Add(len(userIDs))
	for _, id := range userIDs {
		go func(uid string) {
			defer wg.Done()
			user, err := u.userRepo.FindByID(ctx, uid)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				u.logger.Error("error finding user in parallel", zap.String("id", uid), zap.Error(err))
			} else if err == nil {
				results = append(results, user)
				u.logger.Debug("user found in parallel", zap.String("id", uid))
			}
		}(id)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

// for testing, batch query can be more efficient
func (u *UserUseCaseImpl) GetUsersWithConcurrencyLimit(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	u.logger.Info("GetUsersWithConcurrencyLimit called", zap.Int("count", len(userIDs)), zap.Int("maxWorkers", maxWorkers))
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	if maxWorkers > len(userIDs) {
		maxWorkers = len(userIDs)
	}
	sem := make(chan struct{}, maxWorkers)
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  = make([]*domain.User, len(userIDs))
		firstErr error
	)
	for i, id := range userIDs {
		wg.Add(1)
		go func(idx int, uid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				return
			default:
			}
			user, err := u.userRepo.FindByID(ctx, uid)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				u.logger.Error("error finding user with concurrency limit", zap.String("id", uid), zap.Error(err))
				return
			}
			if err == nil {
				results[idx] = user
				u.logger.Debug("user found with concurrency limit", zap.String("id", uid))
			}
		}(i, id)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

// for testing, batch query can be more efficient
func (u *UserUseCaseImpl) GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	u.logger.Info("GetUsersFailFast called", zap.Int("count", len(userIDs)), zap.Int("maxWorkers", maxWorkers))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	if maxWorkers > len(userIDs) {
		maxWorkers = len(userIDs)
	}
	sem := make(chan struct{}, maxWorkers)
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  = make([]*domain.User, len(userIDs))
		firstErr error
	)
	for i, id := range userIDs {
		wg.Add(1)
		go func(idx int, uid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				return
			default:
			}
			user, err := u.userRepo.FindByID(ctx, uid)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				u.logger.Error("error finding user (fail-fast)", zap.String("id", uid), zap.Error(err))
				cancel()
				return
			}
			if err == nil {
				results[idx] = user
				u.logger.Debug("user found (fail-fast)", zap.String("id", uid))
			}
		}(i, id)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (u *UserUseCaseImpl) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	u.logger.Info("GetAllUsers called")
	users, err := u.userRepo.FindAll(ctx)
	if err != nil {
		u.logger.Error("failed to get all users", zap.Error(err))
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	u.logger.Info("all users retrieved", zap.Int("count", len(users)))
	return users, nil
}
