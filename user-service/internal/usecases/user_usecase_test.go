package usecases

import (
	"context"
	"errors"
	"testing"
	"user-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

var _ UserRepository = (*MockUserRepository)(nil)

func TestUserUseCase(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Log("➡️  Starting Test: CreateUser Success")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()

			mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
				return user.Username == "testuser" &&
					user.Email == "test@example.com" &&
					user.ID != ""
			})).Return(nil).Once()

			user, err := userUseCase.CreateUser(ctx, "testuser", "test@example.com")
			assert.NoError(t, err)
			assert.NotEmpty(t, user.ID)
			assert.Equal(t, "testuser", user.Username)
			assert.Equal(t, "test@example.com", user.Email)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: CreateUser Success")
		})
		t.Run("InvalidEmail", func(t *testing.T) {
			t.Log("➡️  Starting Test: CreateUser InvalidEmail")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			_, err := userUseCase.CreateUser(ctx, "user", "")
			assert.Error(t, err)
			assert.Equal(t, domain.ErrInvalidEmail, err)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: CreateUser InvalidEmail")
		})
	})
	t.Run("GetUserInParallel", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUserInParallel Success")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			users := []*domain.User{
				{ID: "1", Username: "user1", Email: "user1@example.com"},
				{ID: "2", Username: "user2", Email: "user2@example.com"},
			}
			mockRepo.On("FindByID", mock.Anything, "1").Return(users[0], nil).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(users[1], nil).Once()
			result, err := userUseCase.GetUserInParallel(ctx, []string{"1", "2"})
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.ElementsMatch(t, users, result)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUserInParallel Success")
		})
		t.Run("Error", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUserInParallel Error")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			mockRepo.On("FindByID", mock.Anything, "1").Return(nil, errors.New("user not found")).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(&domain.User{ID: "2"}, nil).Maybe()
			result, err := userUseCase.GetUserInParallel(ctx, []string{"1", "2"})
			assert.Error(t, err)
			assert.Nil(t, result)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUserInParallel Error")
		})
	})
	t.Run("GetUsersWithConcurrencyLimit", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUsersWithConcurrencyLimit Success")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			u1 := &domain.User{ID: "1"}
			u2 := &domain.User{ID: "2"}
			mockRepo.On("FindByID", mock.Anything, "1").Return(u1, nil).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(u2, nil).Once()
			result, err := userUseCase.GetUsersWithConcurrencyLimit(ctx, []string{"1", "2"}, 2)
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUsersWithConcurrencyLimit Success")
		})
		t.Run("Error", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUsersWithConcurrencyLimit Error")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			mockRepo.On("FindByID", mock.Anything, "1").Return(nil, errors.New("not found")).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(&domain.User{ID: "2"}, nil).Maybe()
			result, err := userUseCase.GetUsersWithConcurrencyLimit(ctx, []string{"1", "2"}, 2)
			assert.Error(t, err)
			assert.Nil(t, result)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUsersWithConcurrencyLimit Error")
		})
	})
	t.Run("GetUsersFailFast", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUsersFailFast Success")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			u1 := &domain.User{ID: "1"}
			u2 := &domain.User{ID: "2"}
			mockRepo.On("FindByID", mock.Anything, "1").Return(u1, nil).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(u2, nil).Once()
			result, err := userUseCase.GetUsersFailFast(ctx, []string{"1", "2"}, 2)
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUsersFailFast Success")
		})
		t.Run("Error", func(t *testing.T) {
			t.Log("➡️  Starting Test: GetUsersFailFast Error")
			mockRepo := new(MockUserRepository)
			logger, _ := zap.NewDevelopment()
			userUseCase := NewUserUseCase(mockRepo, logger)
			ctx := context.Background()
			mockRepo.On("FindByID", mock.Anything, "1").Return(nil, errors.New("some error")).Once()
			mockRepo.On("FindByID", mock.Anything, "2").Return(&domain.User{ID: "2"}, nil).Maybe()
			result, err := userUseCase.GetUsersFailFast(ctx, []string{"1", "2"}, 2)
			assert.Error(t, err)
			assert.Nil(t, result)
			mockRepo.AssertExpectations(t)
			t.Log("✅  Finished Test: GetUsersFailFast Error")
		})
	})
}
