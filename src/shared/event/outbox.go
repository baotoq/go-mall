package event

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// MaxRetryAttempts is the number of times a message will be retried via
// RequeueForRetry before being marked as [StatusFailed].
const MaxRetryAttempts = 3

// PendingBatchSize is the maximum number of pending messages a single
// [OutboxDispatcher.DispatchPendingEvents] call will lock and process.
const PendingBatchSize = 100

// OutboxDispatcher persists events to an [OutboxStore] before publishing them
// through an inner [Dispatcher]. It implements the transactional outbox
// pattern: PublishEvent only writes to the store, and a separate
// DispatchPendingEvents goroutine drains the store to the inner dispatcher.
type OutboxDispatcher[T Event] struct {
	inner Dispatcher[T]
	store OutboxStore
}

// NewOutboxDispatcher returns an [OutboxDispatcher] that writes to store and
// forwards drained events to inner.
func NewOutboxDispatcher[T Event](inner Dispatcher[T], store OutboxStore) *OutboxDispatcher[T] {
	return &OutboxDispatcher[T]{inner: inner, store: store}
}

// PublishEvent serialises event to JSON and persists it as a pending outbox
// row. It does not call the inner dispatcher; DispatchPendingEvents does that.
func (d *OutboxDispatcher[T]) PublishEvent(ctx context.Context, event T) error {
	logx.WithContext(ctx).Infow("publishing event through outbox dispatcher", logx.Field("event", event))

	payload, err := json.Marshal(event)
	if err != nil {
		logx.WithContext(ctx).Errorw("failed to marshal event payload", logx.Field("error", err))
		return err
	}

	eventName := reflect.TypeOf(event).Name()
	if err := d.store.CreatePending(ctx, eventName, payload); err != nil {
		logx.WithContext(ctx).Errorw("failed to create outbox message", logx.Field("error", err))
		return err
	}
	return nil
}

// DispatchPendingEvents locks a batch of pending messages, marks them
// processing inside one transaction, then publishes each via the inner
// dispatcher. On success messages are marked sent; on failure they are either
// requeued for retry or marked failed once [MaxRetryAttempts] is reached.
func (d *OutboxDispatcher[T]) DispatchPendingEvents(ctx context.Context) error {
	logger := logx.WithContext(ctx)

	var messages []Message
	err := d.store.RunInTx(ctx, func(tx OutboxStore) error {
		var err error
		messages, err = tx.ClaimPending(ctx, PendingBatchSize)
		if err != nil {
			logger.Errorw("failed to query pending outbox messages", logx.Field("error", err))
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("locking pending outbox messages: %w", err)
	}

	wg := sync.WaitGroup{}
	for _, msg := range messages {
		wg.Go(func() {
			d.dispatchOne(ctx, msg, logger)
		})
	}
	wg.Wait()
	return nil
}

func (d *OutboxDispatcher[T]) dispatchOne(ctx context.Context, msg Message, logger logx.Logger) {
	l := logger.WithFields(logx.Field("message_id", msg.ID))

	var event T
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		l.Errorw("failed to unmarshal outbox message payload", logx.Field("error", err))
		d.markFailed(ctx, msg, l)
		return
	}

	if err := d.inner.PublishEvent(ctx, event); err != nil {
		if msg.RetryAttempts >= MaxRetryAttempts {
			l.Errorw("failed to publish event from outbox message max retry attempts reached", logx.Field("error", err))
			d.markFailed(ctx, msg, l)
			return
		}
		l.Infow("failed to publish event from outbox message, will retry dispatching",
			logx.Field("error", err),
			logx.Field("retry_attempts", msg.RetryAttempts+1))
		if rerr := d.store.RequeueForRetry(ctx, msg.ID); rerr != nil {
			l.Errorw("failed to update outbox message retry attempts", logx.Field("error", rerr))
		}
		return
	}

	if err := d.store.MarkSent(ctx, msg.ID, time.Now().UTC()); err != nil {
		l.Errorw("failed to update outbox message status to sent", logx.Field("error", err))
	}
}

func (d *OutboxDispatcher[T]) markFailed(ctx context.Context, msg Message, l logx.Logger) {
	if err := d.store.MarkFailed(ctx, msg.ID); err != nil {
		l.Errorw("failed to update outbox message status to failed", logx.Field("error", err))
	}
}
