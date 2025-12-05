package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	bag2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// bagUseCaseAdapter 实现 BagUseCase 接口，供其他系统依赖
type bagUseCaseAdapter struct {
	addItemUseCase    *bag2.AddItemUseCase
	removeItemUseCase *bag2.RemoveItemUseCase
	hasItemUseCase    *bag2.HasItemUseCase
}

// NewBagUseCaseAdapter 创建 BagUseCase 适配器
func NewBagUseCaseAdapter() interfaces.BagUseCase {
	return &bagUseCaseAdapter{
		addItemUseCase:    bag2.NewAddItemUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway()),
		removeItemUseCase: bag2.NewRemoveItemUseCase(deps.PlayerGateway(), deps.EventPublisher()),
		hasItemUseCase:    bag2.NewHasItemUseCase(deps.PlayerGateway()),
	}
}

func (a *bagUseCaseAdapter) GetItem(ctx context.Context, roleID uint64, itemID uint32) (*protocol.ItemSt, error) {
	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range bagData.Items {
		if item != nil && item.ItemId == itemID {
			return item, nil
		}
	}
	return nil, nil
}

func (a *bagUseCaseAdapter) RemoveItem(ctx context.Context, roleID uint64, itemID uint32, count uint32) error {
	return a.removeItemUseCase.Execute(ctx, roleID, itemID, count)
}

func (a *bagUseCaseAdapter) AddItem(ctx context.Context, roleID uint64, itemID uint32, count uint32, bind uint32) error {
	return a.addItemUseCase.Execute(ctx, roleID, itemID, count, bind)
}

func (a *bagUseCaseAdapter) HasItem(ctx context.Context, roleID uint64, itemID uint32, count uint32) (bool, error) {
	return a.hasItemUseCase.Execute(ctx, roleID, itemID, count)
}
