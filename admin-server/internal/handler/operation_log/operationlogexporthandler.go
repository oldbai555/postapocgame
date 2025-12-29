// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package operation_log

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/operation_log"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func OperationLogExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OperationLogExportReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := operation_log.NewOperationLogExportLogic(r.Context(), svcCtx)
		err := l.OperationLogExport(w, r, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
		// 导出功能直接写入响应流，不需要返回 JSON
	}
}
