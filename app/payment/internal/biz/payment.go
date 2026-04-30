package biz

import (
	"context"
	"time"

	paymentv1 "gomall/api/payment/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound     = errors.NotFound(paymentv1.ErrorReason_PAYMENT_NOT_FOUND.String(), "payment not found")
	ErrPaymentCannotRefund = errors.BadRequest(paymentv1.ErrorReason_PAYMENT_CANNOT_REFUND.String(), "payment cannot be refunded")
)

type Payment struct {
	ID          uuid.UUID
	OrderID     string
	UserID      string
	AmountCents int64
	Currency    string
	Status      string
	Provider    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Payment, int, error)
	ListByOrder(ctx context.Context, orderID string) ([]*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Payment, error)
}

type PaymentUsecase struct {
	repo PaymentRepo
}

func NewPaymentUsecase(repo PaymentRepo) *PaymentUsecase {
	return &PaymentUsecase{repo: repo}
}

func (uc *PaymentUsecase) Create(ctx context.Context, p *Payment) (*Payment, error) {
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
