package main

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

// ClientHandler Actor消息处理器
type ClientHandler struct {
	*actor.BaseActorHandler
}

func NewClientHandler() *ClientHandler {
	h := &ClientHandler{
		BaseActorHandler: actor.NewBaseActorHandler("client_handler"),
	}
	h.OnInit()

	// 注册消息处理器
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRegisterResult), h.handleRegisterResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLoginResult), h.handleLoginResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CError), h.handleError)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRoleList), h.handleRoleList)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CCreateRoleResult), h.handleCreateRoleResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEnterScene), h.handleEnterScene)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLoginSuccess), h.handleReconnectSuccess)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CReconnectSuccess), h.handleReconnectSuccess)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityMove), h.handleEntityMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityStopMove), h.handleEntityStopMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CSkillCastResult), h.handleSkillCastResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CSkillDamageResult), h.handleSkillDamageResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CTimeSync), h.handleTimeSync)

	return h
}

func (h *ClientHandler) getClient(msg actor.IActorMessage) (*GameClient, bool) {
	actorCtx, ok := msg.GetContext().Value("actorCtx").(actor.IActorContext)
	if !ok {
		return nil, false
	}
	client, ok := actorCtx.GetData("gameClient").(*GameClient)
	return client, ok
}

// handleError 处理错误消息
func (h *ClientHandler) handleError(msg actor.IActorMessage) {
	var errResp protocol.ErrorData
	if err := proto.Unmarshal(msg.GetData(), &errResp); err == nil {
		log.Infof("\n⚠️ 服务器错误: %s\n> ", errResp.Msg)
	}
}

func (h *ClientHandler) handleRegisterResult(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CRegisterResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	client.OnRegisterResult(&resp)
}

func (h *ClientHandler) handleLoginResult(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CLoginResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	client.OnLoginResult(&resp)
}

// handleRoleList 处理角色列表
func (h *ClientHandler) handleRoleList(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}

	var resp protocol.S2CRoleListReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析角色列表失败: %v", err)
		return
	}

	log.Infof("\n[%s] 📜 角色列表:\n", client.GetPlayerID())
	for i, role := range resp.RoleList {
		log.Infof("  [%d] 角色ID: %d, 名字: %s, 职业: %d, 等级: %d\n",
			i+1, role.RoleId, role.RoleName, role.Job, role.Level)
	}
	client.OnRoleList(&resp)
}

func (h *ClientHandler) handleCreateRoleResult(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CCreateRoleResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	client.OnCreateRoleResult(&resp)
}

// handleEnterScene 处理进入场景
func (h *ClientHandler) handleEnterScene(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}

	var resp protocol.S2CEnterSceneReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析进入场景响应失败: %v", err)
		return
	}
	entityData := resp.EntityData
	log.Infof("\n[%s] 🌍 成功进入场景 %d\n", client.GetPlayerID(), entityData.SceneId)
	log.Infof("  位置: (%v, %v) HP=%d MP=%d\n",
		entityData.PosX,
		entityData.PosY,
		attrValueOrZero(entityData.Attrs, attrdef.AttrHP),
		attrValueOrZero(entityData.Attrs, attrdef.AttrMP),
	)
	log.Infof("  角色: %s (Lv.%d)\n", entityData.ShowName, entityData.Level)
	client.OnEnterScene(&resp)
}

func (h *ClientHandler) handleLoginSuccess(msg actor.IActorMessage) {
	var resp protocol.S2CLoginSuccessReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("S2CLoginSuccessReq: %v", err)
		return
	}
	log.Infof("ReconnectKey:%s, roleInfo:%+v", resp.ReconnectKey, resp.RoleData)
}
func (h *ClientHandler) handleReconnectSuccess(msg actor.IActorMessage) {
	var resp protocol.S2CReconnectSuccessReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("S2CReconnectSuccessReq: %v", err)
		return
	}
	log.Infof("ReconnectKey:%s, roleInfo:%+v", resp.ReconnectKey, resp.RoleData)
}

func (h *ClientHandler) handleEntityMove(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityMove 失败: %v", err)
		return
	}
	client.OnEntityMove(&resp)
}

func (h *ClientHandler) handleEntityStopMove(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityStopMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityStopMove 失败: %v", err)
		return
	}
	client.OnEntityStop(&resp)
}

func (h *ClientHandler) handleSkillCastResult(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CSkillCastResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 SkillCastResult 失败: %v", err)
		return
	}
	client.OnSkillCastResult(&resp)
}

func (h *ClientHandler) handleSkillDamageResult(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CSkillDamageResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 SkillDamageResult 失败: %v", err)
		return
	}
	client.OnSkillDamageResult(&resp)
}

func (h *ClientHandler) handleTimeSync(msg actor.IActorMessage) {
	client, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CTimeSyncReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 TimeSync 失败: %v", err)
		return
	}
	client.OnTimeSync(&resp)
}

// NetworkMessageHandler 网络消息处理器（转发到Actor）
type NetworkMessageHandler struct {
	client *GameClient
}

func (h *NetworkMessageHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	// 解码客户端消息
	clientMsg, err := h.client.codec.DecodeClientMessage(msg.Payload)
	if err != nil {
		log.Errorf("[%s] ❌ 解析消息失败: %v\n", h.client.GetPlayerID(), err)
		return customerr.Wrap(err)
	}

	// 创建Actor消息
	actorMsg := actor.NewBaseMessage(
		context.WithValue(ctx, "actorCtx", h.client.actorCtx),
		clientMsg.MsgId,
		clientMsg.Data,
	)

	// 发送到Actor处理
	if err := h.client.actorMgr.SendMessageAsync(h.client.GetPlayerID(), actorMsg); err != nil {
		log.Errorf("[%s] 发送消息到Actor失败: %v", h.client.GetPlayerID(), err)
		return customerr.Wrap(err)
	}

	return nil
}
