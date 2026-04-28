// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package catalog

import (
	"context"
	"fmt"

	"catalog/internal/domain"
	"catalog/internal/lib"
	"catalog/internal/svc"
	"catalog/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProductLogic) CreateProduct(req *types.CreateProductRequest) (resp *types.CreateProductResponse, err error) {
	l.Logger.Infow("handling create product", logx.Field("req", req))

	var categoryID uuid.UUID
	if req.CategoryId != "" {
		categoryID, err = uuid.Parse(req.CategoryId)
		if err != nil {
			return nil, fmt.Errorf("invalid category id: %w", err)
		}
	}

	p, err := domain.NewProduct(req.Name, req.Slug, req.Description, req.ImageUrl, req.Price, req.TotalStock, categoryID)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	create := l.svcCtx.Db.Product.Create().
		SetName(p.Name).
		SetSlug(p.Slug).
		SetDescription(p.Description).
		SetImageURL(p.ImageURL).
		SetPrice(p.Price).
		SetTotalStock(p.TotalStock).
		SetRemainingStock(p.RemainingStock)

	if categoryID != uuid.Nil {
		create = create.SetCategoryID(categoryID)
	}

	saved, err := create.Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductCreatedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  saved.ID,
	}); err != nil {
		l.Logger.Errorw("failed to publish product created event", logx.Field("id", saved.ID))
		return nil, fmt.Errorf("publish event: %w", err)
	}

	l.Logger.Infow("create product success", logx.Field("id", saved.ID))
	return &types.CreateProductResponse{Id: saved.ID.String()}, nil
}
