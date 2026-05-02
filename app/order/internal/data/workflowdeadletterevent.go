package data

import (
	"context"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/data/ent"

	"github.com/go-kratos/kratos/v2/log"
)

type workflowDLQRepo struct {
	data *Data
	log  *log.Helper
}

// NewWorkflowDeadLetterEventRepo creates a WorkflowDeadLetterEventRepo data impl.
func NewWorkflowDeadLetterEventRepo(data *Data, logger log.Logger) biz.WorkflowDeadLetterEventRepo {
	return &workflowDLQRepo{data: data, log: log.NewHelper(logger)}
}

// Insert is idempotent: a UNIQUE(workflow_instance_id, topic) violation is
// silently dropped to bound DLQ row growth from repeated orphan events for
// the same workflow.
func (r *workflowDLQRepo) Insert(ctx context.Context, topic string, payload []byte, workflowInstanceID string, reason string) error {
	err := r.data.db.WorkflowDeadLetterEvent.Create().
		SetTopic(topic).
		SetPayloadJSON(payload).
		SetWorkflowInstanceID(workflowInstanceID).
		SetReason(reason).
		Exec(ctx)
	if ent.IsConstraintError(err) {
		return nil
	}
	return err
}
