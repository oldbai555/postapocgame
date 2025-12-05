package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	bag2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/bag"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// BagController 背包控制器
type BagController struct {
	addItemUseCase    *bag2.AddItemUseCase
	removeItemUseCase *bag2.RemoveItemUseCase
	hasItemUseCase    *bag2.HasItemUseCase
	presenter         *presenter.BagPresenter
}

// NewBagController 创建背包控制器
func NewBagController() *BagController {
	return &BagController{
		addItemUseCase:    bag2.NewAddItemUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway()),
		removeItemUseCase: bag2.NewRemoveItemUseCase(deps.PlayerGateway(), deps.EventPublisher()),
		hasItemUseCase:    bag2.NewHasItemUseCase(deps.PlayerGateway()),
		presenter:         presenter.NewBagPresenter(deps.NetworkGateway()),
	}
}

// HandleOpenBag 处理打开背包请求
func (c *BagController) HandleOpenBag(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	bagSys := system.GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "背包系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 获取背包数据
	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送响应
	return c.presenter.SendBagData(ctx, sessionID, bagData)
}

// HandleAddItem 处理添加物品请求（RPC，来自 DungeonServer）
func (c *BagController) HandleAddItem(ctx context.Context, sessionID string, data []byte) error {
	// 检查系统是否开启
	bagSys := system.GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "背包系统未开启")
	}

	var req protocol.D2GAddItemReq
	if err := proto.Unmarshal(data, &req); err != nil {
		return customerr.Wrap(err)
	}

	// 执行添加物品用例
	err := c.addItemUseCase.Execute(ctx, req.RoleId, req.ItemId, req.Count, 0) // bind=0 表示非绑定
	if err != nil {
		return err
	}

	// 获取更新后的背包数据并发送
	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送背包数据更新
	return c.presenter.SendBagData(ctx, sessionID, bagData)
}
