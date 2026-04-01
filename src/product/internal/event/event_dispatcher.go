package event

import (
	"context"
)

type EventDispatcher[T Event] interface {
	Publish(ctx context.Context, event T) error
}
