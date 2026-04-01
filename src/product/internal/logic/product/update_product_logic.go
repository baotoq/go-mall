// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"product/internal/svc"
	"product/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update product
func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductRequest) error {
	l.Logger.Infow("handling update product", logx.Field("req", req))

	_, err := l.svcCtx.Db.Product.UpdateOneID(req.Id).
		SetName(req.Name).
		SetDescription(req.Description).
		SetPrice(req.Price).
		Save(l.ctx)

	if err != nil {
		return err
	}

	l.Logger.Infof("handled update product successfully")

	return nil
}
