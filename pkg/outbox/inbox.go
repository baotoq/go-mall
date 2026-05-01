package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dapr/go-sdk/service/common"
	"github.com/google/uuid"
)

// Message is delivered to inbox consumer handlers.
type Message struct {
	ID        string
	Topic     string
	Payload   json.RawMessage
	Headers   map[string]string
	CreatedAt time.Time
}

// Handler processes an inbox message.
type Handler func(ctx context.Context, msg Message) error

// Subscribe wraps handler with inbox deduplication keyed on e.ID.
// Returns a Dapr TopicEventHandler to register on a pub/sub subscription.
func (c *Client) Subscribe(topic string, handler Handler) common.TopicEventHandler {
	return func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		messageID := e.ID
		if messageID == "" {
			messageID = uuid.NewString()
			c.log.Warnf("outbox inbox: received event without ID on topic %s; generated fallback %s", topic, messageID)
		}

		var processedAt sql.NullTime
		var inserted bool
		row := c.db.QueryRowContext(ctx, inboxClaimSQL, messageID, topic, c.cfg.ConsumerID)
		if err := row.Scan(&processedAt, &inserted); err != nil {
			return true, err
		}

		if !inserted {
			if processedAt.Valid {
				// True duplicate — already processed successfully.
				return false, nil
			}
			// Mid-flight: another worker is processing or crashed; ask Dapr to redeliver.
			return true, nil
		}

		// Build the message payload.
		var payload json.RawMessage
		if len(e.RawData) > 0 {
			payload = json.RawMessage(e.RawData)
		} else if e.Data != nil {
			b, marshalErr := json.Marshal(e.Data)
			if marshalErr != nil {
				return true, fmt.Errorf("outbox inbox: marshal payload for %s: %w", messageID, marshalErr)
			}
			payload = json.RawMessage(b)
		}

		msg := Message{
			ID:        messageID,
			Topic:     topic,
			Payload:   payload,
			CreatedAt: time.Now(),
		}

		if handlerErr := handler(ctx, msg); handlerErr != nil {
			// Roll back the inbox claim using a fresh context — the handler ctx may be cancelled.
			rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if _, dbErr := c.db.ExecContext(rollbackCtx, inboxRollbackSQL, messageID); dbErr != nil {
				c.log.Errorf("outbox inbox: rollback failed for %s: %v", messageID, dbErr)
			}
			rollbackCancel()
			return true, handlerErr
		}

		// Mark done using a fresh context so a cancelled request ctx doesn't orphan the row.
		markCtx, markCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if _, dbErr := c.db.ExecContext(markCtx, inboxMarkDoneSQL, messageID); dbErr != nil {
			c.log.Errorf("outbox inbox: mark done failed for %s: %v", messageID, dbErr)
		}
		markCancel()
		return false, nil
	}
}
