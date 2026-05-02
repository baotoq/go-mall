package server

import (
	"context"
	"fmt"
	"net"

	"gomall/app/payment/internal/biz"
	"gomall/app/payment/internal/conf"
	"gomall/pkg/outbox"

	daprcommon "github.com/dapr/go-sdk/service/common"
	daprhttp "github.com/dapr/go-sdk/service/http"
	"github.com/go-kratos/kratos/v2/log"
)

// paymentRequestedEvent is the payload published on the payment.requested topic.
type paymentRequestedEvent struct {
	WorkflowInstanceID string `json:"workflow_instance_id"`
	OrderID            string `json:"order_id"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	Attempt            int32  `json:"attempt"`
}

// PaymentSubscriber is a kratos transport.Server that hosts the Dapr pub/sub
// subscription for the payment.requested topic.
type PaymentSubscriber struct {
	uc    *biz.PaymentUsecase
	inbox *outbox.Client
	addr  string
	log   *log.Helper
	svc   daprcommon.Service
}

// NewPaymentSubscriber creates the subscriber, binding to a dedicated address
// so the Dapr sidecar can post pub/sub deliveries to it.
func NewPaymentSubscriber(c *conf.Server, uc *biz.PaymentUsecase, inbox *outbox.Client, logger log.Logger) *PaymentSubscriber {
	addr := ":8001"
	if c.Http != nil && c.Http.Addr != "" {
		addr = deriveDaprListenAddr(c.Http.Addr)
	}
	return &PaymentSubscriber{
		uc:    uc,
		inbox: inbox,
		addr:  addr,
		log:   log.NewHelper(logger),
	}
}

// deriveDaprListenAddr bumps the HTTP port by 1 for the Dapr pub/sub endpoint.
func deriveDaprListenAddr(httpAddr string) string {
	host, portStr, err := net.SplitHostPort(httpAddr)
	if err != nil {
		return ":8001"
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return ":8001"
	}
	return fmt.Sprintf("%s:%d", host, port+1)
}

// Start implements kratos transport.Server. Registers the subscription handler
// and begins serving in the background.
func (s *PaymentSubscriber) Start(_ context.Context) error {
	svc := daprhttp.NewService(s.addr)
	s.svc = svc

	sub := &daprcommon.Subscription{
		PubsubName: "pubsub",
		Topic:      "payment.requested",
		Route:      "/payment/requested",
	}
	handler := s.inbox.Subscribe("payment.requested", outbox.TypedHandler(s.handlePaymentRequested))
	if err := svc.AddTopicEventHandler(sub, handler); err != nil {
		return fmt.Errorf("payment subscriber: register handler: %w", err)
	}

	go func() {
		if err := svc.Start(); err != nil {
			s.log.Errorf("payment subscriber: serve error: %v", err)
		}
	}()
	s.log.Infof("payment subscriber: listening on %s", s.addr)
	return nil
}

// Stop implements kratos transport.Server.
func (s *PaymentSubscriber) Stop(_ context.Context) error {
	if s.svc != nil {
		return s.svc.Stop()
	}
	return nil
}

func (s *PaymentSubscriber) handlePaymentRequested(ctx context.Context, evt paymentRequestedEvent) error {
	if evt.WorkflowInstanceID == "" {
		s.log.Warnf("payment subscriber: missing workflow_instance_id, skipping")
		return nil
	}

	// Idempotency: check if a payment for (workflow_instance_id, attempt) already exists.
	_, err := s.uc.GetByWorkflowAndAttempt(ctx, evt.WorkflowInstanceID, evt.Attempt)
	if err == nil {
		s.log.Infof("payment subscriber: duplicate delivery workflow=%s attempt=%d, ACK",
			evt.WorkflowInstanceID, evt.Attempt)
		return nil
	}
	if err != biz.ErrPaymentNotFound {
		// Transient DB error → NACK for redelivery.
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
