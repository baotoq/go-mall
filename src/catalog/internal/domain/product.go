package domain

import (
	"errors"

	"catalog/ent"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	Description    string
	ImageURL       string
	Price          float64
	TotalStock     int64
	RemainingStock int64
	CategoryID     uuid.UUID
}

func NewProduct(name, slug, desc, imageURL string, price float64, initialStock int64, categoryID uuid.UUID) (*Product, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	if price < 0 {
		return nil, errors.New("price must be >= 0")
	}
	if initialStock < 0 {
		return nil, errors.New("initial stock must be >= 0")
	}

	return &Product{
		Name:           name,
		Slug:           slug,
		Description:    desc,
		ImageURL:       imageURL,
		Price:          price,
		TotalStock:     initialStock,
		RemainingStock: initialStock,
		CategoryID:     categoryID,
	}, nil
}

func NewFromEnt(p *ent.Product) *Product {
	out := &Product{
		ID:             p.ID,
		Name:           p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		ImageURL:       p.ImageURL,
		Price:          p.Price,
		TotalStock:     p.TotalStock,
		RemainingStock: p.RemainingStock,
	}
	if p.Edges.Category != nil {
		out.CategoryID = p.Edges.Category.ID
	}
	return out
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
