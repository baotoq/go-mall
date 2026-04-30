package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "gomall/api/catalog/v1"
	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/service"
)

// minimal stub repos to feed real usecases
type nopProductRepo struct{ err error }

func (r *nopProductRepo) Save(_ context.Context, p *biz.Product) (*biz.Product, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return p, r.err
}
func (r *nopProductRepo) FindByID(_ context.Context, _ uuid.UUID) (*biz.Product, error) {
	return &biz.Product{}, r.err
}
func (r *nopProductRepo) Update(_ context.Context, p *biz.Product) (*biz.Product, error) {
	return p, r.err
}
func (r *nopProductRepo) Delete(_ context.Context, _ uuid.UUID) error { return r.err }
func (r *nopProductRepo) List(_ context.Context, f biz.ListProductsFilter) (*biz.ListProductsResult, error) {
	return &biz.ListProductsResult{Page: f.Page, PageSize: f.PageSize}, r.err
}

type nopCategoryRepo struct{}

func (r *nopCategoryRepo) Save(_ context.Context, c *biz.Category) (*biz.Category, error) {
	c.ID = uuid.New()
	return c, nil
}
func (r *nopCategoryRepo) FindByID(_ context.Context, _ uuid.UUID) (*biz.Category, error) {
	return &biz.Category{}, nil
}
func (r *nopCategoryRepo) Update(_ context.Context, c *biz.Category) (*biz.Category, error) {
	return c, nil
}
func (r *nopCategoryRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *nopCategoryRepo) List(_ context.Context, f biz.ListCategoriesFilter) (*biz.ListCategoriesResult, error) {
	return &biz.ListCategoriesResult{Page: f.Page, PageSize: f.PageSize}, nil
}

func newSvc(prodErr error) *service.CatalogService {
	return service.NewCatalogService(
		biz.NewProductUsecase(&nopProductRepo{err: prodErr}),
		biz.NewCategoryUsecase(&nopCategoryRepo{}),
	)
}

func TestCatalogService_ListProducts_validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *v1.ListProductsRequest
		wantErr bool
	}{
		{
			name:    "page_size too large",
			req:     &v1.ListProductsRequest{PageSize: 101},
			wantErr: true,
		},
		{
			name:    "invalid sort value",
			req:     &v1.ListProductsRequest{Sort: "invalid_sort"},
			wantErr: true,
		},
		{
			name:    "valid request",
			req:     &v1.ListProductsRequest{Page: 1, PageSize: 20, Sort: "price_asc"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange + Act
			_, err := newSvc(nil).ListProducts(context.Background(), tc.req)

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCatalogService_GetProduct_invalidUUID(t *testing.T) {
	// Arrange
	svc := newSvc(nil)

	// Act
	_, err := svc.GetProduct(context.Background(), &v1.GetProductRequest{Id: "not-a-uuid"})

	// Assert
	require.Error(t, err)
}

func TestCatalogService_DeleteProduct_returnsEmpty(t *testing.T) {
	// Arrange
	svc := newSvc(nil)

	// Act
	got, err := svc.DeleteProduct(context.Background(), &v1.DeleteProductRequest{Id: "00000000-0000-0000-0000-000000000001"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &emptypb.Empty{}, got)
}

func TestCatalogService_CreateProduct_emptyName(t *testing.T) {
	// Arrange
	svc := newSvc(nil)

	// Act
	_, err := svc.CreateProduct(context.Background(), &v1.CreateProductRequest{Name: ""})

	// Assert
	assert.Error(t, err)
}
