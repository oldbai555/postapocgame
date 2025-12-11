package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	bag2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/bag"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

type BagSystemAdapter struct {
	*BaseSystemAdapter
	addItemUseCase    *bag2.AddItemUseCase
	removeItemUseCase *bag2.RemoveItemUseCase
	hasItemUseCase    *bag2.HasItemUseCase // TODO(adapter-phaseA-A2): 纯校验型逻辑后续可统一通过 BagUseCase 接口暴露，SystemAdapter 只做路由
}

// NewBagSystemAdapter 创建背包系统适配器
func NewBagSystemAdapter() *BagSystemAdapter {
	return &BagSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysBag)),
		addItemUseCase:    bag2.NewAddItemUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway()),
		removeItemUseCase: bag2.NewRemoveItemUseCase(deps.PlayerGateway(), deps.EventPublisher()),
		hasItemUseCase:    bag2.NewHasItemUseCase(deps.PlayerGateway()),
	}
}

// OnInit 系统初始化
func (a *BagSystemAdapter) OnInit(ctx context.Context) {
	// 初始化可选：提前构建索引用于查询
	_, _ = bag2.NewBagSnapshot(ctx, deps.PlayerGateway())
}

// AddItem 添加物品（对外接口，供其他系统调用）
func (a *BagSystemAdapter) AddItem(ctx context.Context, itemID uint32, count uint32, bind uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.addItemUseCase.Execute(ctx, roleID, itemID, count, bind)
}

// RemoveItem 移除物品（对外接口，供其他系统调用）
func (a *BagSystemAdapter) RemoveItem(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.removeItemUseCase.Execute(ctx, roleID, itemID, count)
}

// HasItem 检查是否拥有足够数量的指定物品
func (a *BagSystemAdapter) HasItem(ctx context.Context, itemID uint32, count uint32) (bool, error) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	return a.hasItemUseCase.Execute(ctx, roleID, itemID, count)
}

// GetItem 获取物品（使用辅助索引优化查找）
func (a *BagSystemAdapter) GetItem(ctx context.Context, itemID uint32) (*protocol.ItemSt, error) {
	acc, err := bag2.NewBagSnapshot(ctx, deps.PlayerGateway())
	if err != nil {
		return nil, err
	}
	return acc.Find(itemID, 0), nil
}

// GetAllItems 获取所有物品
func (a *BagSystemAdapter) GetAllItems(ctx context.Context) ([]*protocol.ItemSt, error) {
	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return nil, err
	}
	return bagData.Items, nil
}

// GetItemsSnapshot 获取物品快照（用于事务回滚）
// 说明：该方法主要用于兼容旧的经济系统逻辑（ApplyConsume/GrantRewards），不会触发事件，
// 仅操作 BinaryData 中的内存数据，符合当前"数据只存 BinaryData、不直接访问数据库"的约束。
func (a *BagSystemAdapter) GetItemsSnapshot(ctx context.Context) (map[uint32]*protocol.ItemSt, error) {
	acc, err := bag2.NewBagSnapshot(ctx, deps.PlayerGateway())
	if err != nil {
		return nil, err
	}
	return acc.Snapshot(), nil
}

// RestoreItemsSnapshot 恢复物品快照（用于事务回滚）
func (a *BagSystemAdapter) RestoreItemsSnapshot(ctx context.Context, snapshot map[uint32]*protocol.ItemSt) error {
	acc, err := bag2.NewBagSnapshot(ctx, deps.PlayerGateway())
	if err != nil {
		return err
	}
	acc.Restore(snapshot)
	return nil
}

// GetBagData 获取背包数据（用于协议）
func (a *BagSystemAdapter) GetBagData(ctx context.Context) (*protocol.SiBagData, error) {
	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return nil, err
	}
	return bagData, nil
}

// EnsureISystem 确保 BagSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*BagSystemAdapter)(nil)

// GetBagSys 获取背包系统
func GetBagSys(ctx context.Context) *BagSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysBag)
		return nil
	}
	sys, ok := system.(*BagSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysBag)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysBag)
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSystemAdapter()
	})
}
