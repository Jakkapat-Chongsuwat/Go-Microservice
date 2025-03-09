package interfaces

import (
	"context"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
)

type IOrderUseCase interface {
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
	GetOrdersInParallel(ctx context.Context, orderIDs []string) ([]*domain.Order, error)
	CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error)
}

type IOrderRepository interface {
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
	CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error)
}

type IOrderEventProducer interface {
	SendOrderEvent(event models.OrderEvent) error
}
