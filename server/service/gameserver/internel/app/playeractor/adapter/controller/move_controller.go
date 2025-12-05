package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/gshare"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// MoveController 负责将客户端移动相关协议转发给 DungeonActor
// 说明：所有真实的移动校验与广播逻辑仍在 DungeonActor 的 MoveSys 中执行。
type MoveController struct{}

// NewMoveController 创建移动控制器
func NewMoveController() *MoveController {
	return &MoveController{}
}

// HandleStartMove 处理 C2SStartMove 请求
func (c *MoveController) HandleStartMove(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SStartMove payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdStartMove), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}

// HandleUpdateMove 处理 C2SUpdateMove 请求
func (c *MoveController) HandleUpdateMove(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SUpdateMove payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUpdateMove), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}

// HandleEndMove 处理 C2SEndMove 请求
func (c *MoveController) HandleEndMove(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SEndMove payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdEndMove), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}

// HandleChangeScene 处理 C2SChangeScene 请求
func (c *MoveController) HandleChangeScene(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	if len(msg.Data) == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "empty C2SChangeScene payload")
	}

	ctxWithSession := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
	actorMsg := actor.NewBaseMessage(ctxWithSession, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdChangeScene), msg.Data)
	return gshare.SendDungeonMessageAsync("global", actorMsg)
}
