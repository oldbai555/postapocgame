package entitysystem

import (
	"postapocgame/server/internal/argsdef"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/internal/servertime"
	"postapocgame/server/pkg/customerr"
	dungeonattrcalc "postapocgame/server/service/gameserver/internel/app/dungeonactor/entitysystem/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
	"time"
)

// BuffSys 负责管理实体身上的 Buff/DOT，并驱动其生命周期与效果。
type BuffSys struct {
	owner iface.IEntity
	buffs map[uint32]*argsdef.BData
	dots  map[uint32][]*dotRuntime
}

type dotRuntime struct {
	stateId      uint32
	tickInterval time.Duration
	tickValue    int32
	nextTick     time.Time
}

func NewBuffSystem(owner iface.IEntity) *BuffSys {
	return &BuffSys{
		owner: owner,
		buffs: make(map[uint32]*argsdef.BData),
		dots:  make(map[uint32][]*dotRuntime),
	}
}

// AddBuff 添加Buff
func (bs *BuffSys) AddBuff(buffId uint32, caster iface.IEntity) error {
	buffInfo := jsonconf.GetConfigManager().GetBuffConfig(buffId)
	if buffInfo == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "buff %d not found", buffId)
	}

	existingBuff, exists := bs.buffs[buffId]
	now := servertime.Now()

	if exists {
		if existingBuff.StackCount < existingBuff.MaxStack {
			existingBuff.StackCount++
		}
		existingBuff.EndTime = now.Add(existingBuff.Duration)
		bs.applyEffects(existingBuff, buffInfo, now, false)
		return nil
	}

	casterId := uint64(0)
	if caster != nil {
		casterId = caster.GetHdl()
	}

	buffInstance := &argsdef.BData{
		BuffId:     buffId,
		BuffName:   buffInfo.Name,
		BuffType:   buffInfo.Type,
		StackCount: 1,
		MaxStack:   buffInfo.StackLimit,
		Duration:   time.Duration(buffInfo.Duration) * time.Millisecond,
		StartTime:  now,
		EndTime:    now.Add(time.Duration(buffInfo.Duration) * time.Millisecond),
		CasterId:   casterId,
	}

	bs.buffs[buffId] = buffInstance
	bs.applyEffects(buffInstance, buffInfo, now, true)

	// 触发 Buff 属性重算
	if bs.owner != nil && bs.owner.GetAttrSys() != nil {
		bs.owner.GetAttrSys().ResetSysAttr(uint32(protocol.SaAttrSys_SaBuff))
	}

	return nil
}

// RemoveBuff 移除Buff
func (bs *BuffSys) RemoveBuff(buffId uint32) error {
	if _, ok := bs.buffs[buffId]; !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "buff not found")
	}
	bs.cleanupBuff(buffId)
	delete(bs.buffs, buffId)

	// 触发 Buff 属性重算
	if bs.owner != nil && bs.owner.GetAttrSys() != nil {
		bs.owner.GetAttrSys().ResetSysAttr(uint32(protocol.SaAttrSys_SaBuff))
	}

	return nil
}

// HasBuff 检查是否有某个Buff
func (bs *BuffSys) HasBuff(buffId uint32) bool {
	_, ok := bs.buffs[buffId]
	return ok
}

func (bs *BuffSys) ClearAllBuffs() {
	for buffId := range bs.buffs {
		bs.cleanupBuff(buffId)
	}
	bs.buffs = make(map[uint32]*argsdef.BData)
	bs.dots = make(map[uint32][]*dotRuntime)

	// 触发 Buff 属性重算
	if bs.owner != nil && bs.owner.GetAttrSys() != nil {
		bs.owner.GetAttrSys().ResetSysAttr(uint32(protocol.SaAttrSys_SaBuff))
	}
}

func (bs *BuffSys) RunOne(now time.Time) {
	for buffId, b := range bs.buffs {
		if now.After(b.EndTime) {
			bs.cleanupBuff(buffId)
			delete(bs.buffs, buffId)
		}
	}
	bs.tickDots(now)
}

func (bs *BuffSys) applyEffects(instance *argsdef.BData, cfg *jsonconf.BuffConfig, now time.Time, fresh bool) {
	if instance == nil || cfg == nil {
		return
	}
	if fresh {
		delete(bs.dots, instance.BuffId)
	}
	for _, effect := range cfg.Effects {
		switch effect.EffectType {
		case jsonconf.BuffEffectTypeState:
			bs.applyStateEffect(effect, instance.Duration)
		case jsonconf.BuffEffectTypeDot:
			bs.applyStateEffect(effect, instance.Duration)
			if fresh {
				bs.registerDot(instance.BuffId, effect, now)
			}
		default:
			// 属性增减等效果后续扩展
		}
	}
}

func (bs *BuffSys) applyStateEffect(effect jsonconf.BuffEffect, buffDuration time.Duration) {
	if bs.owner == nil || effect.StateId == 0 {
		return
	}
	duration := buffDuration
	if effect.StateDuration > 0 {
		duration = time.Duration(effect.StateDuration) * time.Millisecond
	}
	bs.owner.ApplyExtraState(effect.StateId, duration)
}

func (bs *BuffSys) registerDot(buffId uint32, effect jsonconf.BuffEffect, now time.Time) {
	if effect.TickValue == 0 {
		return
	}
	interval := time.Duration(effect.TickInterval) * time.Millisecond
	if interval <= 0 {
		interval = time.Second
	}
	timer := &dotRuntime{
		stateId:      effect.StateId,
		tickInterval: interval,
		tickValue:    effect.TickValue,
		nextTick:     now.Add(interval),
	}
	bs.dots[buffId] = append(bs.dots[buffId], timer)
}

func (bs *BuffSys) tickDots(now time.Time) {
	for buffId, timers := range bs.dots {
		instance, ok := bs.buffs[buffId]
		if !ok {
			delete(bs.dots, buffId)
			continue
		}
		for _, timer := range timers {
			if timer == nil || timer.tickValue == 0 {
				continue
			}
			if now.Before(timer.nextTick) {
				continue
			}
			damage := int64(timer.tickValue) * int64(instance.StackCount)
			if damage > 0 {
				applyPeriodicDamage(bs.owner, damage)
			}
			timer.nextTick = now.Add(timer.tickInterval)
		}
	}
}

func (bs *BuffSys) cleanupBuff(buffId uint32) {
	buffConfig := jsonconf.GetConfigManager().GetBuffConfig(buffId)
	if buffConfig == nil {
		return
	}
	for _, effect := range buffConfig.Effects {
		if effect.StateId != 0 && bs.owner != nil {
			bs.owner.RemoveExtraState(effect.StateId)
		}
	}
	delete(bs.dots, buffId)
}

// GetBuffSys 获取 Buff 系统（用于属性计算器）
func GetBuffSys(owner iface.IEntity) iface.IBuffSys {
	if owner == nil {
		return nil
	}
	return owner.GetBuffSys()
}

// buffAttrCalc Buff 属性计算
func buffAttrCalc(owner iface.IEntity, calc *icalc.FightAttrCalc) {
	// 使用类型断言获取具体的 BuffSys 实例
	buffSys, ok := owner.GetBuffSys().(*BuffSys)
	if !ok || buffSys == nil {
		return
	}

	// 遍历所有 Buff，汇总属性加成
	for _, buffInstance := range buffSys.buffs {
		if buffInstance == nil {
			continue
		}
		buffConfig := jsonconf.GetConfigManager().GetBuffConfig(buffInstance.BuffId)
		if buffConfig == nil {
			continue
		}

		// 遍历 Buff 的所有效果
		for _, effect := range buffConfig.Effects {
			// 只处理属性类型的效果
			if effect.EffectType != jsonconf.BuffEffectTypeAttr {
				continue
			}

			attrType := attrdef.AttrType(effect.AttrType)
			if attrType == 0 {
				continue
			}

			// 计算属性值（考虑叠加层数）
			value := int64(effect.Value) * int64(buffInstance.StackCount)

			// 根据加成类型计算
			if effect.AddType == 2 {
				// 百分比加成：需要基于基础值计算，这里先简单累加
				// 注意：百分比加成应该在 ResetProperty 中通过 ApplyPercentages 处理
				// 这里先按固定值处理，后续可以扩展
				calc.AddValue(attrType, attrdef.AttrValue(value))
			} else {
				// 固定值加成
				calc.AddValue(attrType, attrdef.AttrValue(value))
			}
		}
	}
}

func init() {
	// 注册 Buff 属性计算器
	dungeonattrcalc.RegIncAttrCalcFn(uint32(protocol.SaAttrSys_SaBuff), buffAttrCalc)
}
