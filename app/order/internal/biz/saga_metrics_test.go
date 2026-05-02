package biz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gomall/app/order/internal/biz"
)

// TestSagaMetrics_Counters uses delta snapshots so tests are safe against the
// global SagaMetrics singleton being mutated by other tests in this package.

func TestSagaMetrics_RecordCompleted_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.RecordCompleted()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.AttemptsCompleted-before.AttemptsCompleted)
}

func TestSagaMetrics_RecordFailed_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.RecordFailed()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.AttemptsFailed-before.AttemptsFailed)
}

func TestSagaMetrics_RecordFailedAfterPivot_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.RecordFailedAfterPivot()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.AttemptsFailedAfterPivot-before.AttemptsFailedAfterPivot)
}

func TestSagaMetrics_RecordCompensation_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.RecordCompensation()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.Compensations-before.Compensations)
}

func TestSagaMetrics_RecordOrphanPayment_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.RecordOrphanPayment()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.OrphanPayments-before.OrphanPayments)
}

func TestSagaMetrics_IncDrift_incrementsByOne(t *testing.T) {
	before := biz.SagaMetrics.Snapshot()
	biz.SagaMetrics.IncDrift()
	after := biz.SagaMetrics.Snapshot()
	assert.Equal(t, int64(1), after.ReconcileDrift-before.ReconcileDrift)
}

func TestSagaMetrics_Snapshot_capturesAllCounters(t *testing.T) {
	// Arrange — record one of each, capture delta
	before := biz.SagaMetrics.Snapshot()

	biz.SagaMetrics.RecordCompleted()
	biz.SagaMetrics.RecordFailed()
	biz.SagaMetrics.RecordFailedAfterPivot()
	biz.SagaMetrics.RecordCompensation()
	biz.SagaMetrics.RecordOrphanPayment()
	biz.SagaMetrics.IncDrift()

	// Act
	after := biz.SagaMetrics.Snapshot()

	// Assert — each counter advanced by exactly 1 relative to the before snapshot
	assert.Equal(t, int64(1), after.AttemptsCompleted-before.AttemptsCompleted)
	assert.Equal(t, int64(1), after.AttemptsFailed-before.AttemptsFailed)
	assert.Equal(t, int64(1), after.AttemptsFailedAfterPivot-before.AttemptsFailedAfterPivot)
	assert.Equal(t, int64(1), after.Compensations-before.Compensations)
	assert.Equal(t, int64(1), after.OrphanPayments-before.OrphanPayments)
	assert.Equal(t, int64(1), after.ReconcileDrift-before.ReconcileDrift)
}
