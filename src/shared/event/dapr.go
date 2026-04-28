package event

import (
	"context"
	"reflect"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/zeromicro/go-zero/core/logx"
)

// DaprPubSubName is the Dapr component name used for publishing events.
const DaprPubSubName = "pubsub"

// DaprDispatcher publishes events through a Dapr pubsub component, using the
// Go type name of T as the topic.
type DaprDispatcher[T Event] struct {
	dapr dapr.Client
}

// NewDaprDispatcher returns a [DaprDispatcher] bound to the given Dapr client.
func NewDaprDispatcher[T Event](client dapr.Client) *DaprDispatcher[T] {
	return &DaprDispatcher[T]{dapr: client}
}

// PublishEvent publishes event to Dapr on the topic matching the Go type name
// of T. Errors from Dapr are returned unwrapped so callers can decide on retry
// behaviour (the outbox dispatcher uses them to drive its retry policy).
func (d *DaprDispatcher[T]) PublishEvent(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("dapr publishing event", logx.Field("event", event))
	return d.dapr.PublishEvent(ctx, DaprPubSubName, reflect.TypeOf(event).Name(), event)
}
