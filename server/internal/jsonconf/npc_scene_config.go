/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc: NPC场景配置
**/

package jsonconf

// NPCSceneConfig NPC场景配置
type NPCSceneConfig struct {
	NpcId      uint32 `json:"npcId"`      // NPC ID
	SceneId    uint32 `json:"sceneId"`    // 场景ID
	Name       string `json:"name"`       // NPC名称
	PosX       uint32 `json:"posX"`       // X坐标
	PosY       uint32 `json:"posY"`       // Y坐标
	Function   uint32 `json:"function"`   // NPC功能: 1=商店 2=任务 3=传送 4=对话
	DialogId   uint32 `json:"dialogId"`   // 对话ID（如果是对话NPC）
	ShopId     uint32 `json:"shopId"`     // 商店ID（如果是商店NPC）
	TeleportId uint32 `json:"teleportId"` // 传送点ID（如果是传送NPC）
}
