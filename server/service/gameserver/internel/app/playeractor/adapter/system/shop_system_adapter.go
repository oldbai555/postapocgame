package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/consume"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/shop"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// ShopSystemAdapter 商城系统适配器
//
// 生命周期职责：
// - OnInit: 暂未使用（商城数据无需初始化）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：所有业务逻辑（购买商品、消耗/奖励构建）均在 UseCase 层实现
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
type ShopSystemAdapter struct {
	*BaseSystemAdapter
	buyItemUseCase *shop.BuyItemUseCase
}

// NewShopSystemAdapter 创建商城系统适配器
func NewShopSystemAdapter() *ShopSystemAdapter {
	buyItemUC := shop.NewBuyItemUseCase(deps.PlayerGateway(), deps.ConfigGateway())

	// 注入依赖
	consumeUseCase := consume.NewConsumeUseCase(deps.PlayerGateway(), deps.EventPublisher())
	rewardUseCase := reward.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
	buyItemUC.SetDependencies(consumeUseCase, rewardUseCase)

	return &ShopSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysShop)),
		buyItemUseCase:    buyItemUC,
	}
}

// OnInit 系统初始化
func (a *ShopSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("shop sys OnInit get role err:%v", err)
		return
	}

	// 注意：原始代码中 purchaseCounters 是存储在系统实例中的内存数据，不是持久化的
	// 这里暂时不实现持久化，保持与原始代码一致
	// TODO(adapter-phaseA-A2): 限购/统计等策略应由 Shop 领域/UseCase 统一管理，适配层仅作为“本局运行期内存态”或查询入口

	log.Infof("ShopSys initialized: RoleID=%d", roleID)
}

// Buy 购买商品（对外接口，供其他系统调用）
func (a *ShopSystemAdapter) Buy(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	return a.buyItemUseCase.Execute(ctx, roleID, itemID, count)
}

// GetPurchaseCount 获取购买次数（对外接口，供其他系统调用）
// 注意：原始代码中 purchaseCounters 是存储在系统实例中的内存数据，不是持久化的
// 这里暂时返回 0，保持与原始代码一致
func (a *ShopSystemAdapter) GetPurchaseCount(ctx context.Context, itemID uint32) (uint32, error) {
	// 当前版本沿用原始实现：购买次数仅用于运行期限购校验，不做数据库持久化，
	// 具体的限购/审计需求将在后续版本通过独立的统计与审计系统实现。
	return 0, nil
}

// EnsureISystem 确保 ShopSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*ShopSystemAdapter)(nil)

// GetShopSys 获取商城系统
func GetShopSys(ctx context.Context) *ShopSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get ShopSys player role failed: %v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysShop))
	if system == nil {
		return nil
	}
	sys, ok := system.(*ShopSystemAdapter)
	if !ok || !sys.IsOpened() {
		return nil
	}
	return sys
}

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysShop), func() iface.ISystem {
		return NewShopSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
