package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"payment/ent/schema/mixin"
)

type Payment struct {
	ent.Schema
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.String("idempotency_key").NotEmpty(),
		field.Float("total_amount").Positive(),
		field.String("currency").Default("USD"),
		field.String("status").Default("pending"),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("idempotency_key").Unique(),
	}
}

func (Payment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
