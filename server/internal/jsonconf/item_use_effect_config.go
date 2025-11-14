/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 物品使用效果配置
**/

package jsonconf

// ItemUseEffectConfig 物品使用效果配置
type ItemUseEffectConfig struct {
	ItemId     uint32   `json:"itemId"`     // 物品ID
	EffectType uint32   `json:"effectType"` // 效果类型: 1=恢复HP 2=恢复MP 3=增加经验
	Values     []uint32 `json:"values"`     // 效果值数组，支持多个效果值
}
