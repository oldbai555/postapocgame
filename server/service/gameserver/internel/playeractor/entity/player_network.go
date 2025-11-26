/**
 * @Author: zjj
 * @Date: 2025/11/11
 * @Desc:
**/

package entity

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
)

func logInfo(ctx context.Context, format string, v ...interface{}) {
	gshare.InfofCtx(ctx, format, v...)
}

func logWarn(ctx context.Context, format string, v ...interface{}) {
	gshare.WarnfCtx(ctx, format, v...)
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

// handleRegister 处理账号注册
func handleRegister(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleRegister: SessionId=%s", sessionId)

	// 解析注册请求
	var req protocol.C2SRegisterReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		logError(ctx, "unmarshal register request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 验证用户名和密码
	if len(req.Username) < 3 || len(req.Username) > 32 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRegisterResult), &protocol.S2CRegisterResultReq{
			Success: false,
			Message: "用户名长度必须在3-32个字符之间",
		})
	}
	if len(req.Password) < 6 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRegisterResult), &protocol.S2CRegisterResultReq{
			Success: false,
			Message: "密码长度至少6个字符",
		})
	}

	// 检查用户名是否已存在
	_, err = database.GetAccountByUsername(req.Username)
	if err == nil {
		// 用户名已存在
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRegisterResult), &protocol.S2CRegisterResultReq{
			Success: false,
			Message: "用户名已存在",
		})
	}

	// 创建账号
	account, err := database.CreateAccount(req.Username, req.Password)
	if err != nil {
		logError(ctx, "create account failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRegisterResult), &protocol.S2CRegisterResultReq{
			Success: false,
			Message: "注册失败，请稍后重试",
		})
	}

	// 生成token
	token := database.GenerateToken(account.ID)

	// 设置Session的账号ID和Token
	session := gatewaylink.GetSession(sessionId)
	if session != nil {
		session.SetAccountID(account.ID)
		session.SetToken(token)
	}

	logInfo(ctx, "Account registered: AccountID=%d, Username=%s", account.ID, account.Username)

	// 返回注册成功
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRegisterResult), &protocol.S2CRegisterResultReq{
		Success: true,
		Message: "注册成功",
		Token:   token,
	})
}

// handleLogin 处理账号登录
func handleLogin(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleLogin: SessionId=%s", sessionId)

	// 解析登录请求
	var req protocol.C2SLoginReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		logError(ctx, "unmarshal login request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 查找账号
	account, err := database.GetAccountByUsername(req.Username)
	if err != nil {
		logError(ctx, "account not found: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLoginResult), &protocol.S2CLoginResultReq{
			Success: false,
			Message: "用户名或密码错误",
		})
	}

	// 验证密码
	if !account.CheckPassword(req.Password) {
		logWarn(ctx, "password incorrect for account: %s", req.Username)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLoginResult), &protocol.S2CLoginResultReq{
			Success: false,
			Message: "用户名或密码错误",
		})
	}

	// 生成token
	token := database.GenerateToken(account.ID)

	// 设置Session的账号ID和Token
	session := gatewaylink.GetSession(sessionId)
	if session != nil {
		session.SetAccountID(account.ID)
		session.SetToken(token)
	}

	logInfo(ctx, "Account logged in: AccountID=%d, Username=%s", account.ID, account.Username)

	// 返回登录成功
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLoginResult), &protocol.S2CLoginResultReq{
		Success: true,
		Message: "登录成功",
		Token:   token,
	})
}

func handleVerify(_ context.Context, _ *network.ClientMessage) error {
	return nil
}

func handleQueryRoles(ctx context.Context, _ *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleQueryRoles: SessionId=%s", sessionId)

	// 获取Session
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	// 获取账号ID
	accountID := session.GetAccountID()
	if accountID == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRoleList), &protocol.S2CRoleListReq{
			RoleList: []*protocol.PlayerSimpleData{},
		})
	}

	// 从数据库查询角色列表
	dbPlayers, err := database.GetPlayersByAccountID(accountID)
	if err != nil {
		logError(ctx, "query roles failed: %v", err)
		return customerr.Wrap(err)
	}

	// 转换为协议格式
	roleList := make([]*protocol.PlayerSimpleData, 0, len(dbPlayers))
	for _, dbPlayer := range dbPlayers {
		roleList = append(roleList, &protocol.PlayerSimpleData{
			RoleId:   uint64(dbPlayer.ID),
			Job:      uint32(dbPlayer.Job),
			Sex:      uint32(dbPlayer.Sex),
			RoleName: dbPlayer.RoleName,
			Level:    uint32(dbPlayer.Level),
		})
	}

	resp := &protocol.S2CRoleListReq{
		RoleList: roleList,
	}

	// 发送给客户端
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRoleList), resp)
}

func handleEnterGame(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleSelectRole: SessionId=%s", sessionId)

	// 解析选择角色请求
	var req protocol.C2SEnterGameReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		logError(ctx, "unmarshal select player role request failed: %v", err)
		return err
	}

	// 获取Session
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	// 从数据库查找角色
	dbPlayer, err := database.GetPlayerByID(uint(req.RoleId))
	if err != nil {
		logError(ctx, "player not found: RoleId=%d, err=%v", req.RoleId, err)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不存在")
	}

	// 验证角色是否属于当前账号
	if dbPlayer.AccountID != session.GetAccountID() {
		logError(ctx, "role not belong to account: RoleId=%d, AccountID=%d, SessionAccountID=%d",
			req.RoleId, dbPlayer.AccountID, session.GetAccountID())
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不属于当前账号")
	}

	// 转换为协议格式
	selectedRole := &protocol.PlayerSimpleData{
		RoleId:   uint64(dbPlayer.ID),
		Job:      uint32(dbPlayer.Job),
		Sex:      uint32(dbPlayer.Sex),
		RoleName: dbPlayer.RoleName,
		Level:    uint32(dbPlayer.Level),
	}

	logInfo(ctx, "Selected player role: RoleId=%d, Name=%s", selectedRole.RoleId, selectedRole.RoleName)

	// 进入游戏
	err = enterGame(sessionId, selectedRole)
	if err != nil {
		logError(ctx, "enterGame failed: %v", err)
		return err
	}
	return nil
}

func handleReconnect(_ context.Context, _ *network.ClientMessage) error {
	return nil
}

func handleCreateRole(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	logInfo(ctx, "handleCreateRole: SessionId=%s", sessionId)

	// 解析创建角色请求
	var req protocol.C2SCreateRoleReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		logError(ctx, "unmarshal create role request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取Session
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session not found")
	}

	// 验证是否已登录
	accountID := session.GetAccountID()
	if accountID == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateRoleResult), &protocol.S2CCreateRoleResultReq{
			Job:      0,
			Sex:      0,
			RoleName: "",
		})
	}

	// 验证角色名
	if req.RoleData == nil || len(req.RoleData.RoleName) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateRoleResult), &protocol.S2CCreateRoleResultReq{
			Job:      0,
			Sex:      0,
			RoleName: "",
		})
	}

	// 检查角色名是否已存在
	exists, err := database.CheckRoleNameExists(req.RoleData.RoleName)
	if err != nil {
		logError(ctx, "check role name failed: %v", err)
		return customerr.Wrap(err)
	}
	if exists {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateRoleResult), &protocol.S2CCreateRoleResultReq{
			Job:      0,
			Sex:      0,
			RoleName: "",
		})
	}

	// 检查角色数量限制（每个账号最多3个角色）
	players, err := database.GetPlayersByAccountID(accountID)
	if err != nil {
		logError(ctx, "query players failed: %v", err)
		return customerr.Wrap(err)
	}
	if len(players) >= 3 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateRoleResult), &protocol.S2CCreateRoleResultReq{
			Job:      0,
			Sex:      0,
			RoleName: "",
		})
	}

	// 创建角色
	dbPlayer, err := database.CreatePlayer(accountID, req.RoleData.RoleName, int(req.RoleData.Job), int(req.RoleData.Sex))
	if err != nil {
		logError(ctx, "create player failed: %v", err)
		return customerr.Wrap(err)
	}

	logInfo(ctx, "Player created: AccountID=%d, RoleId=%d, RoleName=%s", accountID, dbPlayer.ID, dbPlayer.RoleName)

	// 返回创建成功
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateRoleResult), &protocol.S2CCreateRoleResultReq{
		Job:      uint32(dbPlayer.Job),
		Sex:      uint32(dbPlayer.Sex),
		RoleName: dbPlayer.RoleName,
	})
}

// enterGame 进入游戏
func enterGame(sessionId string, roleInfo *protocol.PlayerSimpleData) error {
	ctx := newSessionContext(sessionId)
	if roleInfo != nil {
		ctx = context.WithValue(ctx, gshare.ContextKeyRole, roleInfo.RoleId)
	}
	logInfo(ctx, "enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

	// 创建PlayerRole实例
	playerRole := NewPlayerRole(sessionId, roleInfo)

	// 添加到PlayerRole管理器
	manager.GetPlayerRoleManager().Add(playerRole)
	session := gatewaylink.GetSession(sessionId)
	session.SetRoleId(playerRole.GetPlayerRoleId())

	// 设置玩家所在的DungeonServer类型(默认为3)
	srvType := uint8(protocol.SrvType_SrvTypeDungeonServer)
	playerRole.SetDungeonSrvType(srvType)

	// 先调用OnLogin初始化系统
	err := playerRole.OnLogin()
	if err != nil {
		logError(ctx, "player OnLogin failed: %v", err)
		return customerr.Wrap(err)
	}

	// 汇总所有属性（首次登录时计算所有属性）
	roleCtx := playerRole.WithContext(context.Background())
	attrSys := entitysystem.GetAttrSys(roleCtx)
	var syncAttrData *protocol.SyncAttrData
	if attrSys != nil {
		allAttrs := attrSys.CalculateAllAttrs(roleCtx)
		if len(allAttrs) > 0 {
			syncAttrData = &protocol.SyncAttrData{
				AttrData: allAttrs,
			}
			attrSys.PushSyncDataToClient(roleCtx, syncAttrData)
		}
	}

	// 获取技能列表
	skillSys := entitysystem.GetSkillSys(roleCtx)
	var skillMap map[uint32]uint32
	if skillSys != nil {
		skillMap = skillSys.GetSkillMap()
	} else {
		skillMap = make(map[uint32]uint32)
	}

	// 构造进入副本请求

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

	// 使用带SessionId的异步RPC调用
	err = dungeonserverlink.AsyncCall(context.Background(), srvType, sessionId, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
	if err != nil {
		logError(ctx, "call dungeon service enter scene failed: %v", err)
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}

	return nil
}

func handleDoNetWorkMsg(message actor.IActorMessage) {
	sessionId := message.GetContext().Value(gshare.ContextKeySession).(string)
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return
	}

	cliMsg, err := network.DefaultCodec().DecodeClientMessage(message.GetData())
	if err != nil {
		ctx := newSessionContext(sessionId)
		logError(ctx, "decode client message failed: %v", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)

	// 优先检查是否可以在GameServer处理
	getFunc := clientprotocol.GetFunc(cliMsg.MsgId)
	if getFunc != nil {
		// GameServer可以处理此协议
		var buildPlayerRoleCtx = func(ctx context.Context, roleId uint64) context.Context {
			if roleId == 0 {
				return ctx
			}
			pr := manager.GetPlayerRole(roleId)
			if pr == nil {
				return ctx
			}
			return pr.WithContext(ctx)
		}

		roleId := session.GetRoleId()
		ctx = buildPlayerRoleCtx(ctx, roleId)
		err = getFunc(ctx, cliMsg)

		if err == nil {
			return
		}

		logError(ctx, "handleDoNetWorkMsg failed: %v", err)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  err.Error(),
		})
		if err != nil {
			logError(ctx, "send error proto failed: %v", err)
		}
		return
	}

	// GameServer无法处理,检查是否需要转发到DungeonServer
	protocolMgr := dungeonserverlink.GetProtocolManager()
	if !protocolMgr.IsDungeonProtocol(cliMsg.MsgId) {
		// 协议既不在GameServer也不在DungeonServer
		logError(ctx, "protocol %d not found in GameServer or DungeonServer", cliMsg.MsgId)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  fmt.Sprintf("protocol %d not supported", cliMsg.MsgId),
		})
		if err != nil {
			logError(ctx, "send unsupported protocol error failed: %v", err)
		}
		return
	}

	// 需要转发到DungeonServer
	srvType, protocolType, _ := protocolMgr.GetSrvTypeForProtocol(cliMsg.MsgId)

	// 判断转发到哪个DungeonServer
	var targetSrvType uint8
	if protocolType == dungeonserverlink.ProtocolTypeUnique {
		// 独有协议,转发到指定的srvType
		targetSrvType = srvType
		log.Debugf("forwarding protocol %d to unique DungeonServer: srvType=%d", cliMsg.MsgId, targetSrvType)
	} else {
		// 通用协议,需要根据角色所在的DungeonServer来决定
		roleId := session.GetRoleId()
		pr := manager.GetPlayerRole(roleId)
		if pr == nil {
			logError(ctx, "player role not found: roleId=%d", roleId)
			err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "player role not found",
			})
			if err != nil {
				logError(ctx, "send player-not-found error failed: %v", err)
			}
			return
		}

		// 获取角色所在的DungeonServer类型
		targetSrvType = pr.GetDungeonSrvType()
		if targetSrvType == 0 {
			// 如果角色还没有进入DungeonServer,使用协议注册的默认srvType
			targetSrvType = srvType
		}
		log.Debugf("forwarding protocol %d to common DungeonServer: srvType=%d, roleId=%d", cliMsg.MsgId, targetSrvType, roleId)
	}

	// 转发到DungeonServer
	// 将原始消息数据转发(包含完整的ClientMessage)
	err = dungeonserverlink.AsyncCall(ctx, targetSrvType, sessionId, 0, message.GetData())
	if err != nil {
		logError(ctx, "forward to DungeonServer failed: srvType=%d, err:%v", targetSrvType, err)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  fmt.Sprintf("forward to DungeonServer failed: %v", err),
		})
		if err != nil {
			logError(ctx, "send forward error failed: %v", err)
		}
		return
	}

	log.Debugf("successfully forwarded protocol %d to DungeonServer: srvType=%d", cliMsg.MsgId, targetSrvType)
}

func handlePlayerMessageMsg(message actor.IActorMessage) {
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

func handleRunOneMsg(message actor.IActorMessage) {
	sessionId := message.GetContext().Value(gshare.ContextKeySession).(string)
	session := gatewaylink.GetSession(sessionId)
	if session == nil {
		return
	}
	iPlayerRole := manager.GetPlayerRoleManager().GetBySession(sessionId)
	if iPlayerRole == nil {
		return
	}
	iPlayerRole.RunOne()
}

// handleQueryRank 处理查询排行榜
func handleQueryRank(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryRank: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SQueryRankReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleQueryRank: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	topN := req.TopN
	if topN <= 0 || topN > 100 {
		topN = 100
	}

	// 发送到 PublicActor 查询排行榜
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
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleQueryRank: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleRegisterProtocols 处理DungeonServer注册协议的RPC请求
func handleRegisterProtocols(_ context.Context, _ string, data []byte) error {
	var req protocol.D2GRegisterProtocolsReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal register protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol registration from DungeonServer: srvType=%d, protocols=%d", srvType, len(req.Protocols))

	// 转换协议信息
	protocols := make([]struct {
		ProtoId  uint16
		IsCommon bool
	}, len(req.Protocols))

	for i, protocolInfo := range req.Protocols {
		protocols[i].ProtoId = uint16(protocolInfo.ProtoId)
		protocols[i].IsCommon = protocolInfo.IsCommon
		log.Debugf("  - protoId=%d, isCommon=%v", protocolInfo.ProtoId, protocolInfo.IsCommon)
	}

	// 注册到协议管理器
	if err := dungeonserverlink.GetProtocolManager().RegisterProtocols(srvType, protocols); err != nil {
		log.Errorf("register protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully registered %d protocols for srvType=%d", len(protocols), srvType)
	return nil
}

// handleUnregisterProtocols 处理DungeonServer注销协议的RPC请求
func handleUnregisterProtocols(_ context.Context, _ string, data []byte) error {
	var req protocol.D2GUnregisterProtocolsReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal unregister protocols request failed: %v", err)
		return customerr.Wrap(err)
	}

	srvType := uint8(req.SrvType)
	log.Infof("received protocol unregistration from DungeonServer: srvType=%d", srvType)

	// 从协议管理器注销
	if err := dungeonserverlink.GetProtocolManager().UnregisterProtocols(srvType); err != nil {
		log.Errorf("unregister protocols failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("successfully unregistered protocols for srvType=%d", srvType)
	return nil
}

// handleSyncPosition 处理坐标同步的RPC请求
func handleSyncPosition(_ context.Context, _ string, data []byte) error {
	var req protocol.D2GSyncPositionReq
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Errorf("unmarshal sync position request failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Debugf("received position sync: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)

	// 获取玩家角色
	playerRole := manager.GetPlayerRole(req.RoleId)
	if playerRole == nil {
		log.Warnf("player role not found for position sync: RoleId=%d", req.RoleId)
		// 不返回错误，坐标同步失败不影响游戏流程
		return nil
	}

	// 更新角色坐标（如果需要存储的话）
	// 注意：当前GameServer不存储角色坐标，坐标由DungeonServer管理
	// 这里只是记录日志，如果需要可以扩展PlayerRole存储坐标信息

	log.Debugf("position synced: RoleId=%d, SceneId=%d, Pos=(%d,%d)", req.RoleId, req.SceneId, req.PosX, req.PosY)
	return nil
}

// handleDungeonSyncAttrs 处理副本属性回传
func handleDungeonSyncAttrs(_ context.Context, _ string, data []byte) error {
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
	attrSys := entitysystem.GetAttrSys(playerRole.WithContext(nil))
	if attrSys == nil {
		log.Warnf("attr sys not found for RoleId=%d", req.RoleId)
		return nil
	}
	attrSys.ApplyDungeonSyncData(req.SyncData)
	return nil
}

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		gshare.RegisterHandler(gshare.DoNetWorkMsg, handleDoNetWorkMsg)
		gshare.RegisterHandler(gshare.DoRunOneMsg, handleRunOneMsg)
		gshare.RegisterHandler(gshare.PlayerMessageMsg, handlePlayerMessageMsg)

		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRegister), handleRegister)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLogin), handleLogin)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SVerify), handleVerify)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryRoles), handleQueryRoles)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateRole), handleCreateRole)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterGame), handleEnterGame)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SReconnect), handleReconnect)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryRank), handleQueryRank)

		// 注册协议注册的RPC处理器
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GRegisterProtocols), handleRegisterProtocols)
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GUnregisterProtocols), handleUnregisterProtocols)
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSyncPosition), handleSyncPosition)
		dungeonserverlink.RegisterRPCHandler(uint16(protocol.D2GRpcProtocol_D2GSyncAttrs), handleDungeonSyncAttrs)

	})
}
