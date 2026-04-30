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

type stubCategoryRepo struct {
	cats map[uuid.UUID]*biz.Category
}

func newStubCategoryRepo() *stubCategoryRepo {
	return &stubCategoryRepo{cats: make(map[uuid.UUID]*biz.Category)}
}

func (r *stubCategoryRepo) Save(_ context.Context, c *biz.Category) (*biz.Category, error) {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	r.cats[c.ID] = c
	return c, nil
}

func (r *stubCategoryRepo) FindByID(_ context.Context, id uuid.UUID) (*biz.Category, error) {
	c, ok := r.cats[id]
	if !ok {
		return nil, biz.ErrCategoryNotFound
	}
	return c, nil
}

func (r *stubCategoryRepo) Update(_ context.Context, c *biz.Category) (*biz.Category, error) {
	if _, ok := r.cats[c.ID]; !ok {
		return nil, biz.ErrCategoryNotFound
	}
	c.UpdatedAt = time.Now()
	r.cats[c.ID] = c
	return c, nil
}

func (r *stubCategoryRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.cats[id]; !ok {
		return biz.ErrCategoryNotFound
	}
	delete(r.cats, id)
	return nil
}

func (r *stubCategoryRepo) List(_ context.Context, f biz.ListCategoriesFilter) (*biz.ListCategoriesResult, error) {
	var cs []*biz.Category
	for _, c := range r.cats {
		cs = append(cs, c)
	}
	return &biz.ListCategoriesResult{Categories: cs, Total: int64(len(cs)), Page: f.Page, PageSize: f.PageSize}, nil
}

func TestCategoryUsecase_Create(t *testing.T) {
	// Arrange
	repo := newStubCategoryRepo()
	uc := biz.NewCategoryUsecase(repo)
	cat := &biz.Category{Name: "Mac", Slug: "mac"}

	// Act
	got, err := uc.Create(context.Background(), cat)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, "Mac", got.Name)
}

func TestCategoryUsecase_Get_notFound(t *testing.T) {
	// Arrange
	repo := newStubCategoryRepo()
	uc := biz.NewCategoryUsecase(repo)

	// Act
	_, err := uc.Get(context.Background(), uuid.New())

	// Assert
	assert.ErrorIs(t, err, biz.ErrCategoryNotFound)
}
