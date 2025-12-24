// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	userrole "postapocgame/admin-server/internal/logic/user_role"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func UserRoleListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserRoleListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userrole.NewUserRoleListLogic(r.Context(), svcCtx)
		resp, err := l.UserRoleList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
