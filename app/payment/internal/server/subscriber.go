package server

import (
	"context"
	"errors"
	"fmt"

	"gomall/app/payment/internal/biz"
	pkgdapr "gomall/pkg/dapr"
	"gomall/pkg/outbox"

	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// paymentRequestedEvent is the payload published on the payment.requested topic.
type paymentRequestedEvent struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	OrderID            string `json:"order_id"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	Attempt            int32  `json:"attempt"`
}

// PaymentSubscriber registers the Dapr pub/sub handler for payment.requested on
// the shared Kratos HTTP server so the Dapr sidecar discovers and delivers
// events on the same app-port (8000) as the REST API.
type PaymentSubscriber struct {
	uc    *biz.PaymentUsecase
	inbox *outbox.Client
	log   *log.Helper
}

// NewPaymentSubscriber constructs a PaymentSubscriber. Call Register to mount
// routes onto the Kratos HTTP server; Wire wires this via NewHTTPServer.
func NewPaymentSubscriber(uc *biz.PaymentUsecase, inbox *outbox.Client, logger log.Logger) *PaymentSubscriber {
	return &PaymentSubscriber{
		uc:    uc,
		inbox: inbox,
		log:   log.NewHelper(logger),
	}
}

// Subscriptions returns the Dapr pub/sub subscriptions served by this subscriber.
// NewHTTPServer aggregates these and registers /dapr/subscribe once.
func (s *PaymentSubscriber) Subscriptions() []pkgdapr.Subscription {
	return []pkgdapr.Subscription{
		{PubsubName: "pubsub", Topic: "payment.requested", Route: "/dapr/events/payment/requested"},
	}
}

// Register mounts the Dapr event delivery route on srv. The route is wrapped
// with LoopbackOnly so only the sidecar (127.0.0.1) can call it.
// Called from NewHTTPServer before the server starts.
func (s *PaymentSubscriber) Register(srv *kratoshttp.Server) {
	handler := s.inbox.Subscribe("payment.requested", outbox.TypedHandler(s.handlePaymentRequested))
	srv.HandleFunc("/dapr/events/payment/requested", pkgdapr.LoopbackOnly(pkgdapr.TopicHandler(handler)))
}

func (s *PaymentSubscriber) handlePaymentRequested(ctx context.Context, evt paymentRequestedEvent) error {
	if evt.WorkflowInstanceID == "" {
		s.log.Warnf("payment subscriber: missing workflow_instance_id, skipping")
		return nil
	}

	_, err := s.uc.GetByWorkflowAndAttempt(ctx, evt.WorkflowInstanceID, evt.Attempt)
	if err == nil {
		s.log.Infof("payment subscriber: duplicate delivery workflow=%s attempt=%d, ACK",
			evt.WorkflowInstanceID, evt.Attempt)
		return nil
	}
	if !errors.Is(err, biz.ErrPaymentNotFound) {
		return fmt.Errorf("payment subscriber: idempotency check: %w", err)
	}

	_, err = s.uc.CreateFromWorkflow(ctx, biz.CreatePaymentInput{
		OrderID:            evt.OrderID,
		AmountCents:        evt.Amount,
		Currency:           evt.Currency,
		Provider:           "dapr-workflow",
		WorkflowInstanceID: evt.WorkflowInstanceID,
		Attempt:            evt.Attempt,
	})
	if err != nil {
		return fmt.Errorf("payment subscriber: create from workflow: %w", err)
	}
	return nil
}
