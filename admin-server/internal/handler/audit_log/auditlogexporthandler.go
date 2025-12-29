// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit_log

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/audit_log"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func AuditLogExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AuditLogExportReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := audit_log.NewAuditLogExportLogic(r.Context(), svcCtx)
		if err := l.AuditLogExport(w, r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
		// 导出逻辑直接写入响应流，不需要返回 JSON
	}
}
