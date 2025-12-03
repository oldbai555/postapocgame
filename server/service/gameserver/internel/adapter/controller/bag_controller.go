package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/adapter/system"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/bag"
)

// BagController 背包控制器
type BagController struct {
	addItemUseCase    *bag.AddItemUseCase
	removeItemUseCase *bag.RemoveItemUseCase
	hasItemUseCase    *bag.HasItemUseCase
	presenter         *presenter.BagPresenter
}

// NewBagController 创建背包控制器
func NewBagController() *BagController {
	container := di.GetContainer()
	return &BagController{
		addItemUseCase:    bag.NewAddItemUseCase(container.PlayerGateway(), container.EventPublisher(), container.ConfigGateway()),
		removeItemUseCase: bag.NewRemoveItemUseCase(container.PlayerGateway(), container.EventPublisher()),
		hasItemUseCase:    bag.NewHasItemUseCase(container.PlayerGateway()),
		presenter:         presenter.NewBagPresenter(container.NetworkGateway()),
	}
}

// HandleOpenBag 处理打开背包请求
func (c *BagController) HandleOpenBag(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	bagSys := system.GetBagSys(ctx)
	if bagSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "背包系统未开启")
	}

	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 获取背包数据
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	var bagData *protocol.SiBagData
	if binaryData != nil && binaryData.BagData != nil {
		bagData = binaryData.BagData
	} else {
		bagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
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
	roleID := req.RoleId
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	var bagData *protocol.SiBagData
	if binaryData != nil && binaryData.BagData != nil {
		bagData = binaryData.BagData
	} else {
		bagData = &protocol.SiBagData{
			Items: make([]*protocol.ItemSt, 0),
		}
	}

	// 通过 Presenter 发送背包数据更新
	return c.presenter.SendBagData(ctx, sessionID, bagData)
}
