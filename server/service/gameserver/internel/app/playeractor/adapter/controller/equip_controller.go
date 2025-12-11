package controller

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/equip"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// EquipController 装备控制器
type EquipController struct {
	equipItemUseCase   *equip.EquipItemUseCase
	unEquipItemUseCase *equip.UnEquipItemUseCase
	presenter          *presenter.EquipPresenter
}

// NewEquipController 创建装备控制器
func NewEquipController() *EquipController {
	equipItemUC := equip.NewEquipItemUseCase(deps.PlayerGateway(), deps.EventPublisher(), deps.ConfigGateway())
	unEquipItemUC := equip.NewUnEquipItemUseCase(deps.PlayerGateway(), deps.EventPublisher())

	// 注入 BagUseCase 依赖（通过适配器）
	bagUseCase := system.NewBagUseCaseAdapter()
	equipItemUC.SetDependencies(bagUseCase)
	unEquipItemUC.SetDependencies(bagUseCase)

	return &EquipController{
		equipItemUseCase:   equipItemUC,
		unEquipItemUseCase: unEquipItemUC,
		presenter:          presenter.NewEquipPresenter(deps.NetworkGateway()),
	}
}

// HandleEquipItem 处理装备物品请求
func (c *EquipController) HandleEquipItem(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	equipSys := system.GetEquipSys(ctx)
	if equipSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "装备系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	var req protocol.C2SEquipItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return customerr.Wrap(err)
	}

	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 执行装备物品用例
	err = c.equipItemUseCase.Execute(ctx, roleID, req.ItemId, req.Slot)

	// 构建响应
	resp := &protocol.S2CEquipResultReq{
		Slot:   req.Slot,
		ItemId: req.ItemId,
	}

	// 发送响应
	if sendErr := c.presenter.SendEquipResult(ctx, sessionID, resp); sendErr != nil {
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
	return nil
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		equipController := NewEquipController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEquipItem), equipController.HandleEquipItem)
	})
}
