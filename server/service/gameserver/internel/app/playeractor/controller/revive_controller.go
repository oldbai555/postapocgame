package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/app/playeractor/fuben"
	"postapocgame/server/service/gameserver/internel/app/playeractor/router"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// ReviveController 复活控制器（PlayerActor → DungeonActor）
type ReviveController struct{}

// NewReviveController 创建复活控制器
func NewReviveController() *ReviveController {
	return &ReviveController{}
}

// HandleRevive 处理 C2SRevive 请求，做系统开关检查后转发到 DungeonActor 的 handleRevive
func (c *ReviveController) HandleRevive(ctx context.Context, msg *network.ClientMessage) error {
	// 复活依赖副本/战斗环境，这里使用 FubenSys 的开启状态作为入口开关
	fubenSys := fuben.GetFubenSys(ctx)
	if fubenSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "副本系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SRevive payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdRevive), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		reviveController := NewReviveController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRevive), reviveController.HandleRevive)
	})
}
