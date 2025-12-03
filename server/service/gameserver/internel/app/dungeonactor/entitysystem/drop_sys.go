/**
 * @Author: zjj
 * @Date: 2025/11/25
 * @Desc:
**/

package entitysystem

import (
	"context"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	gshare "postapocgame/server/service/gameserver/internel/core/gshare"

	"google.golang.org/protobuf/proto"
)

// HandlePickupItem 处理拾取物品请求（Actor 消息版本）
// 约定：msg.Context 中包含 "session" 字段，可通过 EntityMgr 根据 sessionId 查到实体。
func HandlePickupItem(msg actor.IActorMessage) error {
	if msg == nil {
		return nil
	}

	ctx := msg.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	sessionId, _ := ctx.Value("session").(string)
	if sessionId == "" {
		return nil
	}

	entityMgr := entitymgr.GetEntityMgr()
	entityAny, ok := entityMgr.GetBySession(sessionId)
	if !ok || entityAny == nil {
		return nil
	}
	picker, ok := entityAny.(iface.IEntity)
	if !ok {
		return nil
	}

	var req protocol.C2SPickupItemReq
	if err := proto.Unmarshal(msg.GetData(), &req); err != nil {
		return err
	}

	// 获取掉落物实体
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
	if _, ok := picker.(iface.IRole); !ok {
		resp := &protocol.S2CPickupItemResultReq{
			Success: false,
			Message: "只有角色可以拾取",
			ItemHdl: req.ItemHdl,
		}
		return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
	}

	// 拾取物品逻辑：直接移除掉落物，物品添加由 GameServer Controller 层处理
	// 立即移除掉落物
	if scene, ok := entityMgr.GetSceneByHandle(req.ItemHdl); ok && scene != nil {
		if err := scene.RemoveEntity(req.ItemHdl); err != nil {
			log.Warnf("Failed to remove drop item from scene: %v", err)
		}
		entityMgr.UnbindScene(req.ItemHdl)
	}
	entityMgr.Unregister(req.ItemHdl)

	// 发送添加物品消息给 PlayerActor
	roleEntity, ok := picker.(interface {
		GetSessionId() string
		GetRoleId() uint64
	})
	if ok && roleEntity != nil {
		sendAddItemToPlayerActor(roleEntity.GetSessionId(), roleEntity.GetRoleId(), dropItem.GetItemId(), dropItem.GetCount(), req.ItemHdl)
	}

	resp := &protocol.S2CPickupItemResultReq{
		Success: true,
		Message: "拾取成功",
		ItemHdl: req.ItemHdl,
	}
	return picker.SendProtoMessage(uint16(protocol.S2CProtocol_S2CPickupItemResult), resp)
}

// sendAddItemToPlayerActor 发送添加物品消息给 PlayerActor
func sendAddItemToPlayerActor(sessionId string, roleId uint64, itemId uint32, count uint32, itemHdl uint64) {
	req := &protocol.D2GAddItemReq{
		SessionId: sessionId,
		RoleId:    roleId,
		ItemId:    itemId,
		Count:     count,
		ItemHdl:   itemHdl,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		log.Errorf("[drop_sys] marshal D2GAddItemReq failed: %v", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdAddItem), data)
	if err := gshare.SendMessageAsync(sessionId, actorMsg); err != nil {
		log.Errorf("[drop_sys] send AddItem message to PlayerActor failed: sessionId=%s err=%v", sessionId, err)
	}
}
