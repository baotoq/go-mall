package event

import (
	"context"

	"github.com/google/uuid"
)

type Event interface {
	EventID() uuid.UUID
}

type Dispatcher[T Event] interface {
	PublishEvent(ctx context.Context, event T) error
}
