package data

import (
	"context"
	"strings"

	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data/ent"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{data: data, log: log.NewHelper(logger)}
}

func (r *categoryRepo) Save(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	q := r.data.db.Category.Create().
		SetName(c.Name).
		SetSlug(c.Slug)
	if c.Description != "" {
		q = q.SetDescription(c.Description)
	}
	out, err := q.Save(ctx)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, biz.ErrSlugConflict
		}
		return nil, err
	}
	return entToCategory(out), nil
}

func (r *categoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*biz.Category, error) {
	out, err := r.data.db.Category.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrCategoryNotFound
		}
		return nil, err
	}
	return entToCategory(out), nil
}

func (r *categoryRepo) Update(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	q := r.data.db.Category.UpdateOneID(c.ID).
		SetName(c.Name).
		SetSlug(c.Slug)
	if c.Description != "" {
		q = q.SetDescription(c.Description)
	} else {
		q = q.ClearDescription()
	}
	out, err := q.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, biz.ErrCategoryNotFound
		}
		if isUniqueViolation(err) {
			return nil, biz.ErrSlugConflict
		}
		return nil, err
	}
	return entToCategory(out), nil
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.data.db.Category.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return biz.ErrCategoryNotFound
	}
	return err
}

func (r *categoryRepo) List(ctx context.Context, f biz.ListCategoriesFilter) (*biz.ListCategoriesResult, error) {
	page, pageSize := normalizePage(f.Page, f.PageSize)
	offset := int((page - 1) * pageSize)

	total, err := r.data.db.Category.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.data.db.Category.Query().
		Offset(offset).
		Limit(int(pageSize)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	cats := make([]*biz.Category, len(rows))
	for i, row := range rows {
		cats[i] = entToCategory(row)
	}

	return &biz.ListCategoriesResult{
		Categories: cats,
		Total:      int64(total),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func entToCategory(e *ent.Category) *biz.Category {
	c := &biz.Category{
		ID:        e.ID,
		Name:      e.Name,
		Slug:      e.Slug,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
	if e.Description != nil {
		c.Description = *e.Description
	}
	return c
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate")
}

func normalizePage(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
