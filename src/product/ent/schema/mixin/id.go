package mixin

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type IdMixin struct {
	mixin.Schema
}

func (IdMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				return must(uuid.NewV7())
			}).
			Immutable().
			Unique(),
	}
}
