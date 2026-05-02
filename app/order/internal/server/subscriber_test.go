package server

// Tests for the order subscriber payment-event handlers.
//
// The production subscriber holds a *workflow.Client (concrete type with no
// exported interface), so we cannot inject a fake without modifying
// subscriber.go.  Cases that exercise wfc.RaiseEvent are skipped with a
// clear explanation.  Cases that short-circuit before touching wfc
// (missing workflow_instance_id) are fully exercised.

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gomall/app/order/internal/biz"
)

// fakeDLQ records Insert calls.
type fakeDLQ struct {
	calls []dlqCall
}

type dlqCall struct {
	topic              string
	payload            []byte
	workflowInstanceID string
	reason             string
}

func (f *fakeDLQ) Insert(_ context.Context, topic string, payload []byte, workflowInstanceID string, reason string) error {
	f.calls = append(f.calls, dlqCall{topic: topic, payload: payload, workflowInstanceID: workflowInstanceID, reason: reason})
	return nil
}

// newTestSubscriber builds an OrderSubscriber with nil wfc (usable only for
// paths that never reach wfc.RaiseEvent).
func newTestSubscriber(dlq biz.WorkflowDeadLetterEventRepo) *OrderSubscriber {
	return &OrderSubscriber{
		wfc: nil,
		dlq: dlq,
		log: log.NewHelper(log.DefaultLogger),
	}
}

// --- Test: missing workflow_instance_id is ACK'd without touching wfc ---

func TestOrderSubscriber_HandlePaymentCompleted_MissingInstanceID(t *testing.T) {
	// Arrange
	dlq := &fakeDLQ{}
	before := biz.SagaMetrics.Snapshot()
	s := newTestSubscriber(dlq)

	// Act
	err := s.handlePaymentCompleted(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "",
		Attempt:            1,
		PaymentID:          "pay-X",
		OrderID:            "ord-Y",
	})

	// Assert
	require.NoError(t, err, "missing instanceID must ACK (return nil)")
	assert.Empty(t, dlq.calls, "no DLQ row expected for missing instanceID")
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, before.OrphanPayments, after.OrphanPayments, "orphan counter must not increment")
}

func TestOrderSubscriber_HandlePaymentFailed_MissingInstanceID(t *testing.T) {
	// Arrange
	dlq := &fakeDLQ{}
	before := biz.SagaMetrics.Snapshot()
	s := newTestSubscriber(dlq)

	// Act
	err := s.handlePaymentFailed(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "",
		Attempt:            2,
		OrderID:            "ord-Y",
		ReasonCode:         "insufficient_funds",
	})

	// Assert
	require.NoError(t, err, "missing instanceID must ACK (return nil)")
	assert.Empty(t, dlq.calls, "no DLQ row expected for missing instanceID")
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, before.OrphanPayments, after.OrphanPayments, "orphan counter must not increment")
}

// --- Cases that require a fake wfc: skipped (wfc is *workflow.Client, a
//     concrete gRPC-backed type with no injectable interface seam). ---

func TestOrderSubscriber_HappyPath_PaymentCompleted_Skipped(t *testing.T) {
	t.Skip("wfc field is *workflow.Client (concrete); RaiseEvent cannot be intercepted without modifying production code")
}

func TestOrderSubscriber_HappyPath_PaymentFailed_Skipped(t *testing.T) {
	t.Skip("wfc field is *workflow.Client (concrete); RaiseEvent cannot be intercepted without modifying production code")
}

func TestOrderSubscriber_WorkflowNotFound_DLQ_Skipped(t *testing.T) {
	t.Skip("wfc field is *workflow.Client (concrete); NotFound path exercised through wfc.RaiseEvent which cannot be faked")
}

func TestOrderSubscriber_WorkflowPreconditionFailed_DLQ_Skipped(t *testing.T) {
	t.Skip("wfc field is *workflow.Client (concrete); FailedPrecondition path exercised through wfc.RaiseEvent which cannot be faked")
}

func TestOrderSubscriber_TransientError_NACK_Skipped(t *testing.T) {
	t.Skip("wfc field is *workflow.Client (concrete); Unavailable/NACK path exercised through wfc.RaiseEvent which cannot be faked")
}

// --- Verify DLQ Insert signature matches the interface (compile-time check) ---

var _ biz.WorkflowDeadLetterEventRepo = (*fakeDLQ)(nil)
