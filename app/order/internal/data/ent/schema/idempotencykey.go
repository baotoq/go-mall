package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// IdempotencyKey stores checkout idempotency entries so that duplicate
// Schedule calls with the same key + same user return the previously scheduled
// checkout result.
type IdempotencyKey struct{ ent.Schema }

func (IdempotencyKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").Unique().MaxLen(255),
		field.String("response_json").NotEmpty(),
		field.String("user_id").MaxLen(128),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (IdempotencyKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key", "user_id"),
	}
}
