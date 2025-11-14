package entitysystem

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

const shopDefaultMoneyID uint32 = 1

// ShopSys 个人商城系统
type ShopSys struct {
	*BaseSystem
	purchaseCounters map[uint32]uint32
}

func NewShopSys() *ShopSys {
	return &ShopSys{
		BaseSystem:       NewBaseSystem(uint32(protocol.SystemId_SysShop)),
		purchaseCounters: make(map[uint32]uint32),
	}
}

func GetShopSys(ctx context.Context) *ShopSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get ShopSys player role failed: %v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysShop))
	if system == nil {
		return nil
	}
	shopSys, ok := system.(*ShopSys)
	if !ok || !shopSys.IsOpened() {
		return nil
	}
	return shopSys
}

// Buy 购买商城商品
func (ss *ShopSys) Buy(ctx context.Context, itemID, count uint32) error {
	if count == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "count invalid")
	}
	cfg, ok := jsonconf.GetConfigManager().GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item config not found")
	}

	consumes, err := ss.buildConsumeList(cfg, count)
	if err != nil {
		return err
	}
	rewards := ss.buildRewardList(cfg, count)

	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return err
	}

	if err := playerRole.ApplyConsume(ctx, consumes); err != nil {
		return err
	}
	if err := playerRole.GrantRewards(ctx, rewards); err != nil {
		return err
	}

	ss.recordPurchase(itemID, count)
	return nil
}

func (ss *ShopSys) buildConsumeList(cfg *jsonconf.ItemConfig, count uint32) ([]*jsonconf.ItemAmount, error) {
	if cfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config nil")
	}
	manager := jsonconf.GetConfigManager()

	// 从商城配置读取消耗
	shopCfg, ok := manager.GetShopConfig(cfg.ItemId)
	if ok && shopCfg != nil {
		if shopCfg.ConsumeId > 0 {
			if consumeCfg, ok := manager.GetConsumeConfig(shopCfg.ConsumeId); ok && consumeCfg != nil {
				return scaleItemAmounts(consumeCfg.Items, count), nil
			}
		}
	}

	// 如果没有商城配置，返回错误（必须通过商城配置购买）
	return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item not in shop")
}

func (ss *ShopSys) buildRewardList(cfg *jsonconf.ItemConfig, count uint32) []*jsonconf.ItemAmount {
	if cfg == nil {
		return nil
	}
	manager := jsonconf.GetConfigManager()

	// 从商城配置读取奖励
	shopCfg, ok := manager.GetShopConfig(cfg.ItemId)
	if ok && shopCfg != nil {
		if shopCfg.RewardId > 0 {
			if rewardCfg, ok := manager.GetRewardConfig(shopCfg.RewardId); ok && rewardCfg != nil {
				return scaleItemAmounts(rewardCfg.Items, count)
			}
		}
	}

	// 如果没有商城配置，返回默认奖励（物品本身）
	return []*jsonconf.ItemAmount{
		{
			ItemType: uint32(cfg.Type),
			ItemId:   cfg.ItemId,
			Count:    int64(count),
			Bind:     1,
		},
	}
}

func (ss *ShopSys) recordPurchase(itemID, count uint32) {
	ss.purchaseCounters[itemID] += count
}

func scaleItemAmounts(items []*jsonconf.ItemAmount, times uint32) []*jsonconf.ItemAmount {
	if times == 0 || len(items) == 0 {
		return nil
	}
	result := make([]*jsonconf.ItemAmount, 0, len(items))
	for _, item := range items {
		if item == nil || item.Count <= 0 {
			continue
		}
		cp := item.Clone()
		if cp == nil {
			continue
		}
		cp.Count = cp.Count * int64(times)
		result = append(result, cp)
	}
	return result
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysShop), func() iface.ISystem {
		return NewShopSys()
	})
}
