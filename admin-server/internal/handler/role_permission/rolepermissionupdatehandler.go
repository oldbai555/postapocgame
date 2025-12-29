// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role_permission

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	rolepermission "postapocgame/admin-server/internal/logic/role_permission"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/audit"
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
			// 记录审计日志：权限分配（角色-权限关联）
			audit.RecordAuditLog(svcCtx, r.Context(), r, audit.AuditTypePermissionAssign, audit.AuditObjectRolePermission, map[string]interface{}{
				"roleId":        req.RoleId,
				"permissionIds": req.PermissionIds,
			})
			httpx.Ok(w)
		}
	}
}
