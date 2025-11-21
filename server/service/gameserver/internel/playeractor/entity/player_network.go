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
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/clientprotocol"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"strings"
	"time"
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

func handleVerify(ctx context.Context, msg *network.ClientMessage) error {
	return nil
}

func handleQueryRoles(ctx context.Context, msg *network.ClientMessage) error {
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

func handleReconnect(ctx context.Context, msg *network.ClientMessage) error {
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

func handleOpenBag(ctx context.Context, _ *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	bagSys := entitysystem.GetBagSys(ctx)
	var bagData *protocol.SiBagData
	if bagSys != nil {
		bagData = bagSys.GetBagData()
	} else {
		bagData = &protocol.SiBagData{}
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
		BagData: bagData,
	})
}

func handleOpenMoney(ctx context.Context, _ *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	moneySys := entitysystem.GetMoneySys(ctx)
	var moneyData *protocol.SiMoneyData
	if moneySys != nil {
		moneyData = moneySys.GetMoneyData()
	} else {
		moneyData = &protocol.SiMoneyData{
			MoneyMap: map[uint32]int64{},
		}
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CMoneyData), &protocol.S2CMoneyDataReq{
		MoneyData: moneyData,
	})
}

func handleShopBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SShopBuyReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	resp := &protocol.S2CShopBuyResultReq{
		ItemId: req.ItemId,
		Count:  req.Count,
	}
	sendResp := func() error {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CShopBuyResult), resp)
	}

	if req.ItemId == 0 || req.Count == 0 {
		resp.ErrCode = uint32(protocol.ErrorCode_Param_Invalid)
		return sendResp()
	}

	shopSys := entitysystem.GetShopSys(ctx)
	if shopSys == nil {
		resp.ErrCode = uint32(protocol.ErrorCode_Internal_Error)
		return sendResp()
	}
	if err := shopSys.Buy(ctx, req.ItemId, req.Count); err != nil {
		resp.ErrCode = errCodeFromError(err)
		return sendResp()
	}

	resp.ErrCode = uint32(protocol.ErrorCode_Success)
	if err := sendResp(); err != nil {
		return err
	}

	pushBagData(ctx, sessionId)
	pushMoneyData(ctx, sessionId)
	return nil
}

func handleEquipItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SEquipItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}
	resp := &protocol.S2CEquipResultReq{
		Slot:   req.Slot,
		ItemId: req.ItemId,
	}
	equipSys := entitysystem.GetEquipSys(ctx)
	if equipSys == nil {
		resp.ErrCode = uint32(protocol.ErrorCode_Internal_Error)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEquipResult), resp)
	}

	err := equipSys.EquipItem(ctx, req.ItemId, req.Slot)
	resp.ErrCode = errCodeFromError(err)

	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEquipResult), resp); sendErr != nil {
		return sendErr
	}
	if err == nil {
		pushBagData(ctx, sessionId)
	}
	return err
}

// handleUseItem 处理使用物品
func handleUseItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SUseItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 默认使用数量为1
	if req.Count == 0 {
		req.Count = 1
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 获取物品使用系统
	itemUseSys := entitysystem.GetItemUseSys(ctx)
	if itemUseSys == nil {
		resp := &protocol.S2CUseItemResultReq{
			Success: false,
			Message: "物品使用系统未初始化",
			ItemId:  req.ItemId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CUseItemResult), resp)
	}

	// 使用物品
	err := itemUseSys.UseItem(ctx, req.ItemId, req.Count)

	// 构造响应
	resp := &protocol.S2CUseItemResultReq{
		Success:        err == nil,
		ItemId:         req.ItemId,
		RemainingCount: 0,
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Success = false
	} else {
		resp.Message = "使用成功"
		// 获取剩余数量
		bagSys := entitysystem.GetBagSys(ctx)
		if bagSys != nil {
			item := bagSys.GetItem(req.ItemId)
			if item != nil {
				resp.RemainingCount = item.Count
			}
		}
		// 推送背包数据更新
		pushBagData(ctx, sessionId)
	}

	// 发送响应
	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CUseItemResult), resp); sendErr != nil {
		return sendErr
	}

	return err
}

// handleRecycleItem 处理回收物品
func handleRecycleItem(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SRecycleItemReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 默认回收数量为1
	if req.Count == 0 {
		req.Count = 1
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 创建回收系统实例
	recycleSys := entitysystem.NewRecycleSys()

	// 回收物品
	awards, err := recycleSys.RecycleItem(ctx, req.ItemId, req.Count)

	// 构造响应
	resp := &protocol.S2CRecycleItemResultReq{
		Success:       err == nil,
		ItemId:        req.ItemId,
		RecycledCount: req.Count,
		Awards:        awards,
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Success = false
	} else {
		resp.Message = "回收成功"
		// 推送背包数据更新
		pushBagData(ctx, sessionId)
		// 推送货币数据更新（如果有货币奖励）
		pushMoneyData(ctx, sessionId)
	}

	// 发送响应
	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRecycleItemResult), resp); sendErr != nil {
		return sendErr
	}

	return err
}

// handleGMCommand 处理GM命令
func handleGMCommand(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)

	// 解析GM命令请求
	var req protocol.C2SGMCommandReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		log.Errorf("unmarshal GM command request failed: %v", err)
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	roleCtx := playerRole.WithContext(ctx)

	// 执行GM命令
	gmSys := entitysystem.GetGMSys(roleCtx)
	if gmSys == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "gm system not ready")
	}
	success, message := gmSys.ExecuteCommand(roleCtx, req.GmName, req.Args)

	// 发送GM命令结果
	resp := &protocol.S2CGMCommandResultReq{
		Success: success,
		Message: message,
	}

	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGMCommandResult), resp); err != nil {
		log.Errorf("send GM command result failed: %v", err)
		return err
	}

	log.Infof("GM command executed: RoleID=%d, GMName=%s, Success=%v, Message=%s",
		playerRole.GetPlayerRoleId(), req.GmName, success, message)

	return nil
}

// handleTalkToNPC 处理NPC对话
func handleTalkToNPC(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2STalkToNPCReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 获取NPC配置
	configMgr := jsonconf.GetConfigManager()
	npcConfig := configMgr.GetNPCSceneConfig(req.NpcId)
	if npcConfig == nil {
		resp := &protocol.S2CTalkToNPCResultReq{
			Success: false,
			Message: "NPC不存在",
			NpcId:   req.NpcId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CTalkToNPCResult), resp)
	}

	// 触发任务事件（和NPC对话）
	questSys := entitysystem.GetQuestSys(ctx)
	if questSys != nil {
		questSys.UpdateQuestProgressByType(ctx, uint32(protocol.QuestType_QuestTypeTalkToNPC), req.NpcId, 1)
	}

	// 发送对话结果
	resp := &protocol.S2CTalkToNPCResultReq{
		Success: true,
		Message: "对话成功",
		NpcId:   req.NpcId,
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CTalkToNPCResult), resp)
}

// handleLearnSkill 处理学习技能
func handleLearnSkill(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SLearnSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 获取技能系统
	skillSys := entitysystem.GetSkillSys(ctx)
	if skillSys == nil {
		resp := &protocol.S2CLearnSkillResultReq{
			Success: false,
			Message: "技能系统未初始化",
			SkillId: req.SkillId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLearnSkillResult), resp)
	}

	// 学习技能
	err := skillSys.LearnSkill(ctx, req.SkillId)

	// 构造响应
	resp := &protocol.S2CLearnSkillResultReq{
		Success: err == nil,
		SkillId: req.SkillId,
	}

	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "学习成功"
		// 触发任务事件（学习技能）
		questSys := entitysystem.GetQuestSys(ctx)
		if questSys != nil {
			questSys.UpdateQuestProgressByType(ctx, uint32(protocol.QuestType_QuestTypeLearnSkill), 0, 1)
		}
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLearnSkillResult), resp)
}

// handleUpgradeSkill 处理升级技能
func handleUpgradeSkill(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SUpgradeSkillReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "player role not found")
	}

	// 获取技能系统
	skillSys := entitysystem.GetSkillSys(ctx)
	if skillSys == nil {
		resp := &protocol.S2CUpgradeSkillResultReq{
			Success: false,
			Message: "技能系统未初始化",
			SkillId: req.SkillId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CUpgradeSkillResult), resp)
	}

	// 升级技能
	skillLevel, err := skillSys.UpgradeSkill(ctx, req.SkillId)

	// 构造响应
	resp := &protocol.S2CUpgradeSkillResultReq{
		Success:    err == nil,
		SkillId:    req.SkillId,
		SkillLevel: skillLevel,
	}

	if err != nil {
		resp.Message = err.Error()
	} else {
		resp.Message = "升级成功"
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CUpgradeSkillResult), resp)
}

// handleEnterDungeon 处理进入副本请求（限时副本）
func handleEnterDungeon(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SEnterDungeonReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "玩家角色不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取副本配置
	configMgr := jsonconf.GetConfigManager()
	dungeonCfg, ok := configMgr.GetDungeonConfig(req.DungeonId)
	if !ok || dungeonCfg == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查是否为限时副本
	if dungeonCfg.Type != 2 {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "该副本不是限时副本",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查难度是否有效
	var difficultyCfg *jsonconf.DungeonDifficulty
	for i := range dungeonCfg.Difficulties {
		if dungeonCfg.Difficulties[i].Difficulty == req.Difficulty {
			difficultyCfg = dungeonCfg.Difficulties[i]
			break
		}
	}
	if difficultyCfg == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "难度不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取副本系统
	roleCtx := playerRole.WithContext(ctx)
	fubenSys := entitysystem.GetFubenSys(roleCtx)
	if fubenSys == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "副本系统未初始化",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 检查每日进入次数
	record := fubenSys.GetDungeonRecord(req.DungeonId, req.Difficulty)
	if record != nil {
		now := servertime.Now()
		lastResetTime := time.Unix(record.ResetTime, 0)

		// 检查是否需要重置（每日重置）
		if now.Sub(lastResetTime) >= 24*time.Hour {
			record.EnterCount = 0
			record.ResetTime = now.Unix()
		}

		// 检查每日最大进入次数
		if dungeonCfg.MaxEnterPerDay > 0 && record.EnterCount >= dungeonCfg.MaxEnterPerDay {
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "今日进入次数已用完",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
	}

	// 检查消耗物品（如通天令）
	if len(difficultyCfg.ConsumeItems) > 0 {
		// 检查消耗是否足够
		if err := playerRole.CheckConsume(ctx, difficultyCfg.ConsumeItems); err != nil {
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "消耗物品不足",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
		// 扣除消耗（在Actor中执行）
		roleCtx := playerRole.WithContext(ctx)
		if err := playerRole.ApplyConsume(roleCtx, difficultyCfg.ConsumeItems); err != nil {
			log.Errorf("ApplyConsume failed: %v", err)
			resp := &protocol.S2CEnterDungeonResultReq{
				Success:   false,
				Message:   "扣除消耗失败",
				DungeonId: req.DungeonId,
			}
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
		}
	}

	// 更新进入记录
	if err := fubenSys.EnterDungeon(req.DungeonId, req.Difficulty); err != nil {
		log.Errorf("EnterDungeon failed: %v", err)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "进入副本失败",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 获取角色信息
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "角色信息不存在",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 汇总属性
	attrSys := entitysystem.GetAttrSys(roleCtx)
	var syncAttrData *protocol.SyncAttrData
	if attrSys != nil {
		allAttrs := attrSys.CalculateAllAttrs(roleCtx)
		if len(allAttrs) > 0 {
			syncAttrData = &protocol.SyncAttrData{
				AttrData: allAttrs,
			}
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
		DungeonId:    req.DungeonId,  // 传递副本ID
		Difficulty:   req.Difficulty, // 传递难度
	})
	if err != nil {
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "系统错误",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 使用带SessionId的异步RPC调用
	srvType := uint8(protocol.SrvType_SrvTypeDungeonServer)
	err = dungeonserverlink.AsyncCall(context.Background(), srvType, sessionId, uint16(protocol.G2DRpcProtocol_G2DEnterDungeon), reqData)
	if err != nil {
		log.Errorf("call dungeon service enter dungeon failed: %v", err)
		resp := &protocol.S2CEnterDungeonResultReq{
			Success:   false,
			Message:   "进入副本失败",
			DungeonId: req.DungeonId,
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
	}

	// 发送成功响应
	resp := &protocol.S2CEnterDungeonResultReq{
		Success:   true,
		Message:   "进入副本成功",
		DungeonId: req.DungeonId,
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CEnterDungeonResult), resp)
}

// handleClaimOfflineReward 处理领取离线收益请求
func handleClaimOfflineReward(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	var req protocol.C2SClaimOfflineRewardReq
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// 获取玩家角色
	roleMgr := manager.GetPlayerRoleManager()
	playerRole := roleMgr.GetBySession(sessionId)
	if playerRole == nil {
		resp := &protocol.S2CClaimOfflineRewardResultReq{
			Success: false,
			Message: "玩家角色不存在",
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp)
	}

	// 获取离线收益系统
	roleCtx := playerRole.WithContext(ctx)
	offlineRewardSys := entitysystem.GetOfflineRewardSys(roleCtx)
	if offlineRewardSys == nil {
		resp := &protocol.S2CClaimOfflineRewardResultReq{
			Success: false,
			Message: "离线收益系统未初始化",
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp)
	}

	// 获取离线时间
	offlineSeconds := offlineRewardSys.GetOfflineSeconds()

	// 领取收益
	rewards, err := offlineRewardSys.ClaimReward(roleCtx)

	// 构造响应
	resp := &protocol.S2CClaimOfflineRewardResultReq{
		Success:        err == nil,
		OfflineSeconds: offlineSeconds,
		ClaimedTime:    servertime.Now().Unix(),
	}

	if err != nil {
		resp.Message = err.Error()
		resp.Success = false
	} else {
		resp.Message = "领取成功"
		// 转换奖励列表
		if rewards != nil && len(rewards) > 0 {
			resp.Rewards = make([]*protocol.ItemAmount, 0, len(rewards))
			for _, reward := range rewards {
				resp.Rewards = append(resp.Rewards, &protocol.ItemAmount{
					ItemType: reward.ItemType,
					ItemId:   reward.ItemId,
					Count:    reward.Count,
					Bind:     reward.Bind,
				})
			}
		}
		// 推送背包和货币数据更新
		pushBagData(roleCtx, sessionId)
		pushMoneyData(roleCtx, sessionId)
	}

	// 发送响应
	if sendErr := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CClaimOfflineRewardResult), resp); sendErr != nil {
		return sendErr
	}

	return err
}

func pushBagData(ctx context.Context, sessionId string) {
	bagSys := entitysystem.GetBagSys(ctx)
	if bagSys == nil {
		return
	}
	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBagData), &protocol.S2CBagDataReq{
		BagData: bagSys.GetBagData(),
	}); err != nil {
		log.Errorf("push bag data failed: %v", err)
	}
}

func pushMoneyData(ctx context.Context, sessionId string) {
	moneySys := entitysystem.GetMoneySys(ctx)
	if moneySys == nil {
		return
	}
	if err := gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CMoneyData), &protocol.S2CMoneyDataReq{
		MoneyData: moneySys.GetMoneyData(),
	}); err != nil {
		log.Errorf("push money data failed: %v", err)
	}
}

func errCodeFromError(err error) uint32 {
	if err == nil {
		return uint32(protocol.ErrorCode_Success)
	}
	code := customerr.GetErrCode(err)
	if code <= 0 {
		return uint32(protocol.ErrorCode_Internal_Error)
	}
	return uint32(code)
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

func init() {
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		gshare.RegisterHandler(gshare.DoNetWorkMsg, handleDoNetWorkMsg)
		gshare.RegisterHandler(gshare.DoRunOneMsg, handleRunOneMsg)

		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRegister), handleRegister)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLogin), handleLogin)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SVerify), handleVerify)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryRoles), handleQueryRoles)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateRole), handleCreateRole)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterGame), handleEnterGame)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SReconnect), handleReconnect)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenBag), handleOpenBag)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenMoney), handleOpenMoney)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SShopBuy), handleShopBuy)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEquipItem), handleEquipItem)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SGMCommand), handleGMCommand)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUseItem), handleUseItem)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRecycleItem), handleRecycleItem)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2STalkToNPC), handleTalkToNPC)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLearnSkill), handleLearnSkill)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUpgradeSkill), handleUpgradeSkill)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SEnterDungeon), handleEnterDungeon)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SClaimOfflineReward), handleClaimOfflineReward)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChatWorld), handleChatWorld)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SChatPrivate), handleChatPrivate)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAddFriend), handleAddFriend)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRespondFriendReq), handleRespondFriendReq)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryFriendList), handleQueryFriendList)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRemoveFriend), handleRemoveFriend)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryRank), handleQueryRank)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SCreateGuild), handleCreateGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SJoinGuild), handleJoinGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLeaveGuild), handleLeaveGuild)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryGuildInfo), handleQueryGuildInfo)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionPutOn), handleAuctionPutOn)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionBuy), handleAuctionBuy)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAuctionQuery), handleAuctionQuery)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SAddToBlacklist), handleAddToBlacklist)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SRemoveFromBlacklist), handleRemoveFromBlacklist)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SQueryBlacklist), handleQueryBlacklist)
	})
}

// handleChatWorld 处理世界聊天
func handleChatWorld(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleChatWorld: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	// 解析聊天请求
	var req protocol.C2SChatWorldReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleChatWorld: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取角色信息
	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 内容验证
	content := req.Content
	if len(content) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容不能为空",
		})
	}

	// 内容长度限制（最大200字符）
	if len(content) > 200 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容过长，最多200个字符",
		})
	}

	// 频率限制（使用防作弊系统）
	antiCheatSys := entitysystem.GetAntiCheatSys(ctx)
	if antiCheatSys != nil {
		if !antiCheatSys.CheckOperationFrequency("chat_world") {
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "发言过于频繁，请稍后再试",
			})
		}
	}

	// 内容过滤
	filteredContent := filterChatContent(content)
	if filteredContent != content {
		// 包含敏感词，记录可疑行为
		if antiCheatSys != nil {
			antiCheatSys.RecordSuspiciousBehavior("chat_content_filtered")
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容包含敏感词，请重新输入",
		})
	}

	// 发送到 PublicActor 进行广播
	chatMsg := &protocol.ChatWorldMsg{
		SenderId:   roleId,
		SenderName: roleInfo.RoleName,
		Content:    filteredContent,
	}
	msgData, err := proto.Marshal(chatMsg)
	if err != nil {
		log.Errorf("handleChatWorld: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatWorld), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleChatWorld: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleChatPrivate 处理私聊
func handleChatPrivate(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleChatPrivate: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	// 解析聊天请求
	var req protocol.C2SChatPrivateReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleChatPrivate: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	// 获取角色信息
	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 验证目标角色
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "目标角色ID无效",
		})
	}

	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "不能给自己发私聊",
		})
	}

	// 内容验证
	content := req.Content
	if len(content) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容不能为空",
		})
	}

	// 内容长度限制（最大200字符）
	if len(content) > 200 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容过长，最多200个字符",
		})
	}

	// 频率限制（使用防作弊系统）
	antiCheatSys := entitysystem.GetAntiCheatSys(ctx)
	if antiCheatSys != nil {
		if !antiCheatSys.CheckOperationFrequency("chat_private") {
			return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "发言过于频繁，请稍后再试",
			})
		}
	}

	// 内容过滤
	filteredContent := filterChatContent(content)
	if filteredContent != content {
		// 包含敏感词，记录可疑行为
		if antiCheatSys != nil {
			antiCheatSys.RecordSuspiciousBehavior("chat_content_filtered")
		}
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "聊天内容包含敏感词，请重新输入",
		})
	}

	// 发送到 PublicActor 进行转发
	chatMsg := &protocol.ChatPrivateMsg{
		SenderId:   roleId,
		TargetId:   req.TargetId,
		SenderName: roleInfo.RoleName,
		Content:    filteredContent,
	}
	msgData, err := proto.Marshal(chatMsg)
	if err != nil {
		log.Errorf("handleChatPrivate: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdChatPrivate), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleChatPrivate: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// filterChatContent 过滤聊天内容（使用配置化的敏感词库）
func filterChatContent(content string) string {
	// 从配置管理器获取敏感词配置
	configMgr := jsonconf.GetConfigManager()
	if configMgr == nil {
		return content
	}

	sensitiveWordConfig := configMgr.GetSensitiveWordConfig()
	if sensitiveWordConfig == nil || len(sensitiveWordConfig.Words) == 0 {
		return content
	}

	contentLower := strings.ToLower(content)
	for _, word := range sensitiveWordConfig.Words {
		// 简单的字符串包含检查
		if strings.Contains(contentLower, strings.ToLower(word)) {
			// 包含敏感词，返回空字符串表示需要过滤
			return ""
		}
	}

	return content
}

// handleAddFriend 处理添加好友请求
func handleAddFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAddFriend: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAddFriendReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAddFriend: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 验证目标角色
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "目标角色ID无效",
		})
	}

	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "不能添加自己为好友",
		})
	}

	// 检查是否已经是好友
	friendSys := entitysystem.GetFriendSys(ctx)
	if friendSys != nil {
		friendList := friendSys.GetFriendList()
		for _, friendId := range friendList {
			if friendId == req.TargetId {
				return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddFriendResult), &protocol.S2CAddFriendResultReq{
					Success:  false,
					Message:  "已经是好友",
					TargetId: req.TargetId,
				})
			}
		}
	}

	// 发送到 PublicActor 处理
	addFriendMsg := &protocol.AddFriendReqMsg{
		RequesterId:   roleId,
		TargetId:      req.TargetId,
		RequesterName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(addFriendMsg)
	if err != nil {
		log.Errorf("handleAddFriend: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendReq), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAddFriend: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleRespondFriendReq 处理响应好友申请
func handleRespondFriendReq(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRespondFriendReq: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRespondFriendReqReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRespondFriendReq: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	friendSys := entitysystem.GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "好友系统未初始化",
		})
	}

	// 检查是否有该申请
	requestList := friendSys.GetFriendRequestList()
	hasRequest := false
	for _, requesterId := range requestList {
		if requesterId == req.RequesterId {
			hasRequest = true
			break
		}
	}

	if !hasRequest {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRespondFriendReqResult), &protocol.S2CRespondFriendReqResultReq{
			Success:     false,
			Message:     "未找到该好友申请",
			RequesterId: req.RequesterId,
			Accepted:    req.Accepted,
		})
	}

	// 移除申请
	friendSys.RemoveFriendRequest(req.RequesterId)

	// 如果同意，添加好友
	if req.Accepted {
		friendSys.AddFriend(req.RequesterId)
	}

	// 如果同意，需要通知申请者，并让申请者也添加目标为好友
	if req.Accepted {
		// 发送响应消息到 PublicActor，通知申请者
		respMsg := &protocol.AddFriendRespMsg{
			RequesterId: req.RequesterId,
			TargetId:    roleId,
			Accepted:    true,
		}
		msgData, err := proto.Marshal(respMsg)
		if err != nil {
			log.Errorf("handleRespondFriendReq: marshal failed: %v", err)
			return customerr.Wrap(err)
		}

		actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAddFriendResp), msgData)
		err = gshare.SendPublicMessageAsync("global", actorMsg)
		if err != nil {
			log.Errorf("handleRespondFriendReq: send to public actor failed: %v", err)
			return customerr.Wrap(err)
		}
	}

	// 返回结果给客户端
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRespondFriendReqResult), &protocol.S2CRespondFriendReqResultReq{
		Success:     true,
		Message:     "操作成功",
		RequesterId: req.RequesterId,
		Accepted:    req.Accepted,
	})
}

// handleQueryFriendList 处理查询好友列表
func handleQueryFriendList(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryFriendList: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()
	_ = roleId // 暂时未使用，后续完善时使用
	friendSys := entitysystem.GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CFriendList), &protocol.S2CFriendListReq{
			Friends:      []*protocol.PlayerRankSnapshot{},
			OnlineStatus: make(map[uint64]bool),
		})
	}

	// 获取好友列表
	friendList := friendSys.GetFriendList()
	if len(friendList) == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CFriendList), &protocol.S2CFriendListReq{
			Friends:      []*protocol.PlayerRankSnapshot{},
			OnlineStatus: make(map[uint64]bool),
		})
	}

	// 发送到 PublicActor 查询好友快照和在线状态
	queryMsg := &protocol.FriendListQueryMsg{
		RequesterId:        roleId,
		RequesterSessionId: sessionId,
		FriendIds:          friendList,
	}
	msgData, err := proto.Marshal(queryMsg)
	if err != nil {
		log.Errorf("handleQueryFriendList: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdFriendListQuery), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleQueryFriendList: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleRemoveFriend 处理删除好友
func handleRemoveFriend(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRemoveFriend: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRemoveFriendReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRemoveFriend: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	friendSys := entitysystem.GetFriendSys(ctx)
	if friendSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFriendResult), &protocol.S2CRemoveFriendResultReq{
			Success:  false,
			Message:  "好友系统未初始化",
			FriendId: req.FriendId,
		})
	}

	// 移除好友
	success := friendSys.RemoveFriend(req.FriendId)
	message := "删除成功"
	if !success {
		message = "未找到该好友"
	}
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFriendResult), &protocol.S2CRemoveFriendResultReq{
		Success:  success,
		Message:  message,
		FriendId: req.FriendId,
	})
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

// handleCreateGuild 处理创建公会
func handleCreateGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleCreateGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SCreateGuildReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleCreateGuild: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 验证公会名称
	if len(req.GuildName) == 0 || len(req.GuildName) > 20 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), &protocol.S2CCreateGuildResultReq{
			Success: false,
			Message: "公会名称长度必须在1-20个字符之间",
		})
	}

	// 检查是否已有公会
	guildSys := entitysystem.GetGuildSys(ctx)
	if guildSys != nil && guildSys.GetGuildId() > 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CCreateGuildResult), &protocol.S2CCreateGuildResultReq{
			Success: false,
			Message: "您已经加入公会，无法创建新公会",
		})
	}

	// 发送到 PublicActor 处理
	createMsg := &protocol.CreateGuildMsg{
		CreatorId:   roleId,
		GuildName:   req.GuildName,
		CreatorName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(createMsg)
	if err != nil {
		log.Errorf("handleCreateGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdCreateGuild), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleCreateGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	// 这里需要等待 PublicActor 回调，暂时先返回成功（后续完善）
	return nil
}

// handleJoinGuild 处理加入公会
func handleJoinGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleJoinGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SJoinGuildReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleJoinGuild: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	roleInfo := playerRole.GetRoleInfo()
	if roleInfo == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "角色信息不存在",
		})
	}

	// 检查是否已有公会
	guildSys := entitysystem.GetGuildSys(ctx)
	if guildSys != nil && guildSys.GetGuildId() > 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CJoinGuildResult), &protocol.S2CJoinGuildResultReq{
			Success: false,
			Message: "您已经加入公会，无法加入新公会",
		})
	}

	// 发送到 PublicActor 处理
	joinMsg := &protocol.JoinGuildReqMsg{
		RequesterId:   roleId,
		GuildId:       req.GuildId,
		RequesterName: roleInfo.RoleName,
	}
	msgData, err := proto.Marshal(joinMsg)
	if err != nil {
		log.Errorf("handleJoinGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdJoinGuildReq), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleJoinGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleLeaveGuild 处理离开公会
func handleLeaveGuild(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleLeaveGuild: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()
	guildSys := entitysystem.GetGuildSys(ctx)
	if guildSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
			Success: false,
			Message: "公会系统未初始化",
		})
	}

	guildId := guildSys.GetGuildId()
	if guildId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
			Success: false,
			Message: "您未加入任何公会",
		})
	}

	// 发送到 PublicActor 处理
	leaveMsg := &protocol.LeaveGuildMsg{
		RoleId:  roleId,
		GuildId: guildId,
	}
	msgData, err := proto.Marshal(leaveMsg)
	if err != nil {
		log.Errorf("handleLeaveGuild: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdLeaveGuild), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleLeaveGuild: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	// 清除玩家的公会数据
	guildSys.SetGuildId(0)
	guildSys.SetPosition(0)
	guildSys.SetJoinTime(0)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLeaveGuildResult), &protocol.S2CLeaveGuildResultReq{
		Success: true,
		Message: "离开公会成功",
	})
}

// handleQueryGuildInfo 处理查询公会信息
func handleQueryGuildInfo(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryGuildInfo: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	guildSys := entitysystem.GetGuildSys(ctx)
	if guildSys == nil {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
			GuildData: nil,
		})
	}

	guildId := guildSys.GetGuildId()
	if guildId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
			GuildData: nil,
		})
	}

	// 需要从 PublicActor 获取公会数据（这里先返回空，后续完善）
	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CGuildInfo), &protocol.S2CGuildInfoReq{
		GuildData: nil,
	})
}

// handleAuctionPutOn 处理拍卖上架
func handleAuctionPutOn(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionPutOn: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionPutOnReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionPutOn: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	putOnMsg := &protocol.AuctionPutOnMsg{
		SellerId: roleId,
		ItemId:   req.ItemId,
		Count:    req.Count,
		Price:    req.Price,
		Duration: req.Duration,
	}

	msgData, err := proto.Marshal(putOnMsg)
	if err != nil {
		log.Errorf("handleAuctionPutOn: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionPutOn), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionPutOn: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleAuctionBuy 处理拍卖购买
func handleAuctionBuy(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionBuy: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionBuyReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionBuy: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()
	buyMsg := &protocol.AuctionBuyMsg{
		BuyerId:   roleId,
		AuctionId: req.AuctionId,
	}

	msgData, err := proto.Marshal(buyMsg)
	if err != nil {
		log.Errorf("handleAuctionBuy: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionBuy), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionBuy: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleAuctionQuery 处理拍卖查询
func handleAuctionQuery(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	_, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAuctionQuery: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAuctionQueryReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAuctionQuery: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	queryMsg := &protocol.AuctionQueryMsg{
		ItemId:             req.ItemId,
		Page:               req.Page,
		PageSize:           req.PageSize,
		RequesterSessionId: sessionId,
	}

	msgData, err := proto.Marshal(queryMsg)
	if err != nil {
		log.Errorf("handleAuctionQuery: marshal failed: %v", err)
		return customerr.Wrap(err)
	}

	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PublicActorMsgId_PublicActorMsgIdAuctionQuery), msgData)
	err = gshare.SendPublicMessageAsync("global", actorMsg)
	if err != nil {
		log.Errorf("handleAuctionQuery: send to public actor failed: %v", err)
		return customerr.Wrap(err)
	}

	return nil
}

// handleAddToBlacklist 处理添加到黑名单
func handleAddToBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleAddToBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SAddToBlacklistReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleAddToBlacklist: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()

	// 验证目标ID
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "目标角色ID无效",
		})
	}

	// 不能拉黑自己
	if req.TargetId == roleId {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "不能拉黑自己",
		})
	}

	// 添加到黑名单
	err = database.AddToBlacklist(req.TargetId, roleId, req.Reason)
	if err != nil {
		log.Errorf("handleAddToBlacklist: add to blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
			Success: false,
			Message: "添加到黑名单失败",
		})
	}

	log.Infof("Role %d added %d to blacklist, reason: %s", roleId, req.TargetId, req.Reason)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CAddToBlacklistResult), &protocol.S2CAddToBlacklistResultReq{
		Success: true,
		Message: "已添加到黑名单",
	})
}

// handleRemoveFromBlacklist 处理从黑名单移除
func handleRemoveFromBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	var req protocol.C2SRemoveFromBlacklistReq
	err = proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: unmarshal failed: %v", err)
		return customerr.Wrap(err)
	}

	roleId := playerRole.GetPlayerRoleId()

	// 验证目标ID
	if req.TargetId == 0 {
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
			Success: false,
			Message: "目标角色ID无效",
		})
	}

	// 从黑名单移除
	err = database.RemoveFromBlacklist(req.TargetId, roleId)
	if err != nil {
		log.Errorf("handleRemoveFromBlacklist: remove from blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
			Success: false,
			Message: "从黑名单移除失败",
		})
	}

	log.Infof("Role %d removed %d from blacklist", roleId, req.TargetId)

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CRemoveFromBlacklistResult), &protocol.S2CRemoveFromBlacklistResultReq{
		Success: true,
		Message: "已从黑名单移除",
	})
}

// handleQueryBlacklist 处理查询黑名单
func handleQueryBlacklist(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	playerRole, err := entitysystem.GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("handleQueryBlacklist: get player role failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "未登录",
		})
	}

	roleId := playerRole.GetPlayerRoleId()

	// 查询黑名单列表
	blacklists, err := database.GetBlacklist(roleId)
	if err != nil {
		log.Errorf("handleQueryBlacklist: get blacklist failed: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  "查询黑名单失败",
		})
	}

	// 构建黑名单ID列表
	blacklistIds := make([]uint64, 0, len(blacklists))
	for _, blacklist := range blacklists {
		blacklistIds = append(blacklistIds, blacklist.RoleId)
	}

	return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CBlacklist), &protocol.S2CBlacklistReq{
		BlacklistIds: blacklistIds,
	})
}
