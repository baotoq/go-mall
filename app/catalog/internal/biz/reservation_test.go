package biz_test

import (
	"context"
	"testing"
	"time"

	"gomall/app/catalog/internal/biz"

	"github.com/google/uuid"
)

// fakeReservationRepo is a simple in-memory fake for biz-layer tests.
type fakeReservationRepo struct {
	reservations map[string]*biz.Reservation
	stock        map[uuid.UUID]int
}

func newFakeRepo(stock map[uuid.UUID]int) *fakeReservationRepo {
	return &fakeReservationRepo{
		reservations: make(map[string]*biz.Reservation),
		stock:        stock,
	}
}

func key(cartID, productID uuid.UUID) string {
	return cartID.String() + ":" + productID.String()
}

func (f *fakeReservationRepo) reserved(productID uuid.UUID) int {
	total := 0
	for _, r := range f.reservations {
		if r.ProductID == productID && r.Status == biz.ReservationStatusActive {
			total += r.Quantity
		}
	}
	return total
}

func (f *fakeReservationRepo) Reserve(ctx context.Context, cartID, productID uuid.UUID, quantity int, ttl time.Duration) (*biz.Reservation, error) {
	stock := f.stock[productID]
	reserved := f.reserved(productID)

	k := key(cartID, productID)
	if existing, ok := f.reservations[k]; ok && existing.Status == biz.ReservationStatusActive {
		delta := quantity - existing.Quantity
		if delta > 0 && stock-reserved < delta {
			return nil, biz.ErrOutOfStock
		}
		existing.Quantity = quantity
		existing.ExpiresAt = time.Now().Add(ttl)
		return existing, nil
	}

	if stock-reserved < quantity {
		return nil, biz.ErrOutOfStock
	}

	r := &biz.Reservation{
		ID:        uuid.New(),
		CartID:    cartID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    biz.ReservationStatusActive,
		ExpiresAt: time.Now().Add(ttl),
	}
	f.reservations[k] = r
	return r, nil
}

func (f *fakeReservationRepo) Release(ctx context.Context, cartID, productID uuid.UUID) error {
	k := key(cartID, productID)
	r, ok := f.reservations[k]
	if !ok || r.Status != biz.ReservationStatusActive {
		return biz.ErrReservationNotFound
	}
	r.Status = biz.ReservationStatusReleased
	return nil
}

func (f *fakeReservationRepo) Adjust(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*biz.Reservation, error) {
	k := key(cartID, productID)
	r, ok := f.reservations[k]
	if !ok || r.Status != biz.ReservationStatusActive {
		return nil, biz.ErrReservationNotFound
	}
	delta := quantity - r.Quantity
	if delta > 0 {
		stock := f.stock[productID]
		reserved := f.reserved(productID)
		if stock-reserved < delta {
			return nil, biz.ErrOutOfStock
		}
	}
	r.Quantity = quantity
	return r, nil
}

func (f *fakeReservationRepo) Commit(ctx context.Context, cartID, productID uuid.UUID) error {
	k := key(cartID, productID)
	r, ok := f.reservations[k]
	if !ok || r.Status != biz.ReservationStatusActive {
		return biz.ErrReservationNotFound
	}
	r.Status = biz.ReservationStatusCommitted
	return nil
}

func (f *fakeReservationRepo) ReleaseAll(ctx context.Context, cartID uuid.UUID) error {
	for _, r := range f.reservations {
		if r.CartID == cartID && r.Status == biz.ReservationStatusActive {
			r.Status = biz.ReservationStatusReleased
		}
	}
	return nil
}

func (f *fakeReservationRepo) ExpireStale(ctx context.Context) (int, error) {
	n := 0
	for _, r := range f.reservations {
		if r.Status == biz.ReservationStatusActive && time.Now().After(r.ExpiresAt) {
			r.Status = biz.ReservationStatusExpired
			n++
		}
	}
	return n, nil
}

func TestUsecase_Reserve(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID: 10})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	cartID := uuid.New()
	r, err := uc.Reserve(ctx, cartID, productID, 3)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if r.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", r.Quantity)
	}
}

func TestUsecase_Reserve_OutOfStock(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID: 2})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	_, err := uc.Reserve(ctx, uuid.New(), productID, 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUsecase_Release(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID: 10})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	cartID := uuid.New()
	if _, err := uc.Reserve(ctx, cartID, productID, 3); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := uc.Release(ctx, cartID, productID); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

func TestUsecase_Adjust(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID: 10})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	cartID := uuid.New()
	if _, err := uc.Reserve(ctx, cartID, productID, 3); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	r, err := uc.Adjust(ctx, cartID, productID, 7)
	if err != nil {
		t.Fatalf("Adjust: %v", err)
	}
	if r.Quantity != 7 {
		t.Errorf("quantity = %d, want 7", r.Quantity)
	}
}

func TestUsecase_Commit(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID: 10})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	cartID := uuid.New()
	if _, err := uc.Reserve(ctx, cartID, productID, 3); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := uc.Commit(ctx, cartID, productID); err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestUsecase_ReleaseAll(t *testing.T) {
	productID1, productID2 := uuid.New(), uuid.New()
	repo := newFakeRepo(map[uuid.UUID]int{productID1: 10, productID2: 10})
	uc := biz.NewReservationUsecase(repo)
	ctx := context.Background()

	cartID := uuid.New()
	if _, err := uc.Reserve(ctx, cartID, productID1, 2); err != nil {
		t.Fatalf("Reserve 1: %v", err)
	}
	if _, err := uc.Reserve(ctx, cartID, productID2, 3); err != nil {
		t.Fatalf("Reserve 2: %v", err)
	}
	if err := uc.ReleaseAll(ctx, cartID); err != nil {
		t.Fatalf("ReleaseAll: %v", err)
	}
}
