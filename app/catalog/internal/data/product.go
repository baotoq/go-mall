package data

import (
	"context"

	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data/ent"
	"gomall/app/catalog/internal/data/ent/predicate"
	entproduct "gomall/app/catalog/internal/data/ent/product"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type productRepo struct {
	data *Data
	log  *log.Helper
}

func NewProductRepo(data *Data, logger log.Logger) biz.ProductRepo {
	return &productRepo{data: data, log: log.NewHelper(logger)}
}

func (r *productRepo) Save(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	q := r.data.db.Product.Create().
		SetName(p.Name).
		SetSlug(p.Slug).
		SetPriceCents(p.PriceCents).
		SetStock(p.Stock)

	if p.Currency != "" {
		q = q.SetCurrency(p.Currency)
	}
	if p.Description != "" {
		q = q.SetDescription(p.Description)
	}
	if p.ImageURL != "" {
		q = q.SetImageURL(p.ImageURL)
	}
	if p.Theme != "" {
		q = q.SetTheme(entproduct.Theme(p.Theme))
	}
	if p.CategoryID != nil {
		q = q.SetCategoryID(*p.CategoryID)
	}

	out, err := q.Save(ctx)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, biz.ErrSlugConflict
		}
		return nil, err
	}
	return entToProduct(out), nil
}

func (r *productRepo) FindByID(ctx context.Context, id uuid.UUID) (*biz.Product, error) {
	out, err := r.data.db.Product.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrProductNotFound
		}
		return nil, err
	}
	return entToProduct(out), nil
}

func (r *productRepo) Update(ctx context.Context, p *biz.Product) (*biz.Product, error) {
	q := r.data.db.Product.UpdateOneID(p.ID).
		SetName(p.Name).
		SetSlug(p.Slug).
		SetPriceCents(p.PriceCents).
		SetStock(p.Stock)

	if p.Theme != "" {
		q = q.SetTheme(entproduct.Theme(p.Theme))
	}

	if p.Currency != "" {
		q = q.SetCurrency(p.Currency)
	}
	if p.Description != "" {
		q = q.SetDescription(p.Description)
	} else {
		q = q.ClearDescription()
	}
	if p.ImageURL != "" {
		q = q.SetImageURL(p.ImageURL)
	} else {
		q = q.ClearImageURL()
	}
	if p.CategoryID != nil {
		q = q.SetCategoryID(*p.CategoryID)
	} else {
		q = q.ClearCategoryID()
	}

	out, err := q.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrProductNotFound
		}
		if isUniqueViolation(err) {
			return nil, biz.ErrSlugConflict
		}
		return nil, err
	}
	return entToProduct(out), nil
}

func (r *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.data.db.Product.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return biz.ErrProductNotFound
	}
	return err
}

func (r *productRepo) List(ctx context.Context, f biz.ListProductsFilter) (*biz.ListProductsResult, error) {
	page, pageSize := normalizePage(f.Page, f.PageSize)
	offset := int((page - 1) * pageSize)

	var preds []predicate.Product
	if f.Q != "" {
		preds = append(preds,
			entproduct.Or(
				entproduct.NameContainsFold(f.Q),
				entproduct.DescriptionContainsFold(f.Q),
			),
		)
	}
	if f.CategoryID != nil {
		preds = append(preds, entproduct.CategoryID(*f.CategoryID))
	}
	if f.MinPrice != nil {
		preds = append(preds, entproduct.PriceCentsGTE(*f.MinPrice))
	}
	if f.MaxPrice != nil {
		preds = append(preds, entproduct.PriceCentsLTE(*f.MaxPrice))
	}

	q := r.data.db.Product.Query().Where(preds...)

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	q = q.Offset(offset).Limit(int(pageSize))
	switch f.Sort {
	case "price_asc":
		q = q.Order(ent.Asc(entproduct.FieldPriceCents))
	case "price_desc":
		q = q.Order(ent.Desc(entproduct.FieldPriceCents))
	default:
		q = q.Order(ent.Desc(entproduct.FieldCreatedAt))
	}

	rows, err := q.All(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]*biz.Product, len(rows))
	for i, row := range rows {
		products[i] = entToProduct(row)
	}

	return &biz.ListProductsResult{
		Products: products,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func entToProduct(e *ent.Product) *biz.Product {
	p := &biz.Product{
		ID:         e.ID,
		Name:       e.Name,
		Slug:       e.Slug,
		PriceCents: e.PriceCents,
		Currency:   e.Currency,
		Theme:      string(e.Theme),
		Stock:      e.Stock,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
	if e.Description != nil {
		p.Description = *e.Description
	}
	if e.ImageURL != nil {
		p.ImageURL = *e.ImageURL
	}
	if e.CategoryID != nil {
		p.CategoryID = e.CategoryID
	}
	return p
}
