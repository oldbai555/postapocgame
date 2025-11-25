package entitysystem

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/manager"
	"postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

// SkillSys 技能系统
type SkillSys struct {
	*BaseSystem
	skillData *protocol.SiSkillData
}

// NewSkillSys 创建技能系统
func NewSkillSys() *SkillSys {
	return &SkillSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysSkill)),
	}
}

func GetSkillSys(ctx context.Context) *SkillSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysSkill))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysSkill, err)
		return nil
	}
	sys := system.(*SkillSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysSkill, err)
		return nil
	}
	return sys
}

// OnInit 初始化时从数据库加载技能数据
func (ss *SkillSys) OnInit(ctx context.Context) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData := playerRole.GetBinaryData()
	if binaryData == nil {
		log.Errorf("binary data is nil")
		return
	}

	// 如果skill_data不存在，则初始化
	if binaryData.SkillData == nil {
		binaryData.SkillData = &protocol.SiSkillData{
			SkillMap: make(map[uint32]uint32),
		}
		// 根据职业配置初始化初始技能
		jobConfig, ok := jsonconf.GetConfigManager().GetJobConfig(playerRole.GetJob())
		if ok && jobConfig != nil && len(jobConfig.SkillIds) > 0 {
			for _, skillId := range jobConfig.SkillIds {
				binaryData.SkillData.SkillMap[skillId] = 1 // 初始等级为1
			}
		}
	}
	ss.skillData = binaryData.SkillData

	log.Infof("SkillSys initialized: SkillCount=%d", len(ss.skillData.SkillMap))
}

// GetSkillData 获取技能数据
func (ss *SkillSys) GetSkillData() *protocol.SiSkillData {
	return ss.skillData
}

// GetSkillLevel 获取技能等级
func (ss *SkillSys) GetSkillLevel(skillId uint32) uint32 {
	if ss.skillData == nil || ss.skillData.SkillMap == nil {
		return 0
	}
	return ss.skillData.SkillMap[skillId]
}

// LearnSkill 学习技能
func (ss *SkillSys) LearnSkill(ctx context.Context, skillId uint32) error {
	if ss.skillData == nil || ss.skillData.SkillMap == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能系统未初始化")
	}

	// 检查技能是否已学习
	if ss.skillData.SkillMap[skillId] > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能已学习")
	}

	// 检查技能配置是否存在
	configMgr := jsonconf.GetConfigManager()
	skillConfig, ok := configMgr.GetSkillConfig(skillId)
	if !ok || skillConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能配置不存在")
	}

	// 检查等级要求
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "获取玩家角色失败")
	}
	levelSys := GetLevelSys(ctx)
	if levelSys != nil && skillConfig.LevelRequirement > 0 {
		if levelSys.GetLevel() < skillConfig.LevelRequirement {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "等级不足，需要%d级", skillConfig.LevelRequirement)
		}
	}

	// 检查学习消耗
	if len(skillConfig.LearnConsume) > 0 {
		if err := playerRole.CheckConsume(ctx, skillConfig.LearnConsume); err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
		}
		// 扣除消耗
		if err := playerRole.ApplyConsume(ctx, skillConfig.LearnConsume); err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
		}
	}

	// 学习技能（等级1）
	ss.skillData.SkillMap[skillId] = 1

	// 同步到DungeonServer
	ss.syncSkillToDungeonServer(ctx, skillId, 1)

	log.Infof("Skill learned: SkillId=%d, Level=1", skillId)
	return nil
}

// UpgradeSkill 升级技能
func (ss *SkillSys) UpgradeSkill(ctx context.Context, skillId uint32) (uint32, error) {
	if ss.skillData == nil || ss.skillData.SkillMap == nil {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能系统未初始化")
	}

	// 检查技能是否已学习
	currentLevel := ss.skillData.SkillMap[skillId]
	if currentLevel == 0 {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能未学习")
	}

	// 检查技能配置是否存在
	configMgr := jsonconf.GetConfigManager()
	skillConfig, ok := configMgr.GetSkillConfig(skillId)
	if !ok || skillConfig == nil {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能配置不存在")
	}

	// 检查等级上限
	maxLevel := skillConfig.MaxLevel
	if maxLevel == 0 {
		maxLevel = 10 // 默认最大等级为10
	}
	if currentLevel >= maxLevel {
		return currentLevel, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能已达到最高等级")
	}

	// 检查升级消耗
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "获取玩家角色失败")
	}
	if len(skillConfig.UpgradeConsume) > 0 {
		if err := playerRole.CheckConsume(ctx, skillConfig.UpgradeConsume); err != nil {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
		}
		// 扣除消耗
		if err := playerRole.ApplyConsume(ctx, skillConfig.UpgradeConsume); err != nil {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
		}
	}

	// 升级技能
	newLevel := currentLevel + 1
	ss.skillData.SkillMap[skillId] = newLevel

	// 同步到DungeonServer
	ss.syncSkillToDungeonServer(ctx, skillId, newLevel)

	log.Infof("Skill upgraded: SkillId=%d, OldLevel=%d, NewLevel=%d", skillId, currentLevel, newLevel)
	return newLevel, nil
}

// syncSkillToDungeonServer 同步技能到DungeonServer
func (ss *SkillSys) syncSkillToDungeonServer(ctx context.Context, skillId, level uint32) {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return
	}

	// 获取sessionId
	sessionId := playerRole.GetSessionId()
	if sessionId == "" {
		log.Errorf("sessionId is empty")
		return
	}

	// 构造RPC请求
	reqData, err := internal.Marshal(&protocol.G2DUpdateSkillReq{
		SessionId:  sessionId,
		RoleId:     playerRole.GetPlayerRoleId(),
		SkillId:    skillId,
		SkillLevel: level,
	})
	if err != nil {
		log.Errorf("marshal update skill request failed: %v", err)
		return
	}

	// 异步调用DungeonServer更新技能（通过IPlayerRole接口，避免循环依赖）
	err = playerRole.CallDungeonServer(ctx, uint16(protocol.G2DRpcProtocol_G2DUpdateSkill), reqData)
	if err != nil {
		log.Errorf("call dungeon server update skill failed: %v", err)
		// 不返回错误，继续执行
	} else {
		log.Infof("Skill sync to DungeonServer: SkillId=%d, Level=%d", skillId, level)
	}
}

// GetSkillMap 获取技能列表（用于进入副本时同步）
func (ss *SkillSys) GetSkillMap() map[uint32]uint32 {
	if ss.skillData == nil || ss.skillData.SkillMap == nil {
		return make(map[uint32]uint32)
	}
	return ss.skillData.SkillMap
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
	skillSys := GetSkillSys(ctx)
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
		questSys := GetQuestSys(ctx)
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
	skillSys := GetSkillSys(ctx)
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

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysSkill), func() iface.ISystem {
		return NewSkillSys()
	})
	gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SLearnSkill), handleLearnSkill)
		clientprotocol.Register(uint16(protocol.C2SProtocol_C2SUpgradeSkill), handleUpgradeSkill)
	})
}
