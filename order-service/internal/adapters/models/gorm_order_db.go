package models

import (
	"time"

	"gorm.io/gorm"
)

type GormDBOrder struct {
	ID        string             `gorm:"column:id;primaryKey"`
	UserID    string             `gorm:"column:user_id"`
	Status    string             `gorm:"column:status"`
	Items     []*GormDBOrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt time.Time          `gorm:"column:created_at"`
	UpdatedAt time.Time          `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt     `gorm:"column:deleted_at;index"`
}

type GormDBOrderItem struct {
	ID        string         `gorm:"primaryKey"`
	OrderID   string         `gorm:"column:order_id;index"`
	ProductID string         `gorm:"column:product_id"`
	Quantity  int            `gorm:"column:quantity"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (GormDBOrder) TableName() string {
	return "orders"
}

func (GormDBOrderItem) TableName() string {
	return "order_items"
}
