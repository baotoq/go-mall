// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"product/internal/svc"
	"product/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get product by ID
func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductLogic) GetProduct(req *types.GetProductRequest) (resp *types.ProductInfo, err error) {
	l.Logger.Infow("getting product", logx.Field("id", req.Id))

	product, err := l.svcCtx.Db.Product.Get(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	l.Logger.Infow("retrieved product", logx.Field("product", product))

	return &types.ProductInfo{
		Id:             product.ID,
		Name:           product.Name,
		Description:    product.Description,
		Price:          product.Price,
		TotalStock:     product.TotalStock,
		RemainingStock: product.RemainingStock,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	}, nil
}
