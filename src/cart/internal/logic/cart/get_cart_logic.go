package cart

import (
	"context"
	"errors"
	"fmt"

	"cart/ent"
	"cart/ent/cartitem"
	"cart/internal/svc"
	"cart/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCartLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	sessionID string
}

// Get cart items
func NewGetCartLogic(ctx context.Context, svcCtx *svc.ServiceContext, sessionID string) *GetCartLogic {
	return &GetCartLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		sessionID: sessionID,
	}
}

func (l *GetCartLogic) GetCart() (*types.CartItemListResponse, error) {
	if l.sessionID == "" {
		return nil, errors.New("missing session")
	}

	items, err := l.svcCtx.Db.CartItem.Query().
		Where(cartitem.SessionID(l.sessionID)).
		Order(ent.Desc(cartitem.FieldCreatedAt)).
		All(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("query cart items: %w", err)
	}

	var resp []types.CartItemResponse
	for _, item := range items {
		resp = append(resp, *toResponse(item))
	}

	if resp == nil {
		resp = make([]types.CartItemResponse, 0)
	}

	return &types.CartItemListResponse{
		Items: resp,
		Total: int64(len(items)),
	}, nil
}
