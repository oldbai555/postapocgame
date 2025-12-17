package money

import (
	"context"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// MoneyController 货币控制器
type MoneyController struct {
	addMoneyUseCase     *AddMoneyUseCase
	consumeMoneyUseCase *ConsumeMoneyUseCase
	presenter           *MoneyPresenter
}

// NewMoneyController 创建货币控制器
func NewMoneyController(rt *runtime.Runtime) *MoneyController {
	d := depsFromRuntime(rt)
	addMoneyUC := NewAddMoneyUseCase(d)
	consumeUC := NewConsumeMoneyUseCase(d)
	return &MoneyController{
		addMoneyUseCase:     addMoneyUC,
		consumeMoneyUseCase: consumeUC,
		presenter:           NewMoneyPresenter(d.NetworkGateway),
	}
}

// HandleOpenMoney 处理打开货币界面请求
func (c *MoneyController) HandleOpenMoney(ctx context.Context, msg *network.ClientMessage) error {
	// 检查系统是否开启
	moneySys := GetMoneySys(ctx)
	if moneySys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "货币系统未开启")
	}

	sessionID, err := gshare.GetSessionIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 从 Money 系统获取货币数据
	moneyData, err := moneySys.GetMoneyData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 通过 Presenter 发送响应
	return c.presenter.SendMoneyData(ctx, sessionID, moneyData)
}

// init() 函数已移除，注册逻辑迁移至 playeractor/register.RegisterAll()
