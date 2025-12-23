package skill

import (
	"context"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
)

type InitSkillDataUseCase struct {
	rt *deps.Runtime
}

func NewInitSkillDataUseCase(rt *deps.Runtime) *InitSkillDataUseCase {
	return &InitSkillDataUseCase{
		rt: rt,
	}
}

func (uc *InitSkillDataUseCase) Execute(ctx context.Context, roleID uint64, job uint32) error {
	skillData, err := uc.rt.PlayerRepo().GetSkillData(ctx)
	if err != nil {
		return err
	}

	jobConfig := jsonconf.GetConfigManager().GetJobConfig(job)
	if jobConfig == nil {
		return nil
	}

	for _, skillId := range jobConfig.SkillIds {
		skillData.SkillMap[skillId] = 1 // 初始等级为1
	}

	return nil
}
