package data_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

func makeReservationRepo(t *testing.T) biz.ReservationRepo {
	t.Helper()
	return data.NewReservationRepo(data.NewTestData(testClient), log.DefaultLogger)
}

func makeProductRepo(t *testing.T) biz.ProductRepo {
	t.Helper()
	return data.NewProductRepo(data.NewTestData(testClient), log.DefaultLogger)
}

// seedProduct creates a product with the given stock and returns its ID.
func seedProduct(t *testing.T, repo biz.ProductRepo, stock int) uuid.UUID {
	t.Helper()
	p, err := repo.Save(context.Background(), &biz.Product{
		Name:       "Test Product " + uuid.New().String(),
		Slug:       "test-" + uuid.New().String(),
		PriceCents: 1000,
		Stock:      stock,
	})
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}
	return p.ID
}

func TestReserve_Success(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	r, err := rRepo.Reserve(ctx, cartID, productID, 3, 15*time.Minute)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if r.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", r.Quantity)
	}
	if r.Status != biz.ReservationStatusActive {
		t.Errorf("status = %s, want ACTIVE", r.Status)
	}
}

func TestReserve_OutOfStock(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 2)
	cartID := uuid.New()

	_, err := rRepo.Reserve(ctx, cartID, productID, 5, 15*time.Minute)
	if err == nil {
		t.Fatal("expected ErrOutOfStock, got nil")
	}
	if !isOutOfStock(err) {
		t.Errorf("expected OUT_OF_STOCK error, got %v", err)
	}
}

func TestReserve_Upsert(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	r1, err := rRepo.Reserve(ctx, cartID, productID, 3, 15*time.Minute)
	if err != nil {
		t.Fatalf("first Reserve: %v", err)
	}
	r2, err := rRepo.Reserve(ctx, cartID, productID, 5, 15*time.Minute)
	if err != nil {
		t.Fatalf("second Reserve (upsert): %v", err)
	}
	if r2.ID != r1.ID {
		t.Error("upsert should update existing reservation, not create new")
	}
	if r2.Quantity != 5 {
		t.Errorf("quantity after upsert = %d, want 5", r2.Quantity)
	}
}

func TestRelease_Success(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	if _, err := rRepo.Reserve(ctx, cartID, productID, 3, 15*time.Minute); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := rRepo.Release(ctx, cartID, productID); err != nil {
		t.Fatalf("Release: %v", err)
	}
	// After release, stock should be available again.
	cartID2 := uuid.New()
	if _, err := rRepo.Reserve(ctx, cartID2, productID, 10, 15*time.Minute); err != nil {
		t.Fatalf("Reserve after release: %v", err)
	}
}

func TestRelease_NotFound(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	ctx := context.Background()

	err := rRepo.Release(ctx, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected ErrReservationNotFound, got nil")
	}
}

func TestAdjust_Increase(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	if _, err := rRepo.Reserve(ctx, cartID, productID, 3, 15*time.Minute); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	r, err := rRepo.Adjust(ctx, cartID, productID, 7)
	if err != nil {
		t.Fatalf("Adjust: %v", err)
	}
	if r.Quantity != 7 {
		t.Errorf("quantity after Adjust = %d, want 7", r.Quantity)
	}
}

func TestAdjust_Decrease(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	if _, err := rRepo.Reserve(ctx, cartID, productID, 7, 15*time.Minute); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	r, err := rRepo.Adjust(ctx, cartID, productID, 3)
	if err != nil {
		t.Fatalf("Adjust decrease: %v", err)
	}
	if r.Quantity != 3 {
		t.Errorf("quantity after decrease = %d, want 3", r.Quantity)
	}
}

func TestCommit(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	if _, err := rRepo.Reserve(ctx, cartID, productID, 3, 15*time.Minute); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := rRepo.Commit(ctx, cartID, productID); err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestReleaseAll(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID1 := seedProduct(t, pRepo, 10)
	productID2 := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	if _, err := rRepo.Reserve(ctx, cartID, productID1, 2, 15*time.Minute); err != nil {
		t.Fatalf("Reserve 1: %v", err)
	}
	if _, err := rRepo.Reserve(ctx, cartID, productID2, 3, 15*time.Minute); err != nil {
		t.Fatalf("Reserve 2: %v", err)
	}
	if err := rRepo.ReleaseAll(ctx, cartID); err != nil {
		t.Fatalf("ReleaseAll: %v", err)
	}
	// Full stock should now be available on both products.
	cartID2 := uuid.New()
	if _, err := rRepo.Reserve(ctx, cartID2, productID1, 10, 15*time.Minute); err != nil {
		t.Fatalf("Reserve after ReleaseAll (product 1): %v", err)
	}
}

func TestExpireStale(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	productID := seedProduct(t, pRepo, 10)
	cartID := uuid.New()

	// Create a reservation with a very short TTL.
	if _, err := rRepo.Reserve(ctx, cartID, productID, 3, -1*time.Second); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	n, err := rRepo.ExpireStale(ctx)
	if err != nil {
		t.Fatalf("ExpireStale: %v", err)
	}
	if n == 0 {
		t.Error("expected at least 1 expired reservation")
	}
}

func TestReserve_ConcurrentRace(t *testing.T) {
	truncate(t)
	rRepo := makeReservationRepo(t)
	pRepo := makeProductRepo(t)
	ctx := context.Background()

	const stock = 5
	productID := seedProduct(t, pRepo, stock)

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		successes int
		failures  int
	)

	const goroutines = 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cartID := uuid.New()
			_, err := rRepo.Reserve(ctx, cartID, productID, 1, 15*time.Minute)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successes++
			} else {
				failures++
			}
		}()
	}
	wg.Wait()

	if successes > stock {
		t.Errorf("oversold: %d reservations for stock of %d", successes, stock)
	}
	if successes+failures != goroutines {
		t.Errorf("successes(%d)+failures(%d) != goroutines(%d)", successes, failures, goroutines)
	}
}

func isOutOfStock(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == biz.ErrOutOfStock.Error()
}
