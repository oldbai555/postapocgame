package skill

import (
	"context"
)

// InitSkillDataUseCase 初始化技能数据用例（小 service 风格，持有 Deps）
// 负责技能数据的初始化（技能列表结构、按职业配置初始化基础技能）
type InitSkillDataUseCase struct {
	deps Deps
}

// NewInitSkillDataUseCase 创建初始化技能数据用例
func NewInitSkillDataUseCase(deps Deps) *InitSkillDataUseCase {
	return &InitSkillDataUseCase{
		deps: deps,
	}
}

// Execute 执行初始化技能数据用例
func (uc *InitSkillDataUseCase) Execute(ctx context.Context, roleID uint64, job uint32) error {
	// 获取 BinaryData（共享引用）
	skillData, err := uc.deps.PlayerRepo.GetSkillData(ctx)
	if err != nil {
		return err
	}

	// 根据职业配置初始化初始技能
	jobConfig := uc.deps.ConfigManager.GetJobConfig(job)
	if jobConfig == nil {
		return nil
	}

	for _, skillId := range jobConfig.SkillIds {
		skillData.SkillMap[skillId] = 1 // 初始等级为1
	}

	return nil
}
