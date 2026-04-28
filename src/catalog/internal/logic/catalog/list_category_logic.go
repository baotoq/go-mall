package catalog

import (
	"context"
	"fmt"

	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoryLogic {
	return &ListCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCategoryLogic) ListCategory(req *types.ListCategoryRequest) (resp *types.ListCategoryResponse, err error) {
	l.Logger.Infow("handling list category", logx.Field("req", req))

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}

	query := l.svcCtx.Db.Category.Query()
	total, err := query.Clone().Count(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("count categories: %w", err)
	}

	categories, err := query.
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		All(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	infos := make([]types.CategoryInfo, 0, len(categories))
	for _, c := range categories {
		infos = append(infos, mapToCategoryInfo(c))
	}

	return &types.ListCategoryResponse{Total: int64(total), Categories: infos}, nil
}
