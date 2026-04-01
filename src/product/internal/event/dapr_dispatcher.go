package event

import (
	"context"
	"reflect"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/zeromicro/go-zero/core/logx"
)

type DaprDispatcher[T Event] struct {
	dapr dapr.Client
}

func NewDaprDispatcher[T Event](dapr dapr.Client) *DaprDispatcher[T] {
	return &DaprDispatcher[T]{dapr: dapr}
}

func (d *DaprDispatcher[T]) PublishEvent(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("publishing event", logx.Field("event", event))
	return d.dapr.PublishEvent(ctx, "pubsub", reflect.TypeOf(event).Name(), event)
}
