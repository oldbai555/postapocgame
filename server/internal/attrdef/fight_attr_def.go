/**
 * @Author: zjj
 * @Date: 2025/12/23
 * @Desc:
**/

package attrdef

const (
	MaxHP      uint32 = 1 // 最大生命值
	MaxMP      uint32 = 2 // 最大魔法值
	Attack     uint32 = 3 // 攻击力
	Defense    uint32 = 4 // 防御力
	Speed      uint32 = 5 // 速度
	CritRate   uint32 = 6 // 暴击率 (万分比)
	CritDamage uint32 = 7 // 暴击伤害 (万分比)
	DodgeRate  uint32 = 8 // 闪避率 (万分比)
	HitRate    uint32 = 9 // 命中率 (万分比)

	FightAttrBegin = MaxHP
	FightAttrEnd   = HitRate
)

func IsFightAttr(attrType uint32) bool {
	return attrType >= MaxHP && attrType <= HitRate
}
