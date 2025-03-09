package fiber_http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"user-service/internal/adapters/models"
	"user-service/internal/domain"

	"github.com/gofiber/fiber/v2"
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
	return nil, nil
}

func (f *FakeUserUseCase) GetUsersWithConcurrencyLimit(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	return nil, nil
}

func (f *FakeUserUseCase) GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.User, error) {
	return nil, nil
}

func TestCreateUser_Fiber(t *testing.T) {
	fakeUC := new(FakeUserUseCase)
	logger, _ := zap.NewDevelopment()
	handler := NewUserHttpHandler(fakeUC, logger)

	app := fiber.New()
	RegisterUserRoutes(app, handler)

	reqPayload := models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)

	expectedUser := &domain.User{
		ID:       "generated-id",
		Username: "testuser",
		Email:    "test@example.com",
	}
	fakeUC.On("CreateUser", mock.Anything, "testuser", "test@example.com").Return(expectedUser, nil)

	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var userResp domain.User
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)

	fakeUC.AssertExpectations(t)
}
