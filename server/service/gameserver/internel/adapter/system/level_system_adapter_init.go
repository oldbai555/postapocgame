package system

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
	gevent2 "postapocgame/server/service/gameserver/internel/infrastructure/gevent"
)

// 注册系统工厂和属性计算器
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSystemAdapter()
	})

	// 订阅玩家事件（订阅到全局模板）
	gevent2.SubscribePlayerEventH(gevent2.OnPlayerLevelUp, func(ctx context.Context, ev *event.Event) {})
	gevent2.SubscribePlayerEventH(gevent2.OnPlayerExpChange, func(ctx context.Context, ev *event.Event) {})

	// 注册属性计算器
	gameattrcalc.Register(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) gameattrcalc.Calculator {
		return GetLevelSys(ctx)
	})

	// 注册属性加成计算器
	gameattrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) gameattrcalc.AddRateCalculator {
		return &levelAddRateCalculator{}
	})
}
