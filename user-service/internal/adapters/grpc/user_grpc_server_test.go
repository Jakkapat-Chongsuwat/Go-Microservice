package grpc_test

import (
	"context"
	"testing"
	"user-service/internal/adapters/grpc"
	"user-service/internal/domain"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/user_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type FakeUserUseCase struct {
	mock.Mock
}

func (f *FakeUserUseCase) CreateUser(ctx context.Context, username, email string) (*domain.User, error) {
	args := f.Called(ctx, username, email)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (f *FakeUserUseCase) GetUserInParallel(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	args := f.Called(ctx, userIDs)
	if users, ok := args.Get(0).([]*domain.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
}

func (f *FakeUserUseCase) GetUsersWithConcurrencyLimit(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	return nil, nil
}

func (f *FakeUserUseCase) GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	return nil, nil
}

func (f *FakeUserUseCase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}

func TestUserGRPCServer_CreateUser(t *testing.T) {
	fakeUC := new(FakeUserUseCase)
	logger, _ := zap.NewDevelopment()
	server := grpc.NewUserGRPCServer(fakeUC, logger)

	expectedUser := &domain.User{
		ID:       "generated-id",
		Username: "testuser",
		Email:    "test@example.com",
	}
	fakeUC.On("CreateUser", mock.Anything, "testuser", "test@example.com").Return(expectedUser, nil)

	req := &user_service.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	resp, err := server.CreateUser(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, resp.Id)
	assert.Equal(t, expectedUser.Username, resp.Username)
	assert.Equal(t, expectedUser.Email, resp.Email)

	fakeUC.AssertExpectations(t)
}

func TestUserGRPCServer_GetUserByID(t *testing.T) {
	fakeUC := new(FakeUserUseCase)
	logger, _ := zap.NewDevelopment()
	server := grpc.NewUserGRPCServer(fakeUC, logger)

	expectedUser := &domain.User{
		ID:       "123",
		Username: "testuser",
		Email:    "test@example.com",
	}
	fakeUC.On("GetUserInParallel", mock.Anything, []string{"123"}).Return([]*domain.User{expectedUser}, nil)

	req := &user_service.GetUserRequest{Id: "123"}
	resp, err := server.GetUserByID(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, resp.Id)
	assert.Equal(t, expectedUser.Username, resp.Username)
	assert.Equal(t, expectedUser.Email, resp.Email)

	fakeUC.AssertExpectations(t)
}
