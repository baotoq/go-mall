package cart

import (
	"cart/ent"
	"cart/internal/types"
)

func toResponse(item *ent.CartItem) *types.CartItemResponse {
	if item == nil {
		return nil
	}
	resp := &types.CartItemResponse{
		Id:        item.ID.String(),
		ProductId: item.ProductID.String(),
		Quantity:  item.Quantity,
		CreatedAt: item.CreatedAt.Unix(),
	}
	if item.UpdatedAt != nil {
		resp.UpdatedAt = item.UpdatedAt.Unix()
	} else {
		resp.UpdatedAt = item.CreatedAt.Unix()
	}
	return resp
}
