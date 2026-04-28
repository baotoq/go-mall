package event

import (
	"context"
	"time"

	"catalog/ent"
	"catalog/ent/outboxmessage"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	sharedevent "shared/event"
)

type entStore struct {
	client *ent.Client
	tx     *ent.Tx
}

func NewEntStore(client *ent.Client) sharedevent.OutboxStore {
	return &entStore{client: client}
}

func (s *entStore) outboxClient() *ent.OutboxMessageClient {
	if s.tx != nil {
		return s.tx.OutboxMessage
	}
	return s.client.OutboxMessage
}

func (s *entStore) CreatePending(ctx context.Context, eventName string, payload []byte) error {
	return s.outboxClient().Create().
		SetEventName(eventName).
		SetPayload(payload).
		SetStatus(sharedevent.StatusPending).
		Exec(ctx)
}

func (s *entStore) ClaimPending(ctx context.Context, limit int) ([]sharedevent.Message, error) {
	rows, err := s.outboxClient().Query().
		Where(outboxmessage.StatusEQ(sharedevent.StatusPending)).
		Order(ent.Asc(outboxmessage.FieldCreatedAt)).
		Limit(limit).
		ForUpdate(sql.WithLockAction(sql.SkipLocked)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	if _, err := s.outboxClient().Update().
		Where(
			outboxmessage.IDIn(ids...),
			outboxmessage.StatusEQ(sharedevent.StatusPending),
		).
		SetStatus(sharedevent.StatusProcessing).
		Save(ctx); err != nil {
		return nil, err
	}

	out := make([]sharedevent.Message, len(rows))
	for i, r := range rows {
		out[i] = sharedevent.Message{
			ID:            r.ID,
			Payload:       r.Payload,
			RetryAttempts: r.RetryAttempts,
		}
	}
	return out, nil
}

func (s *entStore) MarkSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	return s.outboxClient().UpdateOneID(id).
		SetStatus(sharedevent.StatusSent).
		SetSentAt(sentAt).
		Exec(ctx)
}

func (s *entStore) MarkFailed(ctx context.Context, id uuid.UUID) error {
	return s.outboxClient().UpdateOneID(id).
		SetStatus(sharedevent.StatusFailed).
		Exec(ctx)
}

func (s *entStore) RequeueForRetry(ctx context.Context, id uuid.UUID) error {
	return s.outboxClient().UpdateOneID(id).
		SetStatus(sharedevent.StatusPending).
		AddRetryAttempts(1).
		Exec(ctx)
}

func (s *entStore) RunInTx(ctx context.Context, fn func(tx sharedevent.OutboxStore) error) error {
	if s.tx != nil {
		return fn(s)
	}
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(&entStore{client: s.client, tx: tx}); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return rerr
		}
		return err
	}
	return tx.Commit()
}
