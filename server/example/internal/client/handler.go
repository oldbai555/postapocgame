package client

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

// ClientHandler 负责将网络消息分发到 Core
type ClientHandler struct {
	*actor.BaseActorHandler
}

func NewClientHandler() *ClientHandler {
	h := &ClientHandler{
		BaseActorHandler: actor.NewBaseActorHandler("client_handler"),
	}
	h.OnInit()

	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRegister), h.handleRegisterResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLogin), h.handleLoginResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CError), h.handleError)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRoleList), h.handleRoleList)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CCreateRole), h.handleCreateRoleResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLoginRole), h.handleLoginRole)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CTimeSync), h.handleTimeSync)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEnterScene), h.handleEnterScene)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CStartMove), h.handleStartMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CStopMove), h.handleEndMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityAppear), h.handleEntityAppear)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityDisappear), h.handleEntityDisappear)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CUseSkill), h.handleUseSkill)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CSkillDamage), h.handleSkillDamage)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLevelData), h.handleLevelData)

	return h
}

func (h *ClientHandler) getClient(msg actor.IActorMessage) (*Core, bool) {
	actorCtx, ok := msg.GetContext().Value("actorCtx").(actor.IActorContext)
	if !ok {
		return nil, false
	}
	core, ok := actorCtx.GetData("gameClient").(*Core)
	return core, ok
}

func (h *ClientHandler) handleError(msg actor.IActorMessage) {
	var errResp protocol.ErrorData
	if err := proto.Unmarshal(msg.GetData(), &errResp); err == nil {
		log.Infof("\n⚠️ 服务器错误: %s\n> ", errResp.Msg)
	}
}

func (h *ClientHandler) handleRegisterResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CRegisterReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	core.OnRegisterResult(&resp)
}

func (h *ClientHandler) handleLoginResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CLoginReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	core.OnLoginResult(&resp)
}

func (h *ClientHandler) handleRoleList(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}

	var resp protocol.S2CRoleListReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析角色列表失败: %v", err)
		return
	}

	log.Infof("\n[%s] 📜 角色列表:\n", core.GetPlayerID())
	for i, role := range resp.RoleList {
		log.Infof("  [%d] 角色ID: %d, 名字: %s, 职业: %d, 等级: %d\n",
			i+1, role.RoleId, role.RoleName, role.Job, role.Level)
	}
	core.OnRoleList(&resp)
}

func (h *ClientHandler) handleCreateRoleResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CCreateRoleReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	core.OnCreateRoleResult(&resp)
}

func (h *ClientHandler) handleLoginRole(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CLoginRoleReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 LoginRole 失败: %v", err)
		return
	}
	core.OnLoginRole(&resp)
}

func (h *ClientHandler) handleEnterScene(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}

	var resp protocol.S2CEnterSceneReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析进入场景响应失败: %v", err)
		return
	}
	entityData := resp.EntityData
	log.Infof("\n[%s] 🌍 成功进入场景 %d\n", core.GetPlayerID(), entityData.SceneId)
	log.Infof("  位置: (%v, %v) HP=%d MP=%d\n",
		entityData.PosX,
		entityData.PosY,
		attrValueOrZero(entityData.Attrs, attrdef.HP),
		attrValueOrZero(entityData.Attrs, attrdef.MP),
	)
	log.Infof("  角色: %s (Lv.%d)\n", entityData.ShowName, entityData.Level)
	core.OnEnterScene(&resp)
}

func (h *ClientHandler) handleStartMove(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CStartMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 StartMove 失败: %v", err)
		return
	}
	core.OnStartMove(&resp)
}

func (h *ClientHandler) handleEndMove(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEndMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EndMove 失败: %v", err)
		return
	}
	core.OnEndMove(&resp)
}

func (h *ClientHandler) handleUseSkill(msg actor.IActorMessage) {
	// 当前 S2CUseSkill 为空载荷，仅占位
	_ = msg
}

func (h *ClientHandler) handleSkillDamage(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CSkillDamageReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 SkillDamageResult 失败: %v", err)
		return
	}
	core.OnSkillDamageResult(&resp)
}

func (h *ClientHandler) handleTimeSync(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CTimeSyncReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 TimeSync 失败: %v", err)
		return
	}
	core.OnTimeSync(&resp)
}

func (h *ClientHandler) handleEntityAppear(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityAppearReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityAppear 失败: %v", err)
		return
	}
	core.OnEntityAppear(&resp)
}

func (h *ClientHandler) handleEntityDisappear(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityDisappearReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityDisappear 失败: %v", err)
		return
	}
	core.OnEntityDisappear(&resp)
}

func (h *ClientHandler) handleLevelData(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CLevelDataReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 LevelData 失败: %v", err)
		return
	}
	core.OnLevelData(&resp)
}

// NetworkMessageHandler 负责将网络消息送入 Actor 管理器
type NetworkMessageHandler struct {
	client *Core
}

func (h *NetworkMessageHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	clientMsg, err := h.client.codec.DecodeClientMessage(msg.Payload)
	if err != nil {
		log.Errorf("[%s] ❌ 解析消息失败: %v\n", h.client.GetPlayerID(), err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(
		context.WithValue(ctx, "actorCtx", h.client.actorCtx),
		clientMsg.MsgId,
		clientMsg.Data,
	)

	if err := h.client.actorMgr.SendMessageAsync(h.client.GetPlayerID(), actorMsg); err != nil {
		log.Errorf("[%s] 发送消息到Actor失败: %v", h.client.GetPlayerID(), err)
		return customerr.Wrap(err)
	}

	return nil
}
