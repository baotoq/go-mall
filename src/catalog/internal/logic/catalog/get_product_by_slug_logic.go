package catalog

import (
	"context"
	"fmt"

	"catalog/ent/product"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductBySlugLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductBySlugLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductBySlugLogic {
	return &GetProductBySlugLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductBySlugLogic) GetProductBySlug(req *types.GetProductBySlugRequest) (resp *types.ProductInfo, err error) {
	l.Logger.Infow("getting product by slug", logx.Field("slug", req.Slug))

	p, err := l.svcCtx.Db.Product.Query().
		Where(product.SlugEQ(req.Slug)).
		WithCategory().
		Only(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("get product by slug: %w", err)
	}

	info := mapToProductInfo(p)
	return &info, nil
}
