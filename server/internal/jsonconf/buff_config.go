package jsonconf

// BuffConfig Buff配置
type BuffConfig struct {
	BuffId      uint32       `json:"buffId"`      // Buff Id
	Name        string       `json:"name"`        // Buff名称
	Type        uint32       `json:"type"`        // 类型: 1=增益 2=减益
	Duration    uint32       `json:"duration"`    // 持续时间(毫秒)
	StackLimit  uint32       `json:"stackLimit"`  // 最大叠加层数
	Effects     []BuffEffect `json:"effects"`     // Buff效果
	Description string       `json:"description"` // 描述
}

// BuffEffect Buff效果
type BuffEffect struct {
	AttrType uint32 `json:"attrType"` // 属性类型: 1=HP 2=MP 3=攻击 4=防御 5=速度
	AddType  uint32 `json:"addType"`  // 加成类型: 1=固定值 2=百分比
	Value    int32  `json:"value"`    // 效果值(可正可负)
}
