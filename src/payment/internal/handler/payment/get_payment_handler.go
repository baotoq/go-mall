// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package payment

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"payment/internal/logic/payment"
	"payment/internal/svc"
	"payment/internal/types"
)

// Get payment status
func GetPaymentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPaymentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := payment.NewGetPaymentLogic(r.Context(), svcCtx)
		resp, err := l.GetPayment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
