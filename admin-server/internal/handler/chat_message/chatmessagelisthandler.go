// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_message

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"postapocgame/admin-server/internal/logic/chat_message"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/internal/types"
)

func ChatMessageListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatMessageListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat_message.NewChatMessageListLogic(r.Context(), svcCtx)
		resp, err := l.ChatMessageList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
