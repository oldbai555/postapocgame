package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	money2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/money"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// MoneyController 货币控制器
type MoneyController struct {
	addMoneyUseCase     *money2.AddMoneyUseCase
	consumeMoneyUseCase *money2.ConsumeMoneyUseCase
	presenter           *presenter.MoneyPresenter
}

// NewMoneyController 创建货币控制器
func NewMoneyController() *MoneyController {
	addMoneyUC := money2.NewAddMoneyUseCase(deps.PlayerGateway(), deps.EventPublisher())
	consumeUC := money2.NewConsumeMoneyUseCase(deps.PlayerGateway(), deps.EventPublisher())
	return &MoneyController{
		addMoneyUseCase:     addMoneyUC,
		consumeMoneyUseCase: consumeUC,
		presenter:           presenter.NewMoneyPresenter(deps.NetworkGateway()),
	}
}

// HandleOpenMoney 处理打开货币界面请求
func (c *MoneyController) HandleOpenMoney(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	moneySys := system.GetMoneySys(ctx)
	if moneySys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "货币系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 获取货币数据
	moneyData, err := deps.PlayerGateway().GetMoneyData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送响应
	return c.presenter.SendMoneyData(ctx, sessionID, moneyData)
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		moneyController := NewMoneyController()
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SOpenMoney), moneyController.HandleOpenMoney)
	})
}
