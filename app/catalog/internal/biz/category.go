package biz

import (
	"context"
	"time"

	v1 "gomall/api/catalog/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

var ErrCategoryNotFound = errors.NotFound(v1.ErrorReason_CATEGORY_NOT_FOUND.String(), "category not found")

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListCategoriesFilter struct {
	Page     int32
	PageSize int32
}

type ListCategoriesResult struct {
	Categories []*Category
	Total      int64
	Page       int32
	PageSize   int32
}

type CategoryRepo interface {
	Save(context.Context, *Category) (*Category, error)
	FindByID(context.Context, uuid.UUID) (*Category, error)
	Update(context.Context, *Category) (*Category, error)
	Delete(context.Context, uuid.UUID) error
	DeleteAll(context.Context) (int, error)
	List(context.Context, ListCategoriesFilter) (*ListCategoriesResult, error)
}

type CategoryUsecase struct {
	repo CategoryRepo
}

func NewCategoryUsecase(repo CategoryRepo) *CategoryUsecase {
	return &CategoryUsecase{repo: repo}
}

func (uc *CategoryUsecase) Create(ctx context.Context, c *Category) (*Category, error) {
	return uc.repo.Save(ctx, c)
}

func (uc *CategoryUsecase) Get(ctx context.Context, id uuid.UUID) (*Category, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *CategoryUsecase) Update(ctx context.Context, c *Category) (*Category, error) {
	return uc.repo.Update(ctx, c)
}

func (uc *CategoryUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *CategoryUsecase) DeleteAll(ctx context.Context) (int, error) {
	return uc.repo.DeleteAll(ctx)
}

func (uc *CategoryUsecase) List(ctx context.Context, f ListCategoriesFilter) (*ListCategoriesResult, error) {
	return uc.repo.List(ctx, f)
}
