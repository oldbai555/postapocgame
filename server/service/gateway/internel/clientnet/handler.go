/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package clientnet

import (
	"context"
	"fmt"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"sync"
)

type ClientHandler struct {
	SessionMgr  *SessionManager
	GsConnector IGameServerConnector
	Sessions    map[network.IConnection]*Session
	mu          sync.RWMutex
}

func NewClientHandler(gsConnector IGameServerConnector, sessionMgr *SessionManager) *ClientHandler {
	return &ClientHandler{
		SessionMgr:  sessionMgr,
		GsConnector: gsConnector,
		Sessions:    make(map[network.IConnection]*Session),
	}
}

// HandleMessage 处理消息（实现IMessageHandler接口）
func (h *ClientHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	// 只处理转发消息
	if msg.Type != network.MsgTypeClient {
		log.Warnf("Unexpected message type: %d", msg.Type)
		return nil
	}

	// 获取或创建会话
	session := h.getOrCreateSession(conn)
	if session == nil {
		return fmt.Errorf("failed to create session")
	}

	// 更新活跃时间
	h.SessionMgr.UpdateActivity(session.Id)

	return h.GsConnector.ForwardClientMsg(context.Background(), &network.ForwardMessage{
		SessionId: session.Id,
		Payload:   msg.Payload,
	})
}

// getOrCreateSession 获取或创建会话
func (h *ClientHandler) getOrCreateSession(conn network.IConnection) *Session {
	h.mu.RLock()
	if session, ok := h.Sessions[conn]; ok {
		h.mu.RUnlock()
		return session
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查
	if session, ok := h.Sessions[conn]; ok {
		return session
	}

	// 创建会话（需要适配器）
	adapter := &ConnectionAdapter{conn: conn}
	session, err := h.SessionMgr.CreateSession(adapter)
	if err != nil {
		log.Errorf("Create session failed: %v", err)
		return nil
	}

	h.Sessions[conn] = session

	// 启动发送协程
	go h.handleSend(conn, session)

	return session
}

// handleSend 处理发送
func (h *ClientHandler) handleSend(conn network.IConnection, session *Session) {
	defer func() {
		h.mu.Lock()
		delete(h.Sessions, conn)
		h.mu.Unlock()
		err := h.SessionMgr.CloseSession(session.Id)
		if err != nil {
			log.Errorf("handleSend failed, err:%v", err)
		}
	}()

	for data := range session.SendChan {
		message := network.GetMessage()
		message.Type = network.MsgTypeClient
		message.Payload = data
		if err := conn.SendMessage(message); err != nil {
			log.Errorf("Send message failed: %v", err)
		}
		network.PutMessage(message)
	}
}
