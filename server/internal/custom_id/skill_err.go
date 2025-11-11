/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// 错误码
const (
	ErrSkillSuccess          = 0     // 释放成功
	ErrSkillNotLearned       = 10001 // 未学习
	ErrSkillInCooldown       = 10002 // CD中
	ErrSkillGroupCooldown    = 10003 // 公共CD中
	ErrSkillCasting          = 10004 // 正在施法
	ErrSkillTargetTooFar     = 10005 // 目标过远
	ErrSkillTargetInvalId    = 10006 // 目标无效
	ErrSkillTargetDead       = 10007 // 目标已死亡
	ErrSkillTargetInvincible = 10008 // 目标无敌
	ErrSkillCannotCast       = 10009 // 无法释放
	ErrSkillNotEnoughMP      = 10010 // 魔法值不足
)
