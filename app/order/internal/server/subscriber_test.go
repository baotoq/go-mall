package server

// Tests for the order subscriber payment-event handlers.
//
// OrderSubscriber.wfc is typed as raiseEventer (interface), so fakeWFC can be
// injected to exercise all paths including happy-path, DLQ, and NACK.

import (
	"context"
	"testing"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

// fakeWFC satisfies raiseEventer, recording calls and returning a configured error.
type fakeWFC struct {
	err   error
	calls []fakeRaiseCall
}

type fakeRaiseCall struct {
	instanceID string
	eventName  string
}

func (f *fakeWFC) RaiseEvent(_ context.Context, id, eventName string, _ ...workflow.RaiseEventOptions) error {
	f.calls = append(f.calls, fakeRaiseCall{instanceID: id, eventName: eventName})
	return f.err
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

func newTestSubscriberWithWFC(wfc raiseEventer, dlq biz.WorkflowDeadLetterEventRepo) *OrderSubscriber {
	return &OrderSubscriber{
		wfc: wfc,
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

// --- Happy path: wfc.RaiseEvent succeeds ---

func TestOrderSubscriber_HappyPath_PaymentCompleted(t *testing.T) {
	// Arrange
	wfc := &fakeWFC{}
	dlq := &fakeDLQ{}
	s := newTestSubscriberWithWFC(wfc, dlq)

	// Act
	err := s.handlePaymentCompleted(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "wf-1",
		Attempt:            1,
		PaymentID:          "pay-1",
		OrderID:            "ord-1",
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, wfc.calls, 1)
	assert.Equal(t, "wf-1", wfc.calls[0].instanceID)
	assert.Equal(t, "payment-result-1", wfc.calls[0].eventName)
	assert.Empty(t, dlq.calls)
}

func TestOrderSubscriber_HappyPath_PaymentFailed(t *testing.T) {
	// Arrange
	wfc := &fakeWFC{}
	dlq := &fakeDLQ{}
	s := newTestSubscriberWithWFC(wfc, dlq)

	// Act
	err := s.handlePaymentFailed(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "wf-1",
		Attempt:            2,
		OrderID:            "ord-1",
		ReasonCode:         "insufficient_funds",
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, wfc.calls, 1)
	assert.Equal(t, "wf-1", wfc.calls[0].instanceID)
	assert.Equal(t, "payment-result-2", wfc.calls[0].eventName)
	assert.Empty(t, dlq.calls)
}

// --- DLQ path: workflow not found / terminated ---

func TestOrderSubscriber_WorkflowNotFound_DLQ(t *testing.T) {
	// Arrange
	wfc := &fakeWFC{err: status.Error(codes.NotFound, "workflow not found")}
	dlq := &fakeDLQ{}
	before := biz.SagaMetrics.Snapshot()
	s := newTestSubscriberWithWFC(wfc, dlq)

	// Act
	err := s.handlePaymentCompleted(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "wf-gone",
		Attempt:            1,
		PaymentID:          "pay-1",
		OrderID:            "ord-1",
	})

	// Assert: ACK, one DLQ row, orphan counter incremented
	require.NoError(t, err)
	require.Len(t, dlq.calls, 1)
	assert.Equal(t, "payment.completed", dlq.calls[0].topic)
	assert.Equal(t, "wf-gone", dlq.calls[0].workflowInstanceID)
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, before.OrphanPayments+1, after.OrphanPayments)
}

func TestOrderSubscriber_WorkflowPreconditionFailed_DLQ(t *testing.T) {
	// Arrange
	wfc := &fakeWFC{err: status.Error(codes.FailedPrecondition, "workflow terminated")}
	dlq := &fakeDLQ{}
	before := biz.SagaMetrics.Snapshot()
	s := newTestSubscriberWithWFC(wfc, dlq)

	// Act
	err := s.handlePaymentFailed(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "wf-terminated",
		Attempt:            3,
		OrderID:            "ord-1",
		ReasonCode:         "timeout",
	})

	// Assert: ACK, DLQ row, orphan counter incremented
	require.NoError(t, err)
	require.Len(t, dlq.calls, 1)
	assert.Equal(t, "payment.failed", dlq.calls[0].topic)
	assert.Equal(t, "wf-terminated", dlq.calls[0].workflowInstanceID)
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, before.OrphanPayments+1, after.OrphanPayments)
}

// --- NACK path: transient gRPC error ---

func TestOrderSubscriber_TransientError_NACK(t *testing.T) {
	// Arrange
	wfc := &fakeWFC{err: status.Error(codes.Unavailable, "service unavailable")}
	dlq := &fakeDLQ{}
	s := newTestSubscriberWithWFC(wfc, dlq)

	// Act
	err := s.handlePaymentCompleted(context.Background(), paymentResultEvent{
		WorkflowInstanceID: "wf-1",
		Attempt:            1,
		PaymentID:          "pay-1",
		OrderID:            "ord-1",
	})

	// Assert: NACK (non-nil error), no DLQ
	assert.Error(t, err, "transient error must NACK")
	assert.Empty(t, dlq.calls)
}

// --- compile-time interface checks ---

var _ biz.WorkflowDeadLetterEventRepo = (*fakeDLQ)(nil)
var _ raiseEventer = (*fakeWFC)(nil)
var _ raiseEventer = (*workflow.Client)(nil)
