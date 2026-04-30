package data_test

import (
	"context"
	"testing"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data"
)

func newTestLogger() klog.Logger { return klog.DefaultLogger }

var _ klog.Logger = klog.DefaultLogger

func TestCategoryRepo_CreateGet(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	logger := newTestLogger()
	repo := data.NewCategoryRepo(data.NewTestData(testClient), logger)

	// Arrange
	cat := &biz.Category{Name: "Mac", Slug: "mac", Description: "Apple Mac line"}

	// Act
	created, err := repo.Save(ctx, cat)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.ID)

	got, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mac", got.Name)
	assert.Equal(t, "mac", got.Slug)
}
