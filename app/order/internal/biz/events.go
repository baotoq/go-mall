package biz

const TopicOrderCreated = "order.created"

type OrderCreatedEvent struct {
	OrderID            string `json:"order_id"`
	UserID             string `json:"user_id"`
	TotalCents         int64  `json:"total_cents"`
	Currency           string `json:"currency"`
	WorkflowInstanceID string `json:"workflow_instance_id,omitempty"`
}
