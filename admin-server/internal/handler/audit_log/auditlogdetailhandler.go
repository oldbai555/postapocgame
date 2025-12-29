// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit_log

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/audit_log"
	"postapocgame/admin-server/internal/svc"
)

func AuditLogDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从路径参数获取ID
		// URL格式: /api/v1/audit-logs/:id
		path := r.URL.Path
		parts := strings.Split(path, "/")
		var idStr string
		for i, part := range parts {
			if part == "audit-logs" && i+1 < len(parts) {
				idStr = parts[i+1]
				break
			}
		}

		// 将字符串 ID 转换为 uint64
		var id uint64
		if idStr != "" {
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
		}

		l := audit_log.NewAuditLogDetailLogic(r.Context(), svcCtx)
		resp, err := l.AuditLogDetail(id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
