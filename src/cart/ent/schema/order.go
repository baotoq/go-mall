package schema

import (
	"cart/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type Order struct {
	ent.Schema
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").NotEmpty(),
		field.String("status").Default("pending"),
		field.Int64("total_amount").NonNegative(),
		field.String("reservation_id").Optional(),
		field.String("payment_id").Optional(),
		field.String("transaction_id").Optional(),
		field.String("failure_reason").Optional(),
		field.JSON("items", []OrderItem{}),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("status"),
	}
}

func (Order) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
