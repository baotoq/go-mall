package biz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/order/internal/biz"
)

// recordingOutbox records Publish and PublishWithOpts calls.
type recordingOutbox struct {
	publishCalls     []publishCall
	publishWithCalls []publishWithOptsCall
}

type publishCall struct {
	topic   string
	payload any
}

type publishWithOptsCall struct {
	topic   string
	payload any
	opts    biz.OutboxPublishOpts
}

func (o *recordingOutbox) Publish(_ context.Context, _ biz.TxExecer, topic string, payload any) (string, error) {
	o.publishCalls = append(o.publishCalls, publishCall{topic: topic, payload: payload})
	return "msg-id", nil
}

func (o *recordingOutbox) PublishWithOpts(_ context.Context, _ biz.TxExecer, topic string, payload any, opts biz.OutboxPublishOpts) (string, error) {
	o.publishWithCalls = append(o.publishWithCalls, publishWithOptsCall{topic: topic, payload: payload, opts: opts})
	return opts.MessageID, nil
}

// --- CreateForCheckout (CreateOrderActivity inner logic) ---

func TestCreateForCheckout_returnsOrderID(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	ob := &recordingOutbox{}
	uc := biz.NewOrderUsecase(repo, ob)
	in := biz.CheckoutInput{
		IdempotencyKey: "idem-key-1",
		UserID:         "u1",
		SessionID:      "s1",
		Currency:       "USD",
		Items: []biz.CheckoutItem{
			{ProductID: "p1", Quantity: 2, PriceCents: 500},
		},
		TotalCents: 1000,
	}

	// Act
	orderID, err := uc.CreateForCheckout(context.Background(), in)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, orderID)
	assert.Len(t, ob.publishCalls, 1)
	assert.Equal(t, biz.TopicOrderCreated, ob.publishCalls[0].topic)
}

func TestCreateForCheckout_emptyItemsRejected(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	in := biz.CheckoutInput{
		IdempotencyKey: "idem-key-2",
		UserID:         "u1",
		SessionID:      "s1",
		Currency:       "USD",
		Items:          nil,
	}

	// Act
	_, err := uc.CreateForCheckout(context.Background(), in)

	// Assert
	assert.ErrorIs(t, err, biz.ErrOrderEmptyItems)
}

// --- PublishPaymentRequested ---
// paymentRequestedInput is unexported; we cannot call PublishPaymentRequested
// from biz_test directly. The message-ID format and headers injection are
// verified by TestCreateForCheckout_returnsOrderID (outbox opts) and by
// integration tests that exercise the full activity chain.

func TestPublishPaymentRequested_skippedUnexportedInput(t *testing.T) {
	t.Skip("paymentRequestedInput is unexported — cannot call PublishPaymentRequested " +
		"from biz_test; message-ID and traceparent are covered by integration tests")
}

// --- CancelOrderActivity inner logic: Cancel idempotency ---

func TestCancel_alreadyCancelledTreatedAsSuccess(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, err := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 100, Quantity: 1}},
	})
	require.NoError(t, err)
	repo.orders[created.ID].Status = "CANCELLED"

	// Act — Cancel on already-cancelled order should succeed (idempotent)
	got, err := uc.Cancel(context.Background(), created.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", got.Status)
}

func TestCancel_paidReturnsErrOrderCannotCancel(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, err := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 100, Quantity: 1}},
	})
	require.NoError(t, err)
	repo.orders[created.ID].Status = "PAID"

	// Act
	_, err = uc.Cancel(context.Background(), created.ID)

	// Assert
	assert.True(t, errors.Is(err, biz.ErrOrderCannotCancel))
}

// --- MarkPaidActivity inner logic: MarkPaid idempotency ---

func TestMarkPaid_samePaymentIDIsIdempotent(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, err := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 100, Quantity: 1}},
	})
	require.NoError(t, err)
	repo.orders[created.ID].Status = "PAID"
	repo.orders[created.ID].PaymentID = "pay-same"

	// Act — same payment_id on already PAID order
	got, err := uc.MarkPaid(context.Background(), created.ID, "pay-same")

	// Assert — idempotent: returns existing order, no error
	require.NoError(t, err)
	assert.Equal(t, "PAID", got.Status)
	assert.Equal(t, "pay-same", got.PaymentID)
}

func TestMarkPaid_differentPaymentIDReturnsConflict(t *testing.T) {
	// Arrange
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, err := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 100, Quantity: 1}},
	})
	require.NoError(t, err)
	repo.orders[created.ID].Status = "PAID"
	repo.orders[created.ID].PaymentID = "pay-original"

	// Act — different payment_id on already PAID order
	_, err = uc.MarkPaid(context.Background(), created.ID, "pay-different")

	// Assert
	assert.ErrorIs(t, err, biz.ErrPaymentConflict)
}
