package cart

import (
	"context"
	"errors"
	"fmt"

	"cart/ent"
	"cart/ent/cartitem"
	"cart/internal/svc"
	"cart/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddToCartLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	sessionID string
}

// Add item to cart
func NewAddToCartLogic(ctx context.Context, svcCtx *svc.ServiceContext, sessionID string) *AddToCartLogic {
	return &AddToCartLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		sessionID: sessionID,
	}
}

func (l *AddToCartLogic) AddToCart(req *types.CartItemRequest) (*types.CartItemResponse, error) {
	if l.sessionID == "" {
		return nil, errors.New("missing session")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	pid, err := uuid.Parse(req.ProductId)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	existing, err := l.svcCtx.Db.CartItem.Query().
		Where(cartitem.SessionID(l.sessionID), cartitem.ProductID(pid)).
		First(l.ctx)

	if err == nil {
		updated, err := existing.Update().SetQuantity(existing.Quantity + req.Quantity).Save(l.ctx)
		if err != nil {
			return nil, fmt.Errorf("update cart item: %w", err)
		}
		return toResponse(updated), nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("query cart item: %w", err)
	}

	created, err := l.svcCtx.Db.CartItem.Create().
		SetSessionID(l.sessionID).
		SetProductID(pid).
		SetQuantity(req.Quantity).
		Save(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("create cart item: %w", err)
	}
	return toResponse(created), nil
}
