// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"
	"fmt"

	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete product
func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(req *types.DeleteProductRequest) error {
	l.Logger.Infow("handling delete product", logx.Field("req", req))

	err := l.svcCtx.Db.Product.DeleteOneID(req.Id).Exec(l.ctx)

	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	l.Logger.Infof("handled delete product successfully")
	return nil
}
