// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/file"
	"postapocgame/admin-server/internal/svc"
)

func FileDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从路径参数获取文件ID
		// URL格式: /api/v1/files/:id/download
		path := r.URL.Path
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "files" && i+1 < len(parts) {
				id = parts[i+1]
				break
			}
		}
		l := file.NewFileDownloadLogic(r.Context(), svcCtx)
		err := l.FileDownload(w, r, id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
	}
}
