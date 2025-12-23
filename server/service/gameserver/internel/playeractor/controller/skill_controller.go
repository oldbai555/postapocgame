package controller

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// SkillController 负责将客户端技能请求转发给 DungeonActor
type SkillController struct{}

// NewSkillController 创建技能控制器
func NewSkillController() *SkillController {
	return &SkillController{}
}

// HandleUseSkill 处理 C2SUseSkill 请求
func (c *SkillController) HandleUseSkill(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SUseSkill payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DAMUseSkill), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}
