package shop

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// BuyItemUseCase 购买商品用例
type BuyItemUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces.ConfigManager
	consumeUseCase interfaces.ConsumeUseCase
	rewardUseCase  interfaces.RewardUseCase
}

// NewBuyItemUseCase 创建购买商品用例
func NewBuyItemUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *BuyItemUseCase {
	return &BuyItemUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// SetDependencies 设置依赖（用于注入 ConsumeUseCase、RewardUseCase）
func (uc *BuyItemUseCase) SetDependencies(
	consumeUseCase interfaces.ConsumeUseCase,
	rewardUseCase interfaces.RewardUseCase,
) {
	uc.consumeUseCase = consumeUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行购买商品用例
func (uc *BuyItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	if count == 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "count invalid")
	}

	// 获取物品配置
	itemConfigRaw, ok := uc.configManager.GetItemConfig(itemID)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item config not found")
	}

	itemConfig, ok := itemConfigRaw.(*jsonconf.ItemConfig)
	if !ok || itemConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "invalid item config type")
	}

	// 构建消耗列表
	consumes, err := uc.buildConsumeList(itemConfig, count)
	if err != nil {
		return err
	}

	// 构建奖励列表
	rewards := uc.buildRewardList(itemConfig, count)

	// 检查消耗
	if uc.consumeUseCase != nil {
		if err := uc.consumeUseCase.CheckConsume(ctx, roleID, consumes); err != nil {
			return err
		}
	}

	// 扣除消耗
	if uc.consumeUseCase != nil {
		if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, consumes); err != nil {
			return err
		}
	}

	// 发放奖励
	if uc.rewardUseCase != nil {
		if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewards); err != nil {
			return err
		}
	}

	// 记录购买次数
	if err := uc.recordPurchase(ctx, roleID, itemID, count); err != nil {
		log.Warnf("Record purchase failed: %v", err)
		// 记录失败不影响购买流程
	}

	return nil
}

// buildConsumeList 构建消耗列表
func (uc *BuyItemUseCase) buildConsumeList(cfg *jsonconf.ItemConfig, count uint32) ([]*jsonconf.ItemAmount, error) {
	if cfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config nil")
	}

	// 从商城配置读取消耗
	shopCfgRaw, ok := uc.configManager.GetShopConfig(cfg.ItemId)
	if !ok || shopCfgRaw == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item not in shop")
	}

	shopCfg, ok := shopCfgRaw.(*jsonconf.ShopConfig)
	if !ok || shopCfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "invalid shop config type")
	}

	if shopCfg.ConsumeId > 0 {
		consumeCfgRaw, ok := uc.configManager.GetConsumeConfig(shopCfg.ConsumeId)
		if ok && consumeCfgRaw != nil {
			consumeCfg, ok := consumeCfgRaw.(*jsonconf.ConsumeConfig)
			if ok && consumeCfg != nil {
				return scaleItemAmounts(consumeCfg.Items, count), nil
			}
		}
	}

	// 如果没有商城配置，返回错误（必须通过商城配置购买）
	return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item not in shop")
}

// buildRewardList 构建奖励列表
func (uc *BuyItemUseCase) buildRewardList(cfg *jsonconf.ItemConfig, count uint32) []*jsonconf.ItemAmount {
	if cfg == nil {
		return nil
	}

	// 从商城配置读取奖励
	shopCfgRaw, ok := uc.configManager.GetShopConfig(cfg.ItemId)
	if ok && shopCfgRaw != nil {
		shopCfg, ok := shopCfgRaw.(*jsonconf.ShopConfig)
		if ok && shopCfg != nil && shopCfg.RewardId > 0 {
			rewardCfgRaw, ok := uc.configManager.GetRewardConfig(shopCfg.RewardId)
			if ok && rewardCfgRaw != nil {
				rewardCfg, ok := rewardCfgRaw.(*jsonconf.RewardConfig)
				if ok && rewardCfg != nil {
					return scaleItemAmounts(rewardCfg.Items, count)
				}
			}
		}
	}

	// 如果没有商城配置，返回默认奖励（物品本身）
	return []*jsonconf.ItemAmount{
		{
			ItemType: cfg.Type,
			ItemId:   cfg.ItemId,
			Count:    int64(count),
			Bind:     1,
		},
	}
}

// recordPurchase 记录购买次数
// 注意：原始代码中 purchaseCounters 是存储在系统实例中的内存数据，不是持久化的
// 这里暂时不实现持久化，保持与原始代码一致
func (uc *BuyItemUseCase) recordPurchase(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	// TODO: 如果需要持久化购买次数，可以在这里实现
	// 当前保持与原始代码一致，不持久化购买次数
	return nil
}

// scaleItemAmounts 缩放物品数量
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
