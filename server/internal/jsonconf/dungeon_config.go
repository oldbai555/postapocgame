package jsonconf

// DungeonConfig 副本配置
type DungeonConfig struct {
	DungeonID      uint32               `json:"dungeonId"`      // 副本ID
	Name           string               `json:"name"`           // 副本名称
	Type           uint32               `json:"type"`           // 副本类型: 1=常驻 2=限时
	Difficulties   []*DungeonDifficulty `json:"difficulties"`   // 难度配置
	SceneIds       []uint32             `json:"sceneIds"`       // 场景ID列表
	CDMinutes      uint32               `json:"cdMinutes"`      // 冷却时间（分钟）
	MaxEnterPerDay uint32               `json:"maxEnterPerDay"` // 每日最大进入次数
	Description    string               `json:"description"`    // 描述
}

// DungeonDifficulty 副本难度配置
type DungeonDifficulty struct {
	Difficulty        uint32           `json:"difficulty"`        // 难度: 1=普通 2=精英 3=地狱
	Name              string           `json:"name"`              // 难度名称
	MonsterLevelBonus uint32           `json:"monsterLevelBonus"` // 怪物等级加成
	ExpMultiplier     float32          `json:"expMultiplier"`     // 经验倍率
	DropMultiplier    float32          `json:"dropMultiplier"`    // 掉落倍率
	Rewards           []*DungeonReward `json:"rewards"`           // 结算奖励
	ConsumeItems      []*ItemAmount    `json:"consumeItems"`      // 进入消耗物品（如通天令）
}

// DungeonReward 副本奖励
type DungeonReward struct {
	Type   uint32  `json:"type"`   // 奖励类型: 1=经验 2=金币 3=物品
	ItemID uint32  `json:"itemId"` // 物品ID（如果是物品奖励）
	Count  uint32  `json:"count"`  // 数量
	Rate   float32 `json:"rate"`   // 获得概率（0-1）
}
