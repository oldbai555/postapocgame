/**
 * @Author: zjj
 * @Date: 2025/12/23
 * @Desc:
**/

package attrdef

// ===== 非战斗属性 (1000+) =====
const (
	HP    uint32 = 1000 // 生命值
	MP    uint32 = 1001 // 魔法值
	Level uint32 = 1002 // 等级
	Exp   uint32 = 1003 // 经验值

	ExtraAttrBegin = HP
	ExtraAttrEnd   = Exp
)

// IsExtraAttr 判断是否为非战斗属性
func IsExtraAttr(attrType uint32) bool {
	return attrType >= HP && attrType <= Exp
}
