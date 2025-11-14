/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// LevelConfig 等级配置
type LevelConfig struct {
	Level     uint32   `json:"level"`     // 等级
	ExpNeeded uint64   `json:"expNeeded"` // 升到下一级所需经验
	Rewards   []ItemSt `json:"rewards"`   // 升级奖励
	Attrs     AttrVec  `json:"attrs"`     // 等级属性列表（高等级覆盖低等级）
}
