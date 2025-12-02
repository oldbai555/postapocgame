package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/gshare"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"

	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// handleGMCommand 处理GM命令
func handleGMCommand(ctx context.Context, msg *network.ClientMessage) error {
	// 交给 GM 系统工具函数解析并执行
	sessionId, success, err := HandleGMCommand(ctx, msg)

	// 发送GM命令结果
	resp := &protocol.S2CGMCommandResultReq{
		Success: success,
		Message: "",
	}
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
	}

	if sessionId != "" {
		if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGMCommandResult), resp); sendErr != nil {
			log.Errorf("send GM command result failed: %v", sendErr)
			return sendErr
		}
	}

	gshare.InfofCtx(ctx, "GM command executed: SessionID=%s, Success=%v, Message=%s",
		sessionId, resp.Success, resp.Message)

	return nil
}

func init() {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysGM), func() iface.ISystem {
		return NewGMSystemAdapter()
	})

	gevent2.Subscribe(gevent2.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SGMCommand), handleGMCommand)
	})
}
