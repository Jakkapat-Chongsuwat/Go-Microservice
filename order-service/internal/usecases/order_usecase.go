package usecases

import (
	"context"
	"fmt"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"order-service/internal/domain/interfaces"
	"sync"

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
	orderEventProducer interfaces.IOrderEventProducer
	logger             *zap.Logger
}

var _ interfaces.IOrderUseCase = (*OrderUseCaseImpl)(nil)

func NewOrderUsecase(
	repo interfaces.IOrderRepository,
	userClient UserServiceClient,
	inventoryClient InventoryServiceClient,
	eventProducer interfaces.IOrderEventProducer,
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

func (o *OrderUseCaseImpl) CreateOrder(ctx context.Context, userID string, itemsReq []models.OrderItemRequest) (*domain.Order, error) {
	order := domain.NewOrder(userID)
	var items []*domain.OrderItem

	for _, req := range itemsReq {
		if req.ProductID == "" || req.Quantity <= 0 {
			return nil, fmt.Errorf("invalid order item data")
		}
		item := domain.NewOrderItem(req.ProductID, req.Quantity)
		items = append(items, item)
	}
	order.Items = items

	return o.CreateOrderWithItems(ctx, order, items)
}

// test uncle bob style
func (o *OrderUseCaseImpl) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	o.logger.Info("CreateOrderWithItems called", zap.String("userID", order.UserID))

	now := domain.Clock.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now

	if err := o.verifyUser(ctx, order.UserID); err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	if err := o.verifyInventory(ctx, items); err != nil {
		return nil, fmt.Errorf("failed to verify inventory: %w", err)
	}

	createdOrder, err := o.persistOrder(ctx, order, items)
	if err != nil {
		return nil, fmt.Errorf("failed to create order with items: %w", err)
	}

	o.publishOrderEvent(createdOrder)

	return createdOrder, nil
}

func (o *OrderUseCaseImpl) verifyUser(ctx context.Context, userID string) error {
	return o.userSvc.VerifyUser(ctx, userID)
}

func (o *OrderUseCaseImpl) verifyInventory(ctx context.Context, items []*domain.OrderItem) error {
	for _, item := range items {
		if err := o.inventorySvc.VerifyInventory(ctx, item.ProductID, item.Quantity); err != nil {
			o.logger.Error("Inventory check failed", zap.String("productID", item.ProductID), zap.Error(err))
			return err
		}
	}
	return nil
}

func (o *OrderUseCaseImpl) persistOrder(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	return o.orderRepo.CreateOrderWithItems(ctx, order, items)
}

func (o *OrderUseCaseImpl) publishOrderEvent(order *domain.Order) {
	event := models.OrderEvent{
		OrderID:   order.ID,
		EventType: "CREATED",
		Timestamp: domain.Clock.Now(),
	}
	if err := o.orderEventProducer.SendOrderEvent(event); err != nil {
		o.logger.Error("failed to send order event", zap.Error(err))
	}
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

// test get parallel orders, batch can be more efficient
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
