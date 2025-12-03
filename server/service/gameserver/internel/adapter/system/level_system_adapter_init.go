package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// 注册系统工厂和属性计算器
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSystemAdapter()
	})

	// 注意：等级系统的事件（OnPlayerLevelUp/OnPlayerExpChange）由 UseCase 层发布，
	// 其他系统如需响应这些事件，应在各自的 UseCase 中订阅，不在 SystemAdapter 层处理

	// 注册属性计算器
	gameattrcalc.Register(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) gameattrcalc.Calculator {
		return GetLevelSys(ctx)
	})

	// 注册属性加成计算器
	gameattrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaLevel), func(ctx context.Context) gameattrcalc.AddRateCalculator {
		return &levelAddRateCalculator{}
	})
}
