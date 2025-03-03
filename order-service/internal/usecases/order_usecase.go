package usecases

import (
	"context"
	"fmt"
	"order-service/internal/domain"
	"order-service/internal/domain/interfaces"
	"sync"
	"time"

	"go.uber.org/zap"
)

type UserServiceClient interface {
	VerifyUser(ctx context.Context, userID string) error
}

type InventoryServiceClient interface {
	VerifyInventory(ctx context.Context, productID string, requiredQuantity int) error
}

type OrderUseCaseImpl struct {
	orderRepo          interfaces.IOrderRepository
	userSvc            UserServiceClient
	inventorySvc       InventoryServiceClient
	orderEventProducer interfaces.OrderEventProducer
	logger             *zap.Logger
}

var _ interfaces.IOrderUseCase = (*OrderUseCaseImpl)(nil)

func NewOrderUsecase(
	repo interfaces.IOrderRepository,
	userClient UserServiceClient,
	inventoryClient InventoryServiceClient,
	eventProducer interfaces.OrderEventProducer,
	logger *zap.Logger,
) interfaces.IOrderUseCase {
	return &OrderUseCaseImpl{
		orderRepo:          repo,
		userSvc:            userClient,
		inventorySvc:       inventoryClient,
		orderEventProducer: eventProducer,
		logger:             logger,
	}
}

func (o *OrderUseCaseImpl) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	o.logger.Info("CreateOrderWithItems called", zap.String("userID", order.UserID))

	if err := o.userSvc.VerifyUser(ctx, order.UserID); err != nil {
		o.logger.Error("failed to verify user", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	for _, item := range items {
		if err := o.inventorySvc.VerifyInventory(ctx, item.ProductID, item.Quantity); err != nil {
			o.logger.Error("failed to verify inventory", zap.String("productID", item.ProductID), zap.Error(err))
			return nil, fmt.Errorf("failed to verify inventory: %w", err)
		}
	}

	order.Items = items

	result, err := o.orderRepo.CreateOrderWithItems(ctx, order, items)
	if err != nil {
		o.logger.Error("failed to create order with items", zap.Error(err))
		return nil, fmt.Errorf("failed to create order with items: %w", err)
	}

	o.logger.Info("order with items created", zap.String("orderID", result.ID))

	event := domain.OrderEvent{
		OrderID:   result.ID,
		EventType: "CREATED",
		Timestamp: time.Now(),
	}
	if err := o.orderEventProducer.SendOrderEvent(event); err != nil {
		o.logger.Error("failed to send order event", zap.Error(err))
	}

	return result, nil
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

func (o *OrderUseCaseImpl) GetOrdersInParallel(ctx context.Context, orderIDs []string) ([]*domain.Order, error) {
	o.logger.Info("GetOrdersInParallel called", zap.Int("count", len(orderIDs)))

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  []*domain.Order
		firstErr error
	)

	wg.Add(len(orderIDs))

	for _, id := range orderIDs {
		go func(oid string) {
			defer wg.Done()
			ord, err := o.orderRepo.GetOrder(ctx, oid)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				return
			}
			results = append(results, ord)
		}(id)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}
