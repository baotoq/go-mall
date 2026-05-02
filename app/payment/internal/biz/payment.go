package biz

import (
	"context"
	"database/sql"
	"time"

	paymentv1 "gomall/api/payment/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

// TxExecer is the minimal SQL interface satisfied by *sql.Tx.
type TxExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// OutboxPublisher writes an event into the transactional outbox.
// TODO: This interface is duplicated across services; extract to pkg/outbox once
// a second consumer exists.
type OutboxPublisher interface {
	Publish(ctx context.Context, tx TxExecer, topic string, payload any) (string, error)
}

var (
	ErrPaymentNotFound        = errors.NotFound(paymentv1.ErrorReason_PAYMENT_NOT_FOUND.String(), "payment not found")
	ErrPaymentCannotRefund    = errors.BadRequest(paymentv1.ErrorReason_PAYMENT_CANNOT_REFUND.String(), "payment cannot be refunded")
	ErrPaymentInvalidTransition = errors.BadRequest(paymentv1.ErrorReason_PAYMENT_INVALID_TRANSITION.String(), "payment status transition not allowed")
)

// CreatePaymentInput carries all fields needed to create a payment record.
type CreatePaymentInput struct {
	OrderID            string
	UserID             string
	AmountCents        int64
	Currency           string
	Provider           string
	WorkflowInstanceID string
	Attempt            int32
}

type Payment struct {
	ID                 uuid.UUID
	OrderID            string
	UserID             string
	AmountCents        int64
	Currency           string
	Status             string
	Provider           string
	WorkflowInstanceID *string
	Attempt            int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// completedPayload is the outbox payload for payment.completed events.
type completedPayload struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	Attempt            int32  `json:"attempt"`
	PaymentID          string `json:"payment_id"`
	OrderID            string `json:"order_id"`
}

// failedPayload is the outbox payload for payment.failed events.
type failedPayload struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	Attempt            int32  `json:"attempt"`
	OrderID            string `json:"order_id"`
	ReasonCode         string `json:"reason_code"`
}

type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Payment, int, error)
	ListByOrder(ctx context.Context, orderID string) ([]*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Payment, error)
	GetByWorkflowAndAttempt(ctx context.Context, workflowInstanceID string, attempt int32) (*Payment, error)
	UpdateStatusInTx(ctx context.Context, id uuid.UUID, status string, emit func(ctx context.Context, tx TxExecer, p *Payment) error) (*Payment, error)
}

type PaymentUsecase struct {
	repo    PaymentRepo
	outbox  OutboxPublisher
}

func NewPaymentUsecase(repo PaymentRepo, outbox OutboxPublisher) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, outbox: outbox}
}

func (uc *PaymentUsecase) Create(ctx context.Context, p *Payment) (*Payment, error) {
	p.Status = "PENDING"
	return uc.repo.Create(ctx, p)
}

// CreateFromWorkflow creates a PENDING payment keyed by (workflow_instance_id, attempt).
func (uc *PaymentUsecase) CreateFromWorkflow(ctx context.Context, in CreatePaymentInput) (*Payment, error) {
	p := &Payment{
		OrderID:     in.OrderID,
		UserID:      in.UserID,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		Provider:    in.Provider,
		Attempt:     in.Attempt,
	}
	if in.WorkflowInstanceID != "" {
		p.WorkflowInstanceID = &in.WorkflowInstanceID
	}
	p.Status = "PENDING"
	return uc.repo.Create(ctx, p)
}

func (uc *PaymentUsecase) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PaymentUsecase) ListPayments(ctx context.Context, userID, orderID string, page, pageSize int) ([]*Payment, int, error) {
	if orderID != "" {
		payments, err := uc.repo.ListByOrder(ctx, orderID)
		return payments, len(payments), err
	}
	return uc.repo.ListByUser(ctx, userID, page, pageSize)
}

func (uc *PaymentUsecase) GetByWorkflowAndAttempt(ctx context.Context, workflowInstanceID string, attempt int32) (*Payment, error) {
	return uc.repo.GetByWorkflowAndAttempt(ctx, workflowInstanceID, attempt)
}

func (uc *PaymentUsecase) Refund(ctx context.Context, id uuid.UUID) (*Payment, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != "COMPLETED" {
		return nil, ErrPaymentCannotRefund
	}
	return uc.repo.UpdateStatus(ctx, id, "REFUNDED")
}

// CompletePayment transitions PENDING → COMPLETED and publishes payment.completed.
func (uc *PaymentUsecase) CompletePayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	return uc.repo.UpdateStatusInTx(ctx, id, "COMPLETED", func(ctx context.Context, tx TxExecer, p *Payment) error {
		if p.Status != "PENDING" {
			return ErrPaymentInvalidTransition
		}
		wid := ""
		if p.WorkflowInstanceID != nil {
			wid = *p.WorkflowInstanceID
		}
		payload := completedPayload{
			WorkflowInstanceID: wid,
			Attempt:            p.Attempt,
			PaymentID:          p.ID.String(),
			OrderID:            p.OrderID,
		}
		_, err := uc.outbox.Publish(ctx, tx, "payment.completed", payload)
		return err
	})
}

// FailPayment transitions PENDING → FAILED and publishes payment.failed.
func (uc *PaymentUsecase) FailPayment(ctx context.Context, id uuid.UUID, reasonCode string) (*Payment, error) {
	return uc.repo.UpdateStatusInTx(ctx, id, "FAILED", func(ctx context.Context, tx TxExecer, p *Payment) error {
		if p.Status != "PENDING" {
			return ErrPaymentInvalidTransition
		}
		wid := ""
		if p.WorkflowInstanceID != nil {
			wid = *p.WorkflowInstanceID
		}
		payload := failedPayload{
			WorkflowInstanceID: wid,
			Attempt:            p.Attempt,
			OrderID:            p.OrderID,
			ReasonCode:         reasonCode,
		}
		_, err := uc.outbox.Publish(ctx, tx, "payment.failed", payload)
		return err
	})
}
