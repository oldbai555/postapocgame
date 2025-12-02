package controller

import (
	"context"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/presenter"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/dailyactivity"
	"postapocgame/server/service/gameserver/internel/usecase/money"
	vipusecase "postapocgame/server/service/gameserver/internel/usecase/vip"
)

// MoneyController 货币控制器
type MoneyController struct {
	addMoneyUseCase     *money.AddMoneyUseCase
	consumeMoneyUseCase *money.ConsumeMoneyUseCase
	presenter           *presenter.MoneyPresenter
}

// NewMoneyController 创建货币控制器
func NewMoneyController() *MoneyController {
	container := di.GetContainer()
	addMoneyUC := money.NewAddMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	// 为特殊货币注入对应用例（VIP 经验 + 活跃点）
	vipUC := vipusecase.NewVipMoneyUseCaseImpl(container.PlayerGateway(), container.ConfigGateway())
	activeUC := dailyactivity.NewPointsUseCase(container.PlayerGateway(), container.EventPublisher())
	addMoneyUC.SetDependencies(nil, vipUC, activeUC)
	consumeUC := money.NewConsumeMoneyUseCase(container.PlayerGateway(), container.EventPublisher())
	consumeUC.SetDependencies(nil, activeUC)
	return &MoneyController{
		addMoneyUseCase:     addMoneyUC,
		consumeMoneyUseCase: consumeUC,
		presenter:           presenter.NewMoneyPresenter(container.NetworkGateway()),
	}
}

// HandleOpenMoney 处理打开货币界面请求
func (c *MoneyController) HandleOpenMoney(ctx context.Context, msg *network.ClientMessage) error {
	sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 获取货币数据
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	var moneyData *protocol.SiMoneyData
	if binaryData != nil && binaryData.MoneyData != nil {
		moneyData = binaryData.MoneyData
	} else {
		moneyData = &protocol.SiMoneyData{
			MoneyMap: make(map[uint32]int64),
		}
	}

	// 通过 Presenter 发送响应
	return c.presenter.SendMoneyData(ctx, sessionID, moneyData)
}
