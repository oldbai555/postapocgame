/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 装备附魔配置
**/

package jsonconf

// EquipEnchantConfig 装备附魔配置
type EquipEnchantConfig struct {
	ItemId          uint32        `json:"itemId"`          // 装备物品ID
	EnchantCost     []*ItemAmount `json:"enchantCost"`     // 附魔消耗
	EnchantAttrs    []*Attr       `json:"enchantAttrs"`    // 附魔属性列表（随机选择）
	MaxEnchantCount uint32        `json:"maxEnchantCount"` // 最大附魔属性数量
}
