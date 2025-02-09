package grpc_test

import (
	"context"
	"testing"
	"user-service/internal/adapters/grpc"
	"user-service/internal/domain"
	"user-service/proto/helloworld"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) CreateUser(ctx context.Context, id, username, email string) (*domain.User, error) {
	args := m.Called(ctx, id, username, email)
	user, _ := args.Get(0).(*domain.User)
	return user, args.Error(1)
}

func (m *MockUserUseCase) GetUserInParallel(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	args := m.Called(ctx, userIDs)
	users, _ := args.Get(0).([]*domain.User)
	return users, args.Error(1)
}

func (m *MockUserUseCase) GetUsersWithConcurrencyLimit(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	args := m.Called(ctx, userIDs, maxWorkers)
	users, _ := args.Get(0).([]*domain.User)
	return users, args.Error(1)
}

func (m *MockUserUseCase) GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	args := m.Called(ctx, userIDs, maxWorkers)
	users, _ := args.Get(0).([]*domain.User)
	return users, args.Error(1)
}

func TestSayHello(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockUserUsecase := new(MockUserUseCase)
	srv := grpc.NewServer(mockUserUsecase, logger)
	req := &helloworld.HelloRequest{Name: "John"}
	resp, err := srv.SayHello(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "Hello John", resp.Message)
}
