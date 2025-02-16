package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormDBOrder struct {
	ID        string             `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string             `gorm:"column:user_id"`
	Status    string             `gorm:"column:status"`
	Items     []*GormDBOrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt time.Time          `gorm:"column:created_at"`
	UpdatedAt time.Time          `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt     `gorm:"column:deleted_at;index"`
}

type GormDBOrderItem struct {
	ID        string         `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID   string         `gorm:"column:order_id;index;type:uuid"`
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

func (o *GormDBOrder) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	return nil
}

func (oi *GormDBOrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if oi.ID == "" {
		oi.ID = uuid.NewString()
	}
	return nil
}
