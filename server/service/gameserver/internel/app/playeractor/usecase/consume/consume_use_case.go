package consume

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/money"
)

// ConsumeUseCaseImpl 将通用消耗下沉到 UseCase 层，基于 Bag/Money 用例实现。
type ConsumeUseCaseImpl struct {
	playerRepo        repository.PlayerRepository
	bagHasItemUseCase *bag.HasItemUseCase
	bagRemoveUseCase  *bag.RemoveItemUseCase
	moneyConsumeUse   *money.ConsumeMoneyUseCase
}

func NewConsumeUseCase(playerRepo repository.PlayerRepository, eventPublisher interfaces.EventPublisher) interfaces.ConsumeUseCase {
	return &ConsumeUseCaseImpl{
		playerRepo:        playerRepo,
		bagHasItemUseCase: bag.NewHasItemUseCase(playerRepo),
		bagRemoveUseCase:  bag.NewRemoveItemUseCase(playerRepo, eventPublisher),
		moneyConsumeUse:   money.NewConsumeMoneyUseCase(playerRepo, eventPublisher),
	}
}

// CheckConsume 检查是否具备足够的消耗资源
func (uc *ConsumeUseCaseImpl) CheckConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error {
	if len(items) == 0 {
		return nil
	}

	moneyData, err := deps.PlayerGateway().GetMoneyData(ctx)
	if err != nil {
		return err
	}
	for _, it := range items {
		if it == nil || it.Count <= 0 {
			continue
		}
		switch it.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if moneyData.MoneyMap[it.ItemId] < it.Count {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Item_NotEnough), "money not enough")
			}
		case uint32(protocol.ItemType_ItemTypeMaterial), uint32(protocol.ItemType_ItemTypeEquipment):
			ok, err := uc.bagHasItemUseCase.Execute(ctx, roleID, it.ItemId, uint32(it.Count))
			if err != nil {
				return err
			}
			if !ok {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Item_NotEnough), "item not enough")
			}
		default:
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "unsupported item type")
		}
	}
	return nil
}

// ApplyConsume 扣除资源
func (uc *ConsumeUseCaseImpl) ApplyConsume(ctx context.Context, roleID uint64, items []*jsonconf.ItemAmount) error {
	for _, it := range items {
		if it == nil || it.Count <= 0 {
			continue
		}
		switch it.ItemType {
		case uint32(protocol.ItemType_ItemTypeMoney):
			if err := uc.moneyConsumeUse.Execute(ctx, roleID, it.ItemId, it.Count); err != nil {
				return err
			}
		case uint32(protocol.ItemType_ItemTypeMaterial), uint32(protocol.ItemType_ItemTypeEquipment):
			if err := uc.bagRemoveUseCase.Execute(ctx, roleID, it.ItemId, uint32(it.Count)); err != nil {
				return err
			}
		default:
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "unsupported item type")
		}
	}
	return nil
}

var _ interfaces.ConsumeUseCase = (*ConsumeUseCaseImpl)(nil)
