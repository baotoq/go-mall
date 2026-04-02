// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"
	"fmt"
	"time"

	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type IncreaseProductStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Increase the stock of a product
func NewIncreaseProductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncreaseProductStockLogic {
	return &IncreaseProductStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ProductStockIncreasedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
}

func (e ProductStockIncreasedEvent) EventID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func (l *IncreaseProductStockLogic) IncreaseProductStock(req *types.IncreaseProductStockRequest) (resp *types.IncreaseProductStockResponse, err error) {
	l.Logger.Infow("handling increase product stock", logx.Field("req", req))

	err = l.svcCtx.Db.Product.
		UpdateOneID(req.ID).
		AddTotalStock(req.Quantity).
		AddRemainingStock(req.Quantity).
		Exec(l.ctx)

	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductStockIncreasedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  req.ID,
	}); err != nil {
		l.Logger.Errorw("failed to publish product stock increased event", logx.Field("id", req.ID))
		return nil, fmt.Errorf("publish event: %w", err)
	}

	l.Logger.Infow("increase product stock success", logx.Field("id", req.ID))

	return &types.IncreaseProductStockResponse{Id: req.ID}, nil
}
