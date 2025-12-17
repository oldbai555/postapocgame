package presenter

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/money"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"

	"postapocgame/server/pkg/log"
)

// PushBagData 推送背包数据到客户端
func PushBagData(ctx context.Context, sessionID string) {
	bagSys := bag.GetBagSys(ctx)
	if bagSys == nil {
		return
	}
	bagData, err := bagSys.GetBagData(ctx)
	if err != nil {
		log.Warnf("get bag data failed: %v", err)
		return
	}
	// 从 context 获取 Runtime
	rt := runtime.FromContext(ctx)
	if rt == nil {
		log.Warnf("runtime not found in context, cannot send bag data")
		return
	}
	bagPresenter := bag.NewBagPresenter(rt.NetworkGateway())
	if err := bagPresenter.SendBagData(ctx, sessionID, bagData); err != nil {
		log.Warnf("send bag data failed: %v", err)
	}
}

// PushMoneyData 推送货币数据到客户端
func PushMoneyData(ctx context.Context, sessionID string) {
	moneySys := money.GetMoneySys(ctx)
	if moneySys == nil {
		return
	}
	moneyData, err := moneySys.GetMoneyData(ctx)
	if err != nil {
		log.Warnf("get money data failed: %v", err)
		return
	}
	// 从 context 获取 Runtime
	rt := runtime.FromContext(ctx)
	if rt == nil {
		log.Warnf("runtime not found in context, cannot send money data")
		return
	}
	moneyPresenter := money.NewMoneyPresenter(rt.NetworkGateway())
	if err := moneyPresenter.SendMoneyData(ctx, sessionID, moneyData); err != nil {
		log.Warnf("send money data failed: %v", err)
	}
}
