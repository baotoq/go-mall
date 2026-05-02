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
	log      *log.Helper
	cancel   context.CancelFunc
}

// NewReconciliationService constructs a ReconciliationService. The tick
// interval is read from conf.Saga.ReconcileInterval; missing or zero falls
// back to 24h.
func NewReconciliationService(repo ReconciliationRepo, sagaCfg *conf.Saga, logger log.Logger) *ReconciliationService {
	interval := 24 * time.Hour
	if sagaCfg != nil && sagaCfg.ReconcileInterval != nil {
		if d := sagaCfg.ReconcileInterval.AsDuration(); d > 0 {
			interval = d
		}
	}
	return &ReconciliationService{repo: repo, interval: interval, log: log.NewHelper(logger)}
}

// Start blocks until ctx is cancelled, running a reconciliation scan every r.interval.
func (r *ReconciliationService) Start(ctx context.Context) error {
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
