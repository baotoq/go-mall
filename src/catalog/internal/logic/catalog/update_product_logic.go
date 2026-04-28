package catalog

import (
	"context"
	"fmt"

	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductRequest) error {
	l.Logger.Infow("handling update product", logx.Field("req", req))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	update := l.svcCtx.Db.Product.UpdateOneID(id)
	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.Slug != "" {
		update = update.SetSlug(req.Slug)
	}
	if req.Description != "" {
		update = update.SetDescription(req.Description)
	}
	if req.ImageUrl != "" {
		update = update.SetImageURL(req.ImageUrl)
	}
	if req.Price > 0 {
		update = update.SetPrice(req.Price)
	}
	if req.CategoryId != "" {
		categoryID, err := uuid.Parse(req.CategoryId)
		if err != nil {
			return fmt.Errorf("invalid category id: %w", err)
		}
		update = update.SetCategoryID(categoryID)
	}

	if _, err := update.Save(l.ctx); err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductUpdatedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  id,
	}); err != nil {
		l.Logger.Errorw("failed to publish product updated event", logx.Field("id", id))
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}
