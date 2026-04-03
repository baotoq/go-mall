package event

import (
	"catalog/ent"
	"catalog/ent/outboxmessage"
	"catalog/ent/schema"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxRetryAttempts = 3
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

func (d *OutboxDispatcher[T]) DispatchPendingEvents(ctx context.Context) error {
	logger := logx.WithContext(ctx)

	var messages []*ent.OutboxMessage

	err := WithTx(ctx, d.db, func(tx *ent.Tx) error {
		var err error
		messages, err = tx.OutboxMessage.Query().
			Where(outboxmessage.StatusEQ(schema.StatusPending)).
			Order(ent.Asc(outboxmessage.FieldCreatedAt)).
			Limit(100).
			ForUpdate(sql.WithLockAction(sql.SkipLocked)).
			All(ctx)

		if err != nil {
			logger.Errorw("failed to query pending outbox messages", logx.Field("error", err))
			return err
		}

		if len(messages) == 0 {
			return nil
		}

		ids := getMessageIDs(messages)
		_, err = tx.OutboxMessage.Update().
			Where(
				outboxmessage.IDIn(ids...),
				outboxmessage.StatusEQ(schema.StatusPending),
			).
			SetStatus(schema.StatusProcessing).
			Save(ctx)

		if err != nil {
			logger.Errorw("failed to update outbox message status", logx.Field("error", err))
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("querying pending outbox messages: %w", err)
	}

	wg := sync.WaitGroup{}
	for _, msg := range messages {
		wg.Go(func() {
			var err error
			l := logger.WithFields(logx.Field("message_id", msg.ID))

			var event T
			err = json.Unmarshal(msg.Payload, &event)
			if err != nil {
				l.Errorw("failed to unmarshal outbox message payload", logx.Field("error", err))

				d.markFailedStatus(ctx, msg, l)
				return
			}

			err = d.inner.PublishEvent(ctx, event)
			if err != nil {
				if msg.RetryAttempts >= maxRetryAttempts {
					l.Errorw("failed to publish event from outbox message max retry attempts reached", logx.Field("error", err))

					d.markFailedStatus(ctx, msg, l)
					return
				}
				l.Infow("failed to publish event from outbox message, will retry dispatching", logx.Field("error", err), logx.Field("retry_attempts", msg.RetryAttempts+1))

				err = d.db.OutboxMessage.UpdateOneID(msg.ID).
					SetStatus(schema.StatusPending).
					AddRetryAttempts(1).
					Exec(ctx)
				if err != nil {
					l.Errorw("failed to update outbox message retry attempts", logx.Field("error", err))
				}
				return
			}

			err = d.db.OutboxMessage.UpdateOneID(msg.ID).
				SetStatus(schema.StatusSent).
				SetSentAt(time.Now().UTC()).
				Exec(ctx)

			if err != nil {
				l.Errorw("failed to update outbox message status to sent", logx.Field("error", err))
				return
			}
		})
	}

	wg.Wait()

	return nil
}

func (d *OutboxDispatcher[T]) markFailedStatus(ctx context.Context, msg *ent.OutboxMessage, l logx.Logger) {
	err := d.db.OutboxMessage.UpdateOneID(msg.ID).
		SetStatus(schema.StatusFailed).
		Exec(ctx)
	if err != nil {
		l.Errorw("failed to update outbox message status to failed", logx.Field("error", err))
	}
}

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func getMessageIDs(messages []*ent.OutboxMessage) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, msg.ID)
	}
	return ids
}
