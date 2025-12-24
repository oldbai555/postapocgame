// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission_menu

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	permissionmenu "postapocgame/admin-server/internal/logic/permission_menu"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func PermissionMenuUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PermissionMenuUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := permissionmenu.NewPermissionMenuUpdateLogic(r.Context(), svcCtx)
		err := l.PermissionMenuUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
