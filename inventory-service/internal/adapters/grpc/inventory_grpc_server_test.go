package grpc

import (
	"context"
	"errors"
	"testing"

	"inventory-service/internal/domain"

	inventory_service "github.com/jakkapat-chongsuwat/go-microservice/proto/inventory_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockInventoryUseCase struct {
	mock.Mock
}

func (m *MockInventoryUseCase) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if prod, ok := args.Get(0).(*domain.Product); ok {
		return prod, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryUseCase) GetProduct(ctx context.Context, productId string) (*domain.Product, error) {
	args := m.Called(ctx, productId)
	if prod, ok := args.Get(0).(*domain.Product); ok {
		return prod, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryUseCase) UpdateProductMetadata(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, product)
	if prod, ok := args.Get(0).(*domain.Product); ok {
		return prod, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryUseCase) UpdateProductStockQuantity(ctx context.Context, productID string, quantityChange int) (*domain.Product, error) {
	args := m.Called(ctx, productID, quantityChange)
	if prod, ok := args.Get(0).(*domain.Product); ok {
		return prod, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockInventoryUseCase) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	args := m.Called(ctx)
	if prods, ok := args.Get(0).([]*domain.Product); ok {
		return prods, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestInventoryGRPCServer_CreateProduct(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.CreateProductRequest{
		Name:     "Test Product",
		Quantity: 100,
		Price:    9.99,
	}

	mockUC.On("CreateProduct", ctx, mock.MatchedBy(func(p *domain.Product) bool {
		return p.Name == req.GetName() &&
			p.Quantity == int(req.GetQuantity()) &&
			p.Price == req.GetPrice()
	})).Return(&domain.Product{
		ID:       "123",
		Name:     req.GetName(),
		Quantity: int(req.GetQuantity()),
		Price:    req.GetPrice(),
	}, nil)

	resp, err := server.CreateProduct(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "123", resp.Product.Id)
	assert.Equal(t, req.GetName(), resp.Product.Name)
	assert.Equal(t, req.GetQuantity(), int32(resp.Product.Quantity))
	assert.Equal(t, req.GetPrice(), resp.Product.Price)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_CreateProduct_Error(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.CreateProductRequest{
		Name:     "Test Product",
		Quantity: 100,
		Price:    9.99,
	}

	expectedErr := errors.New("create error")
	mockUC.On("CreateProduct", ctx, mock.Anything).Return((*domain.Product)(nil), expectedErr)

	resp, err := server.CreateProduct(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_GetProduct(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.GetProductRequest{
		Id: "123",
	}
	expectedProduct := &domain.Product{
		ID:       "123",
		Name:     "Test Product",
		Quantity: 50,
		Price:    19.99,
	}
	mockUC.On("GetProduct", ctx, req.GetId()).Return(expectedProduct, nil)

	resp, err := server.GetProduct(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedProduct.ID, resp.Product.Id)
	assert.Equal(t, expectedProduct.Name, resp.Product.Name)
	assert.Equal(t, int32(expectedProduct.Quantity), resp.Product.Quantity)
	assert.Equal(t, expectedProduct.Price, resp.Product.Price)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_GetProduct_Error(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.GetProductRequest{
		Id: "123",
	}
	expectedErr := errors.New("get error")
	mockUC.On("GetProduct", ctx, req.GetId()).Return((*domain.Product)(nil), expectedErr)

	resp, err := server.GetProduct(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_UpdateProductMetadata(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.UpdateProductMetadataRequest{
		Id:    "123",
		Name:  "Updated Name",
		Price: 29.99,
	}
	existingProduct := &domain.Product{
		ID:       "123",
		Name:     "Old Name",
		Quantity: 50,
		Price:    19.99,
	}
	mockUC.On("GetProduct", ctx, req.GetId()).Return(existingProduct, nil)

	updatedProduct := &domain.Product{
		ID:       "123",
		Name:     req.GetName(),
		Quantity: existingProduct.Quantity,
		Price:    req.GetPrice(),
	}
	mockUC.On("UpdateProductMetadata", ctx, mock.MatchedBy(func(p *domain.Product) bool {
		return p.ID == "123" &&
			p.Name == req.GetName() &&
			p.Price == req.GetPrice() &&
			p.Quantity == existingProduct.Quantity
	})).Return(updatedProduct, nil)

	resp, err := server.UpdateProductMetadata(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, updatedProduct.ID, resp.Product.Id)
	assert.Equal(t, updatedProduct.Name, resp.Product.Name)
	assert.Equal(t, int32(updatedProduct.Quantity), resp.Product.Quantity)
	assert.Equal(t, updatedProduct.Price, resp.Product.Price)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_UpdateProductMetadata_GetError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.UpdateProductMetadataRequest{
		Id:    "123",
		Name:  "Updated Name",
		Price: 29.99,
	}
	expectedErr := errors.New("get error")
	mockUC.On("GetProduct", ctx, req.GetId()).Return((*domain.Product)(nil), expectedErr)

	resp, err := server.UpdateProductMetadata(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_UpdateProductStockQuantity(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.UpdateProductStockQuantityRequest{
		Id:             "123",
		QuantityChange: 10,
	}
	updatedProduct := &domain.Product{
		ID:       "123",
		Name:     "Test Product",
		Quantity: 60,
		Price:    19.99,
	}
	mockUC.On("UpdateProductStockQuantity", ctx, req.GetId(), int(req.GetQuantityChange())).Return(updatedProduct, nil)

	resp, err := server.UpdateProductStockQuantity(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, updatedProduct.ID, resp.Product.Id)
	assert.Equal(t, updatedProduct.Name, resp.Product.Name)
	assert.Equal(t, int32(updatedProduct.Quantity), resp.Product.Quantity)
	assert.Equal(t, updatedProduct.Price, resp.Product.Price)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_UpdateProductStockQuantity_Error(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.UpdateProductStockQuantityRequest{
		Id:             "123",
		QuantityChange: 10,
	}
	expectedErr := errors.New("update error")
	mockUC.On("UpdateProductStockQuantity", ctx, req.GetId(), int(req.GetQuantityChange())).Return((*domain.Product)(nil), expectedErr)

	resp, err := server.UpdateProductStockQuantity(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_ListProducts(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.ListProductsRequest{}
	products := []*domain.Product{
		{ID: "1", Name: "Prod1", Quantity: 10, Price: 9.99},
		{ID: "2", Name: "Prod2", Quantity: 20, Price: 19.99},
	}
	mockUC.On("ListProducts", ctx).Return(products, nil)

	resp, err := server.ListProducts(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(products), len(resp.Products))
	for i, prod := range resp.Products {
		assert.Equal(t, products[i].ID, prod.Id)
		assert.Equal(t, products[i].Name, prod.Name)
		assert.Equal(t, int32(products[i].Quantity), prod.Quantity)
		assert.Equal(t, products[i].Price, prod.Price)
	}

	mockUC.AssertExpectations(t)
}

func TestInventoryGRPCServer_ListProducts_Error(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	mockUC := new(MockInventoryUseCase)
	server := NewInventoryGRPCServer(mockUC, logger)

	req := &inventory_service.ListProductsRequest{}
	expectedErr := errors.New("list error")
	mockUC.On("ListProducts", ctx).Return(([]*domain.Product)(nil), expectedErr)

	resp, err := server.ListProducts(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUC.AssertExpectations(t)
}
