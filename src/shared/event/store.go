package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MessageStatus is the lifecycle state of an outbox message.
type MessageStatus string

// Outbox message status values. The string values are the canonical
// on-the-wire representation persisted by ent's enum field.
const (
	StatusPending    MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusSent       MessageStatus = "sent"
	StatusFailed     MessageStatus = "failed"
)

// Values reports all valid [MessageStatus] strings. ent's schema.EnumValues
// interface is satisfied by [MessageStatus] via this method.
func (MessageStatus) Values() []string {
	return []string{
		string(StatusPending),
		string(StatusProcessing),
		string(StatusSent),
		string(StatusFailed),
	}
}

// Message is a transport-agnostic projection of an outbox row, decoupling the
// shared dispatcher from any service's ent-generated types.
type Message struct {
	ID            uuid.UUID
	Payload       []byte
	RetryAttempts int32
}

// OutboxStore is the persistence boundary the [OutboxDispatcher] depends on.
// Each service implements it as a thin adapter over its own ent client.
//
// All methods take ctx and return any underlying database error verbatim.
// Implementations must guarantee that [OutboxStore.RunInTx] runs fn inside
// a single database transaction and rolls back on error or panic; the store
// passed to fn must observe that transaction.
type OutboxStore interface {
	CreatePending(ctx context.Context, eventName string, payload []byte) error

	// ClaimPending atomically locks up to limit pending messages and
	// transitions them to [StatusProcessing], returning the claimed rows.
	// Implementations must use FOR UPDATE SKIP LOCKED (or the dialect
	// equivalent) and perform the status update in the same transaction so
	// concurrent dispatchers do not double-process. Must be called inside
	// [OutboxStore.RunInTx].
	ClaimPending(ctx context.Context, limit int) ([]Message, error)

	MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID) error

	// RequeueForRetry sets the message back to pending and increments its
	// retry counter atomically.
	RequeueForRetry(ctx context.Context, id uuid.UUID) error

	// RunInTx runs fn inside a single database transaction. The store passed
	// to fn shares that transaction.
	RunInTx(ctx context.Context, fn func(tx OutboxStore) error) error
}
