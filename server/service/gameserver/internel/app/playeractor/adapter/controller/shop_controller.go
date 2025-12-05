package controller

import (
	"context"
	presenter2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/consume"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/reward"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/shop"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// ShopController 商城控制器
type ShopController struct {
	buyItemUseCase *shop.BuyItemUseCase
	presenter      *presenter2.ShopPresenter
}

// NewShopController 创建商城控制器
func NewShopController() *ShopController {
	buyItemUC := shop.NewBuyItemUseCase(deps.PlayerGateway(), deps.ConfigGateway())

	// 注入依赖
	consumeUseCase := consume.NewConsumeUseCase(deps.PlayerGateway(), deps.EventPublisher())
	rewardUseCase := reward.NewRewardUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
	buyItemUC.SetDependencies(consumeUseCase, rewardUseCase)

	return &ShopController{
		buyItemUseCase: buyItemUC,
		presenter:      presenter2.NewShopPresenter(deps.NetworkGateway()),
	}
}

// HandleShopBuy 处理购买商品请求
func (c *ShopController) HandleShopBuy(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	shopSys := system.GetShopSys(ctx)
	if shopSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "商城系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SShopBuyReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 构建响应
	resp := &protocol.S2CShopBuyResultReq{
		ItemId: req.ItemId,
		Count:  req.Count,
	}

	// 参数校验
	if req.ItemId == 0 || req.Count == 0 {
		resp.ErrCode = uint32(protocol.ErrorCode_Param_Invalid)
		return c.presenter.SendShopBuyResult(ctx, sessionID, resp)
	}

	// 执行购买用例
	err = c.buyItemUseCase.Execute(ctx, roleID, req.ItemId, req.Count)
	if err != nil {
		resp.ErrCode = errCodeFromError(err)
		return c.presenter.SendShopBuyResult(ctx, sessionID, resp)
	}

	resp.ErrCode = uint32(protocol.ErrorCode_Success)
	if err := c.presenter.SendShopBuyResult(ctx, sessionID, resp); err != nil {
		return err
	}

	// 推送背包和货币数据更新
	presenter2.PushBagData(ctx, sessionID)
	presenter2.PushMoneyData(ctx, sessionID)

	return nil
}
