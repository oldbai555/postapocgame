/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

// skill_result.go 定义技能命中类型枚举，供战斗、Buff 系统共享。
package skill

// SkillResultType 技能结果类型
type SkillResultType uint32

const (
	SkillResultTypeDamage     SkillResultType = 1 // 伤害
	SkillResultTypeHeal       SkillResultType = 2 // 治疗
	SkillResultTypeAddBuff    SkillResultType = 3 // 加Buff
	SkillResultTypeRemoveBuff SkillResultType = 4 // 减Buff
)
