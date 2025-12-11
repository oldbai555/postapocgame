package controller

import (
	"context"
	"postapocgame/server/internal/event"
	presenter "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	system "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/item_use"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

// ItemUseController 物品使用控制器
type ItemUseController struct {
	useItemUseCase *item_use.UseItemUseCase
	presenter      *presenter.ItemUsePresenter
}

// NewItemUseController 创建物品使用控制器
func NewItemUseController() *ItemUseController {
	useItemUC := item_use.NewUseItemUseCase(deps.PlayerGateway(), deps.ConfigGateway(), deps.DungeonServerGateway())

	// 注入依赖
	bagUseCase := system.NewBagUseCaseAdapter()
	levelUseCase := system.NewLevelUseCaseAdapter()
	useItemUC.SetDependencies(bagUseCase, levelUseCase)

	return &ItemUseController{
		useItemUseCase: useItemUC,
		presenter:      presenter.NewItemUsePresenter(deps.NetworkGateway()),
	}
}

// HandleUseItem 处理使用物品请求
func (c *ItemUseController) HandleUseItem(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	itemUseSys := system.GetItemUseSys(ctx)
	if itemUseSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "物品使用系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SUseItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 默认使用数量为1
	if req.Count == 0 {
		req.Count = 1
	}

	// 执行使用物品用例
	err = c.useItemUseCase.Execute(ctx, roleID, req.ItemId, req.Count)

	// 构建响应
	resp := &protocol.S2CUseItemResultReq{
		Success:        err == nil,
		ItemId:         req.ItemId,
		RemainingCount: 0,
	}

	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "使用成功"
		// 获取剩余数量（通过 BagUseCase）
		bagUseCase := system.NewBagUseCaseAdapter()
		item, _ := bagUseCase.GetItem(ctx, roleID, req.ItemId)
		if item != nil {
			resp.RemainingCount = item.Count
		}
	}

	// 发送响应
	if sendErr := c.presenter.SendUseItemResult(ctx, sessionID, resp); sendErr != nil {
		return sendErr
	}

	// 如果成功，推送背包数据更新
	if err != nil {
		return customerr.Wrap(err)
	}

	bagData, err := deps.PlayerGateway().GetBagData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}
	bagPresenter := presenter.NewBagPresenter(deps.NetworkGateway())
	err = bagPresenter.SendBagData(ctx, sessionID, bagData)
	if err != nil {
		return customerr.Wrap(err)
	}

	return err
}
func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		itemUseController := NewItemUseController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseItem), itemUseController.HandleUseItem)
	})
}
