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
	h.RegisterMessageHandler(protocol.S2C_Error, h.handleError)
	h.RegisterMessageHandler(protocol.S2C_RoleList, h.handleRoleList)
	h.RegisterMessageHandler(protocol.S2C_EnterScene, h.handleEnterScene)
	h.RegisterMessageHandler(protocol.S2C_ReconnectKey, h.handleReconnectKey)

	return h
}

// handleError 处理错误消息
func (h *ClientHandler) handleError(msg actor.IActorMessage) {
	var errResp protocol.ErrorResponse
	if err := tool.JsonUnmarshal(msg.GetData(), &errResp); err == nil {
		log.Infof("\n⚠️ 服务器错误: %s\n> ", errResp.ErrMsg)
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

	var resp protocol.RoleListResponse
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析角色列表失败: %v", err)
		return
	}

	log.Infof("\n[%s] 📜 角色列表:\n", client.GetPlayerID())
	for i, role := range resp.Roles {
		log.Infof("  [%d] 角色ID: %d, 名字: %s, 职业: %d, 等级: %d\n",
			i+1, role.RoleId, role.Name, role.Job, role.Level)
	}

	// 自动选择第一个角色进入游戏
	if len(resp.Roles) > 0 {
		selectedRole := resp.Roles[0]
		log.Infof("[%s] 🎮 自动进入游戏: RoleID=%d\n", client.GetPlayerID(), selectedRole.RoleId)

		req := protocol.SelectRoleRequest{RoleId: selectedRole.RoleId}
		reqData, err := tool.JsonMarshal(req)
		if err != nil {
			log.Errorf("序列化失败: %v", err)
			return
		}

		if err := client.SendMessage(protocol.C2S_EnterGame, reqData); err != nil {
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

	var resp protocol.EnterSceneResponse
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("解析进入场景响应失败: %v", err)
		return
	}

	log.Infof("\n[%s] 🌍 成功进入场景 %d\n", client.GetPlayerID(), resp.SceneId)
	log.Infof("  位置: (%v, %v)\n", resp.PosX, resp.PosY)
	if resp.RoleInfo != nil {
		log.Infof("  角色: %s (Lv.%d)\n", resp.RoleInfo.Name, resp.RoleInfo.Level)
	}
}

func (h *ClientHandler) handleReconnectKey(msg actor.IActorMessage) {
	var resp protocol.LoginSuccessResponse
	if err := tool.JsonUnmarshal(msg.GetData(), &resp); err != nil {
		log.Errorf("LoginSuccessResponse: %v", err)
		return
	}
	log.Infof("ReconnectKey:%s, roleInfo:%+v", resp.ReconnectKey, resp.RoleInfo)
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
