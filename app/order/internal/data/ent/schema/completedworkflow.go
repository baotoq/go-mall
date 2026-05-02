package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CompletedWorkflow records workflow instances that have reached a terminal
// state (COMPLETED, FAILED, TERMINATED) and are pending state purge from Dapr.
type CompletedWorkflow struct{ ent.Schema }

func (CompletedWorkflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("instance_id").Unique(),
		field.String("terminal_state"),
		field.Time("terminated_at").Default(time.Now).Immutable(),
		field.Time("purged_at").Optional().Nillable(),
	}
}

func (CompletedWorkflow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("purged_at"),
	}
}
