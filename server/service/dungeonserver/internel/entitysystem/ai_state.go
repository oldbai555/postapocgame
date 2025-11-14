package entitysystem

import "postapocgame/server/internal/protocol"

// AIState 表示怪物AI状态 - 使用proto枚举
type AIState = protocol.AIState

var (
	AIStateIdle      = protocol.AIState_AIStateIdle
	AIStatePatrol    = protocol.AIState_AIStatePatrol
	AIStateChase     = protocol.AIState_AIStateChase
	AIStateAttack    = protocol.AIState_AIStateAttack
	AIStateReturning = protocol.AIState_AIStateReturning
)
