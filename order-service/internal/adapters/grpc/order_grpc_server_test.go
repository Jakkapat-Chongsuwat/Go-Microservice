package grpc

import (
	"context"
	"fmt"
	"testing"

	"order-service/internal/adapters/models"
	"order-service/internal/domain"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/order_service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type FakeOrderUseCase struct {
	CreateOrderFunc func(ctx context.Context, userID string, items []models.OrderItemRequest) (*domain.Order, error)
	GetOrderFunc    func(ctx context.Context, orderID string) (*domain.Order, error)
}

func (f *FakeOrderUseCase) CreateOrder(ctx context.Context, userID string, items []models.OrderItemRequest) (*domain.Order, error) {
	if f.CreateOrderFunc != nil {
		return f.CreateOrderFunc(ctx, userID, items)
	}
	return nil, nil
}

func (f *FakeOrderUseCase) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	return nil, fmt.Errorf("not implemented")
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
func (f *FakeOrderUseCase) GetUsersFailFast(ctx context.Context, userIDs []string, maxWorkers int) ([]*domain.Order, error) {
	return nil, nil
}

func TestOrderGRPCServer_CreateOrder_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	fakeUC := &FakeOrderUseCase{
		CreateOrderFunc: func(ctx context.Context, userID string, itemsReq []models.OrderItemRequest) (*domain.Order, error) {
			order := domain.NewOrder(userID)
			var domainItems []*domain.OrderItem
			for _, req := range itemsReq {
				domainItem := domain.NewOrderItem(req.ProductID, req.Quantity)
				domainItems = append(domainItems, domainItem)
			}
			order.Items = domainItems
			order.ID = "order123"
			return order, nil
		},
	}

	server := NewOrderGRPCServer(fakeUC, logger)

	protoReq := &order_service.CreateOrderRequest{
		UserId: "user1",
		Items: []*order_service.OrderItem{
			{
				ProductId: "prod1",
				Quantity:  2,
			},
			{
				ProductId: "prod2",
				Quantity:  3,
			},
		},
	}

	resp, err := server.CreateOrder(context.Background(), protoReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "order123", resp.OrderId)
	assert.Equal(t, "user1", resp.UserId)
	assert.Equal(t, string(domain.OrderStatusCreated), resp.Status)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "prod1", resp.Items[0].ProductId)
	assert.Equal(t, int32(2), resp.Items[0].Quantity)
	assert.Equal(t, "prod2", resp.Items[1].ProductId)
	assert.Equal(t, int32(3), resp.Items[1].Quantity)
}

func TestOrderGRPCServer_CreateOrder_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	fakeUC := &FakeOrderUseCase{
		CreateOrderFunc: func(ctx context.Context, userID string, itemsReq []models.OrderItemRequest) (*domain.Order, error) {
			return nil, fmt.Errorf("use case error")
		},
	}

	server := NewOrderGRPCServer(fakeUC, logger)

	protoReq := &order_service.CreateOrderRequest{
		UserId: "user1",
		Items: []*order_service.OrderItem{
			{
				ProductId: "prod1",
				Quantity:  2,
			},
		},
	}

	resp, err := server.CreateOrder(context.Background(), protoReq)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestOrderGRPCServer_GetOrder_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	fakeUC := &FakeOrderUseCase{
		GetOrderFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			order := domain.NewOrder("user1")
			order.ID = orderID
			item1 := domain.NewOrderItem("prod1", 2)
			item2 := domain.NewOrderItem("prod2", 3)
			order.Items = []*domain.OrderItem{item1, item2}
			return order, nil
		},
	}

	server := NewOrderGRPCServer(fakeUC, logger)

	protoReq := &order_service.GetOrderRequest{
		OrderId: "order123",
	}

	resp, err := server.GetOrder(context.Background(), protoReq)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Order)
	assert.Equal(t, "order123", resp.Order.OrderId)
	assert.Equal(t, "user1", resp.Order.UserId)
	assert.Equal(t, string(domain.OrderStatusCreated), resp.Order.Status)
	assert.Len(t, resp.Order.Items, 2)
	assert.Equal(t, "prod1", resp.Order.Items[0].ProductId)
	assert.Equal(t, int32(2), resp.Order.Items[0].Quantity)
	assert.Equal(t, "prod2", resp.Order.Items[1].ProductId)
	assert.Equal(t, int32(3), resp.Order.Items[1].Quantity)
}

func TestOrderGRPCServer_GetOrder_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	fakeUC := &FakeOrderUseCase{
		GetOrderFunc: func(ctx context.Context, orderID string) (*domain.Order, error) {
			return nil, fmt.Errorf("order not found")
		},
	}

	server := NewOrderGRPCServer(fakeUC, logger)

	protoReq := &order_service.GetOrderRequest{
		OrderId: "nonexistent",
	}

	resp, err := server.GetOrder(context.Background(), protoReq)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
