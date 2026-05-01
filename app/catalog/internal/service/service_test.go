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
func (r *nopProductRepo) Delete(_ context.Context, _ uuid.UUID) error    { return r.err }
func (r *nopProductRepo) DeleteAll(_ context.Context) (int, error)        { return 0, r.err }
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
func (r *nopCategoryRepo) Delete(_ context.Context, _ uuid.UUID) error    { return nil }
func (r *nopCategoryRepo) DeleteAll(_ context.Context) (int, error)        { return 0, nil }
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

func TestCatalogService_ListProducts_invalidCategoryID(t *testing.T) {
	_, err := newSvc(nil).ListProducts(context.Background(), &v1.ListProductsRequest{CategoryId: "bad"})
	assert.Error(t, err)
}

func TestCatalogService_ListProducts_filtersApplied(t *testing.T) {
	got, err := newSvc(nil).ListProducts(context.Background(), &v1.ListProductsRequest{
		CategoryId: uuid.NewString(), MinPrice: 100, MaxPrice: 500, Page: 1, PageSize: 10, Sort: "price_asc",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.Page)
}

func TestCatalogService_CreateProduct_okWithCategory(t *testing.T) {
	got, err := newSvc(nil).CreateProduct(context.Background(), &v1.CreateProductRequest{
		Name: "P", Slug: "p", PriceCents: 100, CategoryId: uuid.NewString(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, got.Id)
}

func TestCatalogService_CreateProduct_invalidCategoryID(t *testing.T) {
	_, err := newSvc(nil).CreateProduct(context.Background(), &v1.CreateProductRequest{Name: "P", CategoryId: "bad"})
	assert.Error(t, err)
}

func TestCatalogService_UpdateProduct_validation(t *testing.T) {
	svc := newSvc(nil)
	cases := []struct {
		name string
		req  *v1.UpdateProductRequest
	}{
		{"invalid id", &v1.UpdateProductRequest{Id: "bad", Name: "x"}},
		{"empty name", &v1.UpdateProductRequest{Id: uuid.NewString(), Name: ""}},
		{"invalid category", &v1.UpdateProductRequest{Id: uuid.NewString(), Name: "x", CategoryId: "bad"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.UpdateProduct(context.Background(), tc.req)
			assert.Error(t, err)
		})
	}
}

func TestCatalogService_UpdateProduct_ok(t *testing.T) {
	got, err := newSvc(nil).UpdateProduct(context.Background(), &v1.UpdateProductRequest{
		Id: uuid.NewString(), Name: "P", CategoryId: uuid.NewString(),
	})
	require.NoError(t, err)
	assert.Equal(t, "P", got.Name)
}

func TestCatalogService_GetProduct_ok(t *testing.T) {
	got, err := newSvc(nil).GetProduct(context.Background(), &v1.GetProductRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCatalogService_DeleteProduct_invalidID(t *testing.T) {
	_, err := newSvc(nil).DeleteProduct(context.Background(), &v1.DeleteProductRequest{Id: "bad"})
	assert.Error(t, err)
}

func TestCatalogService_ListCategories(t *testing.T) {
	svc := newSvc(nil)
	got, err := svc.ListCategories(context.Background(), &v1.ListCategoriesRequest{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.Page)

	_, err = svc.ListCategories(context.Background(), &v1.ListCategoriesRequest{PageSize: 200})
	assert.Error(t, err)
}

func TestCatalogService_GetCategory(t *testing.T) {
	svc := newSvc(nil)
	_, err := svc.GetCategory(context.Background(), &v1.GetCategoryRequest{Id: "bad"})
	assert.Error(t, err)

	got, err := svc.GetCategory(context.Background(), &v1.GetCategoryRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCatalogService_CreateCategory(t *testing.T) {
	svc := newSvc(nil)
	_, err := svc.CreateCategory(context.Background(), &v1.CreateCategoryRequest{Name: ""})
	assert.Error(t, err)

	got, err := svc.CreateCategory(context.Background(), &v1.CreateCategoryRequest{Name: "C"})
	require.NoError(t, err)
	assert.NotEmpty(t, got.Id)
}

func TestCatalogService_UpdateCategory(t *testing.T) {
	svc := newSvc(nil)
	_, err := svc.UpdateCategory(context.Background(), &v1.UpdateCategoryRequest{Id: "bad", Name: "x"})
	assert.Error(t, err)

	_, err = svc.UpdateCategory(context.Background(), &v1.UpdateCategoryRequest{Id: uuid.NewString(), Name: ""})
	assert.Error(t, err)

	got, err := svc.UpdateCategory(context.Background(), &v1.UpdateCategoryRequest{Id: uuid.NewString(), Name: "C"})
	require.NoError(t, err)
	assert.Equal(t, "C", got.Name)
}

func TestCatalogService_DeleteCategory(t *testing.T) {
	svc := newSvc(nil)
	_, err := svc.DeleteCategory(context.Background(), &v1.DeleteCategoryRequest{Id: "bad"})
	assert.Error(t, err)

	got, err := svc.DeleteCategory(context.Background(), &v1.DeleteCategoryRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	assert.Equal(t, &emptypb.Empty{}, got)
}
