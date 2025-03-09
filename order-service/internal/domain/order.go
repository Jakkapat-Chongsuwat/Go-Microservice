package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated OrderStatus = "CREATED"
	OrderStatusPaid    OrderStatus = "PAID"
	OrderStatusShipped OrderStatus = "SHIPPED"
)

type Order struct {
	ID        string
	UserID    string
	Status    OrderStatus
	Items     []*OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(userID string) *Order {
	now := Clock.Now()
	return &Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    OrderStatusCreated,
		Items:     []*OrderItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
