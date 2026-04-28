package event

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type testEvent struct {
	ID    uuid.UUID `json:"id"`
	Value string    `json:"value"`
}

func (e testEvent) EventID() uuid.UUID { return e.ID }

type fakeStore struct {
	mu             sync.Mutex
	rows           map[uuid.UUID]*Message
	rowStatus      map[uuid.UUID]MessageStatus
	insertOrder    []uuid.UUID
	createErr      error
	claimErr       error
	markSentErr    error
	markFailedErr  error
	requeueErr     error
	runInTxErr     error
	claimCalls     int
	markSentCalls  int
	markFailedIDs  []uuid.UUID
	requeueIDs     []uuid.UUID
	requeueCounter map[uuid.UUID]int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		rows:           map[uuid.UUID]*Message{},
		rowStatus:      map[uuid.UUID]MessageStatus{},
		requeueCounter: map[uuid.UUID]int{},
	}
}

func (s *fakeStore) CreatePending(_ context.Context, _ string, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.createErr != nil {
		return s.createErr
	}
	id := uuid.Must(uuid.NewV7())
	s.rows[id] = &Message{ID: id, Payload: append([]byte(nil), payload...), RetryAttempts: 0}
	s.rowStatus[id] = StatusPending
	s.insertOrder = append(s.insertOrder, id)
	return nil
}

func (s *fakeStore) ClaimPending(_ context.Context, limit int) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claimCalls++
	if s.claimErr != nil {
		return nil, s.claimErr
	}
	out := make([]Message, 0, limit)
	for _, id := range s.insertOrder {
		if len(out) >= limit {
			break
		}
		if s.rowStatus[id] == StatusPending {
			s.rowStatus[id] = StatusProcessing
			out = append(out, *s.rows[id])
		}
	}
	return out, nil
}

func (s *fakeStore) MarkSent(_ context.Context, id uuid.UUID, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.markSentCalls++
	if s.markSentErr != nil {
		return s.markSentErr
	}
	s.rowStatus[id] = StatusSent
	return nil
}

func (s *fakeStore) MarkFailed(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.markFailedErr != nil {
		return s.markFailedErr
	}
	s.markFailedIDs = append(s.markFailedIDs, id)
	s.rowStatus[id] = StatusFailed
	return nil
}

func (s *fakeStore) RequeueForRetry(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.requeueErr != nil {
		return s.requeueErr
	}
	s.requeueIDs = append(s.requeueIDs, id)
	s.requeueCounter[id]++
	if row, ok := s.rows[id]; ok {
		row.RetryAttempts++
	}
	s.rowStatus[id] = StatusPending
	return nil
}

func (s *fakeStore) RunInTx(ctx context.Context, fn func(tx OutboxStore) error) error {
	if s.runInTxErr != nil {
		return s.runInTxErr
	}
	return fn(s)
}

func (s *fakeStore) statusOf(id uuid.UUID) MessageStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rowStatus[id]
}

type recordingDispatcher struct {
	mu       sync.Mutex
	received []testEvent
	err      error
}

func (r *recordingDispatcher) PublishEvent(_ context.Context, e testEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.received = append(r.received, e)
	return nil
}

func TestPublishEvent_PersistsMarshalledPayload(t *testing.T) {
	store := newFakeStore()
	d := NewOutboxDispatcher[testEvent](&recordingDispatcher{}, store)

	e := testEvent{ID: uuid.Must(uuid.NewV7()), Value: "hello"}
	if err := d.PublishEvent(context.Background(), e); err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}
	if len(store.rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(store.rows))
	}

	var got testEvent
	for _, row := range store.rows {
		if err := json.Unmarshal(row.Payload, &got); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
	}
	if got != e {
		t.Errorf("payload roundtrip: got %+v want %+v", got, e)
	}
}

func TestPublishEvent_ReturnsStoreError(t *testing.T) {
	wantErr := errors.New("db down")
	store := newFakeStore()
	store.createErr = wantErr
	d := NewOutboxDispatcher[testEvent](&recordingDispatcher{}, store)

	err := d.PublishEvent(context.Background(), testEvent{ID: uuid.Must(uuid.NewV7())})
	if !errors.Is(err, wantErr) {
		t.Errorf("got %v, want %v", err, wantErr)
	}
}

func TestDispatchPendingEvents_HappyPath(t *testing.T) {
	store := newFakeStore()
	inner := &recordingDispatcher{}
	d := NewOutboxDispatcher[testEvent](inner, store)
	ctx := context.Background()

	want := []testEvent{
		{ID: uuid.Must(uuid.NewV7()), Value: "a"},
		{ID: uuid.Must(uuid.NewV7()), Value: "b"},
		{ID: uuid.Must(uuid.NewV7()), Value: "c"},
	}
	for _, e := range want {
		if err := d.PublishEvent(ctx, e); err != nil {
			t.Fatalf("PublishEvent: %v", err)
		}
	}

	if err := d.DispatchPendingEvents(ctx); err != nil {
		t.Fatalf("DispatchPendingEvents: %v", err)
	}

	if got := len(inner.received); got != len(want) {
		t.Fatalf("inner publish count: got %d want %d", got, len(want))
	}
	for id := range store.rows {
		if got := store.statusOf(id); got != StatusSent {
			t.Errorf("row %s status: got %s want %s", id, got, StatusSent)
		}
	}
}

func TestDispatchPendingEvents_PublishFailureRequeues(t *testing.T) {
	store := newFakeStore()
	inner := &recordingDispatcher{err: errors.New("dapr down")}
	d := NewOutboxDispatcher[testEvent](inner, store)
	ctx := context.Background()

	if err := d.PublishEvent(ctx, testEvent{ID: uuid.Must(uuid.NewV7()), Value: "x"}); err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}
	if err := d.DispatchPendingEvents(ctx); err != nil {
		t.Fatalf("DispatchPendingEvents: %v", err)
	}

	if len(store.requeueIDs) != 1 {
		t.Errorf("expected 1 requeue, got %d", len(store.requeueIDs))
	}
	for id := range store.rows {
		if got := store.statusOf(id); got != StatusPending {
			t.Errorf("status after retry: got %s want %s", got, StatusPending)
		}
	}
}

func TestDispatchPendingEvents_MaxRetriesMarksFailed(t *testing.T) {
	store := newFakeStore()
	inner := &recordingDispatcher{err: errors.New("dapr down")}
	d := NewOutboxDispatcher[testEvent](inner, store)
	ctx := context.Background()

	if err := d.PublishEvent(ctx, testEvent{ID: uuid.Must(uuid.NewV7()), Value: "x"}); err != nil {
		t.Fatalf("PublishEvent: %v", err)
	}
	var theID uuid.UUID
	for id := range store.rows {
		theID = id
		store.rows[id].RetryAttempts = MaxRetryAttempts
	}

	if err := d.DispatchPendingEvents(ctx); err != nil {
		t.Fatalf("DispatchPendingEvents: %v", err)
	}

	if got := store.statusOf(theID); got != StatusFailed {
		t.Errorf("status: got %s want %s", got, StatusFailed)
	}
	if len(store.markFailedIDs) != 1 || store.markFailedIDs[0] != theID {
		t.Errorf("MarkFailed not called with %s, got %v", theID, store.markFailedIDs)
	}
}

func TestDispatchPendingEvents_BadPayloadMarkedFailed(t *testing.T) {
	store := newFakeStore()
	inner := &recordingDispatcher{}
	d := NewOutboxDispatcher[testEvent](inner, store)
	ctx := context.Background()

	id := uuid.Must(uuid.NewV7())
	store.rows[id] = &Message{ID: id, Payload: []byte("not-json")}
	store.rowStatus[id] = StatusPending
	store.insertOrder = append(store.insertOrder, id)

	if err := d.DispatchPendingEvents(ctx); err != nil {
		t.Fatalf("DispatchPendingEvents: %v", err)
	}
	if got := store.statusOf(id); got != StatusFailed {
		t.Errorf("status: got %s want %s", got, StatusFailed)
	}
	if len(inner.received) != 0 {
		t.Errorf("inner should not be called with bad payload, got %d", len(inner.received))
	}
}

func TestDispatchPendingEvents_ClaimErrorPropagates(t *testing.T) {
	store := newFakeStore()
	store.claimErr = errors.New("lock timeout")
	d := NewOutboxDispatcher[testEvent](&recordingDispatcher{}, store)

	err := d.DispatchPendingEvents(context.Background())
	if err == nil || !errors.Is(err, store.claimErr) {
		t.Errorf("expected wrapped claim error, got %v", err)
	}
}

func TestMessageStatus_ValuesAreEnumValues(t *testing.T) {
	want := []string{"pending", "processing", "sent", "failed"}
	got := MessageStatus("").Values()
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("idx %d: got %s want %s", i, got[i], want[i])
		}
	}
}
