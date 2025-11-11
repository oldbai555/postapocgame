/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// EntityType 实体类型
type EntityType uint32

const (
	EntityTypeRole    EntityType = 1 // 角色
	EntityTypeMonster EntityType = 2 // 怪物
	EntityTypeNPC     EntityType = 3 // NPC
)
