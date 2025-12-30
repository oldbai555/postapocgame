// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login_log

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/login_log"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func LoginLogDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginLogDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := login_log.NewLoginLogDetailLogic(r.Context(), svcCtx)
		resp, err := l.LoginLogDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
