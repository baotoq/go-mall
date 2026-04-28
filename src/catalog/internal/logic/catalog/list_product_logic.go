package catalog

import (
	"context"
	"fmt"

	"catalog/ent/category"
	"catalog/ent/predicate"
	"catalog/ent/product"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductLogic {
	return &ListProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductLogic) ListProduct(req *types.ListProductRequest) (resp *types.ListProductResponse, err error) {
	l.Logger.Infow("handling list product", logx.Field("req", req))

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	preds := []predicate.Product{}
	if req.Keyword != "" {
		preds = append(preds, product.Or(
			product.NameContainsFold(req.Keyword),
			product.DescriptionContainsFold(req.Keyword),
		))
	}
	if req.CategoryId != "" {
		categoryID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid category id: %w", err)
		}
		preds = append(preds, product.HasCategoryWith(category.ID(categoryID)))
	}

	query := l.svcCtx.Db.Product.Query()
	if len(preds) > 0 {
		query = query.Where(preds...)
	}

	total, err := query.Clone().Count(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	products, err := query.
		WithCategory().
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	infos := make([]types.ProductInfo, 0, len(products))
	for _, p := range products {
		infos = append(infos, mapToProductInfo(p))
	}

	return &types.ListProductResponse{Total: int64(total), Products: infos}, nil
}
