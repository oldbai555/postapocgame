package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/bag"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// bagUseCaseAdapter 实现 BagUseCase 接口，供其他系统依赖
type bagUseCaseAdapter struct {
	addItemUseCase    *bag.AddItemUseCase
	removeItemUseCase *bag.RemoveItemUseCase
	hasItemUseCase    *bag.HasItemUseCase
}

// NewBagUseCaseAdapter 创建 BagUseCase 适配器
func NewBagUseCaseAdapter() interfaces.BagUseCase {
	container := di.GetContainer()
	return &bagUseCaseAdapter{
		addItemUseCase:    bag.NewAddItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway()),
		removeItemUseCase: bag.NewRemoveItemUseCase(container.PlayerGateway(), container.EventPublisher()),
		hasItemUseCase:    bag.NewHasItemUseCase(container.PlayerGateway()),
	}
}

func (a *bagUseCaseAdapter) GetItem(ctx context.Context, roleID uint64, itemID uint32) (*protocol.ItemSt, error) {
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.BagData == nil || binaryData.BagData.Items == nil {
		return nil, nil
	}
	for _, item := range binaryData.BagData.Items {
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
