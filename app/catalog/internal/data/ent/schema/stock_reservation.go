package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type StockReservation struct {
	ent.Schema
}

func (StockReservation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("cart_id", uuid.UUID{}),
		field.UUID("product_id", uuid.UUID{}),
		field.Int("quantity").Positive(),
		field.Enum("status").Values("ACTIVE", "RELEASED", "EXPIRED", "COMMITTED").Default("ACTIVE"),
		field.Time("expires_at"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (StockReservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id", "product_id", "status"),
		index.Fields("product_id", "status"),
		index.Fields("expires_at", "status"),
	}
}
