package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrderItem(productID string, quantity int) *OrderItem {
	now := Clock.Now()
	return &OrderItem{
		ID:        uuid.New().String(),
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
