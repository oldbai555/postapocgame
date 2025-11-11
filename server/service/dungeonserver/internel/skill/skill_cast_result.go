/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package skill

// CastResult 技能释放结果
type CastResult struct {
	Success bool
	ErrCode int

	// 命中的目标和结果
	HitResults []*SkillHitResult
}
