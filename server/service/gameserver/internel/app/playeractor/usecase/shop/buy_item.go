package shop

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// BuyItemUseCase 购买商品用例
type BuyItemUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces2.ConfigManager
	consumeUseCase interfaces2.ConsumeUseCase
	rewardUseCase  interfaces2.RewardUseCase
}

// NewBuyItemUseCase 创建购买商品用例
func NewBuyItemUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces2.ConfigManager,
) *BuyItemUseCase {
	return &BuyItemUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// SetDependencies 设置依赖（用于注入 ConsumeUseCase、RewardUseCase）
func (uc *BuyItemUseCase) SetDependencies(
	consumeUseCase interfaces2.ConsumeUseCase,
	rewardUseCase interfaces2.RewardUseCase,
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
	itemConfig := uc.configManager.GetItemConfig(itemID)
	if itemConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item config not found")
	}

	if uc.consumeUseCase == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case nil")
	}

	if uc.rewardUseCase == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "reward use case nil")
	}

	// 构建消耗列表
	consumes, err := uc.buildConsumeList(itemConfig, count)
	if err != nil {
		return err
	}

	// 检查消耗
	if err := uc.consumeUseCase.CheckConsume(ctx, roleID, consumes); err != nil {
		return err
	}

	// 扣除消耗
	if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, consumes); err != nil {
		return err
	}

	// 构建奖励列表
	rewards := uc.buildRewardList(itemConfig, count)

	// 发放奖励
	if err := uc.rewardUseCase.GrantRewards(ctx, roleID, rewards); err != nil {
		return err
	}

	return nil
}

// buildConsumeList 构建消耗列表
func (uc *BuyItemUseCase) buildConsumeList(cfg *jsonconf.ItemConfig, count uint32) ([]*jsonconf.ItemAmount, error) {
	if cfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config nil")
	}

	// 从商城配置读取消耗
	shopCfg := uc.configManager.GetShopConfig(cfg.ItemId)
	if shopCfg == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "item not in shop")
	}

	if shopCfg.ConsumeId > 0 {
		consumeCfg := uc.configManager.GetConsumeConfig(shopCfg.ConsumeId)
		if consumeCfg != nil {
			return scaleItemAmounts(consumeCfg.Items, count), nil
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
	shopCfg := uc.configManager.GetShopConfig(cfg.ItemId)
	if shopCfg == nil || shopCfg.RewardId == 0 {
		return nil
	}
	rewardCfg := uc.configManager.GetRewardConfig(shopCfg.RewardId)
	if rewardCfg == nil {
		return nil
	}
	return scaleItemAmounts(rewardCfg.Items, count)
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
