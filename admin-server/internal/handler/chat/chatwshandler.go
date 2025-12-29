package chat

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"postapocgame/admin-server/internal/hub"
	"postapocgame/admin-server/internal/repository"
	"postapocgame/admin-server/internal/svc"
	"postapocgame/admin-server/pkg/errs"
	jwthelper "postapocgame/admin-server/pkg/jwt"
	"postapocgame/admin-server/pkg/response"

	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境应该限制）
		return true
	},
}

// ChatWSHandler WebSocket 连接处理器
func ChatWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从查询参数获取 token（WebSocket 无法使用 Authorization header）
		token := r.URL.Query().Get("token")
		if token == "" {
			// 尝试从 Authorization header 获取
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					token = parts[1]
				}
			}
		}

		if token == "" {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "未提供认证信息"))
			return
		}

		// 验证 token
		claims, err := jwthelper.ParseToken(token, svcCtx.Config.JWT.AccessSecret)
		if err != nil || claims.IsRefresh {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "访问令牌无效或已过期"))
			return
		}

		// 检查黑名单
		blackRepo := repository.NewTokenBlacklistRepository(svcCtx.Repository)
		blacklisted, err := blackRepo.IsBlacklisted(r.Context(), token)
		if err != nil {
			response.ErrorCtx(r.Context(), w, errs.Wrap(errs.CodeInternalError, "检查令牌黑名单失败", err))
			return
		}
		if blacklisted {
			response.ErrorCtx(r.Context(), w, errs.New(errs.CodeUnauthorized, "令牌已失效"))
			return
		}

		// 升级到 WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("WebSocket 升级失败: %v", err)
			return
		}

		// 获取房间ID（可选）
		roomID := r.URL.Query().Get("roomId")
		if roomID == "" {
			roomID = "default" // 默认房间
		}

		// 生成连接ID
		connectionID := uuid.New().String()

		// 创建客户端
		client := &hub.Client{
			Hub:          svcCtx.ChatHub,
			Conn:         conn,
			Send:         make(chan []byte, 256),
			UserID:       claims.UserID,
			Username:     claims.Username,
			RoomID:       roomID,
			ConnectionID: connectionID,
		}

		// 注册客户端
		client.Hub.Register() <- client

		// 启动读写协程
		go client.WritePump()
		go client.ReadPump()

		// 发送加入消息
		joinMsg := &hub.ChatMessage{
			Type:     "join",
			FromID:   claims.UserID,
			FromName: claims.Username,
			RoomID:   roomID,
			Content:  claims.Username + " 加入了聊天室",
		}
		client.Hub.BroadcastChatMessage(joinMsg)
	}
}
