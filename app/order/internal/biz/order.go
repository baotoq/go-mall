package biz

import (
	"context"
	"time"

	orderv1 "gomall/api/order/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

var (
	ErrOrderNotFound     = errors.NotFound(orderv1.ErrorReason_ORDER_NOT_FOUND.String(), "order not found")
	ErrOrderCannotCancel = errors.BadRequest(orderv1.ErrorReason_ORDER_CANNOT_CANCEL.String(), "order cannot be cancelled")
	ErrOrderAlreadyPaid  = errors.BadRequest(orderv1.ErrorReason_ORDER_ALREADY_PAID.String(), "order already paid")
	ErrOrderEmptyItems   = errors.BadRequest(orderv1.ErrorReason_ORDER_EMPTY_ITEMS.String(), "order must have at least one item")
)

type OrderItem struct {
	ProductID     string
	Name          string
	PriceCents    int64
	Currency      string
	ImageURL      string
	Quantity      int32
	SubtotalCents int64
}

type Order struct {
	ID         uuid.UUID
	UserID     string
	SessionID  string
	Items      []OrderItem
	TotalCents int64
	Currency   string
	Status     string
	PaymentID  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderRepo interface {
	Create(ctx context.Context, o *Order) (*Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByUser(ctx context.Context, userID, status string, page, pageSize int) ([]*Order, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Order, error)
	SetPaymentID(ctx context.Context, id uuid.UUID, paymentID string) (*Order, error)
}

type OrderUsecase struct{ repo OrderRepo }

func NewOrderUsecase(repo OrderRepo) *OrderUsecase { return &OrderUsecase{repo: repo} }

func (uc *OrderUsecase) Create(ctx context.Context, o *Order) (*Order, error) {
	if len(o.Items) == 0 {
		return nil, ErrOrderEmptyItems
	}
	if o.Currency == "" {
		o.Currency = "USD"
	}
	var total int64
	for i := range o.Items {
		sub := o.Items[i].PriceCents * int64(o.Items[i].Quantity)
		o.Items[i].SubtotalCents = sub
		if o.Items[i].Currency == "" {
			o.Items[i].Currency = o.Currency
		}
		total += sub
	}
	o.TotalCents = total
	o.Status = "PENDING"
	return uc.repo.Create(ctx, o)
}

func (uc *OrderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, userID, status string, page, pageSize int) ([]*Order, int, error) {
	return uc.repo.ListByUser(ctx, userID, status, page, pageSize)
}

func isValidStatus(s string) bool {
	switch s {
	case "PENDING", "PAID", "SHIPPED", "DELIVERED", "CANCELLED":
		return true
	}
	return false
}

func (uc *OrderUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Order, error) {
	if !isValidStatus(status) {
		return nil, errors.BadRequest(orderv1.ErrorReason_INVALID_ARGUMENT.String(), "invalid status")
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}

func (uc *OrderUsecase) Cancel(ctx context.Context, id uuid.UUID) (*Order, error) {
	cur, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch cur.Status {
	case "PAID", "SHIPPED", "DELIVERED":
		return nil, ErrOrderCannotCancel
	case "CANCELLED":
		return cur, nil
	}
	return uc.repo.UpdateStatus(ctx, id, "CANCELLED")
}

func (uc *OrderUsecase) MarkPaid(ctx context.Context, id uuid.UUID, paymentID string) (*Order, error) {
	cur, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.Status == "PAID" {
		return nil, ErrOrderAlreadyPaid
	}
	if cur.Status == "CANCELLED" {
		return nil, ErrOrderCannotCancel
	}
	if _, err := uc.repo.SetPaymentID(ctx, id, paymentID); err != nil {
		return nil, err
	}
	return uc.repo.UpdateStatus(ctx, id, "PAID")
}
