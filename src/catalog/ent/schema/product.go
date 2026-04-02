package schema

import (
	"catalog/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

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
