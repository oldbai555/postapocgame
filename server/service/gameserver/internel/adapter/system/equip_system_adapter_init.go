package system

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	gameattrcalc "postapocgame/server/service/gameserver/internel/adapter/system/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/infrastructure/gevent"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysEquip), func() iface.ISystem {
		return NewEquipSystemAdapter()
	})

	// 订阅玩家事件（订阅到全局模板）
	// 注意：装备变更/升级事件的处理属于框架层面的状态管理（标记属性系统需要重算），
	// 符合 Clean Architecture 原则，保留在 SystemAdapter 层
	gevent.SubscribePlayerEvent(gevent.OnEquipChange, func(ctx context.Context, ev *event.Event) {
		// 装备变更时标记属性系统需要重算（框架状态管理，非业务逻辑）
		playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
		if err != nil {
			return
		}
		attrCalcRaw := playerRole.GetAttrCalculator()
		if attrCalc, ok := attrCalcRaw.(interfaces.IAttrCalculator); ok && attrCalc != nil {
			attrCalc.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
		}
	})

	// 订阅装备升级事件
	gevent.SubscribePlayerEvent(gevent.OnEquipUpgrade, func(ctx context.Context, ev *event.Event) {
		// 装备升级时标记属性系统需要重算（框架状态管理，非业务逻辑）
		playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
		if err != nil {
			return
		}
		attrCalcRaw := playerRole.GetAttrCalculator()
		if attrCalc, ok := attrCalcRaw.(interfaces.IAttrCalculator); ok && attrCalc != nil {
			attrCalc.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
		}
	})

	// 注册属性计算器
	gameattrcalc.Register(uint32(protocol.SaAttrSys_SaEquip), func(ctx context.Context) gameattrcalc.Calculator {
		return GetEquipSys(ctx)
	})

	// 注册属性加成计算器（暂时不实现，等需要时再添加）
	// gameattrcalc.RegisterAddRate(uint32(protocol.SaAttrSys_SaEquip), func(ctx context.Context) gameattrcalc.AddRateCalculator {
	// 	return &equipAddRateCalculator{}
	// })

	// 协议注册由 controller 包负责
}
