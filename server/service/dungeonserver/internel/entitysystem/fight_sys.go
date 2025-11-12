/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entitysystem

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"postapocgame/server/service/dungeonserver/internel/skill"
)

var _ iface.IFightSys = (*FightSys)(nil)

type FightSys struct {
	et              iface.IEntity
	skills          map[uint32]*skill.Skill //手动释放技能列表
	passivitySkills map[uint32]*skill.Skill //被动技能列表

	CommonCd map[uint32]int64 // 技能组公共CD
}

func NewFightSys() *FightSys {
	return &FightSys{}
}

func (s *FightSys) LearnSkill(skillId, skillLv uint32) error {
	return nil
}

func (s *FightSys) HasSkill(skillId uint32) bool {
	return false
}

func (s *FightSys) UseSkill(ctx *argsdef.SkillCastContext) int {
	caster := s.et
	log.Infof("=== Skill Cast Start === Caster=%d, SkillId=%d", caster.GetHdl(), ctx.SkillId)

	skillId := ctx.SkillId
	sk := s.skills[skillId]
	if sk == nil {
		return int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if !sk.CheckCd() {
		return int(protocol.SkillUseErr_ErrSkillInCooldown)
	}

	skillCfg := sk.GetConfig()
	if skillCfg == nil {
		return int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if caster.GetMP() < int64(skillCfg.ManaCost) {
		return int(protocol.SkillUseErr_ErrSkillNotEnoughMP)
	}

	return sk.Use(ctx, caster)
}
