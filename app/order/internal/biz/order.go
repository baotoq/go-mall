package biz

import (
	"context"
	"database/sql"
	"time"

	orderv1 "gomall/api/order/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

// TxExecer is the minimal SQL interface satisfied by *sql.Tx.
type TxExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// OutboxPublishOpts carries optional overrides for a Publish call.
// Kept in biz to avoid importing pkg/outbox from biz.
type OutboxPublishOpts struct {
	MessageID string
	Headers   map[string]string
}

// OutboxPublisher writes an event into the transactional outbox.
// TODO: This interface is duplicated across services; extract to pkg/outbox once
// a second consumer exists.
type OutboxPublisher interface {
	Publish(ctx context.Context, tx TxExecer, topic string, payload any) (string, error)
	// PublishWithOpts is like Publish but forwards message-ID and headers for
	// idempotent, trace-correlated delivery (used by saga activities).
	PublishWithOpts(ctx context.Context, tx TxExecer, topic string, payload any, opts OutboxPublishOpts) (string, error)
}

var (
	ErrOrderNotFound     = errors.NotFound(orderv1.ErrorReason_ORDER_NOT_FOUND.String(), "order not found")
	ErrOrderCannotCancel = errors.BadRequest(orderv1.ErrorReason_ORDER_CANNOT_CANCEL.String(), "order cannot be cancelled")
	ErrOrderAlreadyPaid  = errors.BadRequest(orderv1.ErrorReason_ORDER_ALREADY_PAID.String(), "order already paid")
	ErrOrderEmptyItems   = errors.BadRequest(orderv1.ErrorReason_ORDER_EMPTY_ITEMS.String(), "order must have at least one item")
	ErrPaymentConflict   = errors.BadRequest(orderv1.ErrorReason_ORDER_ALREADY_PAID.String(), "order already paid with different payment id")
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
	ID                 uuid.UUID
	UserID             string
	SessionID          string
	Items              []OrderItem
	TotalCents         int64
	Currency           string
	Status             string
	PaymentID          string
	WorkflowInstanceID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrderRepo interface {
	Create(ctx context.Context, o *Order) (*Order, error)
	CreateWithEvent(ctx context.Context, o *Order, emit func(context.Context, TxExecer, *Order) error) (*Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	// GetByWorkflowInstanceID returns the order tied to a Dapr workflow instance.
	// found=false (with nil error) means no such order exists yet.
	GetByWorkflowInstanceID(ctx context.Context, workflowInstanceID string) (*Order, bool, error)
	ListByUser(ctx context.Context, userID, status string, page, pageSize int) ([]*Order, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Order, error)
	MarkPaid(ctx context.Context, id uuid.UUID, paymentID string) (*Order, error)
	// RunInTx executes fn inside a new database transaction. Used by saga
	// activities that need to commit outbox rows atomically.
	RunInTx(ctx context.Context, fn func(tx TxExecer) error) error
}

type OrderUsecase struct {
	repo OrderRepo
	ob   OutboxPublisher
}

func NewOrderUsecase(repo OrderRepo, ob OutboxPublisher) *OrderUsecase {
	return &OrderUsecase{repo: repo, ob: ob}
}

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
	return uc.repo.CreateWithEvent(ctx, o, func(ctx context.Context, tx TxExecer, created *Order) error {
		_, err := uc.ob.Publish(ctx, tx, TopicOrderCreated, OrderCreatedEvent{
			OrderID:    created.ID.String(),
			UserID:     created.UserID,
			TotalCents: created.TotalCents,
			Currency:   created.Currency,
		})
		return err
	})
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
		// Idempotent replay: same payment_id is success; different payment_id is a conflict.
		if cur.PaymentID == paymentID {
			return cur, nil
		}
		return nil, ErrPaymentConflict
	}
	if cur.Status == "CANCELLED" {
		return nil, ErrOrderCannotCancel
	}
	return uc.repo.MarkPaid(ctx, id, paymentID)
}
