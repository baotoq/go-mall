package biz

import (
	"context"
	"time"

	cartv1 "gomall/api/cart/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

var (
	ErrCartNotFound = errors.NotFound(cartv1.ErrorReason_CART_NOT_FOUND.String(), "cart not found")
	ErrItemNotFound = errors.NotFound(cartv1.ErrorReason_ITEM_NOT_FOUND.String(), "item not found")
)

type CartItem struct {
	ID            uuid.UUID
	CartID        uuid.UUID
	ProductID     uuid.UUID
	Name          string
	PriceCents    int64
	Currency      string
	ImageURL      string
	Quantity      int
	SubtotalCents int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Cart struct {
	ID         uuid.UUID
	SessionID  string
	Items      []*CartItem
	TotalCents int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CartRepo interface {
	FindOrCreateBySession(ctx context.Context, sessionID string) (*Cart, error)
	AddItem(ctx context.Context, sessionID string, item *CartItem) (*Cart, error)
	UpdateItem(ctx context.Context, sessionID string, productID uuid.UUID, quantity int) (*Cart, error)
	RemoveItem(ctx context.Context, sessionID string, productID uuid.UUID) (*Cart, error)
	Clear(ctx context.Context, sessionID string) error
}

type CartUsecase struct {
	repo CartRepo
}

func NewCartUsecase(repo CartRepo) *CartUsecase {
	return &CartUsecase{repo: repo}
}

func (uc *CartUsecase) GetOrCreate(ctx context.Context, sessionID string) (*Cart, error) {
	return uc.repo.FindOrCreateBySession(ctx, sessionID)
}

func (uc *CartUsecase) AddItem(ctx context.Context, sessionID string, item *CartItem) (*Cart, error) {
	return uc.repo.AddItem(ctx, sessionID, item)
}

func (uc *CartUsecase) UpdateItem(ctx context.Context, sessionID string, productID uuid.UUID, quantity int) (*Cart, error) {
	return uc.repo.UpdateItem(ctx, sessionID, productID, quantity)
}

func (uc *CartUsecase) RemoveItem(ctx context.Context, sessionID string, productID uuid.UUID) (*Cart, error) {
	return uc.repo.RemoveItem(ctx, sessionID, productID)
}

func (uc *CartUsecase) Clear(ctx context.Context, sessionID string) error {
	return uc.repo.Clear(ctx, sessionID)
}
