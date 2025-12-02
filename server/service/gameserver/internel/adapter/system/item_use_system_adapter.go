package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/item_use"
)

// ItemUseSystemAdapter 物品使用系统适配器
type ItemUseSystemAdapter struct {
	*BaseSystemAdapter
	useItemUseCase *item_use.UseItemUseCase
}

// NewItemUseSystemAdapter 创建物品使用系统适配器
func NewItemUseSystemAdapter() *ItemUseSystemAdapter {
	container := di.GetContainer()
	useItemUC := item_use.NewUseItemUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())

	return &ItemUseSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysItemUse)),
		useItemUseCase:    useItemUC,
	}
}

// OnInit 系统初始化
func (a *ItemUseSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("item use sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		log.Errorf("item use sys OnInit get binary data err:%v", err)
		return
	}

	// 如果item_use_data不存在，则初始化
	if binaryData.ItemUseData == nil {
		binaryData.ItemUseData = &protocol.SiItemUseData{
			CooldownMap: make(map[uint32]int64),
		}
	}

	log.Infof("ItemUseSys initialized")
}

// UseItem 使用物品（对外接口，供其他系统调用）
func (a *ItemUseSystemAdapter) UseItem(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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

// CheckCooldown 检查物品是否在冷却中
func (a *ItemUseSystemAdapter) CheckCooldown(ctx context.Context, itemID uint32) (bool, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return false, err
	}
	if binaryData.ItemUseData == nil || binaryData.ItemUseData.CooldownMap == nil {
		return false, nil
	}
	// TODO: 使用 servertime 获取当前时间
	// now := servertime.Now().Unix()
	// if cooldownEnd, exists := binaryData.ItemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
	// 	return true, nil
	// }
	return false, nil
}

// EnsureISystem 确保 ItemUseSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*ItemUseSystemAdapter)(nil)
