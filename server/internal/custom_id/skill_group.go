/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// SkillGroup 技能分组
type SkillGroup uint32

const (
	SkillGroupNormal SkillGroup = 1 // 普通技能组
	SkillGroupElite  SkillGroup = 2 // 精英技能组
)
