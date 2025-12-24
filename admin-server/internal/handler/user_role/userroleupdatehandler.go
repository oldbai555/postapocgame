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

func UserRoleUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserRoleUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userrole.NewUserRoleUpdateLogic(r.Context(), svcCtx)
		err := l.UserRoleUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
