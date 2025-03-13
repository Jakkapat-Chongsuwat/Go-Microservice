package fiber_http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (f *FakeUserUseCase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	args := f.Called(ctx)
	if users, ok := args.Get(0).([]*domain.User); ok {
		return users, args.Error(1)
	}
	return nil, args.Error(1)
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

	var userResp models.UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, userResp.ID)
	assert.Equal(t, expectedUser.Username, userResp.Username)
	assert.Equal(t, expectedUser.Email, userResp.Email)

	fakeUC.AssertExpectations(t)
}

func TestCreateUser_Fiber_Error(t *testing.T) {
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

	fakeUC.On("CreateUser", mock.Anything, "testuser", "test@example.com").Return(nil, errors.New("service error"))

	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "failed to create user", respBody["error"])

	fakeUC.AssertExpectations(t)
}

func TestGetUsers_Fiber(t *testing.T) {
	fakeUC := new(FakeUserUseCase)
	logger, _ := zap.NewDevelopment()
	handler := NewUserHttpHandler(fakeUC, logger)

	app := fiber.New()
	RegisterUserRoutes(app, handler)

	expectedUsers := []*domain.User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
		{ID: "2", Username: "user2", Email: "user2@example.com"},
	}

	fakeUC.On("GetAllUsers", mock.Anything).Return(expectedUsers, nil)

	req := httptest.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response models.Response[[]models.UserResponse]
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedUsers), len(response.Data))
	assert.Equal(t, len(expectedUsers), response.Meta.Total)

	for i, user := range expectedUsers {
		assert.Equal(t, user.ID, response.Data[i].ID)
		assert.Equal(t, user.Username, response.Data[i].Username)
		assert.Equal(t, user.Email, response.Data[i].Email)
	}

	fakeUC.AssertExpectations(t)
}
