package repository

import (
	"context"
	"fmt"
	"order-service/internal/adapters/columns"
	mappersgen "order-service/internal/adapters/mappers_gen"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"order-service/internal/usecases"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GormOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
	mapper *mappersgen.ConverterImpl
}

var _ usecases.OrderRepository = (*GormOrderRepository)(nil)

func NewGormOrderRepo(db *gorm.DB, logger *zap.Logger) *GormOrderRepository {
	return &GormOrderRepository{
		db:     db,
		logger: logger,
		mapper: &mappersgen.ConverterImpl{},
	}
}

func (r *GormOrderRepository) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Convert domain order to GORM order
		dbOrder := r.mapper.DomainToGorm(*order)

		// Omit the "Items" association to prevent auto-insertion
		if err := tx.Omit("Items").Create(&dbOrder).Error; err != nil {
			r.logger.Error("failed to create order in transaction", zap.Error(err))
			return fmt.Errorf("failed to insert order: %w", err)
		}

		for _, it := range items {
			dbItem := r.mapper.DomainToGormOrderItem(*it)
			dbItem.OrderID = dbOrder.ID
			if err := tx.Create(&dbItem).Error; err != nil {
				r.logger.Error("failed to insert order item", zap.Error(err), zap.Any("orderItem", dbItem))
				return fmt.Errorf("failed to insert order item: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		r.logger.Error("Transaction failed", zap.Error(err))
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	r.logger.Info("Transaction successful", zap.String("orderID", order.ID))
	return order, nil
}

func (r *GormOrderRepository) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	var dbOrder models.GormDBOrder

	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&dbOrder, columns.ColumnID+" = ?", orderID).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Error("order not found", zap.String("orderID", orderID))
			return nil, err
		}
		r.logger.Error("failed to get order", zap.String("orderID", orderID), zap.Error(err))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	domainOrder := r.mapper.GormToDomain(dbOrder)
	return &domainOrder, nil
}

func (r *GormOrderRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	var dbOrders []models.GormDBOrder

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where(columns.ColumnUserID+" = ?", userID).
		Find(&dbOrders).Error
	if err != nil {
		r.logger.Error("failed to get orders by userID", zap.String("userID", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	results := make([]*domain.Order, 0, len(dbOrders))
	for _, dbo := range dbOrders {
		domainOrder := r.mapper.GormToDomain(dbo)
		results = append(results, &domainOrder)
	}
	return results, nil
}
