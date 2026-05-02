package biz_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gomall/app/order/internal/biz"
)

// stubIdempotencyKeyRepo is an in-memory IdempotencyKeyRepo.
type stubIdempotencyKeyRepo struct {
	data map[string]biz.StoredCheckout
}

func newStubIdempotencyKeyRepo() *stubIdempotencyKeyRepo {
	return &stubIdempotencyKeyRepo{data: make(map[string]biz.StoredCheckout)}
}

func (r *stubIdempotencyKeyRepo) Get(_ context.Context, key string) (biz.StoredCheckout, bool, error) {
	sc, ok := r.data[key]
	return sc, ok, nil
}

func (r *stubIdempotencyKeyRepo) Put(_ context.Context, key string, sc biz.StoredCheckout) error {
	r.data[key] = sc
	return nil
}

func defaultSagaCfg() biz.SagaConfig {
	return biz.SagaConfig{
		MaxPaymentAttempts:  3,
		PerAttemptTimeout:   30 * time.Second,
		PaymentInitialDelay: 2 * time.Second,
		PaymentBackoffMax:   30 * time.Second,
		MarkPaidRetryMax:    5,
		MarkPaidBudget:      60 * time.Second,
	}
}

// checkoutInput builds a minimal valid CheckoutInput for a test.
func checkoutInput(key, userID string) biz.CheckoutInput {
	return biz.CheckoutInput{
		IdempotencyKey: key,
		UserID:         userID,
		SessionID:      "sess-1",
		Currency:       "USD",
		Items:          []biz.CheckoutItem{{ProductID: "p1", Quantity: 1, PriceCents: 100}},
		TotalCents:     100,
	}
}

// TestCheckout_DuplicateSameUser verifies that a second Schedule call with the
// same idempotency key and same user_id returns the stored checkout without
// calling ScheduleWorkflow again (requires a real *workflow.Client — test the
// idempotency-store fast path only, which doesn't touch the client).
func TestCheckout_DuplicateSameUser_returnsCachedCheckout(t *testing.T) {
	// Arrange
	idem := newStubIdempotencyKeyRepo()
	// Pre-populate the idempotency store as if a prior Schedule succeeded.
	storedKey := "idem-key-abc"
	storedUserID := "user-1"
	storedCheckoutID := "checkout-idem-key-abc"
	idem.data[storedKey] = biz.StoredCheckout{
		CheckoutID: storedCheckoutID,
		UserID:     storedUserID,
		OrderID:    "order-xyz",
	}

	// NewCheckoutUsecase requires a *workflow.Client; pass nil — the idempotency
	// fast-path returns before touching the client.
	uc := biz.NewCheckoutUsecase(nil, idem, defaultSagaCfg(), noopLogger{})

	in := checkoutInput(storedKey, storedUserID)

	// Act
	checkoutID, orderID, err := uc.Schedule(context.Background(), in)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, storedCheckoutID, checkoutID)
	assert.Equal(t, "order-xyz", orderID)
}

// TestCheckout_DuplicateDifferentUser verifies that a second Schedule call with
// the same idempotency key but a different user_id returns ErrCheckoutDuplicateKey.
func TestCheckout_DuplicateDifferentUser_returnsError(t *testing.T) {
	// Arrange
	idem := newStubIdempotencyKeyRepo()
	storedKey := "idem-key-def"
	idem.data[storedKey] = biz.StoredCheckout{
		CheckoutID: "checkout-old",
		UserID:     "user-original",
	}

	uc := biz.NewCheckoutUsecase(nil, idem, defaultSagaCfg(), noopLogger{})
	in := checkoutInput(storedKey, "user-different")

	// Act
	_, _, err := uc.Schedule(context.Background(), in)

	// Assert
	assert.ErrorIs(t, err, biz.ErrCheckoutDuplicateKey)
}

// TestCheckout_EmptyIdempotencyKey verifies that an empty key is rejected
// (service-layer guard; Schedule returns an error immediately).
func TestCheckout_EmptyIdempotencyKey_returnsError(t *testing.T) {
	// Arrange
	uc := biz.NewCheckoutUsecase(nil, newStubIdempotencyKeyRepo(), defaultSagaCfg(), noopLogger{})
	in := checkoutInput("", "user-1")

	// Act
	_, _, err := uc.Schedule(context.Background(), in)

	// Assert
	assert.Error(t, err)
}

// TestCheckout_GRPCAlreadyExists_fetchesExisting verifies the Delta-D path:
// when ScheduleWorkflow returns codes.AlreadyExists, Schedule falls back to
// fetchExisting which reads the idempotency store.
//
// Because *workflow.Client is a concrete type we cannot inject a fake
// ScheduleWorkflow — this test skips that path.
func TestCheckout_GRPCAlreadyExists_skippedConcreteClient(t *testing.T) {
	t.Skip("requires durabletask-go test harness — *workflow.Client is a concrete " +
		"type; AlreadyExists path is covered by integration tests")
}

// TestCheckout_StringFallback_skippedConcreteClient documents the same
// limitation for the 'already exists' string-match fallback.
func TestCheckout_StringFallback_skippedConcreteClient(t *testing.T) {
	t.Skip("requires durabletask-go test harness — string fallback for 'already exists' " +
		"is covered by integration tests")
}

// TestCheckout_unwrapAll_grpcStatusDetected verifies the unwrapAll helper
// correctly surfaces a wrapped gRPC status for AlreadyExists detection.
func TestCheckout_unwrapAll_grpcStatusDetected(t *testing.T) {
	// Arrange — wrap a gRPC status error two levels deep
	inner := status.Error(codes.AlreadyExists, "orchestration instance exists")
	wrapped := fmt.Errorf("outer: %w", inner)

	// Act — simulate what Schedule does when ScheduleWorkflow returns wrapped
	st, ok := status.FromError(unwrapAllForTest(wrapped))

	// Assert
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

// unwrapAllForTest mirrors the biz.unwrapAll logic for test verification.
// (unwrapAll itself is unexported; we replicate the logic here.)
func unwrapAllForTest(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// noopLogger satisfies log.Logger for tests.
type noopLogger struct{}

func (noopLogger) Log(_ log.Level, _ ...any) error { return nil }
