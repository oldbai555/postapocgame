/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 装备套装配置
**/

package jsonconf

// EquipSetConfig 装备套装配置
type EquipSetConfig struct {
	SetId       uint32       `json:"setId"`       // 套装ID
	Name        string       `json:"name"`        // 套装名称
	Description string       `json:"description"` // 套装描述
	ItemIds     []uint32     `json:"itemIds"`     // 套装物品ID列表
	Effects     []*SetEffect `json:"effects"`     // 套装效果列表（按件数触发）
}

// SetEffect 套装效果
type SetEffect struct {
	Count       uint32  `json:"count"`       // 触发件数（如2件、4件、6件）
	Attrs       AttrVec `json:"attrs"`       // 属性加成
	Description string  `json:"description"` // 效果描述
}
