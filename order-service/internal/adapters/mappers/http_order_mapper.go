package mappers

import (
	"fmt"
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
)

func HTTPCreateOrderRequestToDomain(req models.CreateOrderRequest) (*domain.Order, []*domain.OrderItem, error) {
	if req.UserID == "" {
		return nil, nil, fmt.Errorf("userID is required")
	}
	order := domain.NewOrder(req.UserID)
	var items []*domain.OrderItem
	for _, r := range req.Items {
		if r.ProductID == "" || r.Quantity <= 0 {
			return nil, nil, fmt.Errorf("invalid order item data")
		}
		item := domain.NewOrderItem(r.ProductID, r.Quantity)
		items = append(items, item)
	}
	order.Items = items
	return order, items, nil
}
