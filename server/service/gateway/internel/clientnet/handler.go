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
	"postapocgame/server/pkg/routine"
	"strings"
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
	routine.GoV2(func() error {
		h.handleSend(conn, session)
		return nil
	})

	return session
}

// handleSend 处理发送（优雅关闭版本）
func (h *ClientHandler) handleSend(conn network.IConnection, session *Session) {
	routine.Run(func() {
		defer func() {
			// 清理工作
			h.mu.Lock()
			delete(h.Sessions, conn)
			h.mu.Unlock()
			// 关闭会话
			if err := h.SessionMgr.CloseSession(session.Id); err != nil {
				log.Errorf("CloseSession failed for session %s, err:%v", session.Id, err)
			}
			log.Infof("handleSend goroutine exited for session: %s", session.Id)
		}()

		// 连续发送失败计数器
		consecutiveFailures := 0
		maxConsecutiveFailures := 3 // 连续失败3次后才认为连接已断开

		for {
			select {
			case data, ok := <-session.SendChan:
				// channel 被关闭，退出
				if !ok {
					log.Infof("SendChan closed for session: %s", session.Id)
					return
				}

				message := network.GetMessage()
				message.Type = network.MsgTypeClient
				message.Payload = data

				// 尝试发送消息
				if err := conn.SendMessage(message); err != nil {
					consecutiveFailures++
					log.Warnf("Send message failed (attempt %d/%d) for session %s: %v",
						consecutiveFailures, maxConsecutiveFailures, session.Id, err)

					// 检查是否是致命错误（连接已断开）
					if isConnectionError(err) {
						log.Errorf("Connection error detected for session %s: %v", session.Id, err)
						network.PutMessage(message)
						return
					}

					// 连续失败次数过多，认为连接已不可用
					if consecutiveFailures >= maxConsecutiveFailures {
						log.Errorf("Max consecutive failures reached for session %s, closing connection", session.Id)
						network.PutMessage(message)
						return
					}

					// 发送失败但还没达到阈值，继续尝试下一条消息
					network.PutMessage(message)
					continue
				}

				// 发送成功，重置失败计数器
				consecutiveFailures = 0
				network.PutMessage(message)

			case <-session.stopChan: // 🔧 新增：会话级别的停止信号
				log.Infof("Session stop signal received for session: %s", session.Id)
				return
			}
		}
	})
}

// isConnectionError 判断是否是连接错误（不可恢复的错误）
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 常见的连接断开错误
	connectionErrors := []string{
		"connection reset by peer",
		"broken pipe",
		"use of closed network connection",
		"connection refused",
		"EOF",
		"i/o timeout",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(strings.ToLower(errStr), connErr) {
			return true
		}
	}

	return false
}
