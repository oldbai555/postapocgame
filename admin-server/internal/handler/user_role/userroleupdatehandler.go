// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	userrole "postapocgame/admin-server/internal/logic/user_role"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/audit"
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
			// 记录审计日志：权限分配（用户-角色关联）
			audit.RecordAuditLog(svcCtx, r.Context(), r, audit.AuditTypePermissionAssign, audit.AuditObjectUserRole, map[string]interface{}{
				"userId":  req.UserId,
				"roleIds": req.RoleIds,
			})
			httpx.Ok(w)
		}
	}
}
