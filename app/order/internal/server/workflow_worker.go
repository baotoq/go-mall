package server

import (
	"context"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"

	"gomall/app/order/internal/conf"
)

// WorkflowWorker wraps the Dapr workflow runtime as a kratos transport.Server
// so it participates in the kratos lifecycle (Start / Stop).
type WorkflowWorker struct {
	wfc          *workflow.Client
	reg          *workflow.Registry
	cancel       context.CancelFunc
	drainTimeout time.Duration
	log          *log.Helper
}

// NewWorkflowWorker constructs a WorkflowWorker.
// sagaCfg may be nil (Dapr workflow disabled); in that case the worker becomes
// a no-op transport.Server.
func NewWorkflowWorker(wfc *workflow.Client, reg *workflow.Registry, sagaCfg *conf.Saga, logger log.Logger) *WorkflowWorker {
	drain := 30 * time.Second
	if sagaCfg != nil && sagaCfg.DrainTimeout != nil {
		drain = sagaCfg.DrainTimeout.AsDuration()
	}
	return &WorkflowWorker{
		wfc:          wfc,
		reg:          reg,
		drainTimeout: drain,
		log:          log.NewHelper(logger),
	}
}

// Start blocks until ctx is cancelled. kratos calls Start in a goroutine.
func (w *WorkflowWorker) Start(ctx context.Context) error {
	workerCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.log.Infow("msg", "starting Dapr workflow worker")
	if err := w.wfc.StartWorker(workerCtx, w.reg); err != nil {
		cancel()
		return err
	}
	<-ctx.Done()
	return nil
}

// Stop is best-effort cancel. durabletask-go v0.11.3 has no public drain API
// for the gRPC client path; in-flight activity goroutines detach.
// Safety: outbox WithMessageID UNIQUE + MarkPaid same-payment_id idempotent.
func (w *WorkflowWorker) Stop(ctx context.Context) error {
	w.log.Infow("msg", "stopping Dapr workflow worker (best-effort)")
	if w.cancel != nil {
		w.cancel()
	}
	select {
	case <-ctx.Done():
	case <-time.After(w.drainTimeout):
	}
	return nil
}
