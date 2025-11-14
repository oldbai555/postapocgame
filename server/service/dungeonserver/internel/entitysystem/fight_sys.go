/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entitysystem

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
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
	return &FightSys{
		skills:          make(map[uint32]*skill.Skill),
		passivitySkills: make(map[uint32]*skill.Skill),
		CommonCd:        make(map[uint32]int64),
	}
}

func (s *FightSys) SetEntity(et iface.IEntity) {
	s.et = et
}

func (s *FightSys) LearnSkill(skillId, skillLv uint32) error {
	configMgr := jsonconf.GetConfigManager()
	if _, ok := configMgr.GetSkillConfig(skillId); !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Param_Invalid), "skill config not found:%d", skillId)
	}
	sk := skill.NewSkill(skillId, skillLv)
	sk.SetCd(0)
	s.skills[skillId] = sk
	return nil
}

func (s *FightSys) HasSkill(skillId uint32) bool {
	_, ok := s.skills[skillId]
	return ok
}

func (s *FightSys) UseSkill(ctx *argsdef.SkillCastContext) int {
	_, errCode := s.CastSkill(ctx)
	return errCode
}

func (s *FightSys) CastSkill(ctx *argsdef.SkillCastContext) (*skill.CastResult, int) {
	caster := s.et
	log.Infof("=== Skill Cast Start === Caster=%d, SkillId=%d", caster.GetHdl(), ctx.SkillId)

	skillId := ctx.SkillId
	sk := s.skills[skillId]
	if sk == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotLearned)}, int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if !sk.CheckCd() {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillInCooldown)}, int(protocol.SkillUseErr_ErrSkillInCooldown)
	}

	skillCfg := sk.GetConfig()
	if skillCfg == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotLearned)}, int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	if caster.GetMP() < int64(skillCfg.ManaCost) {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillNotEnoughMP)}, int(protocol.SkillUseErr_ErrSkillNotEnoughMP)
	}

	result := sk.Use(ctx, caster)
	if result == nil {
		return &skill.CastResult{ErrCode: int(protocol.SkillUseErr_ErrSkillCannotCast)}, int(protocol.SkillUseErr_ErrSkillCannotCast)
	}
	return result, result.ErrCode
}
