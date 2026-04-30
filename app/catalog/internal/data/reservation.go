package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/data/ent"
	entproduct "gomall/app/catalog/internal/data/ent/product"
	entreservation "gomall/app/catalog/internal/data/ent/stockreservation"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type reservationRepo struct {
	data *Data
	log  *log.Helper
}

func NewReservationRepo(data *Data, logger log.Logger) biz.ReservationRepo {
	return &reservationRepo{data: data, log: log.NewHelper(logger)}
}

func (r *reservationRepo) Reserve(ctx context.Context, cartID, productID uuid.UUID, quantity int, ttl time.Duration) (*biz.Reservation, error) {
	var res *biz.Reservation
	err := withTx(ctx, r.data.db, func(tx *ent.Tx) error {
		// Lock product row to prevent concurrent overselling.
		product, err := tx.Product.Query().
			Where(
				entproduct.ID(productID),
				func(s *sql.Selector) { s.ForUpdate() },
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return biz.ErrProductNotFound
			}
			return err
		}

		// Sum ACTIVE reservations for this product.
		reserved, err := tx.StockReservation.Query().
			Where(
				entreservation.ProductID(productID),
				entreservation.StatusEQ(entreservation.StatusACTIVE),
			).
			Aggregate(ent.Sum(entreservation.FieldQuantity)).
			Int(ctx)
		if err != nil {
			return err
		}

		available := product.Stock - reserved

		// Check if ACTIVE reservation already exists for (cart, product) — upsert semantics.
		existing, err := tx.StockReservation.Query().
			Where(
				entreservation.CartID(cartID),
				entreservation.ProductID(productID),
				entreservation.StatusEQ(entreservation.StatusACTIVE),
			).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return err
		}

		expiresAt := time.Now().Add(ttl)
		if existing != nil {
			// Only check the incremental delta against available.
			delta := quantity - existing.Quantity
			if delta > 0 && available < delta {
				return biz.ErrOutOfStock
			}
			updated, err := tx.StockReservation.UpdateOne(existing).
				SetQuantity(quantity).
				SetExpiresAt(expiresAt).
				Save(ctx)
			if err != nil {
				return err
			}
			res = entToReservation(updated)
			return nil
		}

		if available < quantity {
			return biz.ErrOutOfStock
		}

		created, err := tx.StockReservation.Create().
			SetCartID(cartID).
			SetProductID(productID).
			SetQuantity(quantity).
			SetStatus(entreservation.StatusACTIVE).
			SetExpiresAt(expiresAt).
			Save(ctx)
		if err != nil {
			return err
		}
		res = entToReservation(created)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *reservationRepo) Release(ctx context.Context, cartID, productID uuid.UUID) error {
	n, err := r.data.db.StockReservation.Update().
		Where(
			entreservation.CartID(cartID),
			entreservation.ProductID(productID),
			entreservation.StatusEQ(entreservation.StatusACTIVE),
		).
		SetStatus(entreservation.StatusRELEASED).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return biz.ErrReservationNotFound
	}
	return nil
}

func (r *reservationRepo) Adjust(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*biz.Reservation, error) {
	var res *biz.Reservation
	err := withTx(ctx, r.data.db, func(tx *ent.Tx) error {
		existing, err := tx.StockReservation.Query().
			Where(
				entreservation.CartID(cartID),
				entreservation.ProductID(productID),
				entreservation.StatusEQ(entreservation.StatusACTIVE),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return biz.ErrReservationNotFound
			}
			return err
		}

		delta := quantity - existing.Quantity
		if delta > 0 {
			// Need more stock — lock product and check availability.
			product, err := tx.Product.Query().
				Where(
					entproduct.ID(productID),
					func(s *sql.Selector) { s.ForUpdate() },
				).
				Only(ctx)
			if err != nil {
				return err
			}
			reserved, err := tx.StockReservation.Query().
				Where(
					entreservation.ProductID(productID),
					entreservation.StatusEQ(entreservation.StatusACTIVE),
				).
				Aggregate(ent.Sum(entreservation.FieldQuantity)).
				Int(ctx)
			if err != nil {
				return err
			}
			available := product.Stock - reserved
			if available < delta {
				return biz.ErrOutOfStock
			}
		}

		updated, err := tx.StockReservation.UpdateOne(existing).
			SetQuantity(quantity).
			Save(ctx)
		if err != nil {
			return err
		}
		res = entToReservation(updated)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *reservationRepo) Commit(ctx context.Context, cartID, productID uuid.UUID) error {
	n, err := r.data.db.StockReservation.Update().
		Where(
			entreservation.CartID(cartID),
			entreservation.ProductID(productID),
			entreservation.StatusEQ(entreservation.StatusACTIVE),
		).
		SetStatus(entreservation.StatusCOMMITTED).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return biz.ErrReservationNotFound
	}
	return nil
}

func (r *reservationRepo) ReleaseAll(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.data.db.StockReservation.Update().
		Where(
			entreservation.CartID(cartID),
			entreservation.StatusEQ(entreservation.StatusACTIVE),
		).
		SetStatus(entreservation.StatusRELEASED).
		Save(ctx)
	return err
}

func (r *reservationRepo) ExpireStale(ctx context.Context) (int, error) {
	n, err := r.data.db.StockReservation.Update().
		Where(
			entreservation.StatusEQ(entreservation.StatusACTIVE),
			entreservation.ExpiresAtLT(time.Now()),
		).
		SetStatus(entreservation.StatusEXPIRED).
		Save(ctx)
	return n, err
}

func entToReservation(e *ent.StockReservation) *biz.Reservation {
	return &biz.Reservation{
		ID:        e.ID,
		CartID:    e.CartID,
		ProductID: e.ProductID,
		Quantity:  e.Quantity,
		Status:    biz.ReservationStatus(e.Status),
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func withTx(ctx context.Context, client *ent.Client, fn func(*ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
