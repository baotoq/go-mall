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

// List payments by idempotency key
func GetPaymentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPaymentsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := payment.NewGetPaymentsLogic(r.Context(), svcCtx)
		resp, err := l.GetPayments(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
