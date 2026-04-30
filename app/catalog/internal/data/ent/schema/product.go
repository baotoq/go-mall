package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty().MaxLen(200),
		field.String("slug").Unique().MaxLen(200),
		field.Text("description").Optional().Nillable(),
		field.Int64("price_cents").Default(0),
		field.String("currency").Default("USD").MaxLen(3),
		field.String("image_url").Optional().Nillable(),
		field.Enum("theme").Values("light", "dark").Default("light"),
		field.Int("stock").Default(0),
		field.UUID("category_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("products").
			Field("category_id").
			Unique(),
	}
}

func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category_id", "created_at"),
	}
}
