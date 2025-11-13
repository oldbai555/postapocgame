package entitysystem

import (
	"fmt"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/service/dungeonserver/internel/buff"
	"sync"
	"time"
)

type BuffSys struct {
	mu          sync.RWMutex
	entityBuffs map[uint64]map[uint32]*buff.BData
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewBuffSystem() *BuffSys {
	bs := &BuffSys{
		entityBuffs: make(map[uint64]map[uint32]*buff.BData),
		stopChan:    make(chan struct{}),
	}
	bs.wg.Add(1)
	routine.GoV2(func() error {
		defer bs.wg.Done()
		bs.cleanupExpiredBuffs()
		return nil
	})
	return bs
}

func (bs *BuffSys) cleanupExpiredBuffs() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bs.stopChan:
			return
		case <-ticker.C:
			bs.mu.Lock()
			now := time.Now()
			for entityId, buffs := range bs.entityBuffs {
				for buffId, b := range buffs {
					if now.After(b.EndTime) {
						delete(buffs, buffId)
					}
				}
				if len(buffs) == 0 {
					delete(bs.entityBuffs, entityId)
				}
			}
			bs.mu.Unlock()
		}
	}
}

func (bs *BuffSys) Close() {
	close(bs.stopChan)
	bs.wg.Wait()
}

// AddBuff 添加Buff
func (bs *BuffSys) AddBuff(entityId uint64, buffId uint32, casterId uint64) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// TODO: 从配置中获取Buff信息
	buffInfo := bs.getBuffInfo(buffId)
	if buffInfo == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), fmt.Sprintf("%d %d %d buff not found", entityId, buffId, casterId))
	}

	// 获取实体的Buff列表
	if _, ok := bs.entityBuffs[entityId]; !ok {
		bs.entityBuffs[entityId] = make(map[uint32]*buff.BData)
	}

	existingBuff, exists := bs.entityBuffs[entityId][buffId]

	if exists {
		// 如果Buff已存在，检查叠加
		if existingBuff.StackCount < existingBuff.MaxStack {
			existingBuff.StackCount++
			existingBuff.EndTime = time.Now().Add(existingBuff.Duration)
		} else {
			// 已达到最大层数，刷新持续时间
			existingBuff.EndTime = time.Now().Add(existingBuff.Duration)
		}
	} else {
		// 添加新Buff
		buffInstance := &buff.BData{
			BuffId:     buffId,
			BuffName:   buffInfo.Name,
			BuffType:   buffInfo.Type,
			StackCount: 1,
			MaxStack:   buffInfo.StackLimit,
			Duration:   time.Duration(buffInfo.Duration) * time.Millisecond,
			StartTime:  time.Now(),
			EndTime:    time.Now().Add(time.Duration(buffInfo.Duration) * time.Millisecond),
			CasterId:   casterId,
			Effects:    buffInfo.Effects,
		}

		bs.entityBuffs[entityId][buffId] = buffInstance
	}

	return nil
}

// RemoveBuff 移除Buff
func (bs *BuffSys) RemoveBuff(entityId uint64, buffId uint32) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	buffs, ok := bs.entityBuffs[entityId]
	if !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "entity has no buffs")
	}

	if _, ok := buffs[buffId]; !ok {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "buff not found on entity")
	}

	delete(buffs, buffId)

	return nil
}

// GetBuffs 获取实体的所有Buff
func (bs *BuffSys) GetBuffs(entityId uint64) []*buff.BData {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	buffs, ok := bs.entityBuffs[entityId]
	if !ok {
		return []*buff.BData{}
	}

	result := make([]*buff.BData, 0, len(buffs))
	for _, buffInstance := range buffs {
		result = append(result, buffInstance)
	}

	return result
}

// HasBuff 检查是否有某个Buff
func (bs *BuffSys) HasBuff(entityId uint64, buffId uint32) bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	buffs, ok := bs.entityBuffs[entityId]
	if !ok {
		return false
	}

	_, ok = buffs[buffId]
	return ok
}

// ClearAllBuffs 清除实体的所有Buff
func (bs *BuffSys) ClearAllBuffs(entityId uint64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.entityBuffs, entityId)
}

func (bs *BuffSys) ClearDeBuffs(entityId uint64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	buffs, ok := bs.entityBuffs[entityId]
	if !ok {
		return
	}

	for buffId, buffInstance := range buffs {
		if buffInstance.BuffType == uint32(protocol.BuffType_BtDeBuff) {
			delete(buffs, buffId)
		}
	}
}

func (bs *BuffSys) getBuffInfo(buffId uint32) *jsonconf.BuffConfig {
	// TODO: 从配置中读取
	// 临时硬编码一些常用Buff
	switch buffId {
	case 2001: // 攻击增益
		return &jsonconf.BuffConfig{
			BuffId:     2001,
			Name:       "攻击增益",
			Type:       1,
			Duration:   10000, // 10秒
			StackLimit: 3,
			Effects: []jsonconf.BuffEffect{
				{AttrType: 3, AddType: 2, Value: 20}, // 攻击力+20%
			},
		}
	case 2002: // 防御增益
		return &jsonconf.BuffConfig{
			BuffId:     2002,
			Name:       "防御增益",
			Type:       1,
			Duration:   10000,
			StackLimit: 3,
			Effects: []jsonconf.BuffEffect{
				{AttrType: 4, AddType: 2, Value: 20}, // 防御力+20%
			},
		}
	case 3001: // 减速
		return &jsonconf.BuffConfig{
			BuffId:     3001,
			Name:       "减速",
			Type:       2,
			Duration:   5000, // 5秒
			StackLimit: 1,
			Effects: []jsonconf.BuffEffect{
				{AttrType: 5, AddType: 2, Value: -30}, // 速度-30%
			},
		}
	case 3002: // 中毒
		return &jsonconf.BuffConfig{
			BuffId:     3002,
			Name:       "中毒",
			Type:       2,
			Duration:   10000,
			StackLimit: 5,
			Effects: []jsonconf.BuffEffect{
				{AttrType: 1, AddType: 1, Value: -50}, // 每秒-50 HP
			},
		}
	default:
		return nil
	}
}

// GetBuffEffect 计算Buff对属性的影响
func (bs *BuffSys) GetBuffEffect(entityId uint64, attrType uint32) int32 {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	buffs, ok := bs.entityBuffs[entityId]
	if !ok {
		return 0
	}

	var totalEffect int32 = 0

	for _, buffInstance := range buffs {
		for _, effect := range buffInstance.Effects {
			if effect.AttrType == attrType {
				effectValue := effect.Value * int32(buffInstance.StackCount)
				totalEffect += effectValue
			}
		}
	}

	return totalEffect
}
