// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role_permission

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	rolepermission "postapocgame/admin-server/internal/logic/role_permission"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func RolePermissionUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RolePermissionUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := rolepermission.NewRolePermissionUpdateLogic(r.Context(), svcCtx)
		err := l.RolePermissionUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
