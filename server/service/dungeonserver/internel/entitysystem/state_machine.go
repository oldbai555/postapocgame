package entitysystem

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

// 战斗状态类型 - 使用proto枚举
var (
	StateIdle       = uint32(protocol.BattleState_BattleStateIdle)
	StateCasting    = uint32(protocol.BattleState_BattleStateCasting)
	StateHardHit    = uint32(protocol.BattleState_BattleStateHardHit)
	StateDown       = uint32(protocol.BattleState_BattleStateDown)
	StateDead       = uint32(protocol.BattleState_BattleStateDead)
	StateInvincible = uint32(protocol.BattleState_BattleStateInvincible)
)

// StateMachine 状态机
type StateMachine struct {
	currentState   uint32
	stateStartTime time.Time
	stateDuration  time.Duration // 状态持续时间
	entity         iface.IEntity
	extraStates    map[uint32]time.Time
}

// NewStateMachine 创建状态机
func NewStateMachine(entity iface.IEntity) *StateMachine {
	return &StateMachine{
		currentState:   StateIdle,
		stateStartTime: time.Now(),
		entity:         entity,
		extraStates:    make(map[uint32]time.Time),
	}
}

// GetState 获取当前状态
func (sm *StateMachine) GetState() uint32 {
	return sm.currentState
}

// SetState 设置状态
func (sm *StateMachine) SetState(state uint32, duration time.Duration) {
	sm.currentState = state
	sm.stateStartTime = time.Now()
	sm.stateDuration = duration

	// 根据状态设置实体标记
	switch state {
	case StateDead:
		// 实体已死亡，由BaseEntity管理
	case StateInvincible:
		if e, ok := sm.entity.(interface{ SetInvincible(bool) }); ok {
			e.SetInvincible(true)
		}
	case StateCasting, StateHardHit, StateDown:
		// 这些状态可能影响移动和攻击
		if e, ok := sm.entity.(interface{ SetCannotAttack(bool) }); ok {
			e.SetCannotAttack(true)
		}
		if e, ok := sm.entity.(interface{ SetCannotMove(bool) }); ok {
			e.SetCannotMove(true)
		}
	case StateIdle:
		// 恢复正常状态
		if e, ok := sm.entity.(interface{ SetCannotAttack(bool) }); ok {
			e.SetCannotAttack(false)
		}
		if e, ok := sm.entity.(interface{ SetCannotMove(bool) }); ok {
			e.SetCannotMove(false)
		}
		if e, ok := sm.entity.(interface{ SetInvincible(bool) }); ok {
			e.SetInvincible(false)
		}
	}
}

// CanChangeState 检查是否可以切换状态
func (sm *StateMachine) CanChangeState(newState uint32) bool {
	currentState := sm.currentState

	// 死亡状态不能切换到其他状态（除了无敌）
	if currentState == StateDead && newState != StateInvincible {
		return false
	}

	// 无敌状态可以切换到任何状态
	if currentState == StateInvincible {
		return true
	}

	// 硬直和倒地状态不能立即切换到待命
	if (currentState == StateHardHit || currentState == StateDown) && newState == StateIdle {
		return false
	}

	return true
}

// Update 更新状态机（检查状态是否过期）
func (sm *StateMachine) Update() {
	// 如果状态有持续时间且已过期，自动切换到待命
	if sm.stateDuration > 0 {
		elapsed := time.Since(sm.stateStartTime)
		if elapsed >= sm.stateDuration {
			// 状态过期，切换到待命
			sm.currentState = StateIdle
			sm.stateDuration = 0

			// 恢复实体状态
			if e, ok := sm.entity.(interface{ SetCannotAttack(bool) }); ok {
				e.SetCannotAttack(false)
			}
			if e, ok := sm.entity.(interface{ SetCannotMove(bool) }); ok {
				e.SetCannotMove(false)
			}
		}
	}

	if len(sm.extraStates) > 0 {
		now := time.Now()
		for state, expire := range sm.extraStates {
			if expire.IsZero() || now.After(expire) {
				delete(sm.extraStates, state)
			}
		}
	}
}

// IsInState 检查是否在指定状态
func (sm *StateMachine) IsInState(state uint32) bool {
	return sm.currentState == state
}

// CanAttack 检查是否可以攻击
func (sm *StateMachine) CanAttack() bool {
	switch sm.currentState {
	case StateIdle, StateCasting:
		return true
	default:
		return false
	}
}

// CanMove 检查是否可以移动
func (sm *StateMachine) CanMove() bool {
	switch sm.currentState {
	case StateIdle:
		return true
	default:
		return false
	}
}

// AddExtraState 添加额外（debuff）状态
func (sm *StateMachine) AddExtraState(state uint32, duration time.Duration) {
	if state == 0 {
		return
	}
	expire := time.Time{}
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	sm.extraStates[state] = expire
}

// RemoveExtraState 移除额外状态
func (sm *StateMachine) RemoveExtraState(state uint32) {
	if state == 0 {
		return
	}
	delete(sm.extraStates, state)
}
