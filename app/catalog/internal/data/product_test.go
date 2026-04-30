package data_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data"
)

var _ = uuid.Nil // ensure uuid imported

func TestProductRepo_CreateGetDelete(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	logger := newTestLogger()
	repo := data.NewProductRepo(data.NewTestData(testClient), logger)

	// Arrange
	p := &biz.Product{Name: "MacBook Pro", Slug: "mbp-14", PriceCents: 199900, Currency: "USD", Theme: "dark"}

	// Act — create
	created, err := repo.Save(ctx, p)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.ID)
	assert.Equal(t, "MacBook Pro", created.Name)

	// Act — get
	got, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Act — delete
	require.NoError(t, repo.Delete(ctx, created.ID))

	// Assert — not found after delete
	_, err = repo.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, biz.ErrProductNotFound)
}

func TestProductRepo_List_filters(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	logger := newTestLogger()
	catRepo := data.NewCategoryRepo(data.NewTestData(testClient), logger)
	repo := data.NewProductRepo(data.NewTestData(testClient), logger)

	// Arrange — category + products
	cat, err := catRepo.Save(ctx, &biz.Category{Name: "Mac", Slug: "mac"})
	require.NoError(t, err)

	for _, p := range []*biz.Product{
		{Name: "MacBook Pro", Slug: "mbp", PriceCents: 199900, Currency: "USD", Theme: "dark", CategoryID: &cat.ID},
		{Name: "MacBook Air", Slug: "mba", PriceCents: 129900, Currency: "USD", Theme: "light", CategoryID: &cat.ID},
		{Name: "iPad", Slug: "ipad", PriceCents: 59900, Currency: "USD", Theme: "light"},
	} {
		_, err := repo.Save(ctx, p)
		require.NoError(t, err)
	}

	// Act — search by q
	res, err := repo.List(ctx, biz.ListProductsFilter{Q: "macbook", Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)

	// Act — filter by category
	res, err = repo.List(ctx, biz.ListProductsFilter{CategoryID: &cat.ID, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)

	// Act — filter by min price
	minPrice := int64(100000)
	res, err = repo.List(ctx, biz.ListProductsFilter{MinPrice: &minPrice, Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
}

func TestProductRepo_SlugConflict(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	logger := newTestLogger()
	repo := data.NewProductRepo(data.NewTestData(testClient), logger)

	// Arrange
	_, err := repo.Save(ctx, &biz.Product{Name: "A", Slug: "dup-slug", PriceCents: 0, Currency: "USD"})
	require.NoError(t, err)

	// Act — duplicate slug
	_, err = repo.Save(ctx, &biz.Product{Name: "B", Slug: "dup-slug", PriceCents: 0, Currency: "USD"})

	// Assert — conflict error
	assert.ErrorIs(t, err, biz.ErrSlugConflict)
}
