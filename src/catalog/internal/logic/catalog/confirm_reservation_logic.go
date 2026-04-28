package catalog

import (
	"context"
	"fmt"

	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmReservationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmReservationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmReservationLogic {
	return &ConfirmReservationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmReservationLogic) ConfirmReservation(req *types.ConfirmReservationRequest) (resp *types.ReservationActionResponse, err error) {
	l.Logger.Infow("handling confirm reservation", logx.Field("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation id: %w", err)
	}

	r, err := l.svcCtx.Db.Reservation.Get(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}

	switch r.Status {
	case "confirmed":
		return &types.ReservationActionResponse{Id: r.ID.String(), Status: r.Status}, nil
	case "cancelled":
		return nil, fmt.Errorf("reservation already cancelled")
	case "pending":
		updated, uerr := l.svcCtx.Db.Reservation.UpdateOneID(id).SetStatus("confirmed").Save(l.ctx)
		if uerr != nil {
			return nil, fmt.Errorf("update reservation: %w", uerr)
		}
		if perr := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ReservationConfirmedEvent{
			OccurredAt:    lib.NowUTC(),
			ReservationID: updated.ID,
		}); perr != nil {
			l.Logger.Errorf("publish ReservationConfirmedEvent: %v", perr)
		}
		return &types.ReservationActionResponse{Id: updated.ID.String(), Status: updated.Status}, nil
	default:
		return nil, fmt.Errorf("unexpected reservation status: %s", r.Status)
	}
}
