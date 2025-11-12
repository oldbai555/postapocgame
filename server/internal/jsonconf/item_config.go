/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// ItemConfig 道具配置
type ItemConfig struct {
	ItemId      uint32         `json:"itemId"`      // 道具Id
	Name        string         `json:"name"`        // 道具名称
	Type        uint32         `json:"type"`        // 道具类型
	Quality     uint32         `json:"quality"`     // 品质: 1=白 2=绿 3=蓝 4=紫 5=橙
	MaxStack    uint32         `json:"maxStack"`    // 最大堆叠数量
	CanUse      bool           `json:"canUse"`      // 是否可使用
	CanSell     bool           `json:"canSell"`     // 是否可出售
	SellPrice   uint32         `json:"sellPrice"`   // 出售价格
	Description string         `json:"description"` // 描述
	UseEffect   *ItemUseEffect `json:"useEffect"`   // 使用效果
}

// ItemUseEffect 道具使用效果
type ItemUseEffect struct {
	EffectType uint32 `json:"effectType"` // 效果类型: 1=恢复HP 2=恢复MP 3=增加经验
	Value      uint32 `json:"value"`      // 效果值
}
