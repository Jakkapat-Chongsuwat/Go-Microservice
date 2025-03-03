package repository

import (
	"context"
	"inventory-service/internal/domain"
	"inventory-service/internal/usecases"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GormInventoryRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

var _ usecases.InventoryRepository = (*GormInventoryRepository)(nil)

func NewGormInventoryRepo(db *gorm.DB, logger *zap.Logger) *GormInventoryRepository {
	return &GormInventoryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *GormInventoryRepository) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		r.logger.Error("failed to create product", zap.Error(err))
		return nil, err
	}
	return product, nil
}

func (r *GormInventoryRepository) GetProduct(ctx context.Context, productId string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.WithContext(ctx).First(&product, "id = ?", productId).Error; err != nil {
		r.logger.Error("failed to get product", zap.String("productId", productId), zap.Error(err))
		return nil, err
	}
	return &product, nil
}

func (r *GormInventoryRepository) UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		r.logger.Error("failed to update product", zap.String("productId", product.ID), zap.Error(err))
		return nil, err
	}
	return product, nil
}

func (r *GormInventoryRepository) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	var products []*domain.Product
	if err := r.db.WithContext(ctx).Find(&products).Error; err != nil {
		r.logger.Error("failed to list products", zap.Error(err))
		return nil, err
	}
	return products, nil
}
