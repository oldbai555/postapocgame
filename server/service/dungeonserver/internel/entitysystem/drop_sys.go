/**
 * @Author: zjj
 * @Date: 2025/11/25
 * @Desc:
**/

package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/clientprotocol"
	"postapocgame/server/service/dungeonserver/internel/devent"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

func handlePickupItem(picker iface.IEntity, msg *network.ClientMessage) error {
	var req protocol.C2SPickupItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
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
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查是否为掉落物实体
	dropItem, ok := dropEntity.(iface.IDrop)
	if !ok {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "不是掉落物",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查归属者
	if !dropItem.IsOwner(picker) {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "不是你的掉落物",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 检查是否为角色实体
	roleEntity, ok := picker.(iface.IRole)
	if !ok {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "只有角色可以拾取",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 通过异步RPC请求GameServer检查背包空间并添加物品
	roleId := roleEntity.GetId()
	sessionId := roleEntity.GetSessionId()

	// 发送异步RPC请求到GameServer（携带SessionId，交给Actor模型处理）
	rpcReq := &protocol.D2GAddItemReq{
		SessionId: sessionId,
		RoleId:    roleId,
		ItemId:    dropItem.GetItemId(),
		Count:     dropItem.GetCount(),
		ItemHdl:   req.ItemHdl,
	}

	reqData, err := proto.Marshal(rpcReq)
	if err != nil {
		log.Errorf("marshal D2GAddItemReq failed: %v", err)
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "系统错误",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 异步调用GameServer
	if err := gameserverlink.CallGameServer(context.Background(), sessionId, uint16(protocol.D2GRpcProtocol_D2GAddItem), reqData); err != nil {
		log.Errorf("call GameServer AddItem failed: %v", err)
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "系统错误",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 发送成功后立即移除掉落物，默认RPC调用成功
	if scene, ok := entityMgr.GetSceneByHandle(req.ItemHdl); ok && scene != nil {
		if err := scene.RemoveEntity(req.ItemHdl); err != nil {
			log.Warnf("Failed to remove drop item from scene: %v", err)
		}
		entityMgr.UnbindScene(req.ItemHdl)
	}
	entityMgr.Unregister(req.ItemHdl)

	resp := &protocol.S2CPickupItemResultReq{
		Success: true,
		Message: "拾取成功",
		ItemHdl: req.ItemHdl,
	}
	return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
}

func init() {
	devent.Subscribe(devent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SPickupItem), handlePickupItem)
	})
}
