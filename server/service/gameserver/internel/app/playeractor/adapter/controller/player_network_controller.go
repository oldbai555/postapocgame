package controller

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/router"
	gatewaylink2 "postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/manager"
	"postapocgame/server/service/gameserver/internel/app/playeractor/adapter/system"
	"postapocgame/server/service/gameserver/internel/app/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entity"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/gshare"
)

func logInfo(ctx context.Context, format string, v ...interface{}) {
	gshare.InfofCtx(ctx, format, v...)
}

func logError(ctx context.Context, format string, v ...interface{}) {
	gshare.ErrorfCtx(ctx, format, v...)
}

func newSessionContext(sessionId string) context.Context {
	if sessionId == "" {
		return context.Background()
	}
	ctx := context.Background()
	return context.WithValue(ctx, gshare.ContextKeySession, sessionId)
}

// HandleEnterGame 处理进入游戏
func HandleEnterGame(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleSelectRole: SessionId=%s", sessionId)

	var req protocol.C2SEnterGameReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		logError(ctx, "unmarshal select player role request failed: %v", err)
		return err
	}

	session := gatewaylink2.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	dbPlayer, err := database.GetPlayerByID(uint(req.RoleId))
	if err != nil {
		logError(ctx, "player not found: RoleId=%d, err=%v", req.RoleId, err)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不存在")
	}

	if dbPlayer.AccountID != session.GetAccountID() {
		logError(ctx, "role not belong to account: RoleId=%d, AccountID=%d, SessionAccountID=%d",
			req.RoleId, dbPlayer.AccountID, session.GetAccountID())
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不属于当前账号")
	}

	selectedRole := &protocol.PlayerSimpleData{
		RoleId:   uint64(dbPlayer.ID),
		Job:      uint32(dbPlayer.Job),
		Sex:      uint32(dbPlayer.Sex),
		RoleName: dbPlayer.RoleName,
		Level:    uint32(dbPlayer.Level),
	}

	logInfo(ctx, "Selected player role: RoleId=%d, Name=%s", selectedRole.RoleId, selectedRole.RoleName)

	if err := enterGame(sessionId, selectedRole); err != nil {
		logError(ctx, "enterGame failed: %v", err)
		return err
	}
	return nil
}

// HandleReconnect 处理重连（预留）
func HandleReconnect(_ context.Context, _ *network.ClientMessage) error {
	return nil
}

func enterGame(sessionId string, roleInfo *protocol.PlayerSimpleData) error {
	ctx := newSessionContext(sessionId)
	if roleInfo != nil {
		ctx = context.WithValue(ctx, gshare.ContextKeyRole, roleInfo.RoleId)
	}
	logInfo(ctx, "enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

	playerRole := entity.NewPlayerRole(sessionId, roleInfo)
	if playerRole == nil {
		logError(ctx, "create player role failed")
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "create player role failed")
	}

	deps.PlayerRoleManager().Add(playerRole)
	session := gatewaylink2.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}
	session.SetRoleId(playerRole.GetPlayerRoleId())

	srvType := uint8(protocol.SrvType_SrvTypeDungeonServer)
	playerRole.SetDungeonSrvType(srvType)

	if err := playerRole.OnLogin(); err != nil {
		logError(ctx, "player OnLogin failed: %v", err)
		return customerr.Wrap(err)
	}

	roleCtx := playerRole.WithContext(context.Background())
	var syncAttrData *protocol.SyncAttrData
	if calc := playerRole.GetAttrCalculator(); calc != nil {
		allAttrs := calc.CalculateAllAttrs(roleCtx)
		if len(allAttrs) > 0 {
			syncAttrData = &protocol.SyncAttrData{AttrData: allAttrs}
			calc.PushSyncDataToClient(roleCtx, syncAttrData)
		}
	}

	skillSys := system.GetSkillSys(roleCtx)
	var skillMap map[uint32]uint32
	if skillSys != nil {
		if m, err := skillSys.GetSkillMap(roleCtx); err == nil {
			skillMap = m
		} else {
			skillMap = make(map[uint32]uint32)
		}
	} else {
		skillMap = make(map[uint32]uint32)
	}

	reqData, err := proto.Marshal(&protocol.G2DEnterDungeonReq{
		SessionId:    sessionId,
		PlatformId:   gshare.GetPlatformId(),
		SrvId:        gshare.GetSrvId(),
		SimpleData:   roleInfo,
		SyncAttrData: syncAttrData,
		SkillMap:     skillMap,
	})
	if err != nil {
		return customerr.Wrap(err)
	}

	if err := deps.DungeonServerGateway().AsyncCall(context.Background(), sessionId, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdEnterDungeon), reqData); err != nil {
		logError(ctx, "call dungeon service enter scene failed: %v", err)
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}

	return nil
}

// HandlePlayerMessageMsg 处理玩家离线消息
func HandlePlayerMessageMsg(message actor.IActorMessage) {
	ctx := message.GetContext()
	var msg protocol.AddActorMessageMsg
	if err := proto.Unmarshal(message.GetData(), &msg); err != nil {
		logError(ctx, "handlePlayerMessageMsg: unmarshal failed: %v", err)
		return
	}

	playerRole := manager.GetPlayerRole(msg.RoleId)
	if playerRole == nil {
		if err := database.SavePlayerActorMessage(msg.RoleId, msg.MsgType, msg.MsgData); err != nil {
			logError(ctx, "handlePlayerMessageMsg: fallback save failed: %v", err)
		}
		return
	}

	roleCtx := playerRole.WithContext(ctx)
	if err := entitysystem.DispatchPlayerMessage(playerRole, msg.MsgType, msg.MsgData); err != nil {
		logError(roleCtx, "handlePlayerMessageMsg: dispatch failed: %v", err)
		if err := database.SavePlayerActorMessage(msg.RoleId, msg.MsgType, msg.MsgData); err != nil {
			logError(roleCtx, "handlePlayerMessageMsg: fallback save failed: %v", err)
		}
	}
}

// HandleRunOneMsg 驱动玩家 Actor RunOne
func HandleRunOneMsg(message actor.IActorMessage) {
	sessionId := message.GetContext().Value(gshare.ContextKeySession).(string)
	session := gatewaylink2.GetSession(sessionId)
	if session == nil {
		return
	}
	iPlayerRole := deps.PlayerRoleManager().GetBySession(sessionId)
	if iPlayerRole == nil {
		return
	}
	iPlayerRole.RunOne()
}

// HandleQueryRank 处理排行榜查询
func HandleQueryRank(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole := deps.PlayerRoleManager().GetBySession(sessionId)
	if playerRole == nil {
		log.Errorf("handleQueryRank: player not found for session=%s", sessionId)
		return gatewaylink2.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SQueryRankReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		log.Errorf("handleQueryRank: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	topN := req.TopN
	if topN <= 0 || topN > 100 {
		topN = 100
	}

	queryMsg := &protocol.QueryRankReqMsg{
		RankType:           req.RankType,
		TopN:               topN,
		RequesterSessionId: sessionId,
		RequesterRoleId:    roleId,
	}
	msgData, err := proto.Marshal(queryMsg)
	if err != nil {
		log.Errorf("handleQueryRank: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdQueryRank), msgData)
	if err := deps.PublicActorGateway().SendMessageAsync(ctx, "global", actorMsg); err != nil {
		log.Errorf("handleQueryRank: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// HandleSyncPosition 处理副本坐标同步
func HandleSyncPosition(_ context.Context, _ string, data []byte) error {
	var req protocol.D2GSyncPositionReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal sync position request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Debugf("received position sync: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)

	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Warnf("player role not found for position sync: RoleId=%d", req.RoleId)
		return nil
	}

	log.Debugf("position synced: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)
	return nil
}

// HandleDungeonSyncAttrs 处理副本属性同步
func HandleDungeonSyncAttrs(_ context.Context, _ string, data []byte) error {
	var req protocol.D2GSyncAttrsReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal dungeon sync attrs failed: %v", err)
		return customerr.Wrap(err)
	}
	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Warnf("player role not found for dungeon sync attrs: RoleId=%d", req.RoleId)
		return nil
	}
	if pr, ok := playerRole.(*entity.PlayerRole); ok {
		if calc := pr.GetAttrCalculator(); calc != nil {
			calc.ApplyDungeonSyncData(req.SyncData)
		}
	} else {
		log.Warnf("attr calculator not found for RoleId=%d", req.RoleId)
	}
	return nil
}

// HandleSendToClient 统一的 S2C 透传
func HandleSendToClient(message actor.IActorMessage) {
	var req protocol.PlayerActorMsgIdSendToClientReq
	if err := proto.Unmarshal(message.GetData(), &req); err != nil {
		log.Errorf("[player-network] handleSendToClient: unmarshal failed: %v", err)
		return
	}

	sessionID, _ := message.GetContext().Value(gshare.ContextKeySession).(string)
	if sessionID == "" {
		log.Warnf("[player-network] handleSendToClient: missing session id")
		return
	}

	if err := gatewaylink2.SendToSession(sessionID, uint16(req.GetMsgId()), req.GetData()); err != nil {
		log.Errorf("[player-network] handleSendToClient: send failed: %v", err)
	}
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, _ *event.Event) {
		protocolRouter := router.NewProtocolRouterController()
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoNetworkMsg), protocolRouter.HandleDoNetworkMsg)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdDoRunOneMsg), HandleRunOneMsg)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdPlayerMessageMsg), HandlePlayerMessageMsg)
		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSendToClient), HandleSendToClient)

		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SEnterGame), HandleEnterGame)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SReconnect), HandleReconnect)
		router.RegisterProtocolHandler(uint16(protocol.C2SProtocol_C2SQueryRank), HandleQueryRank)

		gshare.RegisterHandler(uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSyncAttrs), func(message actor.IActorMessage) {
			msgCtx := message.GetContext()
			sessionID, _ := msgCtx.Value(gshare.ContextKeySession).(string)
			if err := HandleDungeonSyncAttrs(msgCtx, sessionID, message.GetData()); err != nil {
				log.Errorf("[player-network] handleDungeonSyncAttrs failed: %v", err)
			}
		})
	})
}
