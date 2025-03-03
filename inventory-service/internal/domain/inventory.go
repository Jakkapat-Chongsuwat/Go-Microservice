package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        string
	Name      string
	Quantity  int
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProduct(name string, quantity int, price float64) *Product {
	now := time.Now()
	return &Product{
		ID:        uuid.NewString(),
		Name:      name,
		Quantity:  quantity,
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

var ErrInsufficientStock = errors.New("insufficient stock")

func (p *Product) AdjustStock(change int) error {
	newQty := p.Quantity + change
	if newQty < 0 {
		return ErrInsufficientStock
	}
	p.Quantity = newQty
	p.UpdatedAt = time.Now()
	return nil
}
