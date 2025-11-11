/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package argsdef

// SkillCastContext 技能释放上下文
type SkillCastContext struct {
	TargetHdl     uint64   // 目标hdl（单体技能）
	TargetHdlList []uint64 // 目标hdlList（AOE技能）
	SkillId       uint32   // 技能Id
	PosX          uint32   // 释放位置X（AOE技能）
	PosY          uint32   // 释放位置Y（AOE技能）
}
