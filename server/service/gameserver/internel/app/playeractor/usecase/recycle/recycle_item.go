package recycle

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// RecycleItemUseCase 回收物品用例
type RecycleItemUseCase struct {
	configManager interfaces2.ConfigManager
	bagUseCase    interfaces2.BagUseCase
	rewardUseCase interfaces2.RewardUseCase
}

// NewRecycleItemUseCase 创建回收物品用例
func NewRecycleItemUseCase(configManager interfaces2.ConfigManager) *RecycleItemUseCase {
	return &RecycleItemUseCase{
		configManager: configManager,
	}
}

// SetDependencies 设置依赖
func (uc *RecycleItemUseCase) SetDependencies(
	bagUseCase interfaces2.BagUseCase,
	rewardUseCase interfaces2.RewardUseCase,
) {
	uc.bagUseCase = bagUseCase
	uc.rewardUseCase = rewardUseCase
}

// Execute 执行回收逻辑
func (uc *RecycleItemUseCase) Execute(ctx context.Context, roleID uint64, itemID uint32, count uint32) ([]*protocol.ItemAmount, error) {
	if count == 0 {
		count = 1
	}

	if uc.bagUseCase == nil || uc.rewardUseCase == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "recycle dependencies not ready")
	}

	// 检查物品配置
	itemConfig := uc.configManager.GetItemConfig(itemID)
	if itemConfig == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item config not found: %d", itemID)
	}

	// 检查是否可回收
	if itemConfig.Flag&uint64(protocol.ItemFlag_ItemFlagCanDecompose) == 0 {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item cannot be recycled")
	}

	// 检查背包中是否有足够数量
	hasItem, err := uc.bagUseCase.HasItem(ctx, roleID, itemID, count)
	if err != nil {
		return nil, err
	}
	if !hasItem {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item not enough")
	}

	// 获取回收配置
	recycleConfig := uc.configManager.GetItemRecycleConfig(itemID)
	if recycleConfig == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "item recycle config not found: %d", itemID)
	}

	// 计算奖励
	jsonconfAwards := make([]*jsonconf.ItemAmount, 0, len(recycleConfig.Awards))
	protocolAwards := make([]*protocol.ItemAmount, 0, len(recycleConfig.Awards))
	for _, award := range recycleConfig.Awards {
		if award == nil {
			continue
		}
		totalCount := award.Count * int64(count)
		jsonconfAwards = append(jsonconfAwards, &jsonconf.ItemAmount{
			ItemType: award.ItemType,
			ItemId:   award.ItemId,
			Count:    totalCount,
			Bind:     award.Bind,
		})
		protocolAwards = append(protocolAwards, &protocol.ItemAmount{
			ItemType: award.ItemType,
			ItemId:   award.ItemId,
			Count:    totalCount,
			Bind:     award.Bind,
		})
	}

	// 发放奖励
	if len(jsonconfAwards) > 0 {
		if err := uc.rewardUseCase.GrantRewards(ctx, roleID, jsonconfAwards); err != nil {
			log.Errorf("grant recycle rewards failed: %v", err)
			return nil, customerr.Wrap(err)
		}
	}

	// 删除物品
	if err := uc.bagUseCase.RemoveItem(ctx, roleID, itemID, count); err != nil {
		return nil, customerr.Wrap(err)
	}

	log.Infof("Item recycled: role=%d, item=%d, count=%d, awards=%d", roleID, itemID, count, len(protocolAwards))
	return protocolAwards, nil
}
