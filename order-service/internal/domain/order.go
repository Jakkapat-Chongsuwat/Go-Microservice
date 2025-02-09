// order-service/internal/domain/order.go

package domain

const (
	OrderStatusCreated = "CREATED"
	OrderStatusPaid    = "PAID"
	OrderStatusShipped = "SHIPPED"
)

type Order struct {
	ID        string
	UserID    string
	ProductID string
	Quantity  int
	Status    string
}
