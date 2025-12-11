package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	item_use2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/item_use"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

type ItemUseSystemAdapter struct {
	*BaseSystemAdapter
	useItemUseCase *item_use2.UseItemUseCase
}

// NewItemUseSystemAdapter 创建物品使用系统适配器
func NewItemUseSystemAdapter() *ItemUseSystemAdapter {
	useItemUC := item_use2.NewUseItemUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())

	return &ItemUseSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysItemUse)),
		useItemUseCase:    useItemUC,
	}
}

// UseItem 使用物品（对外接口，供其他系统调用）
func (a *ItemUseSystemAdapter) UseItem(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := a.useItemUseCase.Execute(ctx, roleID, itemID, count); err != nil {
		return err
	}
	// HP/MP 等战斗属性的最终同步由 AttrSys + DungeonServer 自身协议负责，
	// ItemUse 用例只负责更新玩家数据与背包，不直接向 DungeonServer 下发属性补丁。
	return nil
}

// CheckCooldown 检查物品是否在冷却中（对外接口，供其他系统调用）
func (a *ItemUseSystemAdapter) CheckCooldown(ctx context.Context, itemID uint32) (bool, error) {
	itemUseData, err := deps.PlayerGateway().GetItemUseData(ctx)
	if err != nil {
		return false, err
	}
	// 使用 servertime 获取当前时间
	now := servertime.Now().Unix()
	if cooldownEnd, exists := itemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return true, nil
	}
	return false, nil
}

// EnsureISystem 确保 ItemUseSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*ItemUseSystemAdapter)(nil)

// GetItemUseSys 获取物品使用系统
func GetItemUseSys(ctx context.Context) *ItemUseSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysItemUse))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysItemUse)
		return nil
	}
	sys, ok := system.(*ItemUseSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysItemUse)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysItemUse)
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysItemUse), func() iface.ISystem {
		return NewItemUseSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
