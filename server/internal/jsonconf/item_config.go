/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 物品配置（保持结构干净，不嵌入其他功能的配置）
**/

package jsonconf

// ItemConfig 道具配置
type ItemConfig struct {
	ItemId      uint32 `json:"itemId"`      // 道具Id
	Name        string `json:"name"`        // 道具名称
	Type        uint32 `json:"type"`        // 道具类型
	SubType     uint32 `json:"subType"`     // 道具子类型（根据Type确定含义：货币类型用MoneyType，装备类型用EquipSlot）
	Quality     uint32 `json:"quality"`     // 品质: 1=白 2=绿 3=蓝 4=紫 5=橙
	Star        uint32 `json:"star"`        // 星级: 0-10
	Tier        uint32 `json:"tier"`        // 阶级: 1-10
	MaxStack    uint32 `json:"maxStack"`    // 最大堆叠数量
	Flag        uint64 `json:"flag"`        // 标志位（二进制位标识：可使用、可出售等）
	Description string `json:"description"` // 描述

	// 属性相关（装备使用）
	NormalAttrs  AttrVec `json:"normalAttrs"`  // 普通属性
	RareAttrs    AttrVec `json:"rareAttrs"`    // 极品属性
	StarAttrs    AttrVec `json:"starAttrs"`    // 星级属性
	QualityAttrs AttrVec `json:"qualityAttrs"` // 品质属性
	TierAttrs    AttrVec `json:"tierAttrs"`    // 阶级属性
}
