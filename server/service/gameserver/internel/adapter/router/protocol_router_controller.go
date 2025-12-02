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

	if !c.dungeonGateway.IsDungeonProtocol(clientMsg.MsgId) {
		log.Warnf("protocol not supported: proto=%d, session=%s", clientMsg.MsgId, sessionID)
		c.sendError(sessionID, fmt.Sprintf("protocol %d not supported", clientMsg.MsgId))
		return
	}

	srvType, protocolType, ok := c.dungeonGateway.GetSrvTypeForProtocol(clientMsg.MsgId)
	if !ok {
		log.Errorf("protocol route not found: proto=%d", clientMsg.MsgId)
		c.sendError(sessionID, fmt.Sprintf("protocol %d route missing", clientMsg.MsgId))
		return
	}

	targetSrvType, err := c.resolveTargetSrvType(protocolType, srvType, session.GetRoleId())
	if err != nil {
		log.Errorf("resolve target srvType failed: session=%s, proto=%d, err=%v", sessionID, clientMsg.MsgId, err)
		c.sendError(sessionID, err.Error())
		return
	}

	if err := c.dungeonGateway.AsyncCall(baseCtx, targetSrvType, sessionID, 0, message.GetData()); err != nil {
		log.Errorf("forward to dungeon server failed: proto=%d, srvType=%d, session=%s, err=%v",
			clientMsg.MsgId, targetSrvType, sessionID, err)
		c.sendError(sessionID, fmt.Sprintf("forward to DungeonServer failed: %v", err))
		return
	}

	log.Debugf("forwarded protocol %d to DungeonServer srvType=%d session=%s", clientMsg.MsgId, targetSrvType, sessionID)
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

func (c *ProtocolRouterController) resolveTargetSrvType(protocolType interfaces.ProtocolType, defaultSrvType uint8, roleID uint64) (uint8, error) {
	if protocolType == interfaces.ProtocolTypeUnique {
		return defaultSrvType, nil
	}

	playerRole := manager.GetPlayerRole(roleID)
	if playerRole == nil {
		return 0, fmt.Errorf("player role not found")
	}

	targetSrvType := playerRole.GetDungeonSrvType()
	if targetSrvType == 0 {
		targetSrvType = defaultSrvType
	}
	return targetSrvType, nil
}

func (c *ProtocolRouterController) sendError(sessionID, message string) {
	_ = c.networkGateway.SendToSessionProto(sessionID, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
		Code: -1,
		Msg:  message,
	})
}
