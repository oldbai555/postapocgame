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

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysEquip), func() iface.ISystem {
		return NewEquipSystemAdapter()
	})

	// 订阅玩家事件（订阅到全局模板）
	gevent2.SubscribePlayerEvent(gevent2.OnEquipChange, func(ctx context.Context, ev *event.Event) {
		// 装备变更时标记属性系统需要重算
		attrSys := GetAttrSys(ctx)
		if attrSys != nil {
			attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
		}
	})

	// 订阅装备升级事件
	gevent2.SubscribePlayerEvent(gevent2.OnEquipUpgrade, func(ctx context.Context, ev *event.Event) {
		// 装备升级时标记属性系统需要重算
		attrSys := GetAttrSys(ctx)
		if attrSys != nil {
			attrSys.MarkDirty(uint32(protocol.SaAttrSys_SaEquip))
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
