// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"

	"product/ent"
	"product/ent/product"
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
	l.Logger.Infow("handling list product", logx.Field("req", req))

	products, err := l.svcCtx.Db.Product.Query().
		Where(
			product.NameContains(req.Keyword),
			product.DescriptionContains(req.Keyword),
		).
		Offset(int((req.Page - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		All(l.ctx)

	if err != nil {
		return nil, err
	}

	l.Logger.Infow("retrieved products", logx.Field("count", len(products)))

	productInfos := make([]types.ProductInfo, 0, len(products))
	for _, p := range products {
		productInfos = append(productInfos, mapToProductInfo(p))
	}

	l.Logger.Infof("handled list product successfully")
	return &types.ListProductResponse{Total: int64(len(productInfos)), Products: productInfos}, nil
}

func mapToProductInfo(p *ent.Product) types.ProductInfo {
	return types.ProductInfo{
		Id:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		Price:          p.Price,
		TotalStock:     p.TotalStock,
		RemainingStock: p.RemainingStock,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
