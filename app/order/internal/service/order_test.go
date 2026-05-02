package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "gomall/api/order/v1"
	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/service"
)

type stubOrderRepo struct {
	orders map[uuid.UUID]*biz.Order
}

func newStubRepo() *stubOrderRepo {
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

func newSvc(repo *stubOrderRepo) *service.OrderService {
	return service.NewOrderService(biz.NewOrderUsecase(repo, &stubOutbox{}), nil, nil)
}

func TestOrderService_CreateOrder_validatesEmptyItems(t *testing.T) {
	svc := newSvc(newStubRepo())
	_, err := svc.CreateOrder(context.Background(), &v1.CreateOrderRequest{
		UserId: "u1", SessionId: "s1",
	})
	assert.Error(t, err)
}

func TestOrderService_CreateOrder_persistsAndReturnsPending(t *testing.T) {
	svc := newSvc(newStubRepo())
	got, err := svc.CreateOrder(context.Background(), &v1.CreateOrderRequest{
		UserId:    "u1",
		SessionId: "s1",
		Currency:  "USD",
		Items: []*v1.CreateOrderItem{
			{ProductId: "p1", Name: "Widget", PriceCents: 100, Quantity: 2},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, v1.OrderStatus_PENDING, got.Status)
	assert.Equal(t, int64(200), got.TotalCents)
}

func TestOrderService_GetOrder_invalidUUIDRejected(t *testing.T) {
	svc := newSvc(newStubRepo())
	_, err := svc.GetOrder(context.Background(), &v1.GetOrderRequest{Id: "not-a-uuid"})
	assert.Error(t, err)
}

func TestOrderService_GetOrder_notFoundPropagates(t *testing.T) {
	svc := newSvc(newStubRepo())
	_, err := svc.GetOrder(context.Background(), &v1.GetOrderRequest{Id: uuid.NewString()})
	assert.Error(t, err)
}

func TestOrderService_UpdateOrderStatus_paidStatus(t *testing.T) {
	repo := newStubRepo()
	svc := newSvc(repo)
	created, err := svc.CreateOrder(context.Background(), &v1.CreateOrderRequest{
		UserId: "u1", SessionId: "s1",
		Items: []*v1.CreateOrderItem{{ProductId: "p1", PriceCents: 50, Quantity: 1}},
	})
	require.NoError(t, err)

	got, err := svc.UpdateOrderStatus(context.Background(), &v1.UpdateOrderStatusRequest{
		Id:     created.Id,
		Status: v1.OrderStatus_PAID,
	})
	require.NoError(t, err)
	assert.Equal(t, v1.OrderStatus_PAID, got.Status)
}

func TestOrderService_CancelOrder_pendingOrder(t *testing.T) {
	repo := newStubRepo()
	svc := newSvc(repo)
	created, err := svc.CreateOrder(context.Background(), &v1.CreateOrderRequest{
		UserId: "u1", SessionId: "s1",
		Items: []*v1.CreateOrderItem{{ProductId: "p1", PriceCents: 50, Quantity: 1}},
	})
	require.NoError(t, err)

	got, err := svc.CancelOrder(context.Background(), &v1.CancelOrderRequest{Id: created.Id})
	require.NoError(t, err)
	assert.Equal(t, v1.OrderStatus_CANCELLED, got.Status)
}

func TestOrderService_ListOrders_returnsList(t *testing.T) {
	repo := newStubRepo()
	svc := newSvc(repo)
	_, err := svc.CreateOrder(context.Background(), &v1.CreateOrderRequest{
		UserId: "u1", SessionId: "s1",
		Items: []*v1.CreateOrderItem{{ProductId: "p1", PriceCents: 50, Quantity: 1}},
	})
	require.NoError(t, err)

	got, err := svc.ListOrders(context.Background(), &v1.ListOrdersRequest{UserId: "u1", Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.Total)
	assert.Len(t, got.Orders, 1)
}
