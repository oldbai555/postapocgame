// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/operation_log"
	"postapocgame/admin-server/internal/svc"
)

func OperationLogDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从路径参数获取ID
		// URL格式: /api/v1/operation-logs/:id
		path := r.URL.Path
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "operation-logs" && i+1 < len(parts) {
				id = parts[i+1]
				break
			}
		}
		l := operation_log.NewOperationLogDetailLogic(r.Context(), svcCtx)
		resp, err := l.OperationLogDetail(id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
