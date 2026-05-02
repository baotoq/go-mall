package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type WorkflowDeadLetterEvent struct{ ent.Schema }

func (WorkflowDeadLetterEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("topic").MaxLen(256),
		field.Bytes("payload_json"),
		field.String("workflow_instance_id").MaxLen(256),
		field.String("reason").Optional().MaxLen(512),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (WorkflowDeadLetterEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("topic"),
		index.Fields("workflow_instance_id"),
		// UNIQUE on (workflow_instance_id, topic) bounds DLQ row growth from
		// crafted orphan events: at most one row per (workflow, topic) pair.
		// Insert path uses ON CONFLICT DO NOTHING for idempotency.
		index.Fields("workflow_instance_id", "topic").Unique(),
	}
}
