package usecases

import (
	"context"
	"errors"
	"inventory-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// --- Mock Implementations ---

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if p, ok := args.Get(0).(*domain.Product); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) GetProduct(ctx context.Context, productId string) (*domain.Product, error) {
	args := m.Called(ctx, productId)
	if p, ok := args.Get(0).(*domain.Product); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if p, ok := args.Get(0).(*domain.Product); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryRepository) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	args := m.Called(ctx)
	if p, ok := args.Get(0).([]*domain.Product); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

type FakeClock struct {
	fixedTime time.Time
}

func (fc FakeClock) Now() time.Time {
	return fc.fixedTime
}

// --- Test Cases ---

func TestInventoryUseCaseImpl_CreateProduct(t *testing.T) {
	ctx := context.Background()
	product := domain.NewProduct("Product 1", 100, 9.99)

	mockRepo := new(MockInventoryRepository)
	mockRepo.On("CreateProduct", ctx, product).Return(product, nil)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	createdProduct, err := usecase.CreateProduct(ctx, product)
	assert.NoError(t, err)
	assert.Equal(t, product, createdProduct)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_CreateProduct_Error(t *testing.T) {
	ctx := context.Background()
	product := domain.NewProduct("Product 1", 100, 9.99)

	mockRepo := new(MockInventoryRepository)
	expectedErr := errors.New("failed to create product")
	mockRepo.On("CreateProduct", ctx, product).Return((*domain.Product)(nil), expectedErr)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	createdProduct, err := usecase.CreateProduct(ctx, product)
	assert.Error(t, err)
	assert.Nil(t, createdProduct)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_GetProduct(t *testing.T) {
	ctx := context.Background()
	product := domain.NewProduct("Product 1", 100, 9.99)

	mockRepo := new(MockInventoryRepository)
	mockRepo.On("GetProduct", ctx, product.ID).Return(product, nil)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	gotProduct, err := usecase.GetProduct(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, product, gotProduct)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_GetProduct_Error(t *testing.T) {
	ctx := context.Background()
	product := domain.NewProduct("Product 1", 100, 9.99)

	mockRepo := new(MockInventoryRepository)
	expectedErr := errors.New("failed to get product")
	mockRepo.On("GetProduct", ctx, product.ID).Return((*domain.Product)(nil), expectedErr)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	gotProduct, err := usecase.GetProduct(ctx, product.ID)
	assert.Error(t, err)
	assert.Nil(t, gotProduct)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_UpdateProductMetadata(t *testing.T) {
	fixedTime := time.Date(2020, time.January, 1, 12, 0, 0, 0, time.UTC)
	oldClock := domain.Clock
	domain.Clock = FakeClock{fixedTime: fixedTime}
	defer func() { domain.Clock = oldClock }()

	ctx := context.Background()
	originalProduct := domain.NewProduct("Product 1", 100, 9.99)
	updatedMetadata := domain.NewProduct("Updated Product", originalProduct.Quantity, 19.99)
	updatedMetadata.ID = originalProduct.ID

	mockRepo := new(MockInventoryRepository)
	mockRepo.On("GetProduct", ctx, originalProduct.ID).Return(originalProduct, nil)
	mockRepo.On("UpdateProduct", ctx, mock.MatchedBy(func(prod *domain.Product) bool {
		return prod.ID == originalProduct.ID &&
			prod.Name == "Updated Product" &&
			prod.Price == 19.99 &&
			prod.Quantity == originalProduct.Quantity &&
			prod.UpdatedAt.Equal(fixedTime)
	})).Return(updatedMetadata, nil)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	result, err := usecase.UpdateProductMetadata(ctx, updatedMetadata)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Product", result.Name)
	assert.Equal(t, 19.99, result.Price)
	assert.Equal(t, originalProduct.Quantity, result.Quantity)
	assert.True(t, result.UpdatedAt.Equal(fixedTime))
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_UpdateProductMetadata_GetError(t *testing.T) {
	ctx := context.Background()
	updatedMetadata := domain.NewProduct("Updated Product", 100, 19.99)

	mockRepo := new(MockInventoryRepository)
	expectedErr := errors.New("failed to get product")
	mockRepo.On("GetProduct", ctx, updatedMetadata.ID).Return((*domain.Product)(nil), expectedErr)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	result, err := usecase.UpdateProductMetadata(ctx, updatedMetadata)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_UpdateProductStockQuantity(t *testing.T) {
	ctx := context.Background()
	originalProduct := domain.NewProduct("Product 1", 100, 9.99)
	quantityChange := 10
	expectedQuantity := originalProduct.Quantity + quantityChange

	adjustedProduct := domain.NewProduct(originalProduct.Name, expectedQuantity, originalProduct.Price)
	adjustedProduct.ID = originalProduct.ID

	mockRepo := new(MockInventoryRepository)
	mockRepo.On("GetProduct", ctx, originalProduct.ID).Return(originalProduct, nil)
	mockRepo.On("UpdateProduct", ctx, mock.MatchedBy(func(prod *domain.Product) bool {
		return prod.ID == originalProduct.ID && prod.Quantity == expectedQuantity
	})).Return(adjustedProduct, nil)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	result, err := usecase.UpdateProductStockQuantity(ctx, originalProduct.ID, quantityChange)
	assert.NoError(t, err)
	assert.Equal(t, expectedQuantity, result.Quantity)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_UpdateProductStockQuantity_GetError(t *testing.T) {
	ctx := context.Background()
	product := domain.NewProduct("Product 1", 100, 9.99)

	mockRepo := new(MockInventoryRepository)
	expectedErr := errors.New("failed to get product")
	mockRepo.On("GetProduct", ctx, product.ID).Return((*domain.Product)(nil), expectedErr)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	result, err := usecase.UpdateProductStockQuantity(ctx, product.ID, 10)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_ListProducts(t *testing.T) {
	ctx := context.Background()
	products := []*domain.Product{
		domain.NewProduct("Product 1", 100, 9.99),
		domain.NewProduct("Product 2", 200, 19.99),
	}

	mockRepo := new(MockInventoryRepository)
	mockRepo.On("ListProducts", ctx).Return(products, nil)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	listedProducts, err := usecase.ListProducts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, products, listedProducts)
	mockRepo.AssertExpectations(t)
}

func TestInventoryUseCaseImpl_ListProducts_Error(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockInventoryRepository)
	expectedErr := errors.New("failed to list products")
	mockRepo.On("ListProducts", ctx).Return(([]*domain.Product)(nil), expectedErr)

	logger := zap.NewNop()
	usecase := NewInventoryUsecase(mockRepo, logger)

	listedProducts, err := usecase.ListProducts(ctx)
	assert.Error(t, err)
	assert.Nil(t, listedProducts)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}
