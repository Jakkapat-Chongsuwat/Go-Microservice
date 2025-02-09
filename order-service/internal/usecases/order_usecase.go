// order-service/internal/usecases/order_usecase.go

package usecases

import (
	"context"
	"fmt"
	"order-service/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
}

type UserServiceClient interface {
	VerifyUser(ctx context.Context, userID string) error
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
}

type OrderUseCaseImpl struct {
	orderRepo OrderRepository
	userSvc   UserServiceClient
	logger    *zap.Logger
}

func NewOrderUsecase(
	repo OrderRepository,
	userClient UserServiceClient,
	logger *zap.Logger,
) OrderUseCase {
	return &OrderUseCaseImpl{
		orderRepo: repo,
		userSvc:   userClient,
		logger:    logger,
	}
}

func (o *OrderUseCaseImpl) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	o.logger.Info("CreateOrder called",
		zap.String("userID", order.UserID),
		zap.String("productID", order.ProductID),
		zap.Int("quantity", order.Quantity),
	)

	if err := o.userSvc.VerifyUser(ctx, order.UserID); err != nil {
		o.logger.Error("failed to verify user", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	order.ID = generateOrderID()
	order.Status = domain.OrderStatusCreated

	savedOrder, err := o.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		o.logger.Error("failed to create order", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	o.logger.Info("order created", zap.String("orderID", savedOrder.ID))
	return savedOrder, nil
}

func (o *OrderUseCaseImpl) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	o.logger.Info("GetOrder called", zap.String("orderID", orderID))

	order, err := o.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		o.logger.Error("failed to get order", zap.String("orderID", orderID), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	o.logger.Info("order found", zap.String("orderID", orderID))
	return order, nil
}

func (o *OrderUseCaseImpl) GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	o.logger.Info("GetOrdersByUserID called", zap.String("userID", userID))

	orders, err := o.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		o.logger.Error("failed to get orders", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	o.logger.Info("orders found", zap.String("userID", userID))
	return orders, nil
}

func generateOrderID() string {
	return "order_" + uuid.New().String()
}
