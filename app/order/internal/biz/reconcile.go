package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"gomall/app/order/internal/conf"
)

// DriftRow describes a single saga reconciliation discrepancy.
type DriftRow struct {
	OrderID   string
	PaymentID string
	Reason    string
}

// ReconciliationRepo detects saga drift between the orders and payments tables.
type ReconciliationRepo interface {
	FindDriftRows(ctx context.Context) ([]DriftRow, error)
}

// ReconciliationService is a kratos transport.Server-compatible background
// service that periodically scans for saga drift and logs each discrepancy.
type ReconciliationService struct {
	repo     ReconciliationRepo
	interval time.Duration
	disabled bool
	log      *log.Helper
	cancel   context.CancelFunc
}

// NewReconciliationService constructs a ReconciliationService. The service is
// disabled (Start returns immediately) when ReconcileInterval is absent or
// zero — an explicit positive interval must be configured to activate polling.
func NewReconciliationService(repo ReconciliationRepo, sagaCfg *conf.Saga, logger log.Logger) *ReconciliationService {
	var interval time.Duration
	disabled := true
	if sagaCfg != nil && sagaCfg.ReconcileInterval != nil {
		if d := sagaCfg.ReconcileInterval.AsDuration(); d > 0 {
			interval = d
			disabled = false
		}
	}
	return &ReconciliationService{repo: repo, interval: interval, disabled: disabled, log: log.NewHelper(logger)}
}

// Start blocks until ctx is cancelled, running a reconciliation scan every r.interval.
// Returns immediately if the service was constructed without a positive ReconcileInterval.
func (r *ReconciliationService) Start(ctx context.Context) error {
	if r.disabled {
		r.log.Infow("msg", "reconciliation service disabled (ReconcileInterval not configured)")
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	r.log.Infow("msg", "reconciliation service started", "interval", r.interval)
	for {
		select {
		case <-runCtx.Done():
			return nil
		case <-ticker.C:
			rows, err := r.repo.FindDriftRows(runCtx)
			if err != nil {
				r.log.Errorw("msg", "reconcile: query failed", "err", err)
				continue
			}
			for _, row := range rows {
				r.log.Warnw("msg", "saga drift detected", "order_id", row.OrderID, "payment_id", row.PaymentID, "reason", row.Reason)
				SagaMetrics.IncDrift()
			}
		}
	}
}

// Stop cancels the reconciliation ticker context.
func (r *ReconciliationService) Stop(ctx context.Context) error {
	r.log.Infow("msg", "reconciliation service stopping")
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
