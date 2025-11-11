/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// SkillTargetType 技能目标类型
type SkillTargetType uint32

const (
	SkillTargetTypeSingle SkillTargetType = 1 // 单体指向性
	SkillTargetTypeAOE    SkillTargetType = 2 // AOE范围
	SkillTargetTypeSelf   SkillTargetType = 3 // 自身
)
