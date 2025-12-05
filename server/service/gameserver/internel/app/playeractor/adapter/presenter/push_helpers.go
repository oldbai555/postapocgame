package presenter

import (
	"context"
	system2 "postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"

	"postapocgame/server/pkg/log"
)

// PushBagData 推送背包数据到客户端
func PushBagData(ctx context.Context, sessionID string) {
	bagSys := system2.GetBagSys(ctx)
	if bagSys == nil {
		return
	}
	bagData, err := bagSys.GetBagData(ctx)
	if err != nil {
		log.Warnf("get bag data failed: %v", err)
		return
	}
	bagPresenter := NewBagPresenter(deps.NetworkGateway())
	if err := bagPresenter.SendBagData(ctx, sessionID, bagData); err != nil {
		log.Warnf("send bag data failed: %v", err)
	}
}

// PushMoneyData 推送货币数据到客户端
func PushMoneyData(ctx context.Context, sessionID string) {
	moneySys := system2.GetMoneySys(ctx)
	if moneySys == nil {
		return
	}
	moneyData, err := moneySys.GetMoneyData(ctx)
	if err != nil {
		log.Warnf("get money data failed: %v", err)
		return
	}
	moneyPresenter := NewMoneyPresenter(deps.NetworkGateway())
	if err := moneyPresenter.SendMoneyData(ctx, sessionID, moneyData); err != nil {
		log.Warnf("send money data failed: %v", err)
	}
}
