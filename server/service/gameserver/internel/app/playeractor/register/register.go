package register

import (
	"context"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/bag"
	"postapocgame/server/service/gameserver/internel/app/playeractor/equip"
	"postapocgame/server/service/gameserver/internel/app/playeractor/fuben"
	"postapocgame/server/service/gameserver/internel/app/playeractor/level"
	"postapocgame/server/service/gameserver/internel/app/playeractor/money"
	"postapocgame/server/service/gameserver/internel/app/playeractor/recycle"
	"postapocgame/server/service/gameserver/internel/app/playeractor/router"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/skill"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
)

// RegisterAll 显式注册所有 PlayerActor 系统（替代各包的 init()）
// 参数 rt 为 Runtime 实例，用于创建各系统的 Controller 和 SystemAdapter
func RegisterAll(rt *runtime.Runtime) {
	// 订阅服务器启动事件，注册所有协议处理器
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		registerBagHandlers(rt)
		registerMoneyHandlers(rt)
		registerEquipHandlers(rt)
		registerSkillHandlers(rt)
		registerFubenHandlers(rt)
		registerRecycleHandlers(rt)
	})

	// 注册所有系统工厂
	level.RegisterSystemFactory(rt)
	bag.RegisterSystemFactory(rt)
	money.RegisterSystemFactory(rt)
	equip.RegisterSystemFactory(rt)
	skill.RegisterSystemFactory(rt)
	fuben.RegisterSystemFactory(rt)
	recycle.RegisterSystemFactory(rt)
}

// registerBagHandlers 注册背包相关协议处理器
func registerBagHandlers(rt *runtime.Runtime) {
	bagController := bag.NewBagController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SOpenBag), bagController.HandleOpenBag)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdAddItem), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := bagController.HandleAddItem(msgCtx, sessionID, message.GetData()); err != nil {
			// log.Errorf("[bag-controller] HandleAddItem failed: %v", err) // 日志记录已在 HandleAddItem 内部处理
		}
	})
}

// registerMoneyHandlers 注册货币相关协议处理器
func registerMoneyHandlers(rt *runtime.Runtime) {
	moneyController := money.NewMoneyController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SOpenMoney), moneyController.HandleOpenMoney)
}

// registerEquipHandlers 注册装备相关协议处理器
func registerEquipHandlers(rt *runtime.Runtime) {
	equipController := equip.NewEquipController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEquipItem), equipController.HandleEquipItem)
	// UnEquip 协议注册已被移除或未实现
}

// registerSkillHandlers 注册技能相关协议处理器
func registerSkillHandlers(rt *runtime.Runtime) {
	skillController := skill.NewSkillController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SLearnSkill), skillController.HandleLearnSkill)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUpgradeSkill), skillController.HandleUpgradeSkill)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseSkill), skillController.HandleUseSkill)
}

// registerFubenHandlers 注册副本相关协议处理器
func registerFubenHandlers(rt *runtime.Runtime) {
	fubenController := fuben.NewFubenController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterDungeon), fubenController.HandleEnterDungeon)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSettleDungeon), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := fubenController.HandleSettleDungeon(msgCtx, sessionID, message.GetData()); err != nil {
			// log.Errorf("[fuben-controller] HandleSettleDungeon failed: %v", err) // 日志记录已在 HandleSettleDungeon 内部处理
		}
	})
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdEnterDungeonSuccess), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := fubenController.HandleEnterDungeonSuccess(msgCtx, sessionID, message.GetData()); err != nil {
			// log.Errorf("[fuben-controller] HandleEnterDungeonSuccess failed: %v", err) // 日志记录已在 HandleEnterDungeonSuccess 内部处理
		}
	})
}

// registerRecycleHandlers 注册回收相关协议处理器
func registerRecycleHandlers(rt *runtime.Runtime) {
	recycleController := recycle.NewRecycleController(rt)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRecycleItem), recycleController.HandleRecycleItem)
}
