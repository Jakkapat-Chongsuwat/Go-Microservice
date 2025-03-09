package fiber_http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"order-service/internal/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type FakeOrderUseCase struct {
	CreateOrderWithItemsFunc func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error)
	GetOrderFunc             func(ctx context.Context, orderID string) (*domain.Order, error)
}

var _ interfaces.IOrderUseCase = (*FakeOrderUseCase)(nil)

func (f *FakeOrderUseCase) CreateOrder(ctx context.Context, userID string, itemsReq []models.OrderItemRequest) (*domain.Order, error) {
	return nil, nil
}

func (f *FakeOrderUseCase) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	if f.CreateOrderWithItemsFunc != nil {
		return f.CreateOrderWithItemsFunc(ctx, order, items)
	}
	return nil, nil
}

func (f *FakeOrderUseCase) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	if f.GetOrderFunc != nil {
		return f.GetOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (f *FakeOrderUseCase) GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	return nil, nil
}

func (f *FakeOrderUseCase) GetOrdersInParallel(ctx context.Context, orderIDs []string) ([]*domain.Order, error) {
	return nil, nil
}

type FakeClock struct {
	FixedTime time.Time
}

func (fc FakeClock) Now() time.Time {
	return fc.FixedTime
}

func TestCreateOrder_Success(t *testing.T) {
	fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
	oldClock := domain.Clock
	domain.Clock = FakeClock{FixedTime: fixedTime}
	defer func() { domain.Clock = oldClock }()

	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeOrderUseCase{
		CreateOrderWithItemsFunc: func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
			order.ID = "order123"
			order.Status = domain.OrderStatusCreated
			return order, nil
		},
	}
	app := fiber.New()
	handler := NewOrderHTTPHandler(fakeUC, logger)
	RegisterOrderRoutes(app, handler)

	reqPayload := models.CreateOrderRequest{
		UserID: "user1",
		Items: []models.OrderItemRequest{
			{ProductID: "prod1", Quantity: 2},
		},
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respPayload models.CreateOrderResponse
	err = json.NewDecoder(resp.Body).Decode(&respPayload)
	assert.NoError(t, err)
	assert.Equal(t, "order123", respPayload.OrderID)
	assert.Equal(t, "user1", respPayload.UserID)
	assert.Equal(t, "CREATED", respPayload.Status)
	assert.Len(t, respPayload.Items, 1)
	assert.Equal(t, "prod1", respPayload.Items[0].ProductID)
	assert.Equal(t, 2, respPayload.Items[0].Quantity)
}

func TestCreateOrder_InvalidPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeOrderUseCase{
		CreateOrderWithItemsFunc: func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
			return nil, nil
		},
	}
	app := fiber.New()
	handler := NewOrderHTTPHandler(fakeUC, logger)
	RegisterOrderRoutes(app, handler)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_MissingUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeOrderUseCase{
		CreateOrderWithItemsFunc: func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
			return nil, nil
		},
	}
	app := fiber.New()
	handler := NewOrderHTTPHandler(fakeUC, logger)
	RegisterOrderRoutes(app, handler)

	reqPayload := models.CreateOrderRequest{
		UserID: "",
		Items: []models.OrderItemRequest{
			{ProductID: "prod1", Quantity: 2},
		},
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_InvalidItemData(t *testing.T) {
	fixedTime := time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC)
	oldClock := domain.Clock
	domain.Clock = FakeClock{FixedTime: fixedTime}
	defer func() { domain.Clock = oldClock }()

	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeOrderUseCase{
		CreateOrderWithItemsFunc: func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
			return nil, nil
		},
	}
	app := fiber.New()
	handler := NewOrderHTTPHandler(fakeUC, logger)
	RegisterOrderRoutes(app, handler)

	// invalid product ID
	reqPayload := models.CreateOrderRequest{
		UserID: "user1",
		Items: []models.OrderItemRequest{
			{ProductID: "", Quantity: 2},
			{ProductID: "prod2", Quantity: 0},
		},
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateOrder_UseCaseError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeOrderUseCase{
		CreateOrderWithItemsFunc: func(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
			return nil, errors.New("use case error")
		},
	}
	app := fiber.New()
	handler := NewOrderHTTPHandler(fakeUC, logger)
	RegisterOrderRoutes(app, handler)

	reqPayload := models.CreateOrderRequest{
		UserID: "user1",
		Items: []models.OrderItemRequest{
			{ProductID: "prod1", Quantity: 2},
		},
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
