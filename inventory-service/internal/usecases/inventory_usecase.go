package usecases

import (
	"context"
	"inventory-service/internal/domain"

	"go.uber.org/zap"
)

type InventoryRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, productId string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error)
	ListProducts(ctx context.Context) ([]*domain.Product, error)
}

type InventoryUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, productId string) (*domain.Product, error)
	UpdateProductMetadata(ctx context.Context, product *domain.Product) (*domain.Product, error)
	UpdateProductStockQuantity(ctx context.Context, productID string, quantityChange int) (*domain.Product, error)
	ListProducts(ctx context.Context) ([]*domain.Product, error)
}

type InventoryUseCaseImpl struct {
	inventoryRepo InventoryRepository
	logger        *zap.Logger
}

func NewInventoryUsecase(
	repo InventoryRepository,
	logger *zap.Logger,
) InventoryUseCase {
	return &InventoryUseCaseImpl{
		inventoryRepo: repo,
		logger:        logger,
	}
}

func (i *InventoryUseCaseImpl) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	i.logger.Info("CreateProduct called", zap.String("productID", product.ID))
	return i.inventoryRepo.CreateProduct(ctx, product)
}

func (i *InventoryUseCaseImpl) GetProduct(ctx context.Context, productId string) (*domain.Product, error) {
	i.logger.Info("GetProduct called", zap.String("productID", productId))
	return i.inventoryRepo.GetProduct(ctx, productId)
}

func (i *InventoryUseCaseImpl) UpdateProductMetadata(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	i.logger.Info("UpdateProductMetadata called", zap.String("productID", product.ID))
	existingProduct, err := i.inventoryRepo.GetProduct(ctx, product.ID)
	if err != nil {
		i.logger.Error("Failed to get product", zap.String("productID", product.ID), zap.Error(err))
		return nil, err
	}
	existingProduct.Name = product.Name
	existingProduct.Price = product.Price
	existingProduct.UpdatedAt = domain.Clock.Now()
	return i.inventoryRepo.UpdateProduct(ctx, existingProduct)
}

func (i *InventoryUseCaseImpl) UpdateProductStockQuantity(ctx context.Context, productID string, quantityChange int) (*domain.Product, error) {
	i.logger.Info("UpdateProductStockQuantity called", zap.String("productID", productID))
	product, err := i.inventoryRepo.GetProduct(ctx, productID)
	if err != nil {
		i.logger.Error("Failed to get product", zap.String("productID", productID), zap.Error(err))
		return nil, err
	}
	if err := product.AdjustStock(quantityChange); err != nil {
		i.logger.Error("Failed to adjust stock", zap.String("productID", productID), zap.Error(err))
		return nil, err
	}
	return i.inventoryRepo.UpdateProduct(ctx, product)
}

func (i *InventoryUseCaseImpl) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	i.logger.Info("ListProducts called")
	return i.inventoryRepo.ListProducts(ctx)
}
