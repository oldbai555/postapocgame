package skill

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/domain/repository"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// InitSkillDataUseCase 初始化技能数据用例
// 负责技能数据的初始化（技能列表结构、按职业配置初始化基础技能）
type InitSkillDataUseCase struct {
	playerRepo    repository.PlayerRepository
	configManager interfaces.ConfigManager
}

// NewInitSkillDataUseCase 创建初始化技能数据用例
func NewInitSkillDataUseCase(
	playerRepo repository.PlayerRepository,
	configManager interfaces.ConfigManager,
) *InitSkillDataUseCase {
	return &InitSkillDataUseCase{
		playerRepo:    playerRepo,
		configManager: configManager,
	}
}

// Execute 执行初始化技能数据用例
func (uc *InitSkillDataUseCase) Execute(ctx context.Context, roleID uint64, job uint32) error {
	// 获取 BinaryData（共享引用）
	binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
	if err != nil {
		return err
	}

	// 如果skill_data不存在，则初始化
	if binaryData.SkillData == nil {
		binaryData.SkillData = &protocol.SiSkillData{
			SkillMap: make(map[uint32]uint32),
		}
		// 根据职业配置初始化初始技能
		jobConfigRaw, ok := uc.configManager.GetJobConfig(job)
		if ok && jobConfigRaw != nil {
			jobConfig, ok := jobConfigRaw.(*jsonconf.JobConfig)
			if ok && jobConfig != nil && len(jobConfig.SkillIds) > 0 {
				for _, skillId := range jobConfig.SkillIds {
					binaryData.SkillData.SkillMap[skillId] = 1 // 初始等级为1
				}
			}
		}
	}

	return nil
}
