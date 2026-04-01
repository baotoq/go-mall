package event

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type OutboxDispatcher[T Event] struct {
	inner Dispatcher[T]
}

func NewOutboxDispatcher[T Event](inner Dispatcher[T]) *OutboxDispatcher[T] {
	return &OutboxDispatcher[T]{inner: inner}
}

func (d *OutboxDispatcher[T]) PublishEvent(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("publishing event through outbox dispatcher", logx.Field("event", event))
	return d.inner.PublishEvent(ctx, event)
}
