// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notice

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/notice"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func NoticeCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NoticeCreateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := notice.NewNoticeCreateLogic(r.Context(), svcCtx)
		resp, err := l.NoticeCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
