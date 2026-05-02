package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type OrderItem struct {
	ProductID     string `json:"product_id"`
	Name          string `json:"name"`
	PriceCents    int64  `json:"price_cents"`
	Currency      string `json:"currency"`
	ImageURL      string `json:"image_url"`
	Quantity      int32  `json:"quantity"`
	SubtotalCents int64  `json:"subtotal_cents"`
}

type Order struct{ ent.Schema }

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("user_id").MaxLen(128),
		field.String("session_id").MaxLen(128),
		field.JSON("items", []OrderItem{}).Optional(),
		field.Int64("total_cents").Default(0),
		field.String("currency").Default("USD").MaxLen(3),
		field.String("status").Default("PENDING").MaxLen(20),
		field.String("payment_id").Optional().MaxLen(64),
		field.String("workflow_instance_id").Optional().Nillable().MaxLen(255),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("workflow_instance_id").Unique(),
	}
}
