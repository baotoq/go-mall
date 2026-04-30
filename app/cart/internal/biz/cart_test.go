package biz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/cart/internal/biz"
)

type stubCartRepo struct {
	carts map[string]*biz.Cart
}

func newStubCartRepo() *stubCartRepo {
	return &stubCartRepo{carts: make(map[string]*biz.Cart)}
}

func (r *stubCartRepo) FindOrCreateBySession(_ context.Context, sessionID string) (*biz.Cart, error) {
	c, ok := r.carts[sessionID]
	if !ok {
		c = &biz.Cart{ID: uuid.New(), SessionID: sessionID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		r.carts[sessionID] = c
	}
	return c, nil
}

func (r *stubCartRepo) AddItem(_ context.Context, sessionID string, item *biz.CartItem) (*biz.Cart, error) {
	c, ok := r.carts[sessionID]
	if !ok {
		c = &biz.Cart{ID: uuid.New(), SessionID: sessionID}
		r.carts[sessionID] = c
	}
	item.ID = uuid.New()
	item.CartID = c.ID
	item.SubtotalCents = item.PriceCents * int64(item.Quantity)
	c.Items = append(c.Items, item)
	c.TotalCents += item.SubtotalCents
	return c, nil
}

func (r *stubCartRepo) UpdateItem(_ context.Context, sessionID string, productID uuid.UUID, quantity int) (*biz.Cart, error) {
	c, ok := r.carts[sessionID]
	if !ok {
		return nil, biz.ErrCartNotFound
	}
	for _, it := range c.Items {
		if it.ProductID == productID {
			it.Quantity = quantity
			it.SubtotalCents = it.PriceCents * int64(quantity)
			return c, nil
		}
	}
	return nil, biz.ErrItemNotFound
}

func (r *stubCartRepo) RemoveItem(_ context.Context, sessionID string, productID uuid.UUID) (*biz.Cart, error) {
	c, ok := r.carts[sessionID]
	if !ok {
		return nil, biz.ErrCartNotFound
	}
	for i, it := range c.Items {
		if it.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return c, nil
		}
	}
	return nil, biz.ErrItemNotFound
}

func (r *stubCartRepo) Clear(_ context.Context, sessionID string) error {
	if _, ok := r.carts[sessionID]; !ok {
		return biz.ErrCartNotFound
	}
	delete(r.carts, sessionID)
	return nil
}

func TestCartUsecase_GetOrCreate(t *testing.T) {
	uc := biz.NewCartUsecase(newStubCartRepo())

	got, err := uc.GetOrCreate(context.Background(), "sess-1")

	require.NoError(t, err)
	assert.Equal(t, "sess-1", got.SessionID)
}

func TestCartUsecase_AddItem_addsAndComputesTotal(t *testing.T) {
	uc := biz.NewCartUsecase(newStubCartRepo())
	item := &biz.CartItem{ProductID: uuid.New(), Name: "X", PriceCents: 1000, Quantity: 2}

	got, err := uc.AddItem(context.Background(), "sess-1", item)

	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, int64(2000), got.TotalCents)
}

func TestCartUsecase_UpdateItem_notFound(t *testing.T) {
	uc := biz.NewCartUsecase(newStubCartRepo())

	_, err := uc.UpdateItem(context.Background(), "missing", uuid.New(), 3)

	assert.ErrorIs(t, err, biz.ErrCartNotFound)
}

func TestCartUsecase_RemoveItem_itemNotFound(t *testing.T) {
	repo := newStubCartRepo()
	uc := biz.NewCartUsecase(repo)
	_, _ = uc.GetOrCreate(context.Background(), "sess-1")

	_, err := uc.RemoveItem(context.Background(), "sess-1", uuid.New())

	assert.ErrorIs(t, err, biz.ErrItemNotFound)
}

func TestCartUsecase_Clear(t *testing.T) {
	repo := newStubCartRepo()
	uc := biz.NewCartUsecase(repo)
	_, _ = uc.GetOrCreate(context.Background(), "sess-1")

	err := uc.Clear(context.Background(), "sess-1")

	require.NoError(t, err)
	assert.ErrorIs(t, uc.Clear(context.Background(), "sess-1"), biz.ErrCartNotFound)
}
