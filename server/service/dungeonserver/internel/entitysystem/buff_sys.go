package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/dungeonserver/internel/buff"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

type BuffSys struct {
	owner iface.IEntity
	buffs map[uint32]*buff.BData
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
		buffs: make(map[uint32]*buff.BData),
		dots:  make(map[uint32][]*dotRuntime),
	}
}

// AddBuff 添加Buff
func (bs *BuffSys) AddBuff(buffId uint32, caster iface.IEntity) error {
	buffInfo, ok := jsonconf.GetConfigManager().GetBuffConfig(buffId)
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), fmt.Sprintf("buff %d not found", buffId))
	}

	existingBuff, exists := bs.buffs[buffId]
	now := time.Now()

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

	buffInstance := &buff.BData{
		BuffId:     buffId,
		BuffName:   buffInfo.Name,
		BuffType:   buffInfo.Type,
		StackCount: 1,
		MaxStack:   buffInfo.StackLimit,
		Duration:   time.Duration(buffInfo.Duration) * time.Millisecond,
		StartTime:  now,
		EndTime:    now.Add(time.Duration(buffInfo.Duration) * time.Millisecond),
		CasterId:   casterId,
		Effects:    buffInfo.Effects,
	}

	bs.buffs[buffId] = buffInstance
	bs.applyEffects(buffInstance, buffInfo, now, true)
	return nil
}

// RemoveBuff 移除Buff
func (bs *BuffSys) RemoveBuff(buffId uint32) error {
	if _, ok := bs.buffs[buffId]; !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "buff not found")
	}
	bs.cleanupBuff(buffId)
	delete(bs.buffs, buffId)
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
	bs.buffs = make(map[uint32]*buff.BData)
	bs.dots = make(map[uint32][]*dotRuntime)
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

func (bs *BuffSys) applyEffects(instance *buff.BData, cfg *jsonconf.BuffConfig, now time.Time, fresh bool) {
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
	instance, ok := bs.buffs[buffId]
	if !ok {
		return
	}
	for _, effect := range instance.Effects {
		if effect.StateId != 0 && bs.owner != nil {
			bs.owner.RemoveExtraState(effect.StateId)
		}
	}
	delete(bs.dots, buffId)
}
