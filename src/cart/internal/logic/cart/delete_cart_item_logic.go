package cart

import (
	"context"
	"errors"
	"fmt"

	"cart/ent/cartitem"
	"cart/internal/svc"
	"cart/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCartItemLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	sessionID string
}

// Delete item from cart
func NewDeleteCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext, sessionID string) *DeleteCartItemLogic {
	return &DeleteCartItemLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		sessionID: sessionID,
	}
}

func (l *DeleteCartItemLogic) DeleteCartItem(req *types.DeleteCartItemRequest) error {
	if l.sessionID == "" {
		return errors.New("missing session")
	}
	pid, err := uuid.Parse(req.ProductId)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	_, err = l.svcCtx.Db.CartItem.Delete().
		Where(cartitem.SessionID(l.sessionID), cartitem.ProductID(pid)).
		Exec(l.ctx)
	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}

	return nil
}
