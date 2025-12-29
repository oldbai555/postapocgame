// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login_log

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/login_log"
	"postapocgame/admin-server/internal/svc"
)

func LoginLogStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := login_log.NewLoginLogStatsLogic(r.Context(), svcCtx)
		resp, err := l.LoginLogStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
