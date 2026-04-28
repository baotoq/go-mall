// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package catalog

import (
	"net/http"

	"catalog/internal/logic/catalog"
	"catalog/internal/svc"
	"catalog/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get category by ID
func GetCategoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCategoryRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := catalog.NewGetCategoryLogic(r.Context(), svcCtx)
		resp, err := l.GetCategory(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
