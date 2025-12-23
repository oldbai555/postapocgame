/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/argsdef"
)

type IFightSys interface {
	LearnSkill(skillId, skillLv uint32) error
	HasSkill(skillId uint32) bool

	UseSkill(ctx *argsdef.SkillCastContext) int
}
