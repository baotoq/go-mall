package event

import (
	"context"
	"encoding/json"
	"product/ent"
	"product/ent/outboxmessage"
	"product/ent/schema"
	"reflect"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type OutboxDispatcher[T Event] struct {
	inner Dispatcher[T]
	db    *ent.Client
}

func NewOutboxDispatcher[T Event](inner Dispatcher[T], db *ent.Client) *OutboxDispatcher[T] {
	return &OutboxDispatcher[T]{inner: inner, db: db}
}

func (d *OutboxDispatcher[T]) PublishEvent(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("publishing event through outbox dispatcher", logx.Field("event", event))

	jsonPayload, err := json.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorw("failed to marshal event payload", logx.Field("error", err))
		return err
	}

	err = d.db.OutboxMessage.Create().
		SetPayload(jsonPayload).
		SetEventName(reflect.TypeOf(event).Name()).
		SetStatus(schema.StatusPending).
		Exec(ctx)

	if err != nil {
		logx.WithContext(ctx).Errorw("failed to create outbox message", logx.Field("error", err))
		return err
	}

	return nil
}

func (d *OutboxDispatcher[T]) DispatchPendingEvents(ctx context.Context, limit int) error {
	messages, err := d.db.OutboxMessage.Query().
		Where(outboxmessage.StatusEQ(schema.StatusPending)).
		Limit(limit).
		All(ctx)

	logger := logx.WithContext(ctx)

	if err != nil {
		logx.WithContext(ctx).Errorw("failed to query pending outbox messages")
		return err
	}

	wg := sync.WaitGroup{}

	for _, msg := range messages {
		wg.Go(func() {
			l := logger.WithFields(logx.Field("message_id", msg.ID))

			d.db.OutboxMessage.UpdateOneID(msg.ID).
				SetStatus(schema.StatusProcessing).
				Exec(ctx)

			var event T
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				l.Errorw("failed to unmarshal outbox message payload", logx.Field("error", err))
				d.db.OutboxMessage.UpdateOneID(msg.ID).
					SetStatus(schema.StatusFailed).
					Exec(ctx)
				return
			}

			err = d.inner.PublishEvent(ctx, event)
			if err != nil {
				l.Errorw("failed to publish event from outbox message", logx.Field("error", err))
				d.db.OutboxMessage.UpdateOneID(msg.ID).
					SetStatus(schema.StatusFailed).
					SetSentAt(time.Now().UTC()).
					Exec(ctx)
				return
			}

			d.db.OutboxMessage.UpdateOneID(msg.ID).SetStatus(schema.StatusSent).Exec(ctx)
		})
	}

	wg.Wait()

	logger.Infow("finished dispatching pending outbox messages", logx.Field("count", len(messages)))

	return nil
}
