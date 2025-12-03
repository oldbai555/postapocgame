package publicactor

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/core/gshare"
)

// 在线状态管理相关逻辑

// RegisterOnline 注册在线状态
func (pr *PublicRole) RegisterOnline(roleId uint64, sessionId string) {
	pr.onlineMap.Store(roleId, sessionId)
	log.Debugf("Role %d registered online with session %s", roleId, sessionId)
}

// UnregisterOnline 注销在线状态
func (pr *PublicRole) UnregisterOnline(roleId uint64) {
	pr.onlineMap.Delete(roleId)
	log.Debugf("Role %d unregistered online", roleId)
}

// GetSessionId 获取角色的SessionId
func (pr *PublicRole) GetSessionId(roleId uint64) (string, bool) {
	value, ok := pr.onlineMap.Load(roleId)
	if !ok {
		return "", false
	}
	sessionId, ok := value.(string)
	return sessionId, ok
}

// IsOnline 检查角色是否在线
func (pr *PublicRole) IsOnline(roleId uint64) bool {
	_, ok := pr.onlineMap.Load(roleId)
	return ok
}

// GetAllOnlineSessionIds 获取所有在线的 SessionId 列表
func (pr *PublicRole) GetAllOnlineSessionIds() []string {
	var sessionIds []string
	pr.onlineMap.Range(func(key, value interface{}) bool {
		if sessionId, ok := value.(string); ok {
			sessionIds = append(sessionIds, sessionId)
		}
		return true
	})
	return sessionIds
}

// ===== 在线状态相关 handler 注册（无闭包捕获 PublicRole） =====

// RegisterOnlineHandlers 注册在线状态相关的消息处理器
func RegisterOnlineHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdRegisterOnline), handleRegisterOnlineMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdUnregisterOnline), handleUnregisterOnlineMsg)
}

// 在线状态 handler 适配为 actor.HandlerMessageFunc
var (
	handleRegisterOnlineMsg   = withPublicRole(handleRegisterOnline)
	handleUnregisterOnlineMsg = withPublicRole(handleUnregisterOnline)
)

// ===== 在线状态业务 handler（从 message_handler.go 迁移）=====

// handleRegisterOnline 处理注册在线状态
func handleRegisterOnline(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	registerMsg := &protocol.RegisterOnlineMsg{}
	if err := proto.Unmarshal(data, registerMsg); err != nil {
		log.Errorf("Failed to unmarshal RegisterOnlineMsg: %v", err)
		return
	}
	publicRole.RegisterOnline(registerMsg.RoleId, registerMsg.SessionId)

	// 推送离线消息
	offlineMessages := publicRole.GetOfflineMessages(registerMsg.RoleId)
	if len(offlineMessages) > 0 {
		log.Debugf("Pushing %d offline messages to role %d", len(offlineMessages), registerMsg.RoleId)
		for _, chatMsg := range offlineMessages {
			broadcastMsg := &protocol.ChatBroadcastMsg{
				ChatMsg: chatMsg,
			}
			broadcastData, err := proto.Marshal(broadcastMsg)
			if err != nil {
				log.Errorf("Failed to marshal offline message: %v", err)
				continue
			}
			if err := sendClientMessageViaPlayerActor(registerMsg.SessionId, uint16(protocol.S2CProtocol_S2CChatMessage), broadcastData); err != nil {
				logSendFailure(registerMsg.SessionId, uint16(protocol.S2CProtocol_S2CChatMessage), err)
			}
		}
	}
}

// handleUnregisterOnline 处理注销在线状态
func handleUnregisterOnline(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	unregisterMsg := &protocol.UnregisterOnlineMsg{}
	if err := proto.Unmarshal(data, unregisterMsg); err != nil {
		log.Errorf("Failed to unmarshal UnregisterOnlineMsg: %v", err)
		return
	}
	publicRole.UnregisterOnline(unregisterMsg.RoleId)
}
