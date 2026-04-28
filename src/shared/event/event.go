// Package event provides shared event-bus and outbox primitives for services.
//
// It defines the [Event] and [Dispatcher] interfaces, a Dapr-backed dispatcher,
// and a generic [OutboxDispatcher] that persists events to a per-service
// outbox table before publishing them. Services adapt their concrete ent
// client to the [OutboxStore] interface to plug into [OutboxDispatcher].
package event

import (
	"context"

	"github.com/google/uuid"
)

// Event is the contract every domain event must satisfy.
type Event interface {
	EventID() uuid.UUID
}

// Dispatcher publishes events to a downstream transport.
type Dispatcher[T Event] interface {
	PublishEvent(ctx context.Context, event T) error
}
