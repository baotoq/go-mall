package data

import (
	"context"

	"gomall/app/order/internal/biz"
)

type reconcileRepo struct {
	data *Data
}

// NewReconciliationRepo creates a ReconciliationRepo data implementation.
func NewReconciliationRepo(data *Data) biz.ReconciliationRepo {
	return &reconcileRepo{data: data}
}

// FindDriftRows queries for saga drift between payments and orders.
// TODO(saga reconcile): SQL join payments(status=COMPLETED) ↔ orders(status=PAID, payment_id).
// For v0 returns empty slice — payments and orders live in separate services/DBs.
func (r *reconcileRepo) FindDriftRows(ctx context.Context) ([]biz.DriftRow, error) {
	return nil, nil
}
