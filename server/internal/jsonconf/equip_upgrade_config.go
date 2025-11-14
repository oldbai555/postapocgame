/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 装备强化消耗配置
**/

package jsonconf

// EquipUpgradeConfig 装备强化消耗配置
type EquipUpgradeConfig struct {
	ItemId      uint32         `json:"itemId"`      // 装备物品ID
	UpgradeCost []*UpgradeCost `json:"upgradeCost"` // 强化消耗（按等级）
}

// UpgradeCost 强化消耗
type UpgradeCost struct {
	Level  uint32 `json:"level"`  // 强化等级
	Gold   uint32 `json:"gold"`   // 消耗金币
	ItemId uint32 `json:"itemId"` // 消耗材料ID
	Count  uint32 `json:"count"`  // 消耗材料数量
}
