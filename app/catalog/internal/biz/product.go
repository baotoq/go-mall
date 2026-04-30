package biz

import (
	"context"
	"time"

	v1 "gomall/api/catalog/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.NotFound(v1.ErrorReason_PRODUCT_NOT_FOUND.String(), "product not found")
	ErrSlugConflict    = errors.Conflict(v1.ErrorReason_CONFLICT.String(), "slug already exists")
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	PriceCents  int64
	Currency    string
	ImageURL    string
	Theme       string
	Stock       int
	CategoryID  *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProductsFilter struct {
	Q          string
	CategoryID *uuid.UUID
	MinPrice   *int64
	MaxPrice   *int64
	Sort       string
	Page       int32
	PageSize   int32
}

type ListProductsResult struct {
	Products []*Product
	Total    int64
	Page     int32
	PageSize int32
}

type ProductRepo interface {
	Save(context.Context, *Product) (*Product, error)
	FindByID(context.Context, uuid.UUID) (*Product, error)
	Update(context.Context, *Product) (*Product, error)
	Delete(context.Context, uuid.UUID) error
	List(context.Context, ListProductsFilter) (*ListProductsResult, error)
}

type ProductUsecase struct {
	repo ProductRepo
}

func NewProductUsecase(repo ProductRepo) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (uc *ProductUsecase) Create(ctx context.Context, p *Product) (*Product, error) {
	return uc.repo.Save(ctx, p)
}

func (uc *ProductUsecase) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *ProductUsecase) Update(ctx context.Context, p *Product) (*Product, error) {
	return uc.repo.Update(ctx, p)
}

func (uc *ProductUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ProductUsecase) List(ctx context.Context, f ListProductsFilter) (*ListProductsResult, error) {
	return uc.repo.List(ctx, f)
}
