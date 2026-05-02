package biz

import "context"

// WorkflowDeadLetterEventRepo persists dead-letter events for workflow
// messages that could not be routed (e.g. workflow instance not found).
type WorkflowDeadLetterEventRepo interface {
	Insert(ctx context.Context, topic string, payload []byte, workflowInstanceID string, reason string) error
}
