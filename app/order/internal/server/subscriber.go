package server

import (
	"context"
	"encoding/json"
	"fmt"

	"gomall/app/order/internal/biz"
	pkgdapr "gomall/pkg/dapr"
	"gomall/pkg/outbox"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// paymentResultEvent is the payload received on payment.completed and
// payment.failed topics from the payment service.
type paymentResultEvent struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	Attempt            int32  `json:"attempt"`
	PaymentID          string `json:"payment_id,omitempty"`
	OrderID            string `json:"order_id"`
	ReasonCode         string `json:"reason_code,omitempty"`
}

// OrderSubscriber registers Dapr pub/sub handlers on the shared Kratos HTTP
// server so the Dapr sidecar can discover subscriptions and deliver events on
// the same app-port (8000) as the REST API.
type OrderSubscriber struct {
	wfc   *workflow.Client
	dlq   biz.WorkflowDeadLetterEventRepo
	inbox *outbox.Client
	log   *log.Helper
}

// NewOrderSubscriber constructs an OrderSubscriber. Call Register to mount
// routes onto the Kratos HTTP server; Wire wires this via NewHTTPServer.
func NewOrderSubscriber(wfc *workflow.Client, dlq biz.WorkflowDeadLetterEventRepo, inbox *outbox.Client, logger log.Logger) *OrderSubscriber {
	return &OrderSubscriber{
		wfc:   wfc,
		dlq:   dlq,
		inbox: inbox,
		log:   log.NewHelper(logger),
	}
}

// Register mounts the Dapr subscription discovery route and event handlers on
// srv. Called from NewHTTPServer so routes are registered before the server starts.
func (s *OrderSubscriber) Register(srv *kratoshttp.Server) {
	subs := []pkgdapr.Subscription{
		{PubsubName: "pubsub", Topic: "payment.completed", Route: "/dapr/events/payment/completed"},
		{PubsubName: "pubsub", Topic: "payment.failed", Route: "/dapr/events/payment/failed"},
	}
	srv.HandleFunc("/dapr/subscribe", pkgdapr.SubscribeHandler(subs))

	completedHandler := s.inbox.Subscribe("payment.completed", outbox.TypedHandler(s.handlePaymentCompleted))
	failedHandler := s.inbox.Subscribe("payment.failed", outbox.TypedHandler(s.handlePaymentFailed))

	srv.HandleFunc("/dapr/events/payment/completed", pkgdapr.TopicHandler(completedHandler))
	srv.HandleFunc("/dapr/events/payment/failed", pkgdapr.TopicHandler(failedHandler))
}

func (s *OrderSubscriber) handlePaymentCompleted(ctx context.Context, evt paymentResultEvent) error {
	if evt.WorkflowInstanceID == "" {
		s.log.Warnf("order subscriber: payment.completed missing workflow_instance_id, skipping")
		return nil
	}
	result := biz.PaymentResult{
		Success:   true,
		PaymentID: evt.PaymentID,
	}
	return s.raisePaymentResult(ctx, "payment.completed", evt, result)
}

func (s *OrderSubscriber) handlePaymentFailed(ctx context.Context, evt paymentResultEvent) error {
	if evt.WorkflowInstanceID == "" {
		s.log.Warnf("order subscriber: payment.failed missing workflow_instance_id, skipping")
		return nil
	}
	result := biz.PaymentResult{
		Success:    false,
		ReasonCode: evt.ReasonCode,
	}
	return s.raisePaymentResult(ctx, "payment.failed", evt, result)
}

// raisePaymentResult routes a PaymentResult to the waiting workflow instance.
// On NotFound/FailedPrecondition (workflow gone): inserts DLQ row, increments
// orphan counter, ACKs the message. On other gRPC errors: NACKs (returns error
// → Dapr redelivery). On success: ACKs.
func (s *OrderSubscriber) raisePaymentResult(ctx context.Context, daprTopic string, evt paymentResultEvent, result biz.PaymentResult) error {
	eventName := fmt.Sprintf("payment-result-%d", evt.Attempt)

	err := s.wfc.RaiseEvent(ctx, evt.WorkflowInstanceID, eventName,
		workflow.WithEventPayload(result),
	)
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok && (st.Code() == codes.NotFound || st.Code() == codes.FailedPrecondition) {
		s.log.Warnf("order subscriber: workflow %s not found for topic %s attempt %d, storing DLQ: %v",
			evt.WorkflowInstanceID, daprTopic, evt.Attempt, err)
		biz.SagaMetrics.RecordOrphanPayment()

		payload, mErr := json.Marshal(evt)
		if mErr != nil {
			s.log.Errorf("order subscriber: marshal DLQ payload workflow=%s: %v", evt.WorkflowInstanceID, mErr)
			payload = []byte("{}")
		}
		if dlqErr := s.dlq.Insert(ctx, daprTopic, payload, evt.WorkflowInstanceID, st.Message()); dlqErr != nil {
			s.log.Errorf("order subscriber: DLQ insert failed workflow=%s: %v", evt.WorkflowInstanceID, dlqErr)
		}
		return nil // ACK
	}

	return fmt.Errorf("order subscriber: RaiseEvent workflow=%s event=%s: %w",
		evt.WorkflowInstanceID, eventName, err)
}
