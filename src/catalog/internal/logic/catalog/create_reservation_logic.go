// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package catalog

import (
	"context"
	"fmt"

	"catalog/ent"
	"catalog/ent/product"
	"catalog/ent/schema"
	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateReservationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateReservationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateReservationLogic {
	return &CreateReservationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateReservationLogic) CreateReservation(req *types.CreateReservationRequest) (resp *types.ReservationInfo, err error) {
	l.Logger.Infow("handling create reservation", logx.Field("sessionId", req.SessionId), logx.Field("itemCount", len(req.Items)))

	if req.SessionId == "" {
		return nil, fmt.Errorf("sessionId is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("items must not be empty")
	}

	parsedItems := make([]struct {
		productID uuid.UUID
		quantity  int64
	}, 0, len(req.Items))
	for i, it := range req.Items {
		pid, perr := uuid.Parse(it.ProductId)
		if perr != nil {
			return nil, fmt.Errorf("invalid productId at index %d: %w", i, perr)
		}
		if it.Quantity <= 0 {
			return nil, fmt.Errorf("quantity must be > 0 at index %d", i)
		}
		parsedItems = append(parsedItems, struct {
			productID uuid.UUID
			quantity  int64
		}{pid, it.Quantity})
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

	for _, it := range parsedItems {
		affected, uerr := tx.Product.Update().
			Where(product.IDEQ(it.productID), product.RemainingStockGTE(it.quantity)).
			AddRemainingStock(-it.quantity).
			Save(l.ctx)
		if uerr != nil {
			return nil, rollback(fmt.Errorf("decrement stock: %w", uerr))
		}
		if affected == 0 {
			return nil, rollback(fmt.Errorf("insufficient stock for product %s", it.productID))
		}
	}

	items := make([]schema.ReservationItem, 0, len(parsedItems))
	for _, it := range parsedItems {
		items = append(items, schema.ReservationItem{
			ProductID: it.productID.String(),
			Quantity:  it.quantity,
		})
	}

	created, err := tx.Reservation.Create().
		SetSessionID(req.SessionId).
		SetStatus("pending").
		SetItems(items).
		Save(l.ctx)
	if err != nil {
		return nil, rollback(fmt.Errorf("create reservation: %w", err))
	}

	if cerr := tx.Commit(); cerr != nil {
		return nil, fmt.Errorf("commit tx: %w", cerr)
	}

	if perr := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ReservationCreatedEvent{
		OccurredAt:    lib.NowUTC(),
		ReservationID: created.ID,
		SessionID:     created.SessionID,
	}); perr != nil {
		l.Logger.Errorf("publish ReservationCreatedEvent: %v", perr)
	}

	info := mapToReservationInfo(created)
	return &info, nil
}

func mapToReservationInfo(r *ent.Reservation) types.ReservationInfo {
	items := make([]types.ReservationItemInfo, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, types.ReservationItemInfo{ProductId: it.ProductID, Quantity: it.Quantity})
	}
	info := types.ReservationInfo{
		Id:        r.ID.String(),
		SessionId: r.SessionID,
		Status:    r.Status,
		Items:     items,
		CreatedAt: r.CreatedAt.UnixMilli(),
	}
	if r.UpdatedAt != nil {
		info.UpdatedAt = r.UpdatedAt.UnixMilli()
	}
	return info
}
