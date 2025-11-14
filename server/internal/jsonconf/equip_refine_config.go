/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 装备精炼配置
**/

package jsonconf

// EquipRefineConfig 装备精炼配置
type EquipRefineConfig struct {
	ItemId      uint32            `json:"itemId"`      // 装备物品ID
	RefineCosts []*RefineCost     `json:"refineCosts"` // 精炼消耗列表（按精炼等级）
	QualityGain map[uint32]uint32 `json:"qualityGain"` // 精炼等级 -> 品质提升（key: refineLevel, value: qualityIncrease）
}

// RefineCost 精炼消耗
type RefineCost struct {
	RefineLevel uint32        `json:"refineLevel"` // 精炼等级（目标等级）
	Consume     []*ItemAmount `json:"consume"`     // 消耗列表
}
