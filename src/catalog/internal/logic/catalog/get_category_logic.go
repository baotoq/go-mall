package catalog

import (
	"context"
	"fmt"

	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCategoryLogic) GetCategory(req *types.GetCategoryRequest) (resp *types.CategoryInfo, err error) {
	l.Logger.Infow("getting category", logx.Field("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid category id: %w", err)
	}

	c, err := l.svcCtx.Db.Category.Get(l.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}

	info := mapToCategoryInfo(c)
	return &info, nil
}
