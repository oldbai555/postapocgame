package clientprotocol

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entity"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func init() {
	Register(uint16(protocol.C2SProtocol_C2SPickupItem), handlePickupItem)
}

func handlePickupItem(picker iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SPickupItemReq
	if err := internal.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取掉落物实体
	entityMgr := entitymgr.GetEntityMgr()
	dropEntity, ok := entityMgr.GetByHdl(req.ItemHdl)
	if !ok || dropEntity == nil {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "掉落物不存在",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查是否为掉落物实体
	dropItem, ok := dropEntity.(*entity.DropItemEntity)
	if !ok {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "不是掉落物",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查归属者
	if !dropItem.IsOwner(picker) {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "不是你的掉落物",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查是否为角色实体
	roleEntity, ok := picker.(*entity.RoleEntity)
	if !ok {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "只有角色可以拾取",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 通过异步RPC请求GameServer检查背包空间并添加物品
	roleId := roleEntity.GetRoleId()
	sessionId := roleEntity.GetSessionId()

	// 发送异步RPC请求到GameServer（携带SessionId，交给Actor模型处理）
	rpcReq := &protocol.D2GAddItemReq{
		SessionId: sessionId,
		RoleId:    roleId,
		ItemId:    dropItem.GetItemId(),
		Count:     dropItem.GetCount(),
		ItemHdl:   req.ItemHdl, // 传递掉落物句柄，用于响应时移除
	}

	reqData, err := internal.Marshal(rpcReq)
	if err != nil {
		log.Errorf("marshal D2GAddItemReq failed: %v", err)
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "系统错误",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 异步调用GameServer（不等待响应，通过RPC响应消息异步处理）
	err = gameserverlink.CallGameServer(context.Background(), sessionId, uint16(protocol.D2GRpcProtocol_D2GAddItem), reqData)
	if err != nil {
		log.Errorf("call GameServer AddItem failed: %v", err)
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "系统错误",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 异步RPC已发送，等待GameServer响应（通过RPC响应消息处理）
	// 先返回处理中状态，实际结果通过RPC响应消息异步通知
	resp := &protocol.S2CPickupItemResultReq{
		Success: true,
		Message: "拾取中...",
		ItemHdl: req.ItemHdl,
	}
	return picker.SendJsonMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
}
