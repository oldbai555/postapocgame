package publicactor

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// 聊天系统相关 handler 注册

// RegisterChatHandlers 注册聊天相关的消息处理器
func RegisterChatHandlers(facade gshare.IPublicActorFacade) {
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatWorld), handleChatWorldMsg)
	facade.RegisterHandler(uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatPrivate), handleChatPrivateMsg)
}

// 聊天 handler 适配
var (
	handleChatWorldMsg   = withPublicRole(handleChatWorld)
	handleChatPrivateMsg = withPublicRole(handleChatPrivate)
)

// ===== 聊天业务 handler（从 message_handler.go 迁移）=====

// handleChatWorld 处理世界聊天
func handleChatWorld(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	chatMsg := &protocol.ChatWorldMsg{}
	if err := proto.Unmarshal(data, chatMsg); err != nil {
		log.Errorf("Failed to unmarshal ChatWorldMsg: %v", err)
		return
	}

	// 验证发送者在线状态
	if !publicRole.IsOnline(chatMsg.SenderId) {
		log.Warnf("handleChatWorld: sender %d is not online", chatMsg.SenderId)
		return
	}

	// 构建聊天消息
	chatMessage := &protocol.ChatMessage{
		ChatType:   protocol.ChatType_ChatTypeWorld,
		SenderId:   chatMsg.SenderId,
		SenderName: chatMsg.SenderName,
		Content:    chatMsg.Content,
		Timestamp:  servertime.UnixMilli(),
	}

	// 广播给所有在线玩家
	sessionIds := publicRole.GetAllOnlineSessionIds()
	broadcastMsg := &protocol.ChatBroadcastMsg{
		ChatMsg: chatMessage,
	}
	broadcastData, err := proto.Marshal(broadcastMsg)
	if err != nil {
		log.Errorf("Failed to marshal ChatBroadcastMsg: %v", err)
		return
	}

	// 通过 gatewaylink 发送给所有在线玩家
	for _, sessionId := range sessionIds {
		err := gatewaylink.SendToSession(sessionId, uint16(protocol.S2CProtocol_S2CChatMessage), broadcastData)
		if err != nil {
			log.Warnf("Failed to send chat message to session %s: %v", sessionId, err)
		}
	}

	log.Debugf("handleChatWorld: broadcasted message from %d to %d players", chatMsg.SenderId, len(sessionIds))
}

// handleChatPrivate 处理私聊
func handleChatPrivate(ctx context.Context, msg actor.IActorMessage, publicRole *PublicRole) {
	data := msg.GetData()
	chatMsg := &protocol.ChatPrivateMsg{}
	if err := proto.Unmarshal(data, chatMsg); err != nil {
		log.Errorf("Failed to unmarshal ChatPrivateMsg: %v", err)
		return
	}

	// 验证发送者在线状态
	if !publicRole.IsOnline(chatMsg.SenderId) {
		log.Warnf("handleChatPrivate: sender %d is not online", chatMsg.SenderId)
		return
	}

	// 检查黑名单
	isBlocked, err := database.IsInBlacklist(chatMsg.SenderId, chatMsg.TargetId)
	if err == nil && isBlocked {
		log.Debugf("handleChatPrivate: sender %d is blocked by target %d", chatMsg.SenderId, chatMsg.TargetId)
		// 不发送消息，也不通知发送者（静默处理）
		return
	}

	// 查找目标玩家的 SessionId
	targetSessionId, ok := publicRole.GetSessionId(chatMsg.TargetId)
	if !ok {
		log.Debugf("handleChatPrivate: target %d is not online, storing offline message", chatMsg.TargetId)
		// 目标不在线，存储离线消息
		chatMessage := &protocol.ChatMessage{
			ChatType:   protocol.ChatType_ChatTypePrivate,
			SenderId:   chatMsg.SenderId,
			SenderName: chatMsg.SenderName,
			TargetId:   chatMsg.TargetId,
			Content:    chatMsg.Content,
			Timestamp:  servertime.UnixMilli(),
		}
		publicRole.AddOfflineMessage(chatMsg.TargetId, chatMessage)

		// 发送确认给发送者（消息已存储）
		senderSessionId, ok := publicRole.GetSessionId(chatMsg.SenderId)
		if ok {
			broadcastMsg := &protocol.ChatBroadcastMsg{
				ChatMsg: chatMessage,
			}
			broadcastData, err := proto.Marshal(broadcastMsg)
			if err == nil {
				gatewaylink.SendToSession(senderSessionId, uint16(protocol.S2CProtocol_S2CChatMessage), broadcastData)
			}
		}
		return
	}

	// 构建聊天消息
	chatMessage := &protocol.ChatMessage{
		ChatType:   protocol.ChatType_ChatTypePrivate,
		SenderId:   chatMsg.SenderId,
		SenderName: chatMsg.SenderName,
		TargetId:   chatMsg.TargetId,
		Content:    chatMsg.Content,
		Timestamp:  servertime.UnixMilli(),
	}

	// 发送给目标玩家
	broadcastMsg := &protocol.ChatBroadcastMsg{
		ChatMsg: chatMessage,
	}
	broadcastData, err := proto.Marshal(broadcastMsg)
	if err != nil {
		log.Errorf("Failed to marshal ChatBroadcastMsg: %v", err)
		return
	}

	err = gatewaylink.SendToSession(targetSessionId, uint16(protocol.S2CProtocol_S2CChatMessage), broadcastData)
	if err != nil {
		log.Warnf("Failed to send private chat message to session %s: %v", targetSessionId, err)
		return
	}

	// 同时发送给发送者（确认消息已发送）
	senderSessionId, ok := publicRole.GetSessionId(chatMsg.SenderId)
	if ok {
		err = gatewaylink.SendToSession(senderSessionId, uint16(protocol.S2CProtocol_S2CChatMessage), broadcastData)
		if err != nil {
			log.Warnf("Failed to send private chat confirmation to sender %s: %v", senderSessionId, err)
		}
	}

	log.Debugf("handleChatPrivate: forwarded message from %d to %d", chatMsg.SenderId, chatMsg.TargetId)
}
