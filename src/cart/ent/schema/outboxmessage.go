package schema

import (
	"cart/ent/schema/mixin"

	"entgo.io/ent"
	"shared/entschema"
)

type OutboxMessage struct {
	ent.Schema
}

func (OutboxMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
		entschema.OutboxMixin{},
	}
}
