package router

import (
	"context"
	"fmt"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/adapter/gateway"
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ProtocolRouterController 负责 C2S 协议路由
type ProtocolRouterController struct {
	dungeonGateway interfaces.DungeonServerGateway
	networkGateway gateway.NetworkGateway
}

// NewProtocolRouterController 创建协议路由控制器
func NewProtocolRouterController() *ProtocolRouterController {
	container := di.GetContainer()
	return &ProtocolRouterController{
		dungeonGateway: container.DungeonServerGateway(),
		networkGateway: container.NetworkGateway(),
	}
}

// HandleDoNetworkMsg 处理客户端消息（原 handleDoNetWorkMsg）
func (c *ProtocolRouterController) HandleDoNetworkMsg(message actor.IActorMessage) {
	ctx := message.GetContext()
	sessionID, ok := ctx.Value(gshare.ContextKeySession).(string)
	if !ok || sessionID == "" {
		return
	}

	session := gatewaylink.GetSession(sessionID)
	if session == nil {
		return
	}

	clientMsg, err := network.DefaultCodec().DecodeClientMessage(message.GetData())
	if err != nil {
		log.Errorf("decode client message failed: session=%s, err=%v", sessionID, err)
		return
	}

	baseCtx := context.WithValue(context.Background(), gshare.ContextKeySession, sessionID)

	if handler := clientprotocol.GetFunc(clientMsg.MsgId); handler != nil {
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
