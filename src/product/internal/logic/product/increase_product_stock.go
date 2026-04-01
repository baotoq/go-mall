// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"
	"fmt"

	"product/ent/product"
	"product/internal/svc"
	"product/internal/types"

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

func (l *IncreaseProductStockLogic) IncreaseProductStock(req *types.IncreaseProductStockRequest) (resp *types.IncreaseProductStockResponse, err error) {
	l.Logger.Infow("handling increase product stock", logx.Field("req", req))

	err = l.svcCtx.Db.Product.
		UpdateOneID(req.Id).
		AddTotalStock(req.Quantity).
		AddRemainingStock(req.Quantity).
		Exec(l.ctx)

	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	l.Logger.Infow("increase product stock success",
		logx.Field("id", product.ID),
	)

	return &types.IncreaseProductStockResponse{Id: req.Id}, nil
}
