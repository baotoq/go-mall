package catalog

import (
	"context"
	"fmt"

	"catalog/internal/domain"
	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCategoryLogic) CreateCategory(req *types.CreateCategoryRequest) (resp *types.CreateCategoryResponse, err error) {
	l.Logger.Infow("handling create category", logx.Field("req", req))

	c, err := domain.NewCategory(req.Name, req.Slug, req.Description)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	saved, err := l.svcCtx.Db.Category.Create().
		SetName(c.Name).
		SetSlug(c.Slug).
		SetDescription(c.Description).
		Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, CategoryCreatedEvent{
		OccurredAt: lib.NowUTC(),
		CategoryID: saved.ID,
	}); err != nil {
		l.Logger.Errorw("failed to publish category created event", logx.Field("id", saved.ID))
		return nil, fmt.Errorf("publish event: %w", err)
	}

	return &types.CreateCategoryResponse{Id: saved.ID.String()}, nil
}
