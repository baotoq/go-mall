package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

func NowUTC() time.Time {
	return time.Now().UTC()
}

type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(NowUTC),

		field.Time("updated_at").
			Optional().
			Nillable().
			UpdateDefault(NowUTC),
	}
}
