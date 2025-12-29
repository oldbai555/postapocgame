// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/config"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
	"postapocgame/admin-server/pkg/audit"
)

func ConfigUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConfigUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := config.NewConfigUpdateLogic(r.Context(), svcCtx)
		err := l.ConfigUpdate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			// 记录审计日志：配置修改
			audit.RecordAuditLog(svcCtx, r.Context(), r, audit.AuditTypeConfigModify, audit.AuditObjectConfig, map[string]interface{}{
				"id":          req.Id,
				"value":       req.Value,
				"description": req.Description,
			})
			httpx.Ok(w)
		}
	}
}
