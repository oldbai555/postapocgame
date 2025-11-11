package skill

import (
	"math"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

type Skill struct {
	Id    uint32 //技能id
	Level uint32 //技能等级
	cd    int64  //技能cd
}

func NewSkill(et iface.IEntity, id, level uint32) *Skill {
	skill := new(Skill)
	skill.Id = id
	skill.Level = level
	return skill
}

func (s *Skill) GetConfig() *jsonconf.SkillConfig {
	configMgr := jsonconf.GetConfigManager()
	skillCfg, ok := configMgr.GetSkillConfig(s.Id)
	if !ok {
		return nil
	}
	return skillCfg
}

// FindSkillTargets 寻找技能释放目标
func (s *Skill) FindSkillTargets(ctx *argsdef.SkillCastContext, caster iface.IEntity) ([]iface.IEntity, int) {
	configMgr := jsonconf.GetConfigManager()
	skillCfg, _ := configMgr.GetSkillConfig(ctx.SkillId)

	var targets []iface.IEntity
	entityMgr := entitymgr.GetEntityMgr()

	switch custom_id.SkillTargetType(skillCfg.TargetType) {
	case custom_id.SkillTargetTypeSingle:
		// 单体指向性技能
		target, ok := entityMgr.GetByHdl(ctx.TargetHdl)
		if !ok {
			return nil, custom_id.ErrSkillTargetInvalId
		}

		// 检查距离
		distance := s.calculateDistance(caster.GetPosition(), target.GetPosition())
		if distance > skillCfg.Range {
			return nil, custom_id.ErrSkillTargetTooFar
		}

		targets = []iface.IEntity{target}

	case custom_id.SkillTargetTypeAOE:
		// AOE技能，找范围内的实体
		targets = s.findAOETargets(caster, ctx.PosX, ctx.PosY, skillCfg.Range, 5)

	case custom_id.SkillTargetTypeSelf:
		// 自身
		targets = []iface.IEntity{caster}
	}

	if len(targets) == 0 {
		return nil, custom_id.ErrSkillTargetInvalId
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
	return time.Now().UnixMilli() >= s.GetCd()
}

func (s *Skill) calculateDistance(pos1, pos2 *argsdef.Position) uint32 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return uint32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func (s *Skill) findAOETargets(caster iface.IEntity, posX, posY, radius uint32, maxCount int) []iface.IEntity {
	entityMgr := entitymgr.GetEntityMgr()
	allEntities := entityMgr.GetAll()

	targets := make([]iface.IEntity, 0)
	targetPos := &argsdef.Position{X: posX, Y: posY}

	for _, et := range allEntities {
		// 跳过施法者自己
		if et.GetHdl() == caster.GetHdl() {
			continue
		}

		// 检查距离
		distance := s.calculateDistance(targetPos, et.GetPosition())
		if distance <= radius {
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
	casterAttr := damageCalc.GetEntityAttr(caster)
	targetAttr := damageCalc.GetEntityAttr(target)

	isDodge := damageCalc.CheckDodge(casterAttr, targetAttr)
	return !isDodge, isDodge
}

func (s *Skill) Use(ctx *argsdef.SkillCastContext, caster iface.IEntity) int {
	// 找目标
	targets, ret := s.FindSkillTargets(ctx, caster)
	if ret != custom_id.ErrSkillSuccess {
		return ret
	}

	// 过滤
	valIdTargets := s.CheckTargetsValId(targets)
	if len(valIdTargets) == 0 {
		log.Warnf("No valId targets after check")
		return custom_id.ErrSkillTargetInvalId
	}

	skillCfg := s.GetConfig()
	if skillCfg == nil {
		return custom_id.ErrSkillNotLearned
	}
	result := &CastResult{
		Success:    true,
		HitResults: make([]*SkillHitResult, 0, len(targets)),
	}

	damageCalc := NewDamageCalculator()

	for _, target := range targets {
		hitResult := &SkillHitResult{
			TargetHdl: target.GetHdl(),
		}

		// 判定技能是否命中
		isHit, isDodge := s.checkHit(caster, target)
		hitResult.IsHit = isHit
		hitResult.IsDodge = isDodge

		if !isHit || isDodge {
			result.HitResults = append(result.HitResults, hitResult)
			continue
		}

		// 根据技能效果类型执行
		for _, effect := range skillCfg.Effects {
			switch SkillResultType(effect.Type) {
			case SkillResultTypeDamage:
				// 造成伤害
				damage, isCrit, _ := damageCalc.Calculate(caster, target, ctx.SkillId)
				hitResult.Damage = damage
				hitResult.IsCrit = isCrit
				hitResult.ResultType = SkillResultTypeDamage

				// 扣除血量
				target.OnAttacked(caster, damage)

			case SkillResultTypeHeal:
				// 治疗
				heal := damageCalc.CalculateHeal(caster, target, ctx.SkillId)
				hitResult.Heal = heal
				hitResult.ResultType = SkillResultTypeHeal

				// 恢复血量
				currentHP := target.GetHP()
				newHP := currentHP + heal
				if newHP > target.GetMaxHP() {
					newHP = target.GetMaxHP()
				}
				target.SetHP(newHP)

			case SkillResultTypeAddBuff:
				// 添加Buff
				buffId := effect.Value
				buffSys := caster.GetBuffSys()
				err := buffSys.AddBuff(target.GetHdl(), buffId, caster.GetHdl())
				if err != nil {
					log.Errorf("AddBuff failed err:%v", err)
				}
				hitResult.AddedBuffs = append(hitResult.AddedBuffs, buffId)
				hitResult.ResultType = SkillResultTypeAddBuff
			}
		}

		result.HitResults = append(result.HitResults, hitResult)
	}

	// 设置技能CD
	s.SetCd(int64(time.Duration(skillCfg.CoolDown) * time.Millisecond))

	// 消耗魔法
	mp := caster.GetMaxHP()
	if mp >= int64(skillCfg.ManaCost) {
		mp -= int64(skillCfg.ManaCost)
		caster.SetMP(mp)
	}

	// 9.1 通知施法者客户端
	//dh.sendSkillCastResult(sessionId, req.SkillId, result)

	// 9.2 AOI广播技能释放
	//dh.broadcastSkillCast(scene, caster, req.SkillId, result)

	// 9.3 广播伤害/治疗结果
	//for _, hitResult := range result.HitResults {
	//	if hitResult.IsHit {
	//		dh.broadcastSkillHitResult(scene, caster.GetHdl(), hitResult)
	//	}
	//}

	return result.ErrCode
}
