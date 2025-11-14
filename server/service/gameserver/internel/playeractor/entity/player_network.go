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
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/clientprotocol"
	"postapocgame/server/service/gameserver/internel/dungeonserverlink"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"time"
)

// handleRegister 处理账号注册
func handleRegister(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	log.Infof("handleRegister: SessionId=%s", sessionId)

	// 解析注册请求
	var req protocol.C2SRegisterReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal register request failed: %v", err)
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
		log.Errorf("create account failed: %v", err)
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

	log.Infof("Account registered: AccountID=%d, Username=%s", account.ID, account.Username)

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
	log.Infof("handleLogin: SessionId=%s", sessionId)

	// 解析登录请求
	var req protocol.C2SLoginReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal login request failed: %v", err)
		return customerr.Wrap(err)
	}

	// 查找账号
	account, err := database.GetAccountByUsername(req.Username)
	if err != nil {
		log.Errorf("account not found: %v", err)
		return gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CLoginResult), &protocol.S2CLoginResultReq{
			Success: false,
			Message: "用户名或密码错误",
		})
	}

	// 验证密码
	if !account.CheckPassword(req.Password) {
		log.Errorf("password incorrect for account: %s", req.Username)
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

	log.Infof("Account logged in: AccountID=%d, Username=%s", account.ID, account.Username)

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
	log.Infof("handleQueryRoles: SessionId=%s", sessionId)

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
		log.Errorf("query roles failed: %v", err)
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
	log.Infof("handleSelectRole: SessionId=%s", sessionId)

	// 解析选择角色请求
	var req protocol.C2SEnterGameReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal select player role request failed: %v", err)
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
		log.Errorf("player not found: RoleId=%d, err=%v", req.RoleId, err)
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "角色不存在")
	}

	// 验证角色是否属于当前账号
	if dbPlayer.AccountID != session.GetAccountID() {
		log.Errorf("role not belong to account: RoleId=%d, AccountID=%d, SessionAccountID=%d",
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

	log.Infof("Selected player role: RoleId=%d, Name=%s", selectedRole.RoleId, selectedRole.RoleName)

	// 进入游戏
	err = enterGame(sessionId, selectedRole)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func handleReconnect(ctx context.Context, msg *network.ClientMessage) error {
	return nil
}

func handleCreateRole(ctx context.Context, msg *network.ClientMessage) error {
	sessionId := ctx.Value(gshare.ContextKeySession).(string)
	log.Infof("handleCreateRole: SessionId=%s", sessionId)

	// 解析创建角色请求
	var req protocol.C2SCreateRoleReq
	err := proto.Unmarshal(msg.Data, &req)
	if err != nil {
		log.Errorf("unmarshal create role request failed: %v", err)
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
		log.Errorf("check role name failed: %v", err)
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
		log.Errorf("query players failed: %v", err)
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
		log.Errorf("create player failed: %v", err)
		return customerr.Wrap(err)
	}

	log.Infof("Player created: AccountID=%d, RoleId=%d, RoleName=%s", accountID, dbPlayer.ID, dbPlayer.RoleName)

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
		now := time.Now()
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
		ClaimedTime:    time.Now().Unix(),
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
	log.Infof("enterGame: SessionId=%s, RoleId=%d", sessionId, roleInfo.RoleId)

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
		log.Errorf("call dungeon service enter scene failed: %v", err)
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
		log.Errorf("err:%v", err)
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

		log.Errorf("handleDoNetWorkMsg failed, err:%v", err)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  err.Error(),
		})
		if err != nil {
			log.Errorf("err:%v", err)
		}
		return
	}

	// GameServer无法处理,检查是否需要转发到DungeonServer
	protocolMgr := dungeonserverlink.GetProtocolManager()
	if !protocolMgr.IsDungeonProtocol(cliMsg.MsgId) {
		// 协议既不在GameServer也不在DungeonServer
		log.Errorf("protocol %d not found in GameServer or DungeonServer", cliMsg.MsgId)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  fmt.Sprintf("protocol %d not supported", cliMsg.MsgId),
		})
		if err != nil {
			log.Errorf("err:%v", err)
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
			log.Errorf("player role not found: roleId=%d", roleId)
			err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
				Code: -1,
				Msg:  "player role not found",
			})
			if err != nil {
				log.Errorf("err:%v", err)
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
		log.Errorf("forward to DungeonServer failed: srvType=%d, err:%v", targetSrvType, err)
		err = gatewaylink.SendToSessionProto(sessionId, uint16(protocol.S2CProtocol_S2CError), &protocol.ErrorData{
			Code: -1,
			Msg:  fmt.Sprintf("forward to DungeonServer failed: %v", err),
		})
		if err != nil {
			log.Errorf("err:%v", err)
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
	})
}
