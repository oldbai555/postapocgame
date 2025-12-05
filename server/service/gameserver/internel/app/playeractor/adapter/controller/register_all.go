package controller

import (
	"context"
	"postapocgame/server/service/gameserver/internel/gevent"

	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	"postapocgame/server/service/gameserver/internel/gshare"
)

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		registerProtocolRouter()
		registerControllers()
	})
}

// registerProtocolRouter 注册通用的网络与内部 Actor 消息入口
func registerProtocolRouter() {
	protocolRouter := router.NewProtocolRouterController()
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoNetworkMsg), protocolRouter.HandleDoNetworkMsg)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoRunOneMsg), HandleRunOneMsg)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdPlayerMessageMsg), HandlePlayerMessageMsg)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSendToClient), HandleSendToClient)

	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterGame), HandleEnterGame)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SReconnect), HandleReconnect)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SQueryRank), HandleQueryRank)

	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSyncPosition), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := HandleSyncPosition(msgCtx, sessionID, message.GetData()); err != nil {
			log.Errorf("[player-network] handleSyncPosition failed: %v", err)
		}
	})
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSyncAttrs), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := HandleDungeonSyncAttrs(msgCtx, sessionID, message.GetData()); err != nil {
			log.Errorf("[player-network] handleDungeonSyncAttrs failed: %v", err)
		}
	})
}

// registerControllers 将所有 Controller 注册到统一路由
func registerControllers() {
	bagController := NewBagController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SOpenBag), bagController.HandleOpenBag)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdAddItem), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := bagController.HandleAddItem(msgCtx, sessionID, message.GetData()); err != nil {
			log.Errorf("[bag-controller] HandleAddItem failed: %v", err)
		}
	})

	chatController := NewChatController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SChatWorld), chatController.HandleWorldChat)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SChatPrivate), chatController.HandlePrivateChat)

	dungeonItemController := NewDungeonItemController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SPickupItem), dungeonItemController.HandlePickupItem)

	equipController := NewEquipController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEquipItem), equipController.HandleEquipItem)

	fubenController := NewFubenController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterDungeon), fubenController.HandleEnterDungeon)
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSettleDungeon), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := fubenController.HandleSettleDungeon(msgCtx, sessionID, message.GetData()); err != nil {
			log.Errorf("[fuben-controller] HandleSettleDungeon failed: %v", err)
		}
	})
	gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdEnterDungeonSuccess), func(message actor.IActorMessage) {
		msgCtx := message.GetContext()
		sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
		if err := fubenController.HandleEnterDungeonSuccess(msgCtx, sessionID, message.GetData()); err != nil {
			log.Errorf("[fuben-controller] HandleEnterDungeonSuccess failed: %v", err)
		}
	})

	itemUseController := NewItemUseController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseItem), itemUseController.HandleUseItem)

	moneyController := NewMoneyController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SOpenMoney), moneyController.HandleOpenMoney)

	moveController := NewMoveController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SStartMove), moveController.HandleStartMove)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUpdateMove), moveController.HandleUpdateMove)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEndMove), moveController.HandleEndMove)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SChangeScene), moveController.HandleChangeScene)

	accountController := NewPlayerAccountController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRegister), accountController.HandleRegister)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SLogin), accountController.HandleLogin)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SVerify), accountController.HandleVerify)

	roleController := NewPlayerRoleController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SQueryRoles), roleController.HandleQueryRoles)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SCreateRole), roleController.HandleCreateRole)

	questController := NewQuestController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2STalkToNPC), questController.HandleTalkToNPC)

	recycleController := NewRecycleController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRecycleItem), recycleController.HandleRecycleItem)

	reviveController := NewReviveController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SRevive), reviveController.HandleRevive)

	shopController := NewShopController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SShopBuy), shopController.HandleShopBuy)

	skillController := NewSkillController()
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SLearnSkill), skillController.HandleLearnSkill)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUpgradeSkill), skillController.HandleUpgradeSkill)
	router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SUseSkill), skillController.HandleUseSkill)
}
