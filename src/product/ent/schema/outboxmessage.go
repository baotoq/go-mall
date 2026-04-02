package schema

import (
	"encoding/json"
	"product/ent/schema/mixin"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type MessageStatus string

const (
	StatusPending MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusSent    MessageStatus = "sent"
	StatusFailed  MessageStatus = "failed"
)

func (MessageStatus) Values() []string {
	return []string{
		string(StatusPending),
		string(StatusProcessing),
		string(StatusSent),
		string(StatusFailed),
	}
}

// OutboxMessage holds the schema definition for the OutboxMessage entity.
type OutboxMessage struct {
	ent.Schema
}

func (OutboxMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_name"),
		field.JSON("payload", json.RawMessage{}),
		field.Enum("status").
			GoType(MessageStatus("")),
		field.Time("sent_at").Optional(),
	}
}

func (OutboxMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IdMixin{},
		mixin.TimeMixin{},
	}
}
