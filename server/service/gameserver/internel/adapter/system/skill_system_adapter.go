package system

import (
	"context"
	"postapocgame/server/internal"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/adapter/usecaseadapter"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/di"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/usecase/skill"
)

// SkillSystemAdapter 技能系统适配器
type SkillSystemAdapter struct {
	*BaseSystemAdapter
	learnSkillUseCase   *skill.LearnSkillUseCase
	upgradeSkillUseCase *skill.UpgradeSkillUseCase
	dungeonGateway      interfaces.DungeonServerGateway
}

// NewSkillSystemAdapter 创建技能系统适配器
func NewSkillSystemAdapter() *SkillSystemAdapter {
	container := di.GetContainer()
	learnSkillUC := skill.NewLearnSkillUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())
	upgradeSkillUC := skill.NewUpgradeSkillUseCase(container.PlayerGateway(), container.ConfigGateway(), container.DungeonServerGateway())

	// 注入依赖
	levelUseCase := NewLevelUseCaseAdapter()
	consumeUseCase := usecaseadapter.NewConsumeUseCaseAdapter()
	learnSkillUC.SetDependencies(levelUseCase, consumeUseCase)
	upgradeSkillUC.SetDependencies(consumeUseCase)

	return &SkillSystemAdapter{
		BaseSystemAdapter:   NewBaseSystemAdapter(uint32(protocol.SystemId_SysSkill)),
		learnSkillUseCase:   learnSkillUC,
		upgradeSkillUseCase: upgradeSkillUC,
		dungeonGateway:      container.DungeonServerGateway(),
	}
}

// OnInit 系统初始化
func (a *SkillSystemAdapter) OnInit(ctx context.Context) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("skill sys OnInit get role err:%v", err)
		return
	}

	// 从PlayerRoleBinaryData获取数据，如果不存在则初始化
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, playerRole.GetPlayerRoleId())
	if err != nil {
		log.Errorf("skill sys OnInit get binary data err:%v", err)
		return
	}

	// 如果skill_data不存在，则初始化
	if binaryData.SkillData == nil {
		binaryData.SkillData = &protocol.SiSkillData{
			SkillMap: make(map[uint32]uint32),
		}
		// 根据职业配置初始化初始技能
		jobConfigRaw, ok := di.GetContainer().ConfigGateway().GetJobConfig(playerRole.GetJob())
		if ok && jobConfigRaw != nil {
			jobConfig, ok := jobConfigRaw.(*jsonconf.JobConfig)
			if ok && jobConfig != nil && len(jobConfig.SkillIds) > 0 {
				for _, skillId := range jobConfig.SkillIds {
					binaryData.SkillData.SkillMap[skillId] = 1 // 初始等级为1
				}
			}
		}
	}

	log.Infof("SkillSys initialized: SkillCount=%d", len(binaryData.SkillData.SkillMap))
}

// LearnSkill 学习技能（对外接口，供其他系统调用）
func (a *SkillSystemAdapter) LearnSkill(ctx context.Context, skillId uint32) error {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return err
	}
	err = a.learnSkillUseCase.Execute(ctx, roleID, skillId)
	if err != nil {
		return err
	}
	// 同步到DungeonServer
	a.syncSkillToDungeonServer(ctx, skillId, 1)
	return nil
}

// UpgradeSkill 升级技能（对外接口，供其他系统调用）
func (a *SkillSystemAdapter) UpgradeSkill(ctx context.Context, skillId uint32) (uint32, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
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
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return 0, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return 0, err
	}
	if binaryData.SkillData == nil || binaryData.SkillData.SkillMap == nil {
		return 0, nil
	}
	return binaryData.SkillData.SkillMap[skillId], nil
}

// GetSkillMap 获取技能列表（用于进入副本时同步）
func (a *SkillSystemAdapter) GetSkillMap(ctx context.Context) (map[uint32]uint32, error) {
	roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	binaryData, err := di.GetContainer().PlayerGateway().GetBinaryData(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if binaryData.SkillData == nil || binaryData.SkillData.SkillMap == nil {
		return make(map[uint32]uint32), nil
	}
	return binaryData.SkillData.SkillMap, nil
}

// syncSkillToDungeonServer 同步技能到DungeonServer
func (a *SkillSystemAdapter) syncSkillToDungeonServer(ctx context.Context, skillId, level uint32) {
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
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

	// 异步调用DungeonServer更新技能
	err = a.dungeonGateway.AsyncCall(ctx, playerRole.GetDungeonSrvType(), sessionId, uint16(protocol.G2DRpcProtocol_G2DUpdateSkill), reqData)
	if err != nil {
		log.Errorf("call dungeon server update skill failed: %v", err)
		// 不返回错误，继续执行
	} else {
		log.Infof("Skill sync to DungeonServer: SkillId=%d, Level=%d", skillId, level)
	}
}

// EnsureISystem 确保 SkillSystemAdapter 实现 ISystem 接口
var _ iface.ISystem = (*SkillSystemAdapter)(nil)
