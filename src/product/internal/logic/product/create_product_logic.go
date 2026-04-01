// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package product

import (
	"context"
	"fmt"
	"time"

	"product/internal/domain"
	"product/internal/lib"
	"product/internal/svc"
	"product/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a new product
func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ProductCreatedEvent struct {
	OccurredAt time.Time `json:"occurred_at"`
	ProductID  uuid.UUID `json:"id"`
}

func (e ProductCreatedEvent) EventID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func (l *CreateProductLogic) CreateProduct(req *types.CreateProductRequest) (resp *types.CreateProductResponse, err error) {
	l.Logger.Infow("handling create product", logx.Field("req", req))

	p, err := domain.NewProduct(req.Name, req.Description, req.Price, req.TotalStock)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	savedProduct, err := l.svcCtx.Db.Product.Create().
		SetName(p.Name).
		SetDescription(p.Description).
		SetPrice(p.Price).
		SetTotalStock(p.TotalStock).
		SetRemainingStock(p.RemainingStock).
		Save(l.ctx)

	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductCreatedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  p.ID,
	})

	l.Logger.Infow("create product success",
		logx.Field("id", p.ID),
	)

	return &types.CreateProductResponse{Id: savedProduct.ID}, nil
}
