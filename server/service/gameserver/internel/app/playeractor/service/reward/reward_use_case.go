package reward

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/money"
	"postapocgame/server/service/gameserver/internel/iface"
)

// RewardUseCaseImpl 通用奖励发放，用 Bag/Money 用例实现。
type RewardUseCaseImpl struct {
	addMoneyUseCase *money.AddMoneyUseCase
	addItemUseCase  *bag.AddItemUseCase
}

func NewRewardUseCase(playerRepo iface.PlayerRepository, eventPublisher iface.EventPublisher, cfg iface.ConfigManager) iface.RewardUseCase {
	bagDeps := bag.Deps{
		PlayerRepo:     playerRepo,
		EventPublisher: eventPublisher,
		ConfigManager:  cfg,
	}
	return &RewardUseCaseImpl{
		addMoneyUseCase: money.NewAddMoneyUseCase(money.Deps{PlayerRepo: playerRepo, EventPublisher: eventPublisher}),
		addItemUseCase:  bag.NewAddItemUseCase(bagDeps),
	}
}

// GrantRewards 发放奖励
func (uc *RewardUseCaseImpl) GrantRewards(ctx context.Context, roleID uint64, rewards []*jsonconf.ItemAmount) error {
	for _, it := range rewards {
		if it == nil || it.Count <= 0 {
			continue
		}
		switch it.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if err := uc.addMoneyUseCase.Execute(ctx, roleID, it.ItemId, it.Count); err != nil {
				return err
			}
		case uint32(protocol.ItemType_ItemTypeMaterial), uint32(protocol.ItemType_ItemTypeEquipment):
			if err := uc.addItemUseCase.Execute(ctx, roleID, it.ItemId, uint32(it.Count), it.Bind); err != nil {
				return err
			}
		default:
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "unsupported item type")
		}
	}
	return nil
}

var _ iface.RewardUseCase = (*RewardUseCaseImpl)(nil)
