/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// LevelConfig 等级配置
type LevelConfig struct {
	Level        uint32   `json:"level"`        // 等级
	ExpNeeded    uint64   `json:"expNeeded"`    // 升到下一级所需经验
	Rewards      []ItemSt `json:"rewards"`      // 升级奖励
	HPBonus      uint32   `json:"hpBonus"`      // HP加成
	MPBonus      uint32   `json:"mpBonus"`      // MP加成
	AttackBonus  uint32   `json:"attackBonus"`  // 攻击加成
	DefenseBonus uint32   `json:"defenseBonus"` // 防御加成
}
