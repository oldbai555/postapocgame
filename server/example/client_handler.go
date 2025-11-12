package main

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
)

// ClientHandler Actor消息处理器
type ClientHandler struct {
	*actor.BaseActorHandler
}

func NewClientHandler() *ClientHandler {
	h := &ClientHandler{
		BaseActorHandler: actor.NewBaseActorHandler(),
	}
	h.OnInit()

	// 注册消息处理器
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CError), h.handleError)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CRoleList), h.handleRoleList)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CEnterScene), h.handleEnterScene)
	h.RegisterMessageHandler(uint16(protocol.S2CProtocol_S2CReconnectKey), h.handleReconnectKey)

	return h
}

// handleError 处理错误消息
func (h *ClientHandler) handleError(msg actor.IActorMessage) {
	var errResp protocol.ErrorData
	if err := tool.JsonUnmarshal(msg.GetData(), &errResp); err == nil {
		log.Infof("\n⚠️ 服务器错误: %s\n> ", errResp.Msg)
	}
}

// handleRoleList 处理角色列表
func (h *ClientHandler) handleRoleList(msg actor.IActorMessage) {
	// 从Actor获取客户端引用
	actorCtx, ok := msg.GetContext().Value("actorCtx").(actor.IActorContext)
	if !ok {
		log.Errorf("无法获取ActorContext")
		return
	}

	client, ok := actorCtx.GetData().(*GameClient)
	if !ok {
		log.Errorf("无法获取GameClient")
		return
	}

	var resp protocol.S2CRoleListReq
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析角色列表失败: %v", err)
		return
	}

	log.Infof("\n[%s] 📜 角色列表:\n", client.GetPlayerID())
	for i, role := range resp.RoleList {
		log.Infof("  [%d] 角色ID: %d, 名字: %s, 职业: %d, 等级: %d\n",
			i+1, role.RoleId, role.RoleName, role.Job, role.Level)
	}

	// 自动选择第一个角色进入游戏
	if len(resp.RoleList) > 0 {
		selectedRole := resp.RoleList[0]
		log.Infof("[%s] 🎮 自动进入游戏: RoleID=%d\n", client.GetPlayerID(), selectedRole.RoleId)

		req := protocol.C2SEnterGameReq{RoleId: selectedRole.RoleId}
		reqData, err := tool.JsonMarshal(&req)
		if err != nil {
			log.Errorf("序列化失败: %v", err)
			return
		}

		if err := client.SendMessage(uint16(protocol.C2SProtocol_C2SEnterGame), reqData); err != nil {
			log.Errorf("发送进入游戏消息失败: %v", err)
		}
	}
}

// handleEnterScene 处理进入场景
func (h *ClientHandler) handleEnterScene(msg actor.IActorMessage) {
	actorCtx, ok := msg.GetContext().Value("actorCtx").(actor.IActorContext)
	if !ok {
		log.Errorf("无法获取ActorContext")
		return
	}

	client, ok := actorCtx.GetData().(*GameClient)
	if !ok {
		log.Errorf("无法获取GameClient")
		return
	}

	var resp protocol.S2CEnterSceneReq
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析进入场景响应失败: %v", err)
		return
	}
	entityData := resp.EntityData
	log.Infof("\n[%s] 🌍 成功进入场景 %d\n", client.GetPlayerID(), entityData.SceneId)
	log.Infof("  位置: (%v, %v)\n", entityData.PosX, entityData.PosY)
	log.Infof("  角色: %s (Lv.%d)\n", entityData.ShowName, entityData.Level)
}

func (h *ClientHandler) handleReconnectKey(msg actor.IActorMessage) {
	var resp protocol.S2CLoginSuccessReq
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("LoginSuccessResponse: %v", err)
		return
	}
	log.Infof("ReconnectKey:%s, roleInfo:%+v", resp.ReconnectKey, resp.RoleData)
}

// NetworkMessageHandler 网络消息处理器（转发到Actor）
type NetworkMessageHandler struct {
	client *GameClient
}

func (h *NetworkMessageHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
	// 解码客户端消息
	clientMsg, err := h.client.codec.DecodeClientMessage(msg.Payload)
	if err != nil {
		log.Errorf("[%s] ❌ 解析消息失败: %v\n", h.client.playerID, err)
		return customerr.Wrap(err)
	}

	// 创建Actor消息
	actorMsg := actor.NewBaseMessage(
		context.WithValue(ctx, "actorCtx", h.client.actorCtx),
		clientMsg.MsgId,
		clientMsg.Data,
	)

	// 发送到Actor处理
	if err := h.client.actorMgr.SendMessageAsync(h.client.playerID, actorMsg); err != nil {
		log.Errorf("[%s] 发送消息到Actor失败: %v", h.client.playerID, err)
		return customerr.Wrap(err)
	}

	return nil
}
