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

type DeleteProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductLogic {
	return &DeleteProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProductLogic) DeleteProduct(req *types.DeleteProductRequest) error {
	l.Logger.Infow("handling delete product", logx.Field("id", req.Id))

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	if err := l.svcCtx.Db.Product.DeleteOneID(id).Exec(l.ctx); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if err := l.svcCtx.Dispatcher.PublishEvent(l.ctx, ProductDeletedEvent{
		OccurredAt: lib.NowUTC(),
		ProductID:  id,
	}); err != nil {
		l.Logger.Errorw("failed to publish product deleted event", logx.Field("id", id))
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}
