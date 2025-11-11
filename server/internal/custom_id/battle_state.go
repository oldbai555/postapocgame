/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// BattleState 战斗状态
type BattleState uint32

const (
	BattleStateIdle     BattleState = 0 // 空闲
	BattleStateFighting BattleState = 1 // 战斗中
	BattleStateEnd      BattleState = 2 // 战斗结束
)
