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

func LoginLogListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginLogListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 检查 URL 参数中是否包含 status，如果没有则设置为 -1 表示不筛选
		// 这样可以区分"未传递"（-1）和"传递了0"（0表示失败）
		if r.URL.Query().Get("status") == "" {
			req.Status = -1
		}

		l := login_log.NewLoginLogListLogic(r.Context(), svcCtx)
		resp, err := l.LoginLogList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
