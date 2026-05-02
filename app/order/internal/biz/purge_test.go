package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/conf"
)

// stubCompletedWorkflowRepo is an in-memory CompletedWorkflowRepo for testing.
type stubCompletedWorkflowRepo struct {
	// listed is signalled each time ListPendingPurge is called.
	listed  chan struct{}
	listIDs []string
	listErr error
	markErr error
	purged  []string
}

func newStubCompletedWorkflowRepo() *stubCompletedWorkflowRepo {
	return &stubCompletedWorkflowRepo{listed: make(chan struct{}, 16)}
}

func (r *stubCompletedWorkflowRepo) ListPendingPurge(_ context.Context, _ time.Duration) ([]string, error) {
	select {
	case r.listed <- struct{}{}:
	default:
	}
	return r.listIDs, r.listErr
}

func (r *stubCompletedWorkflowRepo) MarkPurged(_ context.Context, id string) error {
	r.purged = append(r.purged, id)
	return r.markErr
}

func (r *stubCompletedWorkflowRepo) Insert(_ context.Context, _, _ string) error {
	return nil
}

// awaitOnce blocks until the stub's listed channel receives or times out.
func (r *stubCompletedWorkflowRepo) awaitOnce(t *testing.T) {
	t.Helper()
	select {
	case <-r.listed:
	case <-time.After(2 * time.Second):
		t.Fatal("PurgeService tick did not fire within 2s")
	}
}

func TestNewPurgeService_nilConfig_defaultsTo6h(t *testing.T) {
	// Arrange + Act
	svc := biz.NewPurgeService(nil, newStubCompletedWorkflowRepo(), nil, noopLogger{})

	// Assert — only Stop to observe that the zero-value cancel is safe
	require.NoError(t, svc.Stop(context.Background()))
}

func TestNewPurgeService_zeroInterval_defaultsTo6h(t *testing.T) {
	// Arrange: PurgeInterval = 0 should fall back to 6h (Stop verifies no panic)
	sagaCfg := &conf.Saga{PurgeInterval: durationpb.New(0)}

	// Act + Assert — service constructs without error; Stop is safe
	svc := biz.NewPurgeService(nil, newStubCompletedWorkflowRepo(), sagaCfg, noopLogger{})
	require.NoError(t, svc.Stop(context.Background()))
}

func TestNewPurgeService_customInterval(t *testing.T) {
	// Arrange
	sagaCfg := &conf.Saga{PurgeInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubCompletedWorkflowRepo()

	// Act — start with a 1ms interval; at least one tick should fire quickly
	svc := biz.NewPurgeService(nil, repo, sagaCfg, noopLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	// Assert — ListPendingPurge called at least once
	repo.awaitOnce(t)

	cancel()
	require.NoError(t, <-errCh)
}

func TestPurgeService_Stop_beforeStart_noPanic(t *testing.T) {
	// Arrange — Stop without a preceding Start must not panic (cancel is nil)
	svc := biz.NewPurgeService(nil, newStubCompletedWorkflowRepo(), nil, noopLogger{})

	// Act + Assert
	assert.NotPanics(t, func() {
		require.NoError(t, svc.Stop(context.Background()))
	})
}

func TestPurgeService_Start_contextCancelled_returnsNil(t *testing.T) {
	// Arrange
	sagaCfg := &conf.Saga{PurgeInterval: durationpb.New(1 * time.Hour)}
	svc := biz.NewPurgeService(nil, newStubCompletedWorkflowRepo(), sagaCfg, noopLogger{})
	ctx, cancel := context.WithCancel(context.Background())

	// Act — cancel before any tick fires
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()
	cancel()

	// Assert
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancel")
	}
}

func TestPurgeService_Start_listError_continuesLoop(t *testing.T) {
	// Arrange — repo always errors; service should log + continue, not stop
	sagaCfg := &conf.Saga{PurgeInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubCompletedWorkflowRepo()
	repo.listErr = errors.New("db unavailable")
	svc := biz.NewPurgeService(nil, repo, sagaCfg, noopLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	// Wait for at least two ticks to confirm the loop keeps running after error
	repo.awaitOnce(t)
	repo.awaitOnce(t)

	cancel()
	require.NoError(t, <-errCh)
}

func TestPurgeService_Start_emptyList_noWfcCalled(t *testing.T) {
	// Arrange — wfc is nil; if the empty-list path ever touched wfc it would panic
	sagaCfg := &conf.Saga{PurgeInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubCompletedWorkflowRepo()
	repo.listIDs = []string{}

	svc := biz.NewPurgeService(nil, repo, sagaCfg, noopLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	// Assert — tick fires, empty list, no panic, no wfc call
	repo.awaitOnce(t)
	cancel()
	require.NoError(t, <-errCh)
}
