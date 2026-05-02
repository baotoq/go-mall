package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/propagation"
)

// TopicPaymentRequested is the pub/sub topic for payment requests.
const TopicPaymentRequested = "payment.requested"

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

// paymentRequestedPayload is the JSON body published on payment.requested.
type paymentRequestedPayload struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	OrderID            string `json:"order_id"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	Attempt            int32  `json:"attempt"`
}

// PublishPaymentRequested publishes a payment.requested event via the
// transactional outbox. Opens its own ent.Tx (via repo.RunInTx) so the
// outbox row commits atomically and is isolated from any caller transaction.
// messageID = workflowID+":pay-req:"+attempt ensures idempotent replay
// (UNIQUE outbox id, ON CONFLICT DO NOTHING).
// W3C traceparent is injected into headers from ctx so the payment service
// can continue the trace.
func (uc *OrderUsecase) PublishPaymentRequested(ctx context.Context, in paymentRequestedInput, messageID string) error {
	headers := make(map[string]string)
	propagation.TraceContext{}.Inject(ctx, propagation.MapCarrier(headers))

	evt := paymentRequestedPayload{
		WorkflowInstanceID: in.WorkflowID,
		OrderID:            in.OrderID,
		Amount:             in.Amount,
		Currency:           in.Currency,
		Attempt:            in.Attempt,
	}

	return uc.repo.RunInTx(ctx, func(tx TxExecer) error {
		_, err := uc.ob.PublishWithOpts(ctx, tx, TopicPaymentRequested, evt, OutboxPublishOpts{
			MessageID: messageID,
			Headers:   headers,
		})
		return err
	})
}
