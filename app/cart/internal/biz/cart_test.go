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
	for _, existing := range c.Items {
		if existing.ProductID == item.ProductID {
			existing.Quantity += item.Quantity
			existing.SubtotalCents = existing.PriceCents * int64(existing.Quantity)
			c.TotalCents = 0
			for _, it := range c.Items {
				c.TotalCents += it.SubtotalCents
			}
			return c, nil
		}
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
	// Arrange
	uc := biz.NewCartUsecase(newStubCartRepo())
	item := &biz.CartItem{ProductID: uuid.New(), Name: "X", PriceCents: 1000, Quantity: 2}

	// Act
	got, err := uc.AddItem(context.Background(), "sess-1", item)

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, int64(2000), got.TotalCents)
}

func TestCartUsecase_AddItem_createsCartForNewSession(t *testing.T) {
	// Arrange
	uc := biz.NewCartUsecase(newStubCartRepo())
	item := &biz.CartItem{ProductID: uuid.New(), Name: "Widget", PriceCents: 500, Quantity: 1}

	// Act
	got, err := uc.AddItem(context.Background(), "brand-new-session", item)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "brand-new-session", got.SessionID)
	require.Len(t, got.Items, 1)
	assert.Equal(t, int64(500), got.TotalCents)
}

func TestCartUsecase_AddItem_differentProductsAccumulate(t *testing.T) {
	// Arrange
	uc := biz.NewCartUsecase(newStubCartRepo())
	prodA, prodB := uuid.New(), uuid.New()

	// Act
	_, _ = uc.AddItem(context.Background(), "sess-1", &biz.CartItem{ProductID: prodA, Name: "A", PriceCents: 1000, Quantity: 1})
	got, err := uc.AddItem(context.Background(), "sess-1", &biz.CartItem{ProductID: prodB, Name: "B", PriceCents: 2000, Quantity: 2})

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Items, 2)
	assert.Equal(t, int64(5000), got.TotalCents) // 1×1000 + 2×2000
}

func TestCartUsecase_AddItem_sameProductIncrementsQuantity(t *testing.T) {
	// Arrange
	uc := biz.NewCartUsecase(newStubCartRepo())
	prod := uuid.New()

	// Act
	_, _ = uc.AddItem(context.Background(), "sess-1", &biz.CartItem{ProductID: prod, Name: "X", PriceCents: 800, Quantity: 2})
	got, err := uc.AddItem(context.Background(), "sess-1", &biz.CartItem{ProductID: prod, Name: "X", PriceCents: 800, Quantity: 3})

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, 5, got.Items[0].Quantity)
	assert.Equal(t, int64(4000), got.TotalCents) // 800×5
}

func TestCartUsecase_AddItem_subtotalIsPriceTimesQuantity(t *testing.T) {
	// Arrange
	uc := biz.NewCartUsecase(newStubCartRepo())

	// Act
	got, err := uc.AddItem(context.Background(), "sess-1", &biz.CartItem{
		ProductID: uuid.New(), Name: "Y", PriceCents: 1500, Quantity: 3,
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, int64(4500), got.Items[0].SubtotalCents)
	assert.Equal(t, int64(4500), got.TotalCents)
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
