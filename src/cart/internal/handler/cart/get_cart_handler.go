package cart

import (
	"errors"
	"net/http"

	"cart/internal/logic/cart"
	"cart/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get cart items
func GetCartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-Id")
		if sessionID == "" {
			httpx.ErrorCtx(r.Context(), w, errors.New("missing X-Session-Id header"))
			return
		}

		l := cart.NewGetCartLogic(r.Context(), svcCtx, sessionID)
		resp, err := l.GetCart()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
