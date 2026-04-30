package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CartItem struct{ ent.Schema }

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("cart_id", uuid.UUID{}),
		field.UUID("product_id", uuid.UUID{}),
		field.String("name").MaxLen(200),
		field.Int64("price_cents").Default(0),
		field.String("currency").Default("USD").MaxLen(3),
		field.String("image_url").Optional().Nillable(),
		field.Int("quantity").Default(1),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CartItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cart", Cart.Type).Ref("items").Field("cart_id").Unique().Required(),
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cart_id", "product_id").Unique(),
	}
}
