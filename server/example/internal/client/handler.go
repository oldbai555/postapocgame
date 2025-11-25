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

	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRegisterResult), h.handleRegisterResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLoginResult), h.handleLoginResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CError), h.handleError)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRoleList), h.handleRoleList)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CCreateRoleResult), h.handleCreateRoleResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEnterScene), h.handleEnterScene)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CLoginSuccess), h.handleLoginSuccess)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CReconnectSuccess), h.handleReconnectSuccess)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityMove), h.handleEntityMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEntityStopMove), h.handleEntityStopMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CSkillCastResult), h.handleSkillCastResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CSkillDamageResult), h.handleSkillDamageResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CTimeSync), h.handleTimeSync)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CStartMove), h.handleStartMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEndMove), h.handleEndMove)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CBagData), h.handleBagData)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CUpdateBagData), h.handleBagUpdate)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CMoneyData), h.handleMoneyData)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CGMCommandResult), h.handleGMResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CUseItemResult), h.handleUseItemResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CPickupItemResult), h.handlePickupResult)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEnterDungeonResult), h.handleEnterDungeonResult)

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
	var resp protocol.S2CRegisterResultReq
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
	var resp protocol.S2CLoginResultReq
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
	var resp protocol.S2CCreateRoleResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		return
	}
	core.OnCreateRoleResult(&resp)
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
		attrValueOrZero(entityData.Attrs, attrdef.AttrHP),
		attrValueOrZero(entityData.Attrs, attrdef.AttrMP),
	)
	log.Infof("  角色: %s (Lv.%d)\n", entityData.ShowName, entityData.Level)
	core.OnEnterScene(&resp)
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
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityMove 失败: %v", err)
		return
	}
	core.OnEntityMove(&resp)
}

func (h *ClientHandler) handleEntityStopMove(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEntityStopMoveReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EntityStopMove 失败: %v", err)
		return
	}
	core.OnEntityStop(&resp)
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

func (h *ClientHandler) handleSkillCastResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CSkillCastResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 SkillCastResult 失败: %v", err)
		return
	}
	core.OnSkillCastResult(&resp)
}

func (h *ClientHandler) handleSkillDamageResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CSkillDamageResultReq
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

func (h *ClientHandler) handleBagData(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CBagDataReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 BagData 失败: %v", err)
		return
	}
	core.OnBagData(&resp)
}

func (h *ClientHandler) handleBagUpdate(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	core.OnBagUpdate()
}

func (h *ClientHandler) handleMoneyData(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CMoneyDataReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 MoneyData 失败: %v", err)
		return
	}
	core.OnMoneyData(&resp)
}

func (h *ClientHandler) handleGMResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CGMCommandResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 GMCommandResult 失败: %v", err)
		return
	}
	core.OnGMCommandResult(&resp)
}

func (h *ClientHandler) handleUseItemResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CUseItemResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 UseItemResult 失败: %v", err)
		return
	}
	core.OnUseItemResult(&resp)
}

func (h *ClientHandler) handlePickupResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CPickupItemResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 PickupItemResult 失败: %v", err)
		return
	}
	core.OnPickupItemResult(&resp)
}

func (h *ClientHandler) handleEnterDungeonResult(msg actor.IActorMessage) {
	core, ok := h.getClient(msg)
	if !ok {
		return
	}
	var resp protocol.S2CEnterDungeonResultReq
	if err := proto.Unmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析 EnterDungeonResult 失败: %v", err)
		return
	}
	core.OnEnterDungeonResult(&resp)
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
