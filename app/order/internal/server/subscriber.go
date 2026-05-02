package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/conf"
	"gomall/pkg/outbox"

	"github.com/dapr/durabletask-go/workflow"
	daprcommon "github.com/dapr/go-sdk/service/common"
	daprhttp "github.com/dapr/go-sdk/service/http"
	"github.com/go-kratos/kratos/v2/log"
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

// OrderSubscriber is a kratos transport.Server that hosts Dapr pub/sub
// subscriptions for payment.completed and payment.failed topics.
// It maps each event to a "payment-result-{N}" workflow external event and
// calls wfc.RaiseEvent to unblock the waiting OrderSaga.
type OrderSubscriber struct {
	wfc    *workflow.Client
	dlq    biz.WorkflowDeadLetterEventRepo
	inbox  *outbox.Client
	addr   string
	log    *log.Helper
	svc    daprcommon.Service
}

// NewOrderSubscriber creates the subscriber, binding to a dedicated address so
// the Dapr sidecar can deliver pub/sub events to it.
func NewOrderSubscriber(c *conf.Server, wfc *workflow.Client, dlq biz.WorkflowDeadLetterEventRepo, inbox *outbox.Client, logger log.Logger) *OrderSubscriber {
	addr := ":8002"
	if c.Http != nil && c.Http.Addr != "" {
		addr = deriveOrderSubscriberAddr(c.Http.Addr)
	}
	return &OrderSubscriber{
		wfc:   wfc,
		dlq:   dlq,
		inbox: inbox,
		addr:  addr,
		log:   log.NewHelper(logger),
	}
}

// deriveOrderSubscriberAddr bumps the HTTP port by 2 for the order pub/sub endpoint.
func deriveOrderSubscriberAddr(httpAddr string) string {
	host, portStr, err := net.SplitHostPort(httpAddr)
	if err != nil {
		return ":8002"
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return ":8002"
	}
	return fmt.Sprintf("%s:%d", host, port+2)
}

// Start implements kratos transport.Server.
func (s *OrderSubscriber) Start(_ context.Context) error {
	svc := daprhttp.NewService(s.addr)
	s.svc = svc

	completedSub := &daprcommon.Subscription{
		PubsubName: "pubsub",
		Topic:      "payment.completed",
		Route:      "/payment/completed",
	}
	failedSub := &daprcommon.Subscription{
		PubsubName: "pubsub",
		Topic:      "payment.failed",
		Route:      "/payment/failed",
	}

	completedHandler := s.inbox.Subscribe("payment.completed", outbox.TypedHandler(s.handlePaymentCompleted))
	if err := svc.AddTopicEventHandler(completedSub, completedHandler); err != nil {
		return fmt.Errorf("order subscriber: register payment.completed handler: %w", err)
	}

	failedHandler := s.inbox.Subscribe("payment.failed", outbox.TypedHandler(s.handlePaymentFailed))
	if err := svc.AddTopicEventHandler(failedSub, failedHandler); err != nil {
		return fmt.Errorf("order subscriber: register payment.failed handler: %w", err)
	}

	go func() {
		if err := svc.Start(); err != nil {
			s.log.Errorf("order subscriber: serve error: %v", err)
		}
	}()
	s.log.Infof("order subscriber: listening on %s", s.addr)
	return nil
}

// Stop implements kratos transport.Server.
func (s *OrderSubscriber) Stop(_ context.Context) error {
	if s.svc != nil {
		return s.svc.Stop()
	}
	return nil
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

	// Classify the gRPC error.
	st, ok := status.FromError(err)
	if ok && (st.Code() == codes.NotFound || st.Code() == codes.FailedPrecondition) {
		// Workflow instance is gone (terminated/purged). Store as dead letter
		// and ACK so the message is not redelivered endlessly.
		s.log.Warnf("order subscriber: workflow %s not found for topic %s attempt %d, storing DLQ: %v",
			evt.WorkflowInstanceID, daprTopic, evt.Attempt, err)
		biz.SagaMetrics.RecordOrphanPayment()

		payload, _ := json.Marshal(evt)
		if dlqErr := s.dlq.Insert(ctx, daprTopic, payload, evt.WorkflowInstanceID, st.Message()); dlqErr != nil {
			s.log.Errorf("order subscriber: DLQ insert failed workflow=%s: %v", evt.WorkflowInstanceID, dlqErr)
		}
		return nil // ACK
	}

	// Transient or unknown gRPC error — NACK for Dapr redelivery.
	return fmt.Errorf("order subscriber: RaiseEvent workflow=%s event=%s: %w",
		evt.WorkflowInstanceID, eventName, err)
}
