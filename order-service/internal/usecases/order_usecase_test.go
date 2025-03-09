package usecases

import (
	"context"
	"errors"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, order)
	if o, ok := args.Get(0).(*domain.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	args := m.Called(ctx, orderID)
	if o, ok := args.Get(0).(*domain.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	args := m.Called(ctx, userID)
	if o, ok := args.Get(0).([]*domain.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	args := m.Called(ctx, order, items)
	if o, ok := args.Get(0).(*domain.Order); ok {
		return o, args.Error(1)
	}
	return nil, args.Error(1)
}

type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) VerifyUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockProductServiceClient struct {
	mock.Mock
}

func (m *MockProductServiceClient) VerifyInventory(ctx context.Context, productID string, requiredQuantity int) error {
	args := m.Called(ctx, productID, requiredQuantity)
	return args.Error(0)
}

type MockOrderEventProducer struct {
	mock.Mock
}

func (m *MockOrderEventProducer) SendOrderEvent(event models.OrderEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestOrderUseCase(t *testing.T) {

	t.Run("CreateOrderWithItems", func(t *testing.T) {

		t.Run("Success", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			mockEventProducer.On("SendOrderEvent", mock.Anything).Return(nil).Maybe()

			inputOrder := domain.NewOrder("user123")
			inputItem := domain.NewOrderItem("prodABC", 2)
			inputOrder.Items = []*domain.OrderItem{inputItem}

			expectedOrder := &domain.Order{
				ID:     "order_123",
				UserID: "user123",
				Status: domain.OrderStatusCreated,
				Items: []*domain.OrderItem{
					{
						ID:        "item_123",
						OrderID:   "order_123",
						ProductID: "prodABC",
						Quantity:  2,
					},
				},
			}

			mockUserSvc.On("VerifyUser", mock.Anything, "user123").Return(nil).Once()
			mockInvenSvc.On("VerifyInventory", mock.Anything, "prodABC", 2).Return(nil).Once()
			mockRepo.On("CreateOrderWithItems", mock.Anything,
				mock.MatchedBy(func(o *domain.Order) bool {
					if o.UserID != "user123" || o.Status != domain.OrderStatusCreated || o.ID == "" {
						return false
					}
					if len(o.Items) != 1 {
						return false
					}
					item := o.Items[0]
					if item.ProductID != "prodABC" || item.Quantity != 2 || item.ID == "" {
						return false
					}
					return true
				}),
				mock.Anything).Return(expectedOrder, nil).Once()

			ctx := context.Background()
			result, err := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger).CreateOrderWithItems(ctx, inputOrder, inputOrder.Items)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, expectedOrder.UserID, result.UserID)
			assert.Equal(t, expectedOrder.Status, result.Status)
			assert.Equal(t, expectedOrder.ID, result.ID)
			assert.Len(t, result.Items, 1)
			assert.Equal(t, "prodABC", result.Items[0].ProductID)
			assert.Equal(t, 2, result.Items[0].Quantity)

			mockUserSvc.AssertExpectations(t)
			mockInvenSvc.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})

		t.Run("VerifyUserFail", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			inputOrder := domain.NewOrder("userX")
			inputOrder.Items = []*domain.OrderItem{domain.NewOrderItem("prodY", 1)}

			mockUserSvc.On("VerifyUser", mock.Anything, "userX").
				Return(errors.New("user not found")).Once()

			ctx := context.Background()
			result, err := orderUC.CreateOrderWithItems(ctx, inputOrder, inputOrder.Items)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to verify user")
			mockUserSvc.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
			mockInvenSvc.AssertExpectations(t)
		})

		t.Run("RepoCreateFail", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			inputOrder := domain.NewOrder("userX")
			inputOrder.Items = []*domain.OrderItem{domain.NewOrderItem("prodY", 1)}

			mockUserSvc.On("VerifyUser", mock.Anything, "userX").Return(nil).Once()
			mockInvenSvc.On("VerifyInventory", mock.Anything, "prodY", 1).Return(nil).Once()
			mockRepo.On("CreateOrderWithItems", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("db error")).Once()

			ctx := context.Background()
			result, err := orderUC.CreateOrderWithItems(ctx, inputOrder, inputOrder.Items)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to create order with items")
			mockUserSvc.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
			mockInvenSvc.AssertExpectations(t)
		})
	})

	t.Run("GetOrder", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			expected := &domain.Order{
				ID:     "order123",
				UserID: "userX",
				Items: []*domain.OrderItem{
					{
						ProductID: "prodZ",
						Quantity:  5,
					},
				},
			}

			mockRepo.On("GetOrder", mock.Anything, "order123").Return(expected, nil).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrder(ctx, "order123")
			assert.NoError(t, err)
			assert.Equal(t, expected, result)
			mockRepo.AssertExpectations(t)
		})

		t.Run("RepoError", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			mockRepo.On("GetOrder", mock.Anything, "orderABC").Return(nil, errors.New("not found")).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrder(ctx, "orderABC")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to get order")
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetOrdersByUserID", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			orders := []*domain.Order{
				{ID: "orderA", UserID: "user123", Items: []*domain.OrderItem{{ProductID: "p1", Quantity: 1}}},
				{ID: "orderB", UserID: "user123", Items: []*domain.OrderItem{{ProductID: "p2", Quantity: 2}}},
			}

			mockRepo.On("GetOrdersByUserID", mock.Anything, "user123").Return(orders, nil).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrdersByUserID(ctx, "user123")
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.Equal(t, orders, result)
			mockRepo.AssertExpectations(t)
		})

		t.Run("RepoError", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockInvenSvc := new(MockProductServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := NewOrderUsecase(mockRepo, mockUserSvc, mockInvenSvc, mockEventProducer, logger)
			mockRepo.On("GetOrdersByUserID", mock.Anything, "unknownUser").Return(nil, errors.New("db error")).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrdersByUserID(ctx, "unknownUser")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to get orders")
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetOrdersInParallel", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := &OrderUseCaseImpl{
				orderRepo:          mockRepo,
				userSvc:            mockUserSvc,
				orderEventProducer: mockEventProducer,
				logger:             logger,
			}

			orderIDs := []string{"order1", "order2"}
			orders := []*domain.Order{
				{ID: "order1", UserID: "userX", Items: []*domain.OrderItem{{ProductID: "p1", Quantity: 3}}},
				{ID: "order2", UserID: "userX", Items: []*domain.OrderItem{{ProductID: "p2", Quantity: 4}}},
			}

			mockRepo.On("GetOrder", mock.Anything, "order1").Return(orders[0], nil).Once()
			mockRepo.On("GetOrder", mock.Anything, "order2").Return(orders[1], nil).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrdersInParallel(ctx, orderIDs)
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.ElementsMatch(t, orders, result)
			mockRepo.AssertExpectations(t)
		})

		t.Run("PartialFailure", func(t *testing.T) {
			mockRepo := new(MockOrderRepository)
			mockUserSvc := new(MockUserServiceClient)
			mockEventProducer := new(MockOrderEventProducer)
			logger, _ := zap.NewDevelopment()

			orderUC := &OrderUseCaseImpl{
				orderRepo:          mockRepo,
				userSvc:            mockUserSvc,
				orderEventProducer: mockEventProducer,
				logger:             logger,
			}

			orderIDs := []string{"orderOk", "orderFail"}
			okOrder := &domain.Order{ID: "orderOk", UserID: "userX", Items: []*domain.OrderItem{{ProductID: "p-ok", Quantity: 1}}}
			failErr := errors.New("some DB error")
			mockRepo.On("GetOrder", mock.Anything, "orderOk").Return(okOrder, nil).Once()
			mockRepo.On("GetOrder", mock.Anything, "orderFail").Return(nil, failErr).Once()
			ctx := context.Background()
			result, err := orderUC.GetOrdersInParallel(ctx, orderIDs)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, failErr, err)
			mockRepo.AssertExpectations(t)
		})
	})
}
