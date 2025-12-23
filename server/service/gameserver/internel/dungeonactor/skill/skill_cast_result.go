/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

// skill_cast_result.go 定义技能施法结果在服务器内部的通用结构。
package skill

// CastResult 技能释放结果
type CastResult struct {
	Success bool
	ErrCode int

	Hits []*SkillHitResult
}
