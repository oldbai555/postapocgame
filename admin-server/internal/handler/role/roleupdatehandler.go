// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/role"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/audit"
)

func RoleUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewRoleUpdateLogic(r.Context(), svcCtx)
		err := l.RoleUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 记录审计日志：角色变更（更新角色）
			audit.RecordAuditLog(svcCtx, r.Context(), r, audit.AuditTypeRoleChange, audit.AuditObjectRole, map[string]interface{}{
				"action": "update",
				"id":     req.Id,
				"name":   req.Name,
			})
			httpx.Ok(w)
		}
	}
}
