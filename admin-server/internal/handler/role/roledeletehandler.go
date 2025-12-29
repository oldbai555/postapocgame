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

func RoleDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RoleDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewRoleDeleteLogic(r.Context(), svcCtx)
		err := l.RoleDelete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 记录审计日志：数据删除（删除角色）
			audit.RecordAuditLog(svcCtx, r.Context(), r, audit.AuditTypeDataDelete, audit.AuditObjectRole, map[string]interface{}{
				"action": "delete",
				"id":     req.Id,
			})
			httpx.Ok(w)
		}
	}
}
