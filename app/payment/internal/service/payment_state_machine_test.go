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

// stateMachineRepo is a stub repo that controls per-payment status for state machine tests.
type stateMachineRepo struct {
	payments map[uuid.UUID]*biz.Payment
}

func newStateMachineRepo() *stateMachineRepo {
	return &stateMachineRepo{payments: make(map[uuid.UUID]*biz.Payment)}
}

func (r *stateMachineRepo) Create(_ context.Context, p *biz.Payment) (*biz.Payment, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	cp := *p
	r.payments[p.ID] = &cp
	return &cp, nil
}

func (r *stateMachineRepo) GetByID(_ context.Context, id uuid.UUID) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	cp := *p
	return &cp, nil
}

func (r *stateMachineRepo) ListByUser(_ context.Context, _ string, _, _ int) ([]*biz.Payment, int, error) {
	return nil, 0, nil
}

func (r *stateMachineRepo) ListByOrder(_ context.Context, _ string) ([]*biz.Payment, error) {
	return nil, nil
}

func (r *stateMachineRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	cp := *p
	return &cp, nil
}

func (r *stateMachineRepo) GetByWorkflowAndAttempt(_ context.Context, _ string, _ int32) (*biz.Payment, error) {
	return nil, biz.ErrPaymentNotFound
}

// UpdateStatusInTx is the transactional version used by CompletePayment/FailPayment.
// In this stub we run the emit callback and update status in-memory (no real tx).
func (r *stateMachineRepo) UpdateStatusInTx(_ context.Context, id uuid.UUID, status string, emit func(context.Context, biz.TxExecer, *biz.Payment) error) (*biz.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, biz.ErrPaymentNotFound
	}
	cp := *p

	// Run the emit callback with a nil TxExecer — the no-op outbox won't call it.
	if err := emit(context.Background(), nil, &cp); err != nil {
		return nil, err
	}

	p.Status = status
	p.UpdatedAt = time.Now()
	result := *p
	return &result, nil
}

func newStateMachineSvc(repo biz.PaymentRepo) *service.PaymentService {
	return service.NewPaymentService(biz.NewPaymentUsecase(repo, nopOutbox{}))
}

// seedPayment creates a payment with the given status in the repo and returns its ID.
func seedPayment(t *testing.T, repo *stateMachineRepo, status string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	wid := "wf-1"
	repo.payments[id] = &biz.Payment{
		ID:                 id,
		OrderID:            "order-1",
		UserID:             "user-1",
		AmountCents:        100,
		Currency:           "USD",
		Status:             status,
		Provider:           "stripe",
		WorkflowInstanceID: &wid,
		Attempt:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	return id
}

func TestStateMachine_CompletePayment(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		wantStatus v1.PaymentStatus
		wantErr    bool
	}{
		{
			name:       "PENDING to COMPLETED ok",
			fromStatus: "PENDING",
			wantStatus: v1.PaymentStatus_COMPLETED,
			wantErr:    false,
		},
		{
			name:       "COMPLETED to COMPLETED rejected",
			fromStatus: "COMPLETED",
			wantErr:    true,
		},
		{
			name:       "FAILED to COMPLETED rejected",
			fromStatus: "FAILED",
			wantErr:    true,
		},
		{
			name:       "REFUNDED to COMPLETED rejected",
			fromStatus: "REFUNDED",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := newStateMachineRepo()
			svc := newStateMachineSvc(repo)
			id := seedPayment(t, repo, tc.fromStatus)

			// Act
			got, err := svc.CompletePayment(context.Background(), &v1.CompletePaymentRequest{Id: id.String()})

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantStatus, got.Status)
			}
		})
	}
}

func TestStateMachine_FailPayment(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		wantStatus v1.PaymentStatus
		wantErr    bool
	}{
		{
			name:       "PENDING to FAILED ok",
			fromStatus: "PENDING",
			wantStatus: v1.PaymentStatus_FAILED,
			wantErr:    false,
		},
		{
			name:       "COMPLETED to FAILED rejected",
			fromStatus: "COMPLETED",
			wantErr:    true,
		},
		{
			name:       "FAILED to FAILED rejected",
			fromStatus: "FAILED",
			wantErr:    true,
		},
		{
			name:       "REFUNDED to FAILED rejected",
			fromStatus: "REFUNDED",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := newStateMachineRepo()
			svc := newStateMachineSvc(repo)
			id := seedPayment(t, repo, tc.fromStatus)

			// Act
			got, err := svc.FailPayment(context.Background(), &v1.FailPaymentRequest{
				Id:         id.String(),
				ReasonCode: "test_reason",
			})

			// Assert
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantStatus, got.Status)
			}
		})
	}
}
