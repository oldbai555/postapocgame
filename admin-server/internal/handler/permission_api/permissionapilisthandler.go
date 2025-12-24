// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission_api

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	permissionapi "postapocgame/admin-server/internal/logic/permission_api"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func PermissionApiListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PermissionApiListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := permissionapi.NewPermissionApiListLogic(r.Context(), svcCtx)
		resp, err := l.PermissionApiList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
