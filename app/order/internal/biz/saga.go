package biz

import (
	"errors"
	"fmt"
	"time"

	"github.com/dapr/durabletask-go/task"
	"github.com/dapr/durabletask-go/workflow"
)

// Activity name constants — must match names registered in the workflow registry.
const (
	activityCreateOrder             = "CreateOrderActivity"
	activityPublishPaymentRequested = "PublishPaymentRequestedActivity"
	activityCancelOrder             = "CancelOrderActivity"
	activityMarkPaid                = "MarkPaidActivity"
)

// CheckoutInput is the workflow input for the OrderSaga.
type CheckoutInput struct {
	IdempotencyKey string
	UserID         string
	SessionID      string
	Currency       string
	Items          []CheckoutItem
	TotalCents     int64
}

// CheckoutItem is a single item in a checkout.
type CheckoutItem struct {
	ProductID  string
	Quantity   int32
	PriceCents int64
}

// CheckoutResult is the workflow output for the OrderSaga.
type CheckoutResult struct {
	State     string `json:"state"`      // RUNNING|COMPLETED|FAILED|FAILED_AFTER_PIVOT
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Reason    string `json:"reason"`
	LastError string `json:"last_error"`
}

// PaymentResult is the merged-outcome payload raised by the event router.
// Subscriber merges payment.completed + payment.failed into a single
// "payment-result-{N}" workflow event.
type PaymentResult struct {
	Success    bool   `json:"success"`
	PaymentID  string `json:"payment_id,omitempty"`
	ReasonCode string `json:"reason_code,omitempty"`
}

// SagaConfig holds all tunable saga parameters read from conf.Saga.
type SagaConfig struct {
	MaxPaymentAttempts   int32
	PerAttemptTimeout    time.Duration
	PaymentInitialDelay  time.Duration
	PaymentBackoffMax    time.Duration
	MarkPaidRetryMax     int32
	MarkPaidBudget       time.Duration
	MarkPaidInitialDelay time.Duration
	MarkPaidBackoffMax   time.Duration
}

// cancelInput is the input struct for CancelOrderActivity.
type cancelInput struct {
	WorkflowID string
	OrderID    string
	Reason     string
}

// markPaidInput is the input struct for MarkPaidActivity.
type markPaidInput struct {
	WorkflowID string
	OrderID    string
	PaymentID  string
}

// PaymentRequestedInput is the input struct for PublishPaymentRequestedActivity.
type PaymentRequestedInput struct {
	WorkflowID string
	OrderID    string
	Amount     int64
	Currency   string
	Attempt    int32
}

// NewOrderSagaWorkflow returns an OrderSaga workflow function that captures
// the saga config via closure, satisfying workflow.Workflow.
func NewOrderSagaWorkflow(cfg SagaConfig) workflow.Workflow {
	return func(ctx *workflow.WorkflowContext) (any, error) {
		var in CheckoutInput
		if err := ctx.GetInput(&in); err != nil {
			return nil, err
		}
		workflowID := ctx.ID()

		// Step 1: Create the order.
		var orderID string
		if err := ctx.CallActivity(activityCreateOrder,
			workflow.WithActivityInput(in)).Await(&orderID); err != nil {
			SagaMetrics.RecordFailed()
			return CheckoutResult{State: "FAILED", Reason: "order_create_failed", LastError: err.Error()}, nil
		}

		// Step 2: Payment loop — attempt 1..MaxPaymentAttempts.
		type outcome struct {
			paymentID  string
			success    bool
			reasonCode string
		}
		var out outcome
		delay := cfg.PaymentInitialDelay // time.Duration constant — deterministic

		maxAttempts := int(cfg.MaxPaymentAttempts)
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if err := ctx.CallActivity(activityPublishPaymentRequested,
				workflow.WithActivityInput(PaymentRequestedInput{
					WorkflowID: workflowID,
					OrderID:    orderID,
					Amount:     in.TotalCents,
					Currency:   in.Currency,
					Attempt:    int32(attempt),
				})).Await(nil); err != nil {
				out.reasonCode = "publish_failed"
			} else {
				// Attempt-keyed event name prevents cross-attempt routing race.
				evtName := fmt.Sprintf("payment-result-%d", attempt)
				ev := ctx.WaitForExternalEvent(evtName, cfg.PerAttemptTimeout)
				var r PaymentResult
				err := ev.Await(&r)
				switch {
				case err == nil && r.Success:
					out = outcome{paymentID: r.PaymentID, success: true}
				case err == nil && !r.Success:
					out.reasonCode = r.ReasonCode
				case errors.Is(err, task.ErrTaskCanceled):
					// ErrTaskCanceled is raised on clean workflow shutdown (context
					// cancel). Treating it as a timeout keeps the saga correct: the
					// loop retries if attempts remain, or compensates. Propagating it
					// as an error would orphan the workflow instance.
					out.reasonCode = "timeout"
				default:
					out.reasonCode = "wait_error"
				}
			}

			if out.success {
				break
			}
			if attempt < maxAttempts {
				if err := ctx.CreateTimer(delay).Await(nil); err != nil {
					return nil, err
				}
				if delay < cfg.PaymentBackoffMax {
					delay *= 2
				}
			}
		}

		// Step 3: Compensate on exhausted payment attempts.
		if !out.success {
			cancelRetry := &workflow.RetryPolicy{
				MaxAttempts:          3,
				InitialRetryInterval: 1 * time.Second,
				BackoffCoefficient:   2.0,
				MaxRetryInterval:     30 * time.Second,
			}
			if cancelErr := ctx.CallActivity(activityCancelOrder,
				workflow.WithActivityInput(cancelInput{
					WorkflowID: workflowID,
					OrderID:    orderID,
					Reason:     out.reasonCode,
				}),
				workflow.WithActivityRetryPolicy(cancelRetry)).Await(nil); cancelErr != nil {
				SagaMetrics.RecordFailedCompensation()
				SagaMetrics.RecordFailed()
				return CheckoutResult{
					State:     "FAILED_COMPENSATION",
					OrderID:   orderID,
					Reason:    "compensation_failed",
					LastError: cancelErr.Error(),
				}, nil
			}
			SagaMetrics.RecordCompensation()
			SagaMetrics.RecordFailed()
			return CheckoutResult{
				State:     "FAILED",
				OrderID:   orderID,
				Reason:    "payment_exhausted",
				LastError: out.reasonCode,
			}, nil
		}

		// Step 4: Post-pivot — mark the order paid.
		// MarkPaid is idempotent on same payment_id (Step 3.5 fix), so replay is safe.
		retry := &workflow.RetryPolicy{
			MaxAttempts:          int(cfg.MarkPaidRetryMax),
			InitialRetryInterval: cfg.MarkPaidInitialDelay,
			BackoffCoefficient:   2.0,
			MaxRetryInterval:     cfg.MarkPaidBackoffMax,
			RetryTimeout:         cfg.MarkPaidBudget,
		}
		if err := ctx.CallActivity(activityMarkPaid,
			workflow.WithActivityInput(markPaidInput{
				WorkflowID: workflowID,
				OrderID:    orderID,
				PaymentID:  out.paymentID,
			}),
			workflow.WithActivityRetryPolicy(retry)).Await(nil); err != nil {
			SagaMetrics.RecordFailedAfterPivot()
			return CheckoutResult{
				State:     "FAILED_AFTER_PIVOT",
				OrderID:   orderID,
				PaymentID: out.paymentID,
				LastError: err.Error(),
			}, nil
		}

		SagaMetrics.RecordCompleted()
		return CheckoutResult{
			State:     "COMPLETED",
			OrderID:   orderID,
			PaymentID: out.paymentID,
		}, nil
	}
}
