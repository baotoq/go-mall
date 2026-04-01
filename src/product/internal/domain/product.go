package domain

import (
	"errors"
	"product/ent"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID
	Name           string
	Description    string
	Price          float64
	TotalStock     int64
	RemainingStock int64
}

func NewProduct(name, desc string, price float64, initialStock int64) (*Product, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if price < 0 {
		return nil, errors.New("price must be >= 0")
	}
	if initialStock < 0 {
		return nil, errors.New("initial stock must be >= 0")
	}

	return &Product{
		Name:           name,
		Description:    desc,
		Price:          price,
		TotalStock:     initialStock,
		RemainingStock: initialStock,
	}, nil
}

func NewFromEnt(p *ent.Product) *Product {
	return &Product{
		ID:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		Price:          p.Price,
		TotalStock:     p.TotalStock,
		RemainingStock: p.RemainingStock,
	}
}

func (p *Product) UseRemainingStock(qty int64) error {
	if qty <= 0 {
		return errors.New("invalid quantity")
	}

	if p.RemainingStock < qty {
		return errors.New("not enough stock")
	}

	p.RemainingStock -= qty
	return nil
}

func (p *Product) IncreaseStock(qty int64) error {
	if qty <= 0 {
		return errors.New("invalid quantity")
	}

	p.TotalStock += qty
	p.RemainingStock += qty
	return nil
}
