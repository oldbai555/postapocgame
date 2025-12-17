package bag

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/iface"
)

// bagUseCaseAdapter 实现 BagUseCase 接口，供其他系统依赖
type bagUseCaseAdapter struct {
	addItemUseCase    *AddItemUseCase
	removeItemUseCase *RemoveItemUseCase
	hasItemUseCase    *HasItemUseCase
}

// NewBagUseCaseAdapter 创建 BagUseCase 适配器
// 注意：由于其他系统（Equip/Recycle）需要调用 BagUseCase，这里使用 deps 工厂函数
// TODO: 后续可考虑将 BagUseCase 实例挂载到 Runtime
func NewBagUseCaseAdapter() iface.BagUseCase {
	d := Deps{
		PlayerRepo:     deps.NewPlayerGateway(),
		EventPublisher: deps.NewEventPublisher(),
		ConfigManager:  deps.NewConfigManager(),
		NetworkGateway: deps.NewNetworkGateway(),
	}
	return &bagUseCaseAdapter{
		addItemUseCase:    NewAddItemUseCase(d),
		removeItemUseCase: NewRemoveItemUseCase(d),
		hasItemUseCase:    NewHasItemUseCase(d),
	}
}

func (a *bagUseCaseAdapter) GetItem(ctx context.Context, roleID uint64, itemID uint32) (*protocol.ItemSt, error) {
	// 读取一次当前背包数据，用于简单查询；完整事务/快照逻辑由 BagSystemAdapter/Accessor 负责。
	bagData, err := a.addItemUseCase.deps.PlayerRepo.GetBagData(ctx)
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
