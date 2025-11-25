package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

// RecycleSys 物品回收系统
type RecycleSys struct {
	*BaseSystem
}

// NewRecycleSys 创建物品回收系统
func NewRecycleSys() *RecycleSys {
	return &RecycleSys{
		BaseSystem: NewBaseSystem(0), // 回收系统不需要系统ID，不作为独立系统注册
	}
}

// RecycleItem 回收物品
func (rs *RecycleSys) RecycleItem(ctx context.Context, itemID uint32, count uint32) ([]*protocol.ItemAmount, error) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return nil, customerr.Wrap(err)
	}

	// 检查物品配置
	itemConfig, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 检查物品是否可回收（通过Flag检查）
	if itemConfig.Flag&uint64(protocol.ItemFlag_ItemFlagCanDecompose) == 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item cannot be recycled")
	}

	// 检查背包中是否有该物品
	bagSys := GetBagSys(ctx)
	if bagSys == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "bag system not found")
	}

	if !bagSys.HasItem(itemID, count) {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	// 获取回收配置
	recycleConfig, ok := jsonconf.GetConfigManager().GetItemRecycleConfig(itemID)
	if !ok || recycleConfig == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item recycle config not found: %d", itemID)
	}

	// 计算奖励（按回收数量计算，转换为jsonconf.ItemAmount用于GrantRewards）
	jsonconfAwards := make([]*jsonconf.ItemAmount, 0)
	protocolAwards := make([]*protocol.ItemAmount, 0)
	for _, award := range recycleConfig.Awards {
		if award == nil {
			continue
		}
		// 奖励数量 = 配置数量 * 回收数量
		totalCount := int64(award.Count) * int64(count)
		// 用于GrantRewards的jsonconf格式
		jsonconfAwards = append(jsonconfAwards, &jsonconf.ItemAmount{
			ItemType: award.ItemType,
			ItemId:   award.ItemId,
			Count:    totalCount,
			Bind:     award.Bind,
		})
		// 用于协议返回的protocol格式
		protocolAwards = append(protocolAwards, &protocol.ItemAmount{
			ItemType: award.ItemType,
			ItemId:   award.ItemId,
			Count:    totalCount,
			Bind:     award.Bind,
		})
	}

	// 发放奖励
	if len(jsonconfAwards) > 0 {
		err = playerRole.GrantRewards(ctx, jsonconfAwards)
		if err != nil {
			log.Errorf("grant recycle rewards failed: %v", err)
			return nil, customerr.Wrap(err)
		}
	}

	// 删除物品
	err = bagSys.RemoveItem(ctx, itemID, count)
	if err != nil {
		return nil, customerr.Wrap(err)
	}

	log.Infof("Item recycled: ItemID=%d, Count=%d, Awards=%d", itemID, count, len(protocolAwards))
	return protocolAwards, nil
}

// handleRecycleItem 处理回收物品
func handleRecycleItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SRecycleItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 默认回收数量为1
	if req.Count == 0 {
		req.Count = 1
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 创建回收系统实例
	recycleSys := NewRecycleSys()

	// 回收物品
	awards, err := recycleSys.RecycleItem(ctx, req.ItemId, req.Count)

	// 构造响应
	resp := &protocol.S2CRecycleItemResultReq{
		Success:       err == nil,
		ItemId:        req.ItemId,
		RecycledCount: req.Count,
		Awards:        awards,
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Success = false
	} else {
		resp.Message = "回收成功"
		// 推送背包数据更新
		pushBagData(ctx, sessionId)
		// 推送货币数据更新（如果有货币奖励）
		pushMoneyData(ctx, sessionId)
	}

	// 发送响应
	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRecycleItemResult), resp); sendErr != nil {
		return sendErr
	}

	return err
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRecycleItem), handleRecycleItem)
	})
}
