package mappers

import (
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"time"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ../mappers_gen/order_mapper_gen.go
// goverter:output:package mappers_gen
// goverter:extend OrderStatusToString
// goverter:extend StringToOrderStatus
// goverter:extend ConvertTime
// goverter:extend ZeroTime
type Converter interface {
	// goverter:ignore CreatedAt
	// goverter:ignore UpdatedAt
	// goverter:ignore DeletedAt
	// goverter:map Status Status
	// goverter:map Items Items
	DomainToGorm(source domain.Order) models.GormDBOrder

	// goverter:map Status Status
	// goverter:map Items Items
	GormToDomain(source models.GormDBOrder) domain.Order

	// goverter:ignore CreatedAt
	// goverter:ignore UpdatedAt
	// goverter:ignore DeletedAt
	DomainToGormOrderItem(source domain.OrderItem) models.GormDBOrderItem

	GormToDomainOrderItem(source models.GormDBOrderItem) domain.OrderItem
}

// OrderStatusToString converts an OrderStatus to string.
func OrderStatusToString(status domain.OrderStatus) string {
	return string(status)
}

// StringToOrderStatus converts a string to an OrderStatus.
func StringToOrderStatus(status string) domain.OrderStatus {
	return domain.OrderStatus(status)
}

// ConvertTime is a passthrough for time.Time.
func ConvertTime(t time.Time) time.Time {
	return t
}

// ZeroTime returns the zero value for time.Time.
func ZeroTime(_ time.Time) time.Time {
	return time.Time{}
}
