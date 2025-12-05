package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/domain/repository"
	interfaces2 "postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
)

// UpgradeSkillUseCase 升级技能用例
type UpgradeSkillUseCase struct {
	playerRepo     repository.PlayerRepository
	configManager  interfaces2.ConfigManager
	consumeUseCase interfaces2.ConsumeUseCase
	dungeonGateway interfaces2.DungeonServerGateway
}

// NewUpgradeSkillUseCase 创建升级技能用例
func NewUpgradeSkillUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces2.ConfigManager,
	dungeonGateway interfaces2.DungeonServerGateway,
) *UpgradeSkillUseCase {
	return &UpgradeSkillUseCase{
		playerRepo:     playerRepo,
		configManager:  configManager,
		dungeonGateway: dungeonGateway,
	}
}

// SetDependencies 设置依赖（用于注入 ConsumeUseCase）
func (uc *UpgradeSkillUseCase) SetDependencies(consumeUseCase interfaces2.ConsumeUseCase) {
	uc.consumeUseCase = consumeUseCase
}

// Execute 执行升级技能用例
func (uc *UpgradeSkillUseCase) Execute(ctx context.Context, roleID uint64, skillId uint32) (uint32, error) {
	skillData, err := uc.playerRepo.GetSkillData(ctx)
	if err != nil {
		return 0, customerr.Wrap(err)
	}

	// 检查技能是否已学习
	currentLevel := skillData.SkillMap[skillId]
	if currentLevel == 0 {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能未学习")
	}

	// 检查技能配置是否存在
	skillConfig := uc.configManager.GetSkillConfig(skillId)
	if skillConfig == nil {
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
	if len(skillConfig.UpgradeConsume) > 0 {
		if uc.consumeUseCase != nil {
			if err := uc.consumeUseCase.CheckConsume(ctx, roleID, skillConfig.UpgradeConsume); err != nil {
				return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
			}
			// 扣除消耗
			if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, skillConfig.UpgradeConsume); err != nil {
				return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
			}
		} else {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case not initialized")
		}
	}

	// 升级技能
	newLevel := currentLevel + 1
	skillData.SkillMap[skillId] = newLevel
	// 技能等级变更后，同步到 DungeonServer 的逻辑由 SkillSys SystemAdapter 统一触发，
	// 此处不直接依赖 DungeonServerGateway，保持 UseCase 与框架解耦。
	log.Infof("Skill upgraded: RoleID=%d, SkillId=%d, OldLevel=%d, NewLevel=%d", roleID, skillId, currentLevel, newLevel)
	return newLevel, nil
}
