package biz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/payment/internal/biz"
)

type stubPaymentRepo struct {
	payments map[uuid.UUID]*biz.Payment
}

func newStubPaymentRepo() *stubPaymentRepo {
	return &stubPaymentRepo{payments: make(map[uuid.UUID]*biz.Payment)}
}

func (r *stubPaymentRepo) Create(_ context.Context, p *biz.Payment) (*biz.Payment, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	r.payments[p.ID] = p
	return p, nil
}

func (r *stubPaymentRepo) GetByID(_ context.Context, id uuid.UUID) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	return p, nil
}

func (r *stubPaymentRepo) ListByUser(_ context.Context, userID string, _, _ int) ([]*biz.Payment, int, error) {
	var out []*biz.Payment
	for _, p := range r.payments {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, len(out), nil
}

func (r *stubPaymentRepo) ListByOrder(_ context.Context, orderID string) ([]*biz.Payment, error) {
	var out []*biz.Payment
	for _, p := range r.payments {
		if p.OrderID == orderID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *stubPaymentRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	return p, nil
}

func (r *stubPaymentRepo) GetByWorkflowAndAttempt(_ context.Context, _ string, _ int32) (*biz.Payment, error) {
	return nil, biz.ErrPaymentNotFound
}

func (r *stubPaymentRepo) UpdateStatusInTx(_ context.Context, id uuid.UUID, status string, emit func(context.Context, biz.TxExecer, *biz.Payment) error) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	if err := emit(context.Background(), nil, p); err != nil {
		return nil, err
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	return p, nil
}

type stubNopOutbox struct{}

func (stubNopOutbox) Publish(_ context.Context, _ biz.TxExecer, _ string, _ any) (string, error) {
	return "", nil
}

func newUC(repo *stubPaymentRepo) *biz.PaymentUsecase {
	return biz.NewPaymentUsecase(repo, stubNopOutbox{})
}

func TestPaymentUsecase_Create_setsPending(t *testing.T) {
	uc := newUC(newStubPaymentRepo())

	got, err := uc.Create(context.Background(), &biz.Payment{OrderID: "o1", UserID: "u1", AmountCents: 100})

	require.NoError(t, err)
	assert.Equal(t, "PENDING", got.Status)
}

func TestPaymentUsecase_Refund_rejectsNonCompleted(t *testing.T) {
	repo := newStubPaymentRepo()
	uc := newUC(repo)
	created, _ := uc.Create(context.Background(), &biz.Payment{OrderID: "o", UserID: "u", AmountCents: 1})

	_, err := uc.Refund(context.Background(), created.ID)

	assert.ErrorIs(t, err, biz.ErrPaymentCannotRefund)
}

func TestPaymentUsecase_Refund_completedTransitions(t *testing.T) {
	repo := newStubPaymentRepo()
	uc := newUC(repo)
	created, _ := uc.Create(context.Background(), &biz.Payment{OrderID: "o", UserID: "u", AmountCents: 1})
	created.Status = "COMPLETED"

	got, err := uc.Refund(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, "REFUNDED", got.Status)
}

func TestPaymentUsecase_Refund_notFound(t *testing.T) {
	uc := newUC(newStubPaymentRepo())

	_, err := uc.Refund(context.Background(), uuid.New())

	assert.ErrorIs(t, err, biz.ErrPaymentNotFound)
}

func TestPaymentUsecase_ListPayments_orderTakesPriority(t *testing.T) {
	repo := newStubPaymentRepo()
	uc := newUC(repo)
	_, _ = uc.Create(context.Background(), &biz.Payment{OrderID: "ord-x", UserID: "u1"})
	_, _ = uc.Create(context.Background(), &biz.Payment{OrderID: "ord-x", UserID: "u2"})
	_, _ = uc.Create(context.Background(), &biz.Payment{OrderID: "ord-y", UserID: "u1"})

	got, total, err := uc.ListPayments(context.Background(), "u1", "ord-x", 1, 50)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, got, 2)
}

func TestPaymentUsecase_ListPayments_byUser(t *testing.T) {
	repo := newStubPaymentRepo()
	uc := newUC(repo)
	_, _ = uc.Create(context.Background(), &biz.Payment{OrderID: "o1", UserID: "u1"})
	_, _ = uc.Create(context.Background(), &biz.Payment{OrderID: "o2", UserID: "u2"})

	got, total, err := uc.ListPayments(context.Background(), "u1", "", 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
}
