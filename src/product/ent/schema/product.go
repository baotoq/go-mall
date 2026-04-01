package schema

import (
	"product/ent/schema/mixin"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func NowUTC() time.Time {
	return time.Now().UTC()
}

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Optional(),
		field.Float("price"),
		field.Int64("total_stock"),
		field.Int64("remaining_stock"),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
