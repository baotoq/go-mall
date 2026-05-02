package biz

import (
	"context"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/go-kratos/kratos/v2/log"

	"gomall/app/order/internal/conf"
)

// CompletedWorkflowRepo persists records of terminal workflow instances so
// they can be purged from the Dapr workflow state store.
type CompletedWorkflowRepo interface {
	// ListPendingPurge returns instance IDs that terminated more than
	// olderThan ago and have not yet been purged.
	ListPendingPurge(ctx context.Context, olderThan time.Duration) ([]string, error)
	// MarkPurged records that the given instance has been purged.
	MarkPurged(ctx context.Context, instanceID string) error
	// Insert records a terminal workflow instance. Idempotent via UNIQUE on instance_id.
	Insert(ctx context.Context, instanceID, terminalState string) error
}

// PurgeService is a kratos transport.Server-compatible background service that
// periodically purges terminal workflow instances from the Dapr state store.
//
// v0 skeleton: the CompletedWorkflow table is populated in a follow-up when
// the subscriber observes workflow termination via FetchWorkflowMetadata.
// For now the ticker fires, ListPendingPurge returns empty, and nothing is purged.
type PurgeService struct {
	wfc      *workflow.Client
	repo     CompletedWorkflowRepo
	interval time.Duration
	log      *log.Helper
	cancel   context.CancelFunc
}

// NewPurgeService constructs a PurgeService.
func NewPurgeService(wfc *workflow.Client, repo CompletedWorkflowRepo, sagaCfg *conf.Saga, logger log.Logger) *PurgeService {
	interval := 6 * time.Hour
	return &PurgeService{
		wfc:      wfc,
		repo:     repo,
		interval: interval,
		log:      log.NewHelper(logger),
	}
}

// Start blocks until ctx is cancelled, running a purge ticker every p.interval.
func (p *PurgeService) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	p.log.Infow("msg", "purge service started", "interval", p.interval)
	for {
		select {
		case <-runCtx.Done():
			return nil
		case <-ticker.C:
			ids, err := p.repo.ListPendingPurge(runCtx, time.Hour)
			if err != nil {
				p.log.Errorw("msg", "purge: list pending failed", "err", err)
				continue
			}
			for _, id := range ids {
				// PurgeWorkflowState removes the workflow history from the
				// Dapr state store. The method name in durabletask-go v0.11.3
				// is PurgeWorkflowState (not PurgeWorkflow).
				if err := p.wfc.PurgeWorkflowState(runCtx, id); err != nil {
					p.log.Warnw("msg", "purge: failed", "id", id, "err", err)
					continue
				}
				if err := p.repo.MarkPurged(runCtx, id); err != nil {
					p.log.Warnw("msg", "purge: mark purged failed", "id", id, "err", err)
				}
			}
		}
	}
}

// Stop cancels the purge ticker context.
func (p *PurgeService) Stop(ctx context.Context) error {
	p.log.Infow("msg", "purge service stopping")
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}
