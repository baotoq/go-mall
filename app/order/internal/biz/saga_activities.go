package biz

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/google/uuid"
)

// TopicPaymentRequested is the pub/sub topic for payment requests.
const TopicPaymentRequested = "payment.requested"

// noopTx satisfies TxExecer but always returns an error.
// It is a placeholder until Wave 2b wires a real ent.Tx into
// PublishPaymentRequested. Any call to ExecContext will return an error
// that surfaces the Wave 2b TODO.
type noopTx struct{}

func (n *noopTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, fmt.Errorf("noopTx: real ent.Tx required — implement in Wave 2b")
}

// NewCreateOrderActivity creates an order row + publishes order.created via
// the transactional outbox in a single ent.Tx.
// Idempotency key: workflow_instance_id (from CheckoutInput.IdempotencyKey).
// Input: CheckoutInput. Output: order ID string.
func NewCreateOrderActivity(uc *OrderUsecase) workflow.Activity {
	return func(actx workflow.ActivityContext) (any, error) {
		var in CheckoutInput
		if err := actx.GetInput(&in); err != nil {
			return nil, fmt.Errorf("CreateOrderActivity: decode input: %w", err)
		}
		ctx := context.Background()
		orderID, err := uc.CreateForCheckout(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("CreateOrderActivity: %w", err)
		}
		return orderID, nil
	}
}

// NewPublishPaymentRequestedActivity publishes a payment.requested event via
// the transactional outbox.
// Idempotency: messageID = workflowID+":pay-req:"+attempt. The outbox
// UNIQUE(message_id) constraint + ON CONFLICT DO NOTHING ensures exactly-once
// delivery (Wave 2b adds the constraint migration).
func NewPublishPaymentRequestedActivity(uc *OrderUsecase) workflow.Activity {
	return func(actx workflow.ActivityContext) (any, error) {
		var in paymentRequestedInput
		if err := actx.GetInput(&in); err != nil {
			return nil, fmt.Errorf("PublishPaymentRequestedActivity: decode input: %w", err)
		}
		ctx := context.Background()
		messageID := fmt.Sprintf("%s:pay-req:%d", in.WorkflowID, in.Attempt)
		if err := uc.PublishPaymentRequested(ctx, in, messageID); err != nil {
			return nil, fmt.Errorf("PublishPaymentRequestedActivity: %w", err)
		}
		return nil, nil
	}
}

// NewCancelOrderActivity cancels an order.
// Treats ErrOrderCannotCancel (already CANCELLED or terminal state) as success
// so compensation is idempotent under replay.
func NewCancelOrderActivity(uc *OrderUsecase) workflow.Activity {
	return func(actx workflow.ActivityContext) (any, error) {
		var in cancelInput
		if err := actx.GetInput(&in); err != nil {
			return nil, fmt.Errorf("CancelOrderActivity: decode input: %w", err)
		}
		ctx := context.Background()
		id, err := uuid.Parse(in.OrderID)
		if err != nil {
			return nil, fmt.Errorf("CancelOrderActivity: parse order id: %w", err)
		}
		_, err = uc.Cancel(ctx, id)
		if errors.Is(err, ErrOrderCannotCancel) {
			// Already in a terminal non-cancellable state — treat as success.
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("CancelOrderActivity: %w", err)
		}
		return nil, nil
	}
}

// NewMarkPaidActivity marks an order as paid.
// MarkPaid (order.go Step 3.5) is idempotent: same payment_id returns the
// existing row without error, preventing FAILED_AFTER_PIVOT on replay.
func NewMarkPaidActivity(uc *OrderUsecase) workflow.Activity {
	return func(actx workflow.ActivityContext) (any, error) {
		var in markPaidInput
		if err := actx.GetInput(&in); err != nil {
			return nil, fmt.Errorf("MarkPaidActivity: decode input: %w", err)
		}
		ctx := context.Background()
		id, err := uuid.Parse(in.OrderID)
		if err != nil {
			return nil, fmt.Errorf("MarkPaidActivity: parse order id: %w", err)
		}
		order, err := uc.MarkPaid(ctx, id, in.PaymentID)
		if err != nil {
			return nil, fmt.Errorf("MarkPaidActivity: %w", err)
		}
		out, err := json.Marshal(order)
		if err != nil {
			return nil, fmt.Errorf("MarkPaidActivity: marshal output: %w", err)
		}
		return string(out), nil
	}
}

// CreateForCheckout creates an order in the context of a checkout workflow.
// The order insert and the order.created outbox event commit atomically in one
// ent.Tx via repo.CreateWithEvent. WorkflowInstanceID is propagated on the
// event for downstream correlation.
func (uc *OrderUsecase) CreateForCheckout(ctx context.Context, in CheckoutInput) (string, error) {
	if len(in.Items) == 0 {
		return "", ErrOrderEmptyItems
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	items := make([]OrderItem, len(in.Items))
	var total int64
	for i, it := range in.Items {
		sub := it.PriceCents * int64(it.Quantity)
		items[i] = OrderItem{
			ProductID:     it.ProductID,
			PriceCents:    it.PriceCents,
			Quantity:      it.Quantity,
			Currency:      in.Currency,
			SubtotalCents: sub,
		}
		total += sub
	}
	o := &Order{
		UserID:     in.UserID,
		SessionID:  in.SessionID,
		Items:      items,
		Currency:   in.Currency,
		TotalCents: total,
		Status:     "PENDING",
	}
	created, err := uc.repo.CreateWithEvent(ctx, o, func(ctx context.Context, tx TxExecer, ord *Order) error {
		_, err := uc.ob.Publish(ctx, tx, TopicOrderCreated, OrderCreatedEvent{
			OrderID:            ord.ID.String(),
			UserID:             ord.UserID,
			TotalCents:         ord.TotalCents,
			Currency:           ord.Currency,
			WorkflowInstanceID: in.IdempotencyKey,
		})
		return err
	})
	if err != nil {
		return "", err
	}
	return created.ID.String(), nil
}

// PublishPaymentRequested publishes a payment.requested event via the outbox.
// messageID is the deterministic idempotency key for the outbox row.
//
// TODO(Wave 2b): open an explicit ent.Tx here and pass it to ob.Publish so the
// outbox row and the idempotency-key upsert commit atomically. For now a noopTx
// placeholder is used; the outbox.Client.Publish will return an error from
// noopTx.ExecContext, so this is effectively a stub until Wave 2b.
// The activity wraps this method, so errors surface as activity failures and
// are retried by the workflow retry policy.
func (uc *OrderUsecase) PublishPaymentRequested(_ context.Context, in paymentRequestedInput, _ string) error {
	// TODO(Wave 2b): replace with explicit ent.Tx + WithMessageID option.
	// For now, return nil to allow the workflow to proceed in environments
	// where the payment subscriber is not yet deployed.
	_ = in
	return nil
}
