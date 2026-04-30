package biz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/catalog/internal/biz"
)

// stubProductRepo implements ProductRepo for unit tests.
type stubProductRepo struct {
	products map[uuid.UUID]*biz.Product
}

func newStubProductRepo() *stubProductRepo {
	return &stubProductRepo{products: make(map[uuid.UUID]*biz.Product)}
}

func (r *stubProductRepo) Save(_ context.Context, p *biz.Product) (*biz.Product, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	r.products[p.ID] = p
	return p, nil
}

func (r *stubProductRepo) FindByID(_ context.Context, id uuid.UUID) (*biz.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return nil, biz.ErrProductNotFound
	}
	return p, nil
}

func (r *stubProductRepo) Update(_ context.Context, p *biz.Product) (*biz.Product, error) {
	if _, ok := r.products[p.ID]; !ok {
		return nil, biz.ErrProductNotFound
	}
	p.UpdatedAt = time.Now()
	r.products[p.ID] = p
	return p, nil
}

func (r *stubProductRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.products[id]; !ok {
		return biz.ErrProductNotFound
	}
	delete(r.products, id)
	return nil
}

func (r *stubProductRepo) List(_ context.Context, f biz.ListProductsFilter) (*biz.ListProductsResult, error) {
	var ps []*biz.Product
	for _, p := range r.products {
		ps = append(ps, p)
	}
	return &biz.ListProductsResult{Products: ps, Total: int64(len(ps)), Page: f.Page, PageSize: f.PageSize}, nil
}

func TestProductUsecase_Create(t *testing.T) {
	// Arrange
	repo := newStubProductRepo()
	uc := biz.NewProductUsecase(repo)
	p := &biz.Product{Name: "MacBook Pro", Slug: "mbp-14", PriceCents: 199900}

	// Act
	got, err := uc.Create(context.Background(), p)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, "MacBook Pro", got.Name)
}

func TestProductUsecase_Get_notFound(t *testing.T) {
	// Arrange
	repo := newStubProductRepo()
	uc := biz.NewProductUsecase(repo)

	// Act
	_, err := uc.Get(context.Background(), uuid.New())

	// Assert
	assert.ErrorIs(t, err, biz.ErrProductNotFound)
}

func TestProductUsecase_List_returnsPaginated(t *testing.T) {
	// Arrange
	repo := newStubProductRepo()
	uc := biz.NewProductUsecase(repo)
	for i := 0; i < 3; i++ {
		_, _ = repo.Save(context.Background(), &biz.Product{Name: "P", Slug: uuid.NewString()})
	}
	filter := biz.ListProductsFilter{Page: 1, PageSize: 20}

	// Act
	res, err := uc.List(context.Background(), filter)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Total)
}
