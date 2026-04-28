package catalog

import (
	"context"
	"fmt"

	"catalog/ent/product"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductLogic) GetProduct(req *types.GetProductRequest) (resp *types.ProductInfo, err error) {
	l.Logger.Infow("getting product", logx.Field("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	p, err := l.svcCtx.Db.Product.Query().
		Where(product.ID(id)).
		WithCategory().
		Only(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	info := mapToProductInfo(p)
	return &info, nil
}
