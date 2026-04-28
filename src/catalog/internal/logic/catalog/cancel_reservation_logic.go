package catalog

import (
	"context"
	"fmt"

	"catalog/ent/product"
	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CancelReservationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelReservationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelReservationLogic {
	return &CancelReservationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelReservationLogic) CancelReservation(req *types.CancelReservationRequest) (resp *types.ReservationActionResponse, err error) {
	l.Logger.Infow("handling cancel reservation", logx.Field("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation id: %w", err)
	}

	r, err := l.svcCtx.Db.Reservation.Get(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}

	switch r.Status {
	case "cancelled":
		return &types.ReservationActionResponse{Id: r.ID.String(), Status: r.Status}, nil
	case "confirmed":
		return nil, fmt.Errorf("reservation already confirmed")
	case "pending":
	default:
		return nil, fmt.Errorf("unexpected reservation status: %s", r.Status)
	}

	tx, err := l.svcCtx.Db.Tx(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	rollback := func(cause error) error {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", cause, rerr)
		}
		return cause
	}

	for _, it := range r.Items {
		pid, perr := uuid.Parse(it.ProductID)
		if perr != nil {
			return nil, rollback(fmt.Errorf("invalid productId in reservation: %w", perr))
		}
		if _, uerr := tx.Product.Update().
			Where(product.IDEQ(pid)).
			AddRemainingStock(it.Quantity).
			Save(l.ctx); uerr != nil {
			return nil, rollback(fmt.Errorf("restore stock: %w", uerr))
		}
	}

	updated, err := tx.Reservation.UpdateOneID(id).SetStatus("cancelled").Save(l.ctx)
	if err != nil {
		return nil, rollback(fmt.Errorf("update reservation: %w", err))
	}

	if cerr := tx.Commit(); cerr != nil {
		return nil, fmt.Errorf("commit tx: %w", cerr)
	}

	if perr := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ReservationCancelledEvent{
		OccurredAt:    lib.NowUTC(),
		ReservationID: updated.ID,
	}); perr != nil {
		l.Logger.Errorf("publish ReservationCancelledEvent: %v", perr)
	}

	return &types.ReservationActionResponse{Id: updated.ID.String(), Status: updated.Status}, nil
}
