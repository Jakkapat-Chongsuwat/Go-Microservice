package domain

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderStatusCreated OrderStatus = "CREATED"
	OrderStatusPaid    OrderStatus = "PAID"
	OrderStatusShipped OrderStatus = "SHIPPED"
)

type Order struct {
	ID     string
	UserID string
	Status OrderStatus
	Items  []*OrderItem
}

func NewOrder(userID string) *Order {
	return &Order{
		ID:     uuid.New().String(),
		UserID: userID,
		Status: OrderStatusCreated,
		Items:  []*OrderItem{},
	}
}
