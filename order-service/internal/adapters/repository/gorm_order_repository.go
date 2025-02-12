package repository

import (
	"context"
	"fmt"
	"order-service/internal/adapters/columns"
	mappersgen "order-service/internal/adapters/mappers_gen"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"order-service/internal/usecases"
	"sync"

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

func (r *GormOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	dbOrder := r.mapper.DomainToGorm(*order)

	if err := r.db.WithContext(ctx).Create(&dbOrder).Error; err != nil {
		r.logger.Error("failed to create order", zap.String("orderID", order.ID), zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	r.logger.Info("created order", zap.String("orderID", order.ID))
	return order, nil
}

func (r *GormOrderRepository) CreateOrderWithItems(ctx context.Context, order *domain.Order, items []*domain.OrderItem) (*domain.Order, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbOrder := r.mapper.DomainToGorm(*order)
		if err := tx.Create(&dbOrder).Error; err != nil {
			r.logger.Error("failed to create order in transaction", zap.Error(err))
			return err
		}

		var (
			wg       sync.WaitGroup
			mu       sync.Mutex
			firstErr error
		)
		wg.Add(len(items))

		for _, it := range items {
			localItem := it
			go func() {
				defer wg.Done()
				dbItem := r.mapper.DomainToGormOrderItem(*localItem)
				dbItem.OrderID = dbOrder.ID
				if err2 := tx.Create(&dbItem).Error; err2 != nil {
					mu.Lock()
					defer mu.Unlock()
					if firstErr == nil {
						firstErr = err2
					}
				}
			}()
		}

		wg.Wait()
		if firstErr != nil {
			r.logger.Error("failed to insert order item", zap.Error(firstErr))
			return firstErr
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	r.logger.Info("Created order with items in one transaction", zap.String("orderID", order.ID))
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
