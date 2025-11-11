/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 属性类型定义
**/

package attrdef

// AttrType 属性类型
type AttrType = uint32

// AttrValue 属性值
type AttrValue = int64

// ===== 战斗属性 (1-999) =====
const (
	// 基础属性
	AttrMaxHP      AttrType = 1  // 最大生命值
	AttrMaxMP      AttrType = 2  // 最大魔法值
	AttrAttack     AttrType = 3  // 攻击力
	AttrDefense    AttrType = 4  // 防御力
	AttrSpeed      AttrType = 5  // 速度
	AttrCritRate   AttrType = 6  // 暴击率 (万分比)
	AttrCritDamage AttrType = 7  // 暴击伤害 (万分比)
	AttrDodgeRate  AttrType = 8  // 闪避率 (万分比)
	AttrHitRate    AttrType = 9  // 命中率 (万分比)
	AttrBlock      AttrType = 10 // 格挡值
	AttrBlockRate  AttrType = 11 // 格挡率 (万分比)

	// 伤害加成
	AttrPhysicalDamageAdd AttrType = 20 // 物理伤害加成
	AttrMagicDamageAdd    AttrType = 21 // 魔法伤害加成
	AttrDamageAdd         AttrType = 22 // 总伤害加成

	// 伤害减免
	AttrPhysicalDamageReduce AttrType = 30 // 物理伤害减免
	AttrMagicDamageReduce    AttrType = 31 // 魔法伤害减免
	AttrDamageReduce         AttrType = 32 // 总伤害减免

	// 属性抗性
	AttrFireResistance      AttrType = 40 // 火属性抗性
	AttrIceResistance       AttrType = 41 // 冰属性抗性
	AttrLightningResistance AttrType = 42 // 雷属性抗性
	AttrPoisonResistance    AttrType = 43 // 毒属性抗性

	// 回复
	AttrHPRegen AttrType = 50 // 生命回复
	AttrMPRegen AttrType = 51 // 魔法回复

	// 其他战斗属性
	AttrMoveSpeed      AttrType = 60 // 移动速度
	AttrAttackSpeed    AttrType = 61 // 攻击速度
	AttrCoolDownReduce AttrType = 62 // 冷却缩减 (万分比)
	AttrLifeSteal      AttrType = 63 // 生命偷取 (万分比)
	AttrManaSteal      AttrType = 64 // 法力偷取 (万分比)

	CombatAttrBegin AttrType = 1   // 战斗属性起始
	CombatAttrEnd   AttrType = 999 // 战斗属性结束
)

// ===== 非战斗属性 (1000+) =====
const (
	// 经验相关
	AttrHP     AttrType = 1000 // 生命值
	AttrMP     AttrType = 1001 // 魔法值
	AttrLevel  AttrType = 1002 // 等级
	AttrExp    AttrType = 1003 // 经验值
	AttrExpAdd AttrType = 1004 // 经验加成 (万分比)

	// 货币相关
	AttrGold    AttrType = 1010 // 金币
	AttrDiamond AttrType = 1011 // 钻石
	AttrGoldAdd AttrType = 1012 // 金币获取加成 (万分比)

	// 掉落相关
	AttrDropRate  AttrType = 1020 // 掉落率加成 (万分比)
	AttrLuckValue AttrType = 1021 // 幸运值

	// 资源相关
	AttrEnergy    AttrType = 1030 // 体力
	AttrMaxEnergy AttrType = 1031 // 最大体力

	// 社交相关
	AttrFriendMax AttrType = 1040 // 好友上限
	AttrTeamMax   AttrType = 1041 // 队伍上限

	// 背包相关
	AttrBagCapacity AttrType = 1050 // 背包容量

	ExtraAttrBegin AttrType = 1000 // 非战斗属性起始
	ExtraAttrEnd   AttrType = 9999 // 非战斗属性结束
)

// IsCombatAttr 判断是否为战斗属性
func IsCombatAttr(attrType AttrType) bool {
	return attrType >= CombatAttrBegin && attrType <= CombatAttrEnd
}

// IsExtraAttr 判断是否为非战斗属性
func IsExtraAttr(attrType AttrType) bool {
	return attrType >= ExtraAttrBegin && attrType <= ExtraAttrEnd
}

// GetAttrName 获取属性名称
func GetAttrName(attrType AttrType) string {
	names := map[AttrType]string{
		AttrHP:         "生命值",
		AttrMaxHP:      "最大生命值",
		AttrMP:         "魔法值",
		AttrMaxMP:      "最大魔法值",
		AttrAttack:     "攻击力",
		AttrDefense:    "防御力",
		AttrSpeed:      "速度",
		AttrCritRate:   "暴击率",
		AttrCritDamage: "暴击伤害",
		AttrDodgeRate:  "闪避率",
		AttrHitRate:    "命中率",
		AttrLevel:      "等级",
		AttrExp:        "经验值",
		AttrGold:       "金币",
		AttrDiamond:    "钻石",
		// ... 可以继续添加
	}

	if name, ok := names[attrType]; ok {
		return name
	}
	return "未知属性"
}
