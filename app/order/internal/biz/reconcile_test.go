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

// stubReconciliationRepo is an in-memory ReconciliationRepo for testing.
type stubReconciliationRepo struct {
	called chan struct{}
	rows   []biz.DriftRow
	err    error
}

func newStubReconciliationRepo() *stubReconciliationRepo {
	return &stubReconciliationRepo{called: make(chan struct{}, 16)}
}

func (r *stubReconciliationRepo) FindDriftRows(_ context.Context) ([]biz.DriftRow, error) {
	select {
	case r.called <- struct{}{}:
	default:
	}
	return r.rows, r.err
}

func (r *stubReconciliationRepo) awaitOnce(t *testing.T) {
	t.Helper()
	select {
	case <-r.called:
	case <-time.After(2 * time.Second):
		t.Fatal("ReconciliationService tick did not fire within 2s")
	}
}

func TestNewReconciliationService_nilConfig_disabled(t *testing.T) {
	// Arrange + Act — nil sagaCfg → disabled; Start must return immediately
	svc := biz.NewReconciliationService(newStubReconciliationRepo(), nil, noopLogger{})

	// Assert — Start returns nil without blocking (no ticker started)
	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Start did not return immediately when disabled")
	}
	require.NoError(t, svc.Stop(ctx))
}

func TestNewReconciliationService_zeroInterval_disabled(t *testing.T) {
	// Arrange: zero duration → disabled; Start must return immediately
	sagaCfg := &conf.Saga{ReconcileInterval: durationpb.New(0)}
	svc := biz.NewReconciliationService(newStubReconciliationRepo(), sagaCfg, noopLogger{})

	// Act + Assert
	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Start did not return immediately when disabled")
	}
	require.NoError(t, svc.Stop(ctx))
}

func TestNewReconciliationService_customInterval(t *testing.T) {
	// Arrange
	sagaCfg := &conf.Saga{ReconcileInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubReconciliationRepo()

	// Act — short interval; verify at least one tick fires
	svc := biz.NewReconciliationService(repo, sagaCfg, noopLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	repo.awaitOnce(t)

	cancel()
	require.NoError(t, <-errCh)
}

func TestReconciliationService_Stop_beforeStart_noPanic(t *testing.T) {
	// Arrange — Stop without Start must not panic
	svc := biz.NewReconciliationService(newStubReconciliationRepo(), nil, noopLogger{})

	// Act + Assert
	assert.NotPanics(t, func() {
		require.NoError(t, svc.Stop(context.Background()))
	})
}

func TestReconciliationService_Start_contextCancelled_returnsNil(t *testing.T) {
	// Arrange
	sagaCfg := &conf.Saga{ReconcileInterval: durationpb.New(1 * time.Hour)}
	svc := biz.NewReconciliationService(newStubReconciliationRepo(), sagaCfg, noopLogger{})
	ctx, cancel := context.WithCancel(context.Background())

	// Act — cancel before first tick
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

func TestReconciliationService_Start_driftRows_incrementsMetric(t *testing.T) {
	// Arrange
	sagaCfg := &conf.Saga{ReconcileInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubReconciliationRepo()
	repo.rows = []biz.DriftRow{
		{OrderID: "o1", PaymentID: "p1", Reason: "payment missing"},
		{OrderID: "o2", PaymentID: "p2", Reason: "order missing"},
	}
	svc := biz.NewReconciliationService(repo, sagaCfg, noopLogger{})

	before := biz.SagaMetrics.Snapshot()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	// Wait for at least one full tick to process both rows
	repo.awaitOnce(t)
	// Give the loop body time to call IncDrift before cancelling
	time.Sleep(5 * time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)

	// Assert — at least 2 drift increments (one per row per tick)
	after := biz.SagaMetrics.Snapshot()
	assert.GreaterOrEqual(t, after.ReconcileDrift-before.ReconcileDrift, int64(2))
}

func TestReconciliationService_Start_findDriftError_continuesLoop(t *testing.T) {
	// Arrange — repo always errors; loop must continue, not stop
	sagaCfg := &conf.Saga{ReconcileInterval: durationpb.New(1 * time.Millisecond)}
	repo := newStubReconciliationRepo()
	repo.err = errors.New("db unavailable")
	svc := biz.NewReconciliationService(repo, sagaCfg, noopLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- svc.Start(ctx) }()

	// Two ticks confirms the loop survived the first error
	repo.awaitOnce(t)
	repo.awaitOnce(t)

	cancel()
	require.NoError(t, <-errCh)
}
