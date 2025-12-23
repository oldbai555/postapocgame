// skill.go 实现副本中技能的释放流程、目标筛选与冷却管理。
package skill

import (
	"math"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
	"time"

	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/log"
)

// Skill 封装单个技能的配置、等级以及运行期冷却信息。
type Skill struct {
	Id    uint32 //技能id
	Level uint32 //技能等级
	cd    int64  //技能cd
}

const (
	DamageFlagPhysical uint64 = 1 << 0
	DamageFlagMagical  uint64 = 1 << 1
	DamageFlagTrue     uint64 = 1 << 2
	DamageFlagHeal     uint64 = 1 << 3
)

func NewSkill(id, level uint32) *Skill {
	skill := new(Skill)
	skill.Id = id
	skill.Level = level
	return skill
}

func (s *Skill) GetConfig() *jsonconf.SkillConfig {
	configMgr := jsonconf.GetConfigManager()
	skillCfg := configMgr.GetSkillConfig(s.Id)
	if skillCfg == nil {
		return nil
	}
	return skillCfg
}

// FindSkillTargets 寻找技能释放目标
func (s *Skill) FindSkillTargets(ctx *argsdef.SkillCastContext, caster iface.IEntity) ([]iface.IEntity, int) {
	configMgr := jsonconf.GetConfigManager()
	skillCfg := configMgr.GetSkillConfig(ctx.SkillId)
	if skillCfg == nil {
		return nil, int(protocol.SkillUseErr_ErrSkillNotLearned)
	}

	var targets []iface.IEntity
	entityMgr := entitymgr.GetEntityMgr()

	switch skillCfg.TargetType {
	case uint32(protocol.SkillTargetType_SkillTargetTypeSingle):
		// 单体指向性技能
		target, ok := entityMgr.GetByHdl(ctx.TargetHdl)
		if !ok || target == nil {
			return nil, int(protocol.SkillUseErr_ErrSkillTargetInvalId)
		}

		// 检查距离（格子距离）
		distance := s.calculateDistance(caster.GetPosition(), target.GetPosition())
		if distance > skillCfg.Range { // skillCfg.Range 是格子距离
			return nil, int(protocol.SkillUseErr_ErrSkillTargetTooFar)
		}

		targets = []iface.IEntity{target}

	case uint32(protocol.SkillTargetType_SkillTargetTypeAOE):
		// AOE技能，找范围内的实体
		targets = s.findAOETargets(caster, ctx.PosX, ctx.PosY, skillCfg.Range, 5)

	case uint32(protocol.SkillTargetType_SkillTargetTypeSelf):
		// 自身
		targets = []iface.IEntity{caster}
	}

	if len(targets) == 0 {
		return nil, int(protocol.SkillUseErr_ErrSkillTargetInvalId)
	}

	return targets, 0
}

// CheckTargetsValId 检查目标是否有效
func (s *Skill) CheckTargetsValId(targets []iface.IEntity) []iface.IEntity {
	valIdTargets := make([]iface.IEntity, 0, len(targets))

	for _, target := range targets {
		// 步骤7: 判定对方是否可被攻击
		if !target.CanBeAttacked() {
			log.Debugf("Target cannot be attacked: hdl=%d, dead=%v, invincible=%v",
				target.GetHdl(), target.IsDead(), target.IsInvincible())
			continue
		}

		valIdTargets = append(valIdTargets, target)
	}

	return valIdTargets
}

func (s *Skill) GetCd() int64 {
	return s.cd
}

func (s *Skill) SetCd(cd int64) {
	s.cd = cd
}

func (s *Skill) CheckCd() bool {
	return servertime.UnixMilli() >= s.GetCd()
}

// calculateDistance 计算两个位置之间的格子距离（欧几里得距离）
// 注意：Position 中的 X、Y 是格子坐标，返回的是格子距离
func (s *Skill) calculateDistance(pos1, pos2 *argsdef.Position) uint32 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return uint32(math.Sqrt(float64(dx*dx + dy*dy))) // 格子距离
}

// findAOETargets 查找AOE范围内的目标
// 注意：posX、posY 是格子坐标，radius 是格子距离
func (s *Skill) findAOETargets(caster iface.IEntity, posX, posY, radius uint32, maxCount int) []iface.IEntity {
	entityMgr := entitymgr.GetEntityMgr()
	allEntities := entityMgr.GetAll()

	targets := make([]iface.IEntity, 0)
	targetPos := &argsdef.Position{X: posX, Y: posY} // 格子坐标

	for _, et := range allEntities {
		// 跳过施法者自己
		if et.GetHdl() == caster.GetHdl() {
			continue
		}

		// 检查距离（格子距离）
		distance := s.calculateDistance(targetPos, et.GetPosition())
		if distance <= radius { // radius 是格子距离
			targets = append(targets, et)

			// 最多5个目标
			if len(targets) >= maxCount {
				break
			}
		}
	}

	return targets
}

func (s *Skill) checkHit(caster, target iface.IEntity) (bool, bool) {
	// 简单的命中检查
	// TODO: 考虑命中率、闪避率等
	damageCalc := NewDamageCalculator()
	casterAttr := caster.GetAttrSys()
	targetAttr := target.GetAttrSys()

	isDodge := damageCalc.CheckDodge(casterAttr, targetAttr)
	return !isDodge, isDodge
}

func (s *Skill) Use(ctx *argsdef.SkillCastContext, caster iface.IEntity) *CastResult {
	result := &CastResult{
		Success: false,
		ErrCode: int(protocol.SkillUseErr_SkillUseErrSuccess),
	}

	targets, ret := s.FindSkillTargets(ctx, caster)
	if ret != int(protocol.SkillUseErr_SkillUseErrSuccess) {
		result.ErrCode = ret
		return result
	}

	validTargets := s.CheckTargetsValId(targets)
	if len(validTargets) == 0 {
		log.Warnf("No valid targets after check")
		result.ErrCode = int(protocol.SkillUseErr_ErrSkillTargetInvalId)
		return result
	}

	skillCfg := s.GetConfig()
	if skillCfg == nil {
		result.ErrCode = int(protocol.SkillUseErr_ErrSkillNotLearned)
		return result
	}

	if len(skillCfg.Effects) == 0 {
		log.Warnf("skill %d has empty effects, cast failed, caster=%d", ctx.SkillId, caster.GetHdl())
		result.ErrCode = int(protocol.SkillUseErr_ErrSkillCannotCast)
		return result
	}

	damageCalc := NewDamageCalculator()
	result.Hits = make([]*SkillHitResult, 0, len(validTargets))

	for _, target := range validTargets {
		hit := &SkillHitResult{
			TargetHdl: target.GetHdl(),
			Target:    target,
		}
		isHit, isDodge := s.checkHit(caster, target)
		hit.IsHit = isHit
		hit.IsDodge = isDodge
		hit.ResultType = SkillResultTypeDamage

		if !isHit || isDodge {
			result.Hits = append(result.Hits, hit)
			continue
		}

		for _, effect := range skillCfg.Effects {
			switch SkillResultType(effect.Type) {
			case SkillResultTypeDamage:
				damage, isCrit, _ := damageCalc.Calculate(caster, target, ctx.SkillId)
				hit.Damage += damage
				hit.IsCrit = isCrit
				hit.ResultType = SkillResultTypeDamage
				hit.DamageFlags = buildDamageFlags(skillCfg, SkillResultTypeDamage)
			case SkillResultTypeHeal:
				heal := damageCalc.CalculateHeal(caster, target, ctx.SkillId)
				hit.Heal += heal
				hit.ResultType = SkillResultTypeHeal
				hit.DamageFlags = buildDamageFlags(skillCfg, SkillResultTypeHeal)
			case SkillResultTypeAddBuff:
				hit.AddedBuffs = append(hit.AddedBuffs, effect.Value)
				hit.ResultType = SkillResultTypeAddBuff
			case SkillResultTypeRemoveBuff:
			}
		}

		result.Hits = append(result.Hits, hit)
	}

	if len(result.Hits) == 0 {
		result.ErrCode = int(protocol.SkillUseErr_ErrSkillCannotCast)
		return result
	}

	cdDuration := time.Duration(skillCfg.CoolDown) * time.Millisecond
	s.SetCd(servertime.Now().Add(cdDuration).UnixMilli())

	mp := caster.GetMP()
	if mp >= int64(skillCfg.ManaCost) {
		mp -= int64(skillCfg.ManaCost)
		caster.SetMP(mp)
	}

	result.Success = true
	return result
}

func buildDamageFlags(cfg *jsonconf.SkillConfig, resultType SkillResultType) uint64 {
	switch resultType {
	case SkillResultTypeHeal:
		return DamageFlagHeal
	case SkillResultTypeDamage:
		switch cfg.DamageType {
		case 1:
			return DamageFlagPhysical
		case 2:
			return DamageFlagMagical
		default:
			return DamageFlagTrue
		}
	}
	return 0
}
