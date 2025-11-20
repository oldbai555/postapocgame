package gatewaylink

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"sync"
)

// NetworkHandler 网络消息处理器
type NetworkHandler struct {
	sessions   map[string]*argsdef.SessionInfo
	sessionsMu sync.RWMutex
	codec      *network.Codec
}

func DefaultNetworkHandler() *NetworkHandler {
	if singleSrv == nil {
		once.Do(func() {
			singleSrv = &NetworkHandler{
				sessions: make(map[string]*argsdef.SessionInfo),
				codec:    network.DefaultCodec(),
			}
		})
	}
	return singleSrv
}

// HandleMessage 处理网络消息
func (h *NetworkHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	// 设置 Gateway 连接到全局发送器
	GetMessageSender().SetConn(conn)

	switch msg.Type {
	case network.MsgTypeSessionEvent:
		return h.handleSessionEvent(msg)
	case network.MsgTypeClient:
		return h.handleClientMsg(ctx, msg)
	case network.MsgTypeHeartbeat:
		return nil
	default:
		log.Warnf("unknown message type: %d", msg.Type)
		return nil
	}
}

// handleSessionEvent 处理会话事件
func (h *NetworkHandler) handleSessionEvent(msg *network.Message) error {
	event, err := h.codec.DecodeSessionEvent(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Infof("SessionEvent: Type=%d, SessionId=%s, UserId=%s", event.EventType, event.SessionId, event.UserId)

	switch event.EventType {
	case network.SessionEventNew:
		return h.handleSessionNew(event)
	case network.SessionEventClose:
		return h.handleSessionClose(event)
	}

	return nil
}

// handleSessionNew 处理新会话
func (h *NetworkHandler) handleSessionNew(event *network.SessionEvent) error {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	h.sessions[event.SessionId] = &argsdef.SessionInfo{
		SessionId: event.SessionId,
		CreatedAt: servertime.Now().Unix(),
	}

	log.Infof("New session created: %s", event.SessionId)
	return nil
}

// handleSessionClose 处理会话关闭
func (h *NetworkHandler) handleSessionClose(event *network.SessionEvent) error {
	h.sessionsMu.Lock()
	delete(h.sessions, event.SessionId)
	h.sessionsMu.Unlock()

	// 移除玩家Actor
	err := gshare.RemoveActor(event.SessionId)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Infof("Session closed: %s", event.SessionId)
	return nil
}

// handleClientMsg 处理客户端消息
func (h *NetworkHandler) handleClientMsg(ctx context.Context, msg *network.Message) error {
	// 解码转发消息
	fwdMsg, err := h.codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 解码客户端消息
	clientMsg, err := h.codec.DecodeClientMessage(fwdMsg.Payload)
	if err != nil {
		return customerr.Wrap(err)
	}

	log.Debugf("ClientMsg: SessionId=%s, MsgId=%d, DataLen=%d", fwdMsg.SessionId, clientMsg.MsgId, len(clientMsg.Data))

	newCtx := context.WithValue(ctx, gshare.ContextKeySession, fwdMsg.SessionId)
	message := actor.NewBaseMessage(newCtx, gshare.DoNetWorkMsg, fwdMsg.Payload)

	// 发送到Actor系统处理
	if err := gshare.SendMessageAsync(fwdMsg.SessionId, message); err != nil {
		return customerr.Wrap(err)
	}

	return nil
}

// GetSession 获取会话
func (h *NetworkHandler) GetSession(sessionId string) iface.ISession {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	return h.sessions[sessionId]
}
