// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/operation_log"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func OperationLogDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OperationLogDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := operation_log.NewOperationLogDetailLogic(r.Context(), svcCtx)
		resp, err := l.OperationLogDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
