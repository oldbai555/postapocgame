package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/item_use"
)

// ItemUseSystemAdapter 物品使用系统适配器
//
// 生命周期职责：
// - OnInit: 调用 InitItemUseDataUseCase 初始化物品使用数据（冷却映射结构）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（使用物品、冷却检查）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type ItemUseSystemAdapter struct {
	*BaseSystemAdapter
	useItemUseCase         *item_use.UseItemUseCase
	initItemUseDataUseCase *item_use.InitItemUseDataUseCase
}

// NewItemUseSystemAdapter 创建物品使用系统适配器
func NewItemUseSystemAdapter() *ItemUseSystemAdapter {
	container := di.GetContainer()
	useItemUC := item_use.NewUseItemUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())

	return &ItemUseSystemAdapter{
		BaseSystemAdapter:      NewBaseSystemAdapter(uint32(protocol.SystemId_SysItemUse)),
		useItemUseCase:         useItemUC,
		initItemUseDataUseCase: item_use.NewInitItemUseDataUseCase(container.PlayerGateway()),
	}
}

// OnInit 系统初始化
func (a *ItemUseSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("item use sys OnInit get role err:%v", err)
		return
	}
	// 初始化物品使用数据（包括冷却映射结构等业务逻辑）
	if err := a.initItemUseDataUseCase.Execute(ctx, roleID); err != nil {
		log.Errorf("item use sys OnInit init item use data err:%v", err)
		return
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

// CheckCooldown 检查物品是否在冷却中（对外接口，供其他系统调用）
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
	// 使用 servertime 获取当前时间
	now := servertime.Now().Unix()
	if cooldownEnd, exists := binaryData.ItemUseData.CooldownMap[itemID]; exists && cooldownEnd > now {
		return true, nil
	}
	return false, nil
}

// EnsureISystem 确保 ItemUseSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*ItemUseSystemAdapter)(nil)
