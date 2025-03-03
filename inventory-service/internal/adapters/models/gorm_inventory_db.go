package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormDBProduct struct {
	ID        string         `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string         `gorm:"column:name"`
	Quantity  int            `gorm:"column:quantity"`
	Price     float64        `gorm:"column:price"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (GormDBProduct) TableName() string {
	return "products"
}

func (p *GormDBProduct) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
