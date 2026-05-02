package biz

import (
	"sync/atomic"
)

// sagaMetrics holds simple atomic counters for saga observability.
// These are exported via the Metrics() accessor so the server layer can
// expose them over HTTP or to a future Prometheus scrape endpoint.
//
// When github.com/prometheus/client_golang is added to go.mod (Wave 2b), these
// atomics should be replaced by proper prometheus.Counter / prometheus.Histogram
// registrations using the names below.
//
// Metric names (for future Prometheus registration):
//   saga_duration_seconds     histogram  label: outcome
//   saga_attempts_total       counter    label: outcome
//   saga_compensations_total  counter
//   saga_orphan_payment_total counter
type sagaMetrics struct {
	// saga_attempts_total{outcome="completed"}
	attemptsCompleted atomic.Int64
	// saga_attempts_total{outcome="failed"}
	attemptsFailed atomic.Int64
	// saga_attempts_total{outcome="failed_after_pivot"}
	attemptsFailedAfterPivot atomic.Int64
	// saga_compensations_total — incremented when CancelOrder is called.
	compensations atomic.Int64
	// saga_orphan_payment_total — incremented by subscriber on terminated workflow.
	orphanPayments atomic.Int64
	// saga_reconciliation_drift_total — incremented by ReconciliationService on each drift row.
	reconcileDrift atomic.Int64
}

// SagaMetrics is the package-level singleton.
var SagaMetrics sagaMetrics

// RecordCompleted increments the completed attempt counter.
func (m *sagaMetrics) RecordCompleted() { m.attemptsCompleted.Add(1) }

// RecordFailed increments the failed attempt counter.
func (m *sagaMetrics) RecordFailed() { m.attemptsFailed.Add(1) }

// RecordFailedAfterPivot increments the failed-after-pivot counter.
func (m *sagaMetrics) RecordFailedAfterPivot() { m.attemptsFailedAfterPivot.Add(1) }

// RecordCompensation increments the compensation counter (order cancelled by saga).
func (m *sagaMetrics) RecordCompensation() { m.compensations.Add(1) }

// RecordOrphanPayment increments the orphan-payment counter.
// Called by the event subscriber when a payment event arrives for a terminated workflow.
func (m *sagaMetrics) RecordOrphanPayment() { m.orphanPayments.Add(1) }

// IncDrift increments the saga reconciliation drift counter.
func (m *sagaMetrics) IncDrift() { m.reconcileDrift.Add(1) }

// Snapshot returns a point-in-time copy of all counters.
func (m *sagaMetrics) Snapshot() SagaMetricsSnapshot {
	return SagaMetricsSnapshot{
		AttemptsCompleted:        m.attemptsCompleted.Load(),
		AttemptsFailed:           m.attemptsFailed.Load(),
		AttemptsFailedAfterPivot: m.attemptsFailedAfterPivot.Load(),
		Compensations:            m.compensations.Load(),
		OrphanPayments:           m.orphanPayments.Load(),
		ReconcileDrift:           m.reconcileDrift.Load(),
	}
}

// SagaMetricsSnapshot is a read-only copy of saga counters at a point in time.
type SagaMetricsSnapshot struct {
	AttemptsCompleted        int64
	AttemptsFailed           int64
	AttemptsFailedAfterPivot int64
	Compensations            int64
	OrphanPayments           int64
	ReconcileDrift           int64
}
