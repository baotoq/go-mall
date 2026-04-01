// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"product/internal/svc"
	"product/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List products
func NewListProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductLogic {
	return &ListProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductLogic) ListProduct(req *types.ListProductRequest) (resp *types.ListProductResponse, err error) {
	l.Logger.Infof("ListProduct page: %d, pageSize: %d", req.Page, req.PageSize)

	// TODO: query database via l.svcCtx.ProductModel
	return &types.ListProductResponse{Total: 0, Products: []types.ProductInfo{}}, nil
}
