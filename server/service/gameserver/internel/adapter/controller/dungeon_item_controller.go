package controller

import (
	"context"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/gshare"
)

// DungeonItemController 掉落物拾取控制器（PlayerActor → DungeonActor）
type DungeonItemController struct{}

// NewDungeonItemController 创建拾取控制器
func NewDungeonItemController() *DungeonItemController {
	return &DungeonItemController{}
}

// HandlePickupItem 处理 C2SPickupItem 请求，转发到 DungeonActor 的 handlePickupItem
func (c *DungeonItemController) HandlePickupItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SPickupItem payload")
	}

	ctxWithSession := context.WithValue(ctx, "session", sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdPickupItem), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}
