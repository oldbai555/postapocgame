package controller

import (
	"context"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/shop"
)

// ShopController 商城控制器
type ShopController struct {
	buyItemUseCase *shop.BuyItemUseCase
	presenter      *presenter.ShopPresenter
}

// NewShopController 创建商城控制器
func NewShopController() *ShopController {
	container := di.GetContainer()
	buyItemUC := shop.NewBuyItemUseCase(container.PlayerGateway(), container.ConfigGateway())

	// 注入依赖
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	rewardUseCase := usecaseadapter.NewRewardUseCaseAdapter()
	buyItemUC.SetDependencies(consumeUseCase, rewardUseCase)

	return &ShopController{
		buyItemUseCase: buyItemUC,
		presenter:      presenter.NewShopPresenter(container.NetworkGateway()),
	}
}

// HandleShopBuy 处理购买商品请求
func (c *ShopController) HandleShopBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SShopBuyReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
	presenter.PushBagData(ctx, sessionID)
	presenter.PushMoneyData(ctx, sessionID)

	return nil
}
