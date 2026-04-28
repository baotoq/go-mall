package schema

import (
	"catalog/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReservationItem is the JSON-embedded line item shape inside a Reservation.
type ReservationItem struct {
	ProductID string `json:"productId"`
	Quantity  int64  `json:"quantity"`
}

// Reservation holds a stock hold for a cart at checkout time.
// pending  -> stock has been decremented from remaining_stock
// confirmed -> stock decrement is final (payment succeeded)
// cancelled -> stock has been restored to remaining_stock (payment failed / aborted)
type Reservation struct {
	ent.Schema
}

func (Reservation) Fields() []ent.Field {
	return []ent.Field{
		field.String("session_id").NotEmpty(),
		field.String("status").Default("pending"),
		field.JSON("items", []ReservationItem{}),
	}
}

func (Reservation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}

func (Reservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("status"),
	}
}
