package server

// Tests for the payment subscriber payment.requested handler.
//
// The handler is tested directly (package server) using a stub PaymentRepo.
// Inbox dedup (outbox.Client) requires a real database, so the handler is
// exercised at the biz level: the subscriber's idempotency guard
// (GetByWorkflowAndAttempt) is the seam we test.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/payment/internal/biz"
)

// --- stub repo ---

type stubPaymentRepo struct {
	byWorkflow *biz.Payment // returned by GetByWorkflowAndAttempt (nil = not found)
	byWorkflowErr error     // error to return from GetByWorkflowAndAttempt
	created    []*biz.Payment
}

func (r *stubPaymentRepo) Create(_ context.Context, p *biz.Payment) (*biz.Payment, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	cp := *p
	r.created = append(r.created, &cp)
	return &cp, nil
}

func (r *stubPaymentRepo) GetByID(_ context.Context, id uuid.UUID) (*biz.Payment, error) {
	return &biz.Payment{ID: id, Status: "PENDING"}, nil
}

func (r *stubPaymentRepo) ListByUser(_ context.Context, _ string, _, _ int) ([]*biz.Payment, int, error) {
	return nil, 0, nil
}

func (r *stubPaymentRepo) ListByOrder(_ context.Context, _ string) ([]*biz.Payment, error) {
	return nil, nil
}

func (r *stubPaymentRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) (*biz.Payment, error) {
	return &biz.Payment{ID: id, Status: status}, nil
}

func (r *stubPaymentRepo) GetByWorkflowAndAttempt(_ context.Context, _ string, _ int32) (*biz.Payment, error) {
	if r.byWorkflowErr != nil {
		return nil, r.byWorkflowErr
	}
	if r.byWorkflow == nil {
		return nil, biz.ErrPaymentNotFound
	}
	return r.byWorkflow, nil
}

func (r *stubPaymentRepo) UpdateStatusInTx(_ context.Context, id uuid.UUID, status string, emit func(context.Context, biz.TxExecer, *biz.Payment) error) (*biz.Payment, error) {
	p := &biz.Payment{ID: id, Status: "PENDING"}
	if err := emit(context.Background(), nil, p); err != nil {
		return nil, err
	}
	return &biz.Payment{ID: id, Status: status}, nil
}

// nopTxExecer satisfies biz.TxExecer with a no-op.
type nopTxExecer struct{}

func (nopTxExecer) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}

// nopSubOutbox satisfies biz.OutboxPublisher with a no-op.
type nopSubOutbox struct{}

func (nopSubOutbox) Publish(_ context.Context, _ biz.TxExecer, _ string, _ any) (string, error) {
	return "", nil
}

// newTestPaymentSubscriber builds a PaymentSubscriber wired to a stub repo.
func newTestPaymentSubscriber(repo biz.PaymentRepo) *PaymentSubscriber {
	uc := biz.NewPaymentUsecase(repo, nopSubOutbox{})
	return &PaymentSubscriber{
		uc:  uc,
		log: log.NewHelper(log.DefaultLogger),
	}
}

// --- Test: missing workflow_instance_id is silently ACK'd ---

func TestPaymentSubscriber_HandlePaymentRequested_MissingInstanceID(t *testing.T) {
	// Arrange
	repo := &stubPaymentRepo{}
	s := newTestPaymentSubscriber(repo)

	// Act
	err := s.handlePaymentRequested(context.Background(), paymentRequestedEvent{
		WorkflowInstanceID: "",
		OrderID:            "ord-1",
		Amount:             100,
		Currency:           "USD",
		Attempt:            1,
	})

	// Assert
	require.NoError(t, err, "missing instanceID must ACK (return nil)")
	assert.Empty(t, repo.created, "no payment should be created when instanceID missing")
}

// --- Test: PENDING payment created on first delivery ---

func TestPaymentSubscriber_HandlePaymentRequested_CreatesPayment(t *testing.T) {
	// Arrange: GetByWorkflowAndAttempt returns not-found (first delivery)
	repo := &stubPaymentRepo{byWorkflow: nil}
	s := newTestPaymentSubscriber(repo)

	// Act
	err := s.handlePaymentRequested(context.Background(), paymentRequestedEvent{
		WorkflowInstanceID: "wf-1",
		OrderID:            "ord-1",
		Amount:             500,
		Currency:           "USD",
		Attempt:            1,
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, repo.created, 1, "exactly one payment must be created")
	p := repo.created[0]
	assert.Equal(t, "PENDING", p.Status)
	require.NotNil(t, p.WorkflowInstanceID)
	assert.Equal(t, "wf-1", *p.WorkflowInstanceID)
	assert.Equal(t, int32(1), p.Attempt)
	assert.Equal(t, "ord-1", p.OrderID)
	assert.Equal(t, int64(500), p.AmountCents)
}

// --- Test: idempotent on duplicate delivery (workflow+attempt already exists) ---

func TestPaymentSubscriber_HandlePaymentRequested_IdempotentOnDuplicate(t *testing.T) {
	// Arrange: GetByWorkflowAndAttempt returns an existing payment (duplicate)
	existing := &biz.Payment{
		ID:     uuid.New(),
		Status: "PENDING",
	}
	repo := &stubPaymentRepo{byWorkflow: existing}
	s := newTestPaymentSubscriber(repo)

	// Act
	err := s.handlePaymentRequested(context.Background(), paymentRequestedEvent{
		WorkflowInstanceID: "wf-1",
		OrderID:            "ord-1",
		Amount:             500,
		Currency:           "USD",
		Attempt:            1,
	})

	// Assert: handler ACKs without creating a second payment
	require.NoError(t, err)
	assert.Empty(t, repo.created, "duplicate delivery must not create a second payment")
}

// --- Test: transient DB error on idempotency check causes NACK ---

func TestPaymentSubscriber_HandlePaymentRequested_DBErrorNACK(t *testing.T) {
	// Arrange: GetByWorkflowAndAttempt returns a transient error (not ErrPaymentNotFound)
	repo := &stubPaymentRepo{byWorkflowErr: context.DeadlineExceeded}
	s := newTestPaymentSubscriber(repo)

	// Act
	err := s.handlePaymentRequested(context.Background(), paymentRequestedEvent{
		WorkflowInstanceID: "wf-1",
		OrderID:            "ord-1",
		Amount:             100,
		Currency:           "USD",
		Attempt:            1,
	})

	// Assert: NACK (non-nil error) so Dapr redelivers
	assert.Error(t, err, "transient DB error must NACK")
	assert.Empty(t, repo.created, "no payment must be created on DB error")
}

// --- compile-time interface check ---
var _ biz.PaymentRepo = (*stubPaymentRepo)(nil)
