package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/shop"
)

// ShopSystemAdapter 商城系统适配器
type ShopSystemAdapter struct {
	*BaseSystemAdapter
	buyItemUseCase *shop.BuyItemUseCase
}

// NewShopSystemAdapter 创建商城系统适配器
func NewShopSystemAdapter() *ShopSystemAdapter {
	container := di.GetContainer()
	buyItemUC := shop.NewBuyItemUseCase(container.PlayerGateway(), container.ConfigGateway())

	// 注入依赖
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	buyItemUC.SetDependencies(consumeUseCase, rewardUseCase)

	return &ShopSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysShop)),
		buyItemUseCase:    buyItemUC,
	}
}

// OnInit 系统初始化
func (a *ShopSystemAdapter) OnInit(ctx context.Context) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		log.Errorf("shop sys OnInit get role err:%v", err)
		return
	}

	// 注意：原始代码中 purchaseCounters 是存储在系统实例中的内存数据，不是持久化的
	// 这里暂时不实现持久化，保持与原始代码一致

	log.Infof("ShopSys initialized: RoleID=%d", roleID)
}

// Buy 购买商品（对外接口，供其他系统调用）
func (a *ShopSystemAdapter) Buy(ctx context.Context, itemID uint32, count uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
