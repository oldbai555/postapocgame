package jsonconf

// ItemAmount 通用的道具/货币数量描述，配合 ItemType 使用
type ItemAmount struct {
	ItemType uint32 `json:"itemType"` // 对应 proto/csproto/item_def.proto 中的 ItemType
	ItemId   uint32 `json:"itemId"`   // 货币或道具ID
	Count    int64  `json:"count"`    // 数量，正数代表消耗/奖励的绝对值
	Bind     uint32 `json:"bind"`     // 绑定标记：0=不绑定，1=绑定
}

// Clone 复制一个安全的 ItemAmount
func (ia *ItemAmount) Clone() *ItemAmount {
	if ia == nil {
		return nil
	}
	cp := *ia
	return &cp
}
