package skill

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/runtime"
	"postapocgame/server/service/gameserver/internel/app/playeractor/service/consume"
	"postapocgame/server/service/gameserver/internel/app/playeractor/sysbase"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

type SkillSystemAdapter struct {
	*sysbase.BaseSystem
	deps                 Deps
	learnSkillUseCase    *LearnSkillUseCase
	upgradeSkillUseCase  *UpgradeSkillUseCase
	initSkillDataUseCase *InitSkillDataUseCase
	dungeonGateway       iface.DungeonServerGateway
}

// NewSkillSystemAdapter 创建技能系统适配器
func NewSkillSystemAdapter(rt *runtime.Runtime) *SkillSystemAdapter {
	d := depsFromRuntime(rt)
	consumeUseCase := consume.NewConsumeUseCase(d.PlayerRepo, d.EventPublisher)

	learnSkillUC := NewLearnSkillUseCase(d, consumeUseCase)
	upgradeSkillUC := NewUpgradeSkillUseCase(d, consumeUseCase)
	initSkillDataUC := NewInitSkillDataUseCase(d)

	return &SkillSystemAdapter{
		BaseSystem:           sysbase.NewBaseSystem(uint32(protocol.SystemId_SysSkill)),
		deps:                 d,
		learnSkillUseCase:    learnSkillUC,
		upgradeSkillUseCase:  upgradeSkillUC,
		initSkillDataUseCase: initSkillDataUC,
		dungeonGateway:       d.DungeonGateway,
	}
}

// OnInit 系统初始化
func (a *SkillSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("skill sys OnInit get role err:%v", err)
		return
	}
	// 初始化技能数据（包括技能列表结构、按职业配置初始化基础技能等业务逻辑）
	if err := a.initSkillDataUseCase.Execute(ctx, playerRole.GetPlayerRoleId(), playerRole.GetJob()); err != nil {
		log.Errorf("skill sys OnInit init skill data err:%v", err)
		return
	}
}

// LearnSkill 学习技能（对外接口，供其他系统调用）
func (a *SkillSystemAdapter) LearnSkill(ctx context.Context, skillId uint32) error {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err := a.learnSkillUseCase.Execute(ctx, roleID, skillId); err != nil {
		return err
	}
	// 同步到DungeonServer
	a.syncSkillToDungeonServer(ctx, skillId, 1)
	return nil
}

// UpgradeSkill 升级技能（对外接口，供其他系统调用）
func (a *SkillSystemAdapter) UpgradeSkill(ctx context.Context, skillId uint32) (uint32, error) {
	roleID, err := gshare.GetRoleIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	newLevel, err := a.upgradeSkillUseCase.Execute(ctx, roleID, skillId)
	if err != nil {
		return 0, err
	}
	// 同步到DungeonServer
	a.syncSkillToDungeonServer(ctx, skillId, newLevel)
	return newLevel, nil
}

// GetSkillLevel 获取技能等级
func (a *SkillSystemAdapter) GetSkillLevel(ctx context.Context, skillId uint32) (uint32, error) {
	skillData, err := a.deps.PlayerRepo.GetSkillData(ctx)
	if err != nil {
		return 0, err
	}
	return skillData.SkillMap[skillId], nil
}

// GetSkillMap 获取技能列表（用于进入副本时同步）
func (a *SkillSystemAdapter) GetSkillMap(ctx context.Context) (map[uint32]uint32, error) {
	skillData, err := a.deps.PlayerRepo.GetSkillData(ctx)
	if err != nil {
		return nil, err
	}
	return skillData.SkillMap, nil
}

// syncSkillToDungeonServer 同步技能到DungeonServer
func (a *SkillSystemAdapter) syncSkillToDungeonServer(ctx context.Context, skillId, level uint32) {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
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

	// 异步调用DungeonActor更新技能（通过 DungeonActorMsgId 枚举）
	if a.dungeonGateway == nil {
		log.Errorf("dungeon gateway not initialized")
		return
	}
	err = a.dungeonGateway.AsyncCall(ctx, sessionId, uint16(protocol.DungeonActorMsgId_DungeonActorMsgIdUpdateSkill), reqData)
	if err != nil {
		log.Errorf("call dungeon server update skill failed: %v", err)
		// 不返回错误，继续执行
	} else {
		log.Infof("Skill sync to DungeonServer: SkillId=%d, Level=%d", skillId, level)
	}
}

// EnsureISystem 确保 SkillSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*SkillSystemAdapter)(nil)

// GetSkillSys 获取技能系统
func GetSkillSys(ctx context.Context) *SkillSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysSkill))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysSkill)
		return nil
	}
	sys, ok := system.(*SkillSystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysSkill)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysSkill)
		return nil
	}
	return sys
}

// 注册系统工厂
// RegisterSystemFactory 注册技能系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *runtime.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysSkill), func() iface.ISystem {
		return NewSkillSystemAdapter(rt)
	})
}
