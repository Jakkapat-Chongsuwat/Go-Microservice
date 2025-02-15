package domain

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
		UserID: userID,
		Status: OrderStatusCreated,
		Items:  []*OrderItem{},
	}
}
