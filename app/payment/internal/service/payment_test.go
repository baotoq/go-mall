package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	kratstransport "github.com/go-kratos/kratos/v2/transport"

	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/biz"
	"gomall/app/payment/internal/service"
)

// fakeTransportHeader implements transport.Header for tests.
type fakeTransportHeader map[string]string

func (h fakeTransportHeader) Get(key string) string      { return h[key] }
func (h fakeTransportHeader) Set(key, value string)      { h[key] = value }
func (h fakeTransportHeader) Add(key, value string)      { h[key] = value }
func (h fakeTransportHeader) Keys() []string             { keys := make([]string, 0, len(h)); for k := range h { keys = append(keys, k) }; return keys }
func (h fakeTransportHeader) Values(key string) []string { return []string{h[key]} }

// fakeTransport implements transport.Transporter for tests.
type fakeTransport struct{ header fakeTransportHeader }

func (t *fakeTransport) Kind() kratstransport.Kind            { return kratstransport.KindHTTP }
func (t *fakeTransport) Endpoint() string                     { return "" }
func (t *fakeTransport) Operation() string                    { return "" }
func (t *fakeTransport) RequestHeader() kratstransport.Header { return t.header }
func (t *fakeTransport) ReplyHeader() kratstransport.Header   { return fakeTransportHeader{} }

func ctxWithToken(token string) context.Context {
	h := fakeTransportHeader{"X-Internal-Token": token}
	return kratstransport.NewServerContext(context.Background(), &fakeTransport{header: h})
}

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
func (r *nopPaymentRepo) GetByWorkflowAndAttempt(_ context.Context, _ string, _ int32) (*biz.Payment, error) {
	return nil, biz.ErrPaymentNotFound
}
func (r *nopPaymentRepo) UpdateStatusInTx(_ context.Context, id uuid.UUID, status string, emit func(context.Context, biz.TxExecer, *biz.Payment) error) (*biz.Payment, error) {
	p := &biz.Payment{ID: id, Status: "PENDING"}
	if err := emit(context.Background(), nil, p); err != nil {
		return nil, err
	}
	return &biz.Payment{ID: id, Status: status}, nil
}

// nopOutbox discards all outbox publishes. Shared by all tests in this package.
type nopOutbox struct{}

func (nopOutbox) Publish(_ context.Context, _ biz.TxExecer, _ string, _ any) (string, error) {
	return "", nil
}

func newPaymentSvc(repo *nopPaymentRepo) *service.PaymentService {
	return service.NewPaymentService(biz.NewPaymentUsecase(repo, nopOutbox{}))
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

func TestCompletePayment_noToken_returnsUnauthenticated(t *testing.T) {
	// TDD: CRITICAL-1 — token enforcement, no transport context → 401
	svc := service.NewPaymentServiceWithToken(biz.NewPaymentUsecase(&nopPaymentRepo{}, nopOutbox{}), "secret")
	_, err := svc.CompletePayment(context.Background(), &v1.CompletePaymentRequest{Id: uuid.NewString()})
	require.Error(t, err)
	assert.Equal(t, 401, kerrors.Code(err))
}

func TestCompletePayment_wrongToken_returnsUnauthenticated(t *testing.T) {
	svc := service.NewPaymentServiceWithToken(biz.NewPaymentUsecase(&nopPaymentRepo{}, nopOutbox{}), "secret")
	_, err := svc.CompletePayment(ctxWithToken("wrong"), &v1.CompletePaymentRequest{Id: uuid.NewString()})
	require.Error(t, err)
	assert.Equal(t, 401, kerrors.Code(err))
}

func TestCompletePayment_correctToken_succeeds(t *testing.T) {
	svc := service.NewPaymentServiceWithToken(biz.NewPaymentUsecase(&nopPaymentRepo{getStatus: "PENDING"}, nopOutbox{}), "secret")
	_, err := svc.CompletePayment(ctxWithToken("secret"), &v1.CompletePaymentRequest{Id: uuid.NewString()})
	require.NoError(t, err)
}

func TestCompletePayment_tokenDisabled_succeeds(t *testing.T) {
	// empty token = disabled → existing code path unaffected
	svc := newPaymentSvc(&nopPaymentRepo{getStatus: "PENDING"})
	_, err := svc.CompletePayment(context.Background(), &v1.CompletePaymentRequest{Id: uuid.NewString()})
	require.NoError(t, err)
}

func TestFailPayment_noToken_returnsUnauthenticated(t *testing.T) {
	svc := service.NewPaymentServiceWithToken(biz.NewPaymentUsecase(&nopPaymentRepo{}, nopOutbox{}), "secret")
	_, err := svc.FailPayment(context.Background(), &v1.FailPaymentRequest{Id: uuid.NewString()})
	require.Error(t, err)
	assert.Equal(t, 401, kerrors.Code(err))
}

func TestFailPayment_correctToken_succeeds(t *testing.T) {
	svc := service.NewPaymentServiceWithToken(biz.NewPaymentUsecase(&nopPaymentRepo{getStatus: "PENDING"}, nopOutbox{}), "secret")
	_, err := svc.FailPayment(ctxWithToken("secret"), &v1.FailPaymentRequest{Id: uuid.NewString()})
	require.NoError(t, err)
}
