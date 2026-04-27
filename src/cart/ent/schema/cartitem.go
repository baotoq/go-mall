package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"cart/ent/schema/mixin"
	"github.com/google/uuid"
)

type CartItem struct {
	ent.Schema
}

func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").NotEmpty(),
		field.UUID("product_id", uuid.UUID{}),
		field.Int64("quantity").Positive(),
	}
}

func (CartItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("session_id", "product_id").Unique(),
	}
}

func (CartItem) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
