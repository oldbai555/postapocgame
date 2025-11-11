/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package skill

// SkillHitResult 技能命中结果
type SkillHitResult struct {
	TargetHdl  uint64
	IsHit      bool            // 是否命中
	IsDodge    bool            // 是否闪避
	IsCrit     bool            // 是否暴击
	Damage     int64           // 伤害值
	Heal       int64           // 治疗值
	AddedBuffs []uint32        // 添加的Buff
	ResultType SkillResultType // 结果类型
}
