package cart

import (
	"errors"
	"net/http"

	"cart/internal/logic/cart"
	"cart/internal/svc"
	"cart/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Add item to cart
func AddToCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CartItemRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		sessionID := r.Header.Get("X-Session-Id")
		if sessionID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing X-Session-Id header"))
			return
		}

		l := cart.NewAddToCartLogic(r.Context(), svcCtx, sessionID)
		resp, err := l.AddToCart(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
