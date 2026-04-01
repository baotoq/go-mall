package event

import (
	"context"
	"reflect"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/zeromicro/go-zero/core/logx"
)

type DaprEventDispatcher[T Event] struct {
	dapr dapr.Client
}

func NewDaprEventDispatcher[T Event](dapr dapr.Client) *DaprEventDispatcher[T] {
	return &DaprEventDispatcher[T]{dapr: dapr}
}

func (d *DaprEventDispatcher[T]) Publish(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("publishing event", logx.Field("event", event))
	return d.dapr.PublishEvent(ctx, "pubsub", reflect.TypeOf(event).Name(), event)
}
