package data

import (
	"context"
	"time"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/data/ent/completedworkflow"

	"github.com/go-kratos/kratos/v2/log"
)

type completedWorkflowRepo struct {
	data *Data
	log  *log.Helper
}

// NewCompletedWorkflowRepo creates a CompletedWorkflowRepo data implementation.
func NewCompletedWorkflowRepo(data *Data, logger log.Logger) biz.CompletedWorkflowRepo {
	return &completedWorkflowRepo{data: data, log: log.NewHelper(logger)}
}

// ListPendingPurge returns instance IDs whose terminated_at is older than
// olderThan ago and whose purged_at is still NULL.
// TODO(saga): this returns empty until the subscriber populates the table.
func (r *completedWorkflowRepo) ListPendingPurge(ctx context.Context, olderThan time.Duration) ([]string, error) {
	cutoff := time.Now().Add(-olderThan)
	rows, err := r.data.db.CompletedWorkflow.Query().
		Where(
			completedworkflow.PurgedAtIsNil(),
			completedworkflow.TerminatedAtLTE(cutoff),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.InstanceID)
	}
	return ids, nil
}

// MarkPurged sets purged_at = now for the given instance_id.
// Uses upsert-like update: if the row doesn't exist it is a no-op (the workflow
// may have been removed from the table by a concurrent cleanup).
func (r *completedWorkflowRepo) MarkPurged(ctx context.Context, instanceID string) error {
	now := time.Now()
	return r.data.db.CompletedWorkflow.Update().
		Where(completedworkflow.InstanceIDEQ(instanceID)).
		SetPurgedAt(now).
		Exec(ctx)
}
