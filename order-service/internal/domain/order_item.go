package domain

import "github.com/google/uuid"

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
}

func NewOrderItem(productID string, quantity int) *OrderItem {
	return &OrderItem{
		ID:        uuid.New().String(),
		ProductID: productID,
		Quantity:  quantity,
	}
}
