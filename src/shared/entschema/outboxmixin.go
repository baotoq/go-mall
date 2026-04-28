// Package entschema provides reusable ent schema fragments shared across
// services, notably an [OutboxMixin] that contributes the outbox-message
// fields used by the shared event dispatcher.
package entschema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"shared/event"
)

// OutboxMixin contributes the fields required by the shared outbox
// dispatcher: event_name, payload, retry_attempts, status, and sent_at.
// It does not contribute id or timestamp fields; compose with each service's
// own id/time mixins.
type OutboxMixin struct {
	mixin.Schema
}

// Fields returns the outbox columns. Status uses [event.MessageStatus] as its
// Go type so generated ent code references the shared enum constants.
func (OutboxMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_name"),
		field.JSON("payload", json.RawMessage{}),
		field.Int32("retry_attempts").Default(0),
		field.Enum("status").GoType(event.MessageStatus("")),
		field.Time("sent_at").Optional(),
	}
}
