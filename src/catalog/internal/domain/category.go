package domain

import (
	"errors"

	"catalog/ent"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
}

func NewCategory(name, slug, desc string) (*Category, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return &Category{
		Name:        name,
		Slug:        slug,
		Description: desc,
	}, nil
}

func NewCategoryFromEnt(c *ent.Category) *Category {
	return &Category{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
	}
}
