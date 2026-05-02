package data

import (
	"context"
	"encoding/json"

	"gomall/app/order/internal/biz"
	entidempotency "gomall/app/order/internal/data/ent/idempotencykey"

	"github.com/go-kratos/kratos/v2/log"
)

type idempotencyKeyRepo struct {
	data *Data
	log  *log.Helper
}

// NewIdempotencyKeyRepo creates a new IdempotencyKeyRepo data implementation.
func NewIdempotencyKeyRepo(data *Data, logger log.Logger) biz.IdempotencyKeyRepo {
	return &idempotencyKeyRepo{data: data, log: log.NewHelper(logger)}
}

func (r *idempotencyKeyRepo) Get(ctx context.Context, key string) (biz.StoredCheckout, bool, error) {
	row, err := r.data.db.IdempotencyKey.Query().
		Where(entidempotency.Key(key)).
		Only(ctx)
	if err != nil {
		if isEntNotFound(err) {
			return biz.StoredCheckout{}, false, nil
		}
		return biz.StoredCheckout{}, false, err
	}
	var sc biz.StoredCheckout
	if err := json.Unmarshal([]byte(row.ResponseJSON), &sc); err != nil {
		return biz.StoredCheckout{}, false, err
	}
	return sc, true, nil
}

func (r *idempotencyKeyRepo) Put(ctx context.Context, key string, sc biz.StoredCheckout) error {
	b, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	// Try to insert; if the key already exists (unique constraint), treat as
	// idempotent success (first writer wins).
	err = r.data.db.IdempotencyKey.Create().
		SetKey(key).
		SetResponseJSON(string(b)).
		SetUserID(sc.UserID).
		Exec(ctx)
	if err != nil && !isEntConstraintError(err) {
		return err
	}
	return nil
}

// isEntNotFound returns true if err is an ent not-found error.
func isEntNotFound(err error) bool {
	type notFounder interface{ IsNotFound() bool }
	if nf, ok := err.(notFounder); ok {
		return nf.IsNotFound()
	}
	return false
}

// isEntConstraintError returns true for unique-constraint violations, which
// mean the key already exists (idempotent: OK).
func isEntConstraintError(err error) bool {
	type constraintErr interface{ IsConstraintError() bool }
	if ce, ok := err.(constraintErr); ok {
		return ce.IsConstraintError()
	}
	return false
}
