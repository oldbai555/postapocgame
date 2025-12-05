package router

import (
	"context"
	"fmt"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/gateway"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// ProtocolRouterController 负责 C2S 协议路由
type ProtocolRouterController struct {
	networkGateway gateway.NetworkGateway
	sessionGateway gateway.SessionGateway
}

// NewProtocolRouterController 创建协议路由控制器
func NewProtocolRouterController() *ProtocolRouterController {
	return &ProtocolRouterController{
		networkGateway: deps.NetworkGateway(),
		sessionGateway: deps.SessionGateway(),
	}
}

// HandleDoNetworkMsg 处理客户端消息（原 handleDoNetWorkMsg）
func (c *ProtocolRouterController) HandleDoNetworkMsg(message actor.IActorMessage) {
	ctx := message.GetContext()
	sessionID, ok := ctx.Value(gshare.ContextKeySession).(string)
	if !ok || sessionID == "" {
		return
	}

	session := c.sessionGateway.GetSession(sessionID)
	if session == nil {
		return
	}

	clientMsg, err := network.DefaultCodec().DecodeClientMessage(message.GetData())
	if err != nil {
		log.Errorf("decode client message failed: session=%s, err=%v", sessionID, err)
		return
	}

	baseCtx := context.WithValue(context.Background(), gshare.ContextKeySession, sessionID)

	if handler := getProtocolHandler(clientMsg.MsgId); handler != nil {
		roleCtx := c.withPlayerRoleContext(baseCtx, session.GetRoleId())
		if err := handler(roleCtx, clientMsg); err != nil {
			log.Errorf("handle client protocol failed: proto=%d, session=%s, err=%v", clientMsg.MsgId, sessionID, err)
			c.sendError(sessionID, err.Error())
		}
		return
	}

	// 未注册的协议，返回错误
	log.Warnf("protocol not supported: proto=%d, session=%s", clientMsg.MsgId, sessionID)
	c.sendError(sessionID, fmt.Sprintf("protocol %d not supported", clientMsg.MsgId))
}

func (c *ProtocolRouterController) withPlayerRoleContext(ctx context.Context, roleID uint64) context.Context {
	if roleID == 0 {
		return ctx
	}
	playerRole := manager.GetPlayerRole(roleID)
	if playerRole == nil {
		return ctx
	}
	return playerRole.WithContext(ctx)
}

func (c *ProtocolRouterController) sendError(sessionID, message string) {
	_ = c.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
