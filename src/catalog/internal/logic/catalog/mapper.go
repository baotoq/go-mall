package catalog

import (
	"catalog/ent"
	"catalog/internal/types"
)

func mapToProductInfo(p *ent.Product) types.ProductInfo {
	info := types.ProductInfo{
		Id:             p.ID.String(),
		Name:           p.Name,
		Slug:           p.Slug,
		Description:    p.Description,
		ImageUrl:       p.ImageURL,
		Price:          p.Price,
		TotalStock:     p.TotalStock,
		RemainingStock: p.RemainingStock,
		CreatedAt:      p.CreatedAt.UnixMilli(),
	}
	if p.UpdatedAt != nil {
		info.UpdatedAt = p.UpdatedAt.UnixMilli()
	}
	if p.Edges.Category != nil {
		info.CategoryId = p.Edges.Category.ID.String()
	}
	return info
}

func mapToCategoryInfo(c *ent.Category) types.CategoryInfo {
	info := types.CategoryInfo{
		Id:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.UnixMilli(),
	}
	if c.UpdatedAt != nil {
		info.UpdatedAt = c.UpdatedAt.UnixMilli()
	}
	return info
}
