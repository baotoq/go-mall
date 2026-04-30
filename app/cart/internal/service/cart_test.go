package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "gomall/api/cart/v1"
	"gomall/app/cart/internal/biz"
	"gomall/app/cart/internal/service"
)

type nopCartRepo struct{}

func (r *nopCartRepo) FindOrCreateBySession(_ context.Context, sessionID string) (*biz.Cart, error) {
	return &biz.Cart{ID: uuid.New(), SessionID: sessionID}, nil
}
func (r *nopCartRepo) AddItem(_ context.Context, _ string, item *biz.CartItem) (*biz.Cart, error) {
	item.ID = uuid.New()
	return &biz.Cart{ID: uuid.New(), Items: []*biz.CartItem{item}}, nil
}
func (r *nopCartRepo) UpdateItem(_ context.Context, _ string, _ uuid.UUID, _ int) (*biz.Cart, error) {
	return &biz.Cart{ID: uuid.New()}, nil
}
func (r *nopCartRepo) RemoveItem(_ context.Context, _ string, _ uuid.UUID) (*biz.Cart, error) {
	return &biz.Cart{ID: uuid.New()}, nil
}
func (r *nopCartRepo) Clear(_ context.Context, _ string) error { return nil }

func newCartSvc() *service.CartService {
	return service.NewCartService(biz.NewCartUsecase(&nopCartRepo{}))
}

func TestCartService_GetCart_emptySession(t *testing.T) {
	_, err := newCartSvc().GetCart(context.Background(), &v1.GetCartRequest{SessionId: ""})
	assert.Error(t, err)
}

func TestCartService_AddItem_validation(t *testing.T) {
	svc := newCartSvc()
	cases := []struct {
		name string
		req  *v1.AddItemRequest
	}{
		{"empty session", &v1.AddItemRequest{ProductId: uuid.NewString(), Quantity: 1}},
		{"empty product", &v1.AddItemRequest{SessionId: "s", Quantity: 1}},
		{"invalid product uuid", &v1.AddItemRequest{SessionId: "s", ProductId: "bad", Quantity: 1}},
		{"non-positive quantity", &v1.AddItemRequest{SessionId: "s", ProductId: uuid.NewString(), Quantity: 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.AddItem(context.Background(), tc.req)
			assert.Error(t, err)
		})
	}
}

func TestCartService_AddItem_ok(t *testing.T) {
	got, err := newCartSvc().AddItem(context.Background(), &v1.AddItemRequest{
		SessionId: "s", ProductId: uuid.NewString(), Quantity: 2, PriceCents: 500, Name: "X",
	})
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
}

func TestCartService_UpdateItem_invalidProduct(t *testing.T) {
	_, err := newCartSvc().UpdateItem(context.Background(), &v1.UpdateItemRequest{SessionId: "s", ProductId: "bad"})
	assert.Error(t, err)
}

func TestCartService_RemoveItem_emptySession(t *testing.T) {
	_, err := newCartSvc().RemoveItem(context.Background(), &v1.RemoveItemRequest{ProductId: uuid.NewString()})
	assert.Error(t, err)
}

func TestCartService_ClearCart(t *testing.T) {
	_, err := newCartSvc().ClearCart(context.Background(), &v1.ClearCartRequest{SessionId: "s"})
	require.NoError(t, err)
	_, err = newCartSvc().ClearCart(context.Background(), &v1.ClearCartRequest{SessionId: ""})
	assert.Error(t, err)
}
