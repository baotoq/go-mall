package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/biz"
	"gomall/app/payment/internal/service"
)

type nopPaymentRepo struct {
	getStatus string
	getErr    error
}

func (r *nopPaymentRepo) Create(_ context.Context, p *biz.Payment) (*biz.Payment, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return p, nil
}
func (r *nopPaymentRepo) GetByID(_ context.Context, id uuid.UUID) (*biz.Payment, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	status := r.getStatus
	if status == "" {
		status = "PENDING"
	}
	return &biz.Payment{ID: id, Status: status}, nil
}
func (r *nopPaymentRepo) ListByUser(_ context.Context, _ string, _, _ int) ([]*biz.Payment, int, error) {
	return []*biz.Payment{{ID: uuid.New(), Status: "PENDING"}}, 1, nil
}
func (r *nopPaymentRepo) ListByOrder(_ context.Context, _ string) ([]*biz.Payment, error) {
	return []*biz.Payment{{ID: uuid.New(), Status: "COMPLETED"}}, nil
}
func (r *nopPaymentRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) (*biz.Payment, error) {
	return &biz.Payment{ID: id, Status: status}, nil
}

func newPaymentSvc(repo *nopPaymentRepo) *service.PaymentService {
	return service.NewPaymentService(biz.NewPaymentUsecase(repo))
}

func TestPaymentService_CreatePayment_validation(t *testing.T) {
	svc := newPaymentSvc(&nopPaymentRepo{})
	cases := []struct {
		name string
		req  *v1.CreatePaymentRequest
	}{
		{"missing order", &v1.CreatePaymentRequest{UserId: "u", AmountCents: 1, Provider: "p"}},
		{"missing user", &v1.CreatePaymentRequest{OrderId: "o", AmountCents: 1, Provider: "p"}},
		{"non-positive amount", &v1.CreatePaymentRequest{OrderId: "o", UserId: "u", AmountCents: 0, Provider: "p"}},
		{"missing provider", &v1.CreatePaymentRequest{OrderId: "o", UserId: "u", AmountCents: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreatePayment(context.Background(), tc.req)
			assert.Error(t, err)
		})
	}
}

func TestPaymentService_CreatePayment_ok(t *testing.T) {
	got, err := newPaymentSvc(&nopPaymentRepo{}).CreatePayment(context.Background(), &v1.CreatePaymentRequest{
		OrderId: "o", UserId: "u", AmountCents: 100, Currency: "USD", Provider: "stripe",
	})
	require.NoError(t, err)
	assert.Equal(t, v1.PaymentStatus_PENDING, got.Status)
}

func TestPaymentService_GetPayment_invalidID(t *testing.T) {
	svc := newPaymentSvc(&nopPaymentRepo{})
	_, err := svc.GetPayment(context.Background(), &v1.GetPaymentRequest{Id: ""})
	assert.Error(t, err)
	_, err = svc.GetPayment(context.Background(), &v1.GetPaymentRequest{Id: "not-a-uuid"})
	assert.Error(t, err)
}

func TestPaymentService_GetPayment_unknownStatusMapsToUnspecified(t *testing.T) {
	got, err := newPaymentSvc(&nopPaymentRepo{getStatus: "BOGUS"}).GetPayment(
		context.Background(), &v1.GetPaymentRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	assert.Equal(t, v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED, got.Status)
}

func TestPaymentService_RefundPayment_invalidID(t *testing.T) {
	_, err := newPaymentSvc(&nopPaymentRepo{}).RefundPayment(context.Background(), &v1.RefundPaymentRequest{Id: "bad"})
	assert.Error(t, err)
}

func TestPaymentService_RefundPayment_ok(t *testing.T) {
	got, err := newPaymentSvc(&nopPaymentRepo{getStatus: "COMPLETED"}).RefundPayment(
		context.Background(), &v1.RefundPaymentRequest{Id: uuid.NewString()})
	require.NoError(t, err)
	assert.Equal(t, v1.PaymentStatus_REFUNDED, got.Status)
}

func TestPaymentService_ListPayments_byOrder(t *testing.T) {
	got, err := newPaymentSvc(&nopPaymentRepo{}).ListPayments(context.Background(), &v1.ListPaymentsRequest{OrderId: "o"})
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.Total)
}

func TestPaymentService_ListPayments_byUser(t *testing.T) {
	got, err := newPaymentSvc(&nopPaymentRepo{}).ListPayments(context.Background(), &v1.ListPaymentsRequest{UserId: "u"})
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.Total)
}
