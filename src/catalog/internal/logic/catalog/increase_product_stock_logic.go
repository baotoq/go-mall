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

type IncreaseProductStockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncreaseProductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncreaseProductStockLogic {
	return &IncreaseProductStockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncreaseProductStockLogic) IncreaseProductStock(req *types.IncreaseProductStockRequest) (resp *types.IncreaseProductStockResponse, err error) {
	l.Logger.Infow("handling increase product stock", logx.Field("req", req))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be > 0")
	}

	if err := l.svcCtx.Db.Product.
		UpdateOneID(id).
		AddTotalStock(req.Quantity).
		AddRemainingStock(req.Quantity).
		Exec(l.ctx); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductStockIncreasedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  id,
		Quantity:   req.Quantity,
	}); err != nil {
		l.Logger.Errorw("failed to publish product stock increased event", logx.Field("id", id))
		return nil, fmt.Errorf("publish event: %w", err)
	}

	return &types.IncreaseProductStockResponse{Id: id.String()}, nil
}
