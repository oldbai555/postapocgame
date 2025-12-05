package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// LearnSkillUseCase 学习技能用例
type LearnSkillUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces2.ConfigManager
	levelUseCase   interfaces2.LevelUseCase
	consumeUseCase interfaces2.ConsumeUseCase
	dungeonGateway interfaces2.DungeonServerGateway
}

// NewLearnSkillUseCase 创建学习技能用例
func NewLearnSkillUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces2.ConfigManager,
	dungeonGateway interfaces2.DungeonServerGateway,
) *LearnSkillUseCase {
	return &LearnSkillUseCase{
		playerRepo:     playerRepo,
		configManager:  configManager,
		dungeonGateway: dungeonGateway,
	}
}

// SetDependencies 设置依赖（用于注入 LevelUseCase 和 ConsumeUseCase）
func (uc *LearnSkillUseCase) SetDependencies(levelUseCase interfaces2.LevelUseCase, consumeUseCase interfaces2.ConsumeUseCase) {
	uc.levelUseCase = levelUseCase
	uc.consumeUseCase = consumeUseCase
}

// Execute 执行学习技能用例
func (uc *LearnSkillUseCase) Execute(ctx context.Context, roleID uint64, skillId uint32) error {
	skillData, err := uc.playerRepo.GetSkillData(ctx)
	if err != nil {
		return customerr.Wrap(err)
	}

	// 检查技能是否已学习
	if skillData.SkillMap[skillId] > 0 {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能已学习")
	}

	// 检查技能配置是否存在
	skillConfig := uc.configManager.GetSkillConfig(skillId)
	if skillConfig == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能配置不存在")
	}

	// 检查等级要求
	if skillConfig.LevelRequirement > 0 && uc.levelUseCase != nil {
		currentLevel, err := uc.levelUseCase.GetLevel(ctx, roleID)
		if err != nil {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "获取等级失败: %v", err)
		}
		if currentLevel < skillConfig.LevelRequirement {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "等级不足，需要%d级", skillConfig.LevelRequirement)
		}
	}

	// 检查学习消耗
	if len(skillConfig.LearnConsume) > 0 {
		if uc.consumeUseCase != nil {
			if err := uc.consumeUseCase.CheckConsume(ctx, roleID, skillConfig.LearnConsume); err != nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
			}
			// 扣除消耗
			if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, skillConfig.LearnConsume); err != nil {
				return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
			}
		} else {
			return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case not initialized")
		}
	}

	// 学习技能（等级1）
	skillData.SkillMap[skillId] = 1
	// 技能变更后，技能同步到 DungeonServer 的责任由 SkillSys SystemAdapter 统一处理，
	// UseCase 仅负责更新 BinaryData，不直接发起对 DungeonServer 的 RPC 调用。
	log.Infof("Skill learned: RoleID=%d, SkillId=%d, Level=1", roleID, skillId)
	return nil
}
