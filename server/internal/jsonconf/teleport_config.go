/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: 传送点配置
**/

package jsonconf

// TeleportConfig 传送点配置
type TeleportConfig struct {
	TeleportId   uint32 `json:"teleportId"`   // 传送点ID
	Name         string `json:"name"`         // 传送点名称
	FromSceneId  uint32 `json:"fromSceneId"`  // 源场景ID
	FromPosX     uint32 `json:"fromPosX"`     // 源位置X坐标
	FromPosY     uint32 `json:"fromPosY"`     // 源位置Y坐标
	ToSceneId    uint32 `json:"toSceneId"`    // 目标场景ID
	ToPosX       uint32 `json:"toPosX"`       // 目标位置X坐标
	ToPosY       uint32 `json:"toPosY"`       // 目标位置Y坐标
	LevelRequire uint32 `json:"levelRequire"` // 等级要求（0表示无要求）
	CostType     uint32 `json:"costType"`     // 消耗类型: 0=免费 1=金币 2=物品
	CostId       uint32 `json:"costId"`       // 消耗物品ID（如果是物品消耗）
	CostCount    uint32 `json:"costCount"`    // 消耗数量
}
