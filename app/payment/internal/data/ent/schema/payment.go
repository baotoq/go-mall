package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Payment struct{ ent.Schema }

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("order_id").MaxLen(128),
		field.String("user_id").MaxLen(128),
		field.Int64("amount_cents").Default(0),
		field.String("currency").Default("USD").MaxLen(3),
		field.String("status").Default("PENDING").MaxLen(20),
		field.String("provider").MaxLen(64),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("user_id"),
	}
}
