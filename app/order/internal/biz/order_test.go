package biz_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/order/internal/biz"
)

type stubOrderRepo struct {
	orders map[uuid.UUID]*biz.Order
}

func newStubOrderRepo() *stubOrderRepo {
	return &stubOrderRepo{orders: make(map[uuid.UUID]*biz.Order)}
}

func (r *stubOrderRepo) Create(_ context.Context, o *biz.Order) (*biz.Order, error) {
	o.ID = uuid.New()
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	cp := *o
	r.orders[cp.ID] = &cp
	return &cp, nil
}

func (r *stubOrderRepo) CreateWithEvent(ctx context.Context, o *biz.Order, emit func(context.Context, biz.TxExecer, *biz.Order) error) (*biz.Order, error) {
	created, err := r.Create(ctx, o)
	if err != nil {
		return nil, err
	}
	if err := emit(ctx, nil, created); err != nil {
		return nil, err
	}
	return created, nil
}

type stubOutbox struct{}

func (s *stubOutbox) Publish(_ context.Context, _ biz.TxExecer, _ string, _ any) (string, error) {
	return "stub-id", nil
}

func (r *stubOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*biz.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, biz.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (r *stubOrderRepo) ListByUser(_ context.Context, userID, status string, _, _ int) ([]*biz.Order, int, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.UserID != userID {
			continue
		}
		if status != "" && o.Status != status {
			continue
		}
		cp := *o
		out = append(out, &cp)
	}
	return out, len(out), nil
}

func (r *stubOrderRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) (*biz.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, biz.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	cp := *o
	return &cp, nil
}

func (r *stubOrderRepo) MarkPaid(_ context.Context, id uuid.UUID, paymentID string) (*biz.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, biz.ErrOrderNotFound
	}
	o.PaymentID = paymentID
	o.Status = "PAID"
	o.UpdatedAt = time.Now()
	cp := *o
	return &cp, nil
}

func (r *stubOrderRepo) GetByWorkflowInstanceID(_ context.Context, workflowInstanceID string) (*biz.Order, bool, error) {
	for _, o := range r.orders {
		if o.WorkflowInstanceID == workflowInstanceID {
			cp := *o
			return &cp, true, nil
		}
	}
	return nil, false, nil
}

func (r *stubOrderRepo) RunInTx(_ context.Context, fn func(biz.TxExecer) error) error {
	return fn(nil)
}

func (s *stubOutbox) PublishWithOpts(_ context.Context, _ biz.TxExecer, _ string, _ any, _ biz.OutboxPublishOpts) (string, error) {
	return "stub-id", nil
}

func TestOrderUsecase_Create_setsPendingAndComputesTotals(t *testing.T) {
	uc := biz.NewOrderUsecase(newStubOrderRepo(), &stubOutbox{})

	got, err := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Currency:  "USD",
		Items: []biz.OrderItem{
			{ProductID: "p1", PriceCents: 100, Quantity: 2},
			{ProductID: "p2", PriceCents: 50, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "PENDING", got.Status)
	assert.Equal(t, int64(250), got.TotalCents)
	assert.Equal(t, int64(200), got.Items[0].SubtotalCents)
	assert.Equal(t, int64(50), got.Items[1].SubtotalCents)
}

func TestOrderUsecase_Create_emptyItemsRejected(t *testing.T) {
	uc := biz.NewOrderUsecase(newStubOrderRepo(), &stubOutbox{})

	_, err := uc.Create(context.Background(), &biz.Order{UserID: "u1", SessionID: "s1"})

	assert.ErrorIs(t, err, biz.ErrOrderEmptyItems)
}

func TestOrderUsecase_Cancel_pendingTransitions(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})

	got, err := uc.Cancel(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", got.Status)
}

func TestOrderUsecase_Cancel_paidRejected(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})
	repo.orders[created.ID].Status = "PAID"

	_, err := uc.Cancel(context.Background(), created.ID)

	assert.ErrorIs(t, err, biz.ErrOrderCannotCancel)
}

func TestOrderUsecase_MarkPaid_setsPaidAndPaymentID(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})

	got, err := uc.MarkPaid(context.Background(), created.ID, "pay-123")

	require.NoError(t, err)
	assert.Equal(t, "PAID", got.Status)
	assert.Equal(t, "pay-123", repo.orders[created.ID].PaymentID)
}

func TestOrderUsecase_MarkPaid_alreadyPaidRejected(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})
	repo.orders[created.ID].Status = "PAID"

	_, err := uc.MarkPaid(context.Background(), created.ID, "pay-456")

	assert.ErrorIs(t, err, biz.ErrOrderAlreadyPaid)
}

func TestOrderUsecase_UpdateStatus_invalidStatusRejected(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})
	created, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})

	_, err := uc.UpdateStatus(context.Background(), created.ID, "BOGUS")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestOrderUsecase_ListOrders_filterByStatus(t *testing.T) {
	repo := newStubOrderRepo()
	uc := biz.NewOrderUsecase(repo, &stubOutbox{})

	_, _ = uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s1",
		Items:     []biz.OrderItem{{ProductID: "p1", PriceCents: 10, Quantity: 1}},
	})
	created2, _ := uc.Create(context.Background(), &biz.Order{
		UserID:    "u1",
		SessionID: "s2",
		Items:     []biz.OrderItem{{ProductID: "p2", PriceCents: 20, Quantity: 1}},
	})
	repo.orders[created2.ID].Status = "PAID"

	got, total, err := uc.ListOrders(context.Background(), "u1", "PAID", 1, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)
	assert.Equal(t, "PAID", got[0].Status)
}
