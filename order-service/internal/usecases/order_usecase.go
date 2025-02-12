package usecases

import (
	"context"
	"fmt"
	"order-service/internal/domain"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrderRepository defines the repository interface.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
	CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error)
}

// UserServiceClient verifies a user.
type UserServiceClient interface {
	VerifyUser(ctx context.Context, userID string) error
}

// ProductServiceClient verifies a product.
type ProductServiceClient interface {
	VerifyProduct(ctx context.Context, productID string) error
}

// OrderUseCase defines the order use case interface.
type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
	GetOrdersInParallel(ctx context.Context, orderIDs []string) ([]*domain.Order, error)
	CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error)
}

// OrderUseCaseImpl implements the OrderUseCase.
type OrderUseCaseImpl struct {
	orderRepo  OrderRepository
	userSvc    UserServiceClient
	productSvc ProductServiceClient
	logger     *zap.Logger
}

// NewOrderUsecase creates a new order use case instance.
func NewOrderUsecase(
	repo OrderRepository,
	userClient UserServiceClient,
	productClient ProductServiceClient,
	logger *zap.Logger,
) OrderUseCase {
	return &OrderUseCaseImpl{
		orderRepo:  repo,
		userSvc:    userClient,
		productSvc: productClient,
		logger:     logger,
	}
}

func (o *OrderUseCaseImpl) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	o.logger.Info("CreateOrder called", zap.String("userID", order.UserID))
	if err := o.userSvc.VerifyUser(ctx, order.UserID); err != nil {
		o.logger.Error("failed to verify user", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	order.ID = generateOrderID()
	order.Status = domain.OrderStatusCreated

	order.Items = []*domain.OrderItem{}

	savedOrder, err := o.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		o.logger.Error("failed to create order", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	o.logger.Info("order created", zap.String("orderID", savedOrder.ID))
	return savedOrder, nil
}

func (o *OrderUseCaseImpl) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	o.logger.Info("CreateOrderWithItems called", zap.String("userID", order.UserID))

	if err := o.userSvc.VerifyUser(ctx, order.UserID); err != nil {
		o.logger.Error("failed to verify user", zap.String("userID", order.UserID), zap.Error(err))
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	for _, item := range items {
		if err := o.productSvc.VerifyProduct(ctx, item.ProductID); err != nil {
			o.logger.Error("failed to verify product", zap.String("productID", item.ProductID), zap.Error(err))
			return nil, fmt.Errorf("failed to verify product: %w", err)
		}
	}

	order.ID = generateOrderID()
	order.Status = domain.OrderStatusCreated

	for _, it := range items {
		it.ID = generateItemID()
		it.OrderID = order.ID
	}
	order.Items = items

	result, err := o.orderRepo.CreateOrderWithItems(ctx, order, items)
	if err != nil {
		o.logger.Error("failed to create order with items", zap.Error(err))
		return nil, fmt.Errorf("failed to create order with items: %w", err)
	}

	o.logger.Info("order with items created", zap.String("orderID", result.ID))
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

func generateOrderID() string {
	return "order_" + uuid.New().String()
}

func generateItemID() string {
	return "item_" + uuid.New().String()
}
