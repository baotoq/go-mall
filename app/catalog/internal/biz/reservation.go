package biz

import (
	"context"
	"time"

	v1 "gomall/api/catalog/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

const ReservationTTL = 15 * time.Minute

var (
	ErrOutOfStock          = errors.BadRequest(v1.ErrorReason_OUT_OF_STOCK.String(), "out of stock")
	ErrReservationNotFound = errors.NotFound(v1.ErrorReason_RESERVATION_NOT_FOUND.String(), "reservation not found")
)

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusReleased  ReservationStatus = "RELEASED"
	ReservationStatusExpired   ReservationStatus = "EXPIRED"
	ReservationStatusCommitted ReservationStatus = "COMMITTED"
)

type Reservation struct {
	ID        uuid.UUID
	CartID    uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Status    ReservationStatus
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReservationRepo interface {
	// Reserve creates or re-activates a reservation for (cart_id, product_id).
	// Uses SELECT FOR UPDATE on the product row to prevent overselling.
	Reserve(ctx context.Context, cartID, productID uuid.UUID, quantity int, ttl time.Duration) (*Reservation, error)
	// Release marks an ACTIVE reservation as RELEASED.
	Release(ctx context.Context, cartID, productID uuid.UUID) error
	// Adjust updates the quantity on an ACTIVE reservation (absolute, not delta).
	Adjust(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*Reservation, error)
	// Commit marks an ACTIVE reservation as COMMITTED.
	Commit(ctx context.Context, cartID, productID uuid.UUID) error
	// ReleaseAll releases all ACTIVE reservations for a cart.
	ReleaseAll(ctx context.Context, cartID uuid.UUID) error
	// ExpireStale marks ACTIVE reservations past their expires_at as EXPIRED and restores stock.
	ExpireStale(ctx context.Context) (int, error)
}

type ReservationUsecase struct {
	repo ReservationRepo
}

func NewReservationUsecase(repo ReservationRepo) *ReservationUsecase {
	return &ReservationUsecase{repo: repo}
}

func (uc *ReservationUsecase) Reserve(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*Reservation, error) {
	return uc.repo.Reserve(ctx, cartID, productID, quantity, ReservationTTL)
}

func (uc *ReservationUsecase) Release(ctx context.Context, cartID, productID uuid.UUID) error {
	return uc.repo.Release(ctx, cartID, productID)
}

func (uc *ReservationUsecase) Adjust(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*Reservation, error) {
	return uc.repo.Adjust(ctx, cartID, productID, quantity)
}

func (uc *ReservationUsecase) Commit(ctx context.Context, cartID, productID uuid.UUID) error {
	return uc.repo.Commit(ctx, cartID, productID)
}

func (uc *ReservationUsecase) ReleaseAll(ctx context.Context, cartID uuid.UUID) error {
	return uc.repo.ReleaseAll(ctx, cartID)
}
