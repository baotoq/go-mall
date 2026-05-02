package biz

import (
	"net"
	"os"
	"time"

	"gomall/app/order/internal/conf"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ProviderSet = wire.NewSet(
	NewOrderUsecase,
	NewCheckoutUsecase,
	ProvideSagaConfig,
	NewWorkflowClient,
	NewWorkflowRegistry,
	NewPurgeService,
)

// ProvideSagaConfig converts *conf.Saga (proto Duration fields) to the biz
// SagaConfig value type.
func ProvideSagaConfig(c *conf.Saga) SagaConfig {
	if c == nil {
		return SagaConfig{
			MaxPaymentAttempts:  3,
			PerAttemptTimeout:   60 * time.Second,
			PaymentInitialDelay: 500 * time.Millisecond,
			PaymentBackoffMax:   30 * time.Second,
			MarkPaidRetryMax:    5,
			MarkPaidBudget:      5 * time.Minute,
		}
	}
	cfg := SagaConfig{
		MaxPaymentAttempts: c.MaxPaymentAttempts,
		MarkPaidRetryMax:   c.MarkPaidRetryMax,
	}
	if c.PerAttemptTimeout != nil {
		cfg.PerAttemptTimeout = c.PerAttemptTimeout.AsDuration()
	} else {
		cfg.PerAttemptTimeout = 60 * time.Second
	}
	if c.PaymentInitialDelay != nil {
		cfg.PaymentInitialDelay = c.PaymentInitialDelay.AsDuration()
	} else {
		cfg.PaymentInitialDelay = 500 * time.Millisecond
	}
	if c.PaymentBackoffMax != nil {
		cfg.PaymentBackoffMax = c.PaymentBackoffMax.AsDuration()
	} else {
		cfg.PaymentBackoffMax = 30 * time.Second
	}
	if c.MarkPaidBudget != nil {
		cfg.MarkPaidBudget = c.MarkPaidBudget.AsDuration()
	} else {
		cfg.MarkPaidBudget = 5 * time.Minute
	}
	return cfg
}

// NewWorkflowClient creates a *workflow.Client connected to the Dapr sidecar
// gRPC port (DAPR_GRPC_PORT env var, default 50001).
// The workflow.Client is used by CheckoutUsecase to schedule and query sagas.
func NewWorkflowClient(logger log.Logger) (*workflow.Client, func(), error) {
	port := os.Getenv("DAPR_GRPC_PORT")
	if port == "" {
		port = "50001"
	}
	addr := net.JoinHostPort("127.0.0.1", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	wfc := workflow.NewClient(conn)
	cleanup := func() {
		if cerr := conn.Close(); cerr != nil {
			log.NewHelper(logger).Errorf("workflow grpc conn close: %v", cerr)
		}
	}
	return wfc, cleanup, nil
}

// NewWorkflowRegistry builds a *workflow.Registry with all four saga activities
// and the OrderSaga workflow registered. The registry is passed to the
// WorkflowWorker (Step 6.3) to start the worker.
func NewWorkflowRegistry(uc *OrderUsecase, cfg SagaConfig) (*workflow.Registry, error) {
	r := workflow.NewRegistry()

	if err := r.AddWorkflowN("OrderSaga", NewOrderSagaWorkflow(cfg)); err != nil {
		return nil, err
	}
	if err := r.AddActivityN(activityCreateOrder, NewCreateOrderActivity(uc)); err != nil {
		return nil, err
	}
	if err := r.AddActivityN(activityPublishPaymentRequested, NewPublishPaymentRequestedActivity(uc)); err != nil {
		return nil, err
	}
	if err := r.AddActivityN(activityCancelOrder, NewCancelOrderActivity(uc)); err != nil {
		return nil, err
	}
	if err := r.AddActivityN(activityMarkPaid, NewMarkPaidActivity(uc)); err != nil {
		return nil, err
	}
	return r, nil
}
