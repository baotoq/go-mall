package data

import (
	"context"

	"gomall/app/order/internal/biz"

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

func (r *workflowDLQRepo) Insert(ctx context.Context, topic string, payload []byte, workflowInstanceID string, reason string) error {
	return r.data.db.WorkflowDeadLetterEvent.Create().
		SetTopic(topic).
		SetPayloadJSON(payload).
		SetWorkflowInstanceID(workflowInstanceID).
		SetReason(reason).
		Exec(ctx)
}
