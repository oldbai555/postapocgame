package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LearnSkillUseCase 学习技能用例（小 service 风格，持有 Deps + 依赖用例接口）
type LearnSkillUseCase struct {
	deps           Deps
	consumeUseCase iface.ConsumeUseCase
}

// NewLearnSkillUseCase 创建学习技能用例
func NewLearnSkillUseCase(deps Deps, consumeUseCase iface.ConsumeUseCase) *LearnSkillUseCase {
	return &LearnSkillUseCase{
		deps:           deps,
		consumeUseCase: consumeUseCase,
	}
}

// Execute 执行学习技能用例
func (uc *LearnSkillUseCase) Execute(ctx context.Context, roleID uint64, skillId uint32) error {
	skillData, err := uc.deps.PlayerRepo.GetSkillData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查技能是否已学习
	if skillData.SkillMap[skillId] > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能已学习")
	}

	// 检查技能配置是否存在
	skillConfig := uc.deps.ConfigManager.GetSkillConfig(skillId)
	if skillConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能配置不存在")
	}

	// 检查等级要求：直接从 LevelData 中读取当前等级，避免依赖 LevelUseCase/LevelSys
	if skillConfig.LevelRequirement > 0 {
		levelData, err := uc.deps.PlayerRepo.GetLevelData(ctx)
		if err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "获取等级失败: %v", err)
		}
		currentLevel := levelData.Level
		if currentLevel < skillConfig.LevelRequirement {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "等级不足，需要%d级", skillConfig.LevelRequirement)
		}
	}

	// 检查学习消耗
	if len(skillConfig.LearnConsume) > 0 {
		if uc.consumeUseCase == nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case not initialized")
		}
		if err := uc.consumeUseCase.CheckConsume(ctx, roleID, skillConfig.LearnConsume); err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
		}
		// 扣除消耗
		if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, skillConfig.LearnConsume); err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
		}
	}

	// 学习技能（等级1）
	skillData.SkillMap[skillId] = 1
	// 技能变更后，技能同步到 DungeonServer 的责任由 SkillSystemAdapter 统一处理，
	// UseCase 仅负责更新 BinaryData，不直接发起对 DungeonServer 的 RPC 调用。
	log.Infof("Skill learned: RoleID=%d, SkillId=%d, Level=1", roleID, skillId)
	return nil
}
