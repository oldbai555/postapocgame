package skill

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
)

// UpgradeSkillUseCase 升级技能用例（小 service 风格，持有 Deps + 依赖用例接口）
type UpgradeSkillUseCase struct {
	deps           Deps
	consumeUseCase iface.ConsumeUseCase
}

// NewUpgradeSkillUseCase 创建升级技能用例
func NewUpgradeSkillUseCase(deps Deps, consumeUseCase iface.ConsumeUseCase) *UpgradeSkillUseCase {
	return &UpgradeSkillUseCase{
		deps:           deps,
		consumeUseCase: consumeUseCase,
	}
}

// Execute 执行升级技能用例
func (uc *UpgradeSkillUseCase) Execute(ctx context.Context, roleID uint64, skillId uint32) (uint32, error) {
	skillData, err := uc.deps.PlayerRepo.GetSkillData(ctx)
	if err != nil {
		return 0, customerr.Wrap(err)
	}

	// 检查技能是否已学习
	currentLevel := skillData.SkillMap[skillId]
	if currentLevel == 0 {
		return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "技能未学习")
	}

	// 检查技能配置是否存在
	skillConfig := uc.deps.ConfigManager.GetSkillConfig(skillId)
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
		if uc.consumeUseCase == nil {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "consume use case not initialized")
		}
		if err := uc.consumeUseCase.CheckConsume(ctx, roleID, skillConfig.UpgradeConsume); err != nil {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "消耗不足: %v", err)
		}
		// 扣除消耗
		if err := uc.consumeUseCase.ApplyConsume(ctx, roleID, skillConfig.UpgradeConsume); err != nil {
			return 0, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "扣除消耗失败: %v", err)
		}
	}

	// 升级技能
	newLevel := currentLevel + 1
	skillData.SkillMap[skillId] = newLevel
	// 技能等级变更后，同步到 DungeonServer 的逻辑由 SkillSystemAdapter 统一触发，
	// 此处不直接依赖 DungeonServerGateway，保持 UseCase 与框架解耦。
	log.Infof("Skill upgraded: RoleID=%d, SkillId=%d, OldLevel=%d, NewLevel=%d", roleID, skillId, currentLevel, newLevel)
	return newLevel, nil
}
