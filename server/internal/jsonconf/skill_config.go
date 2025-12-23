/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// SkillConfig 技能配置
type SkillConfig struct {
	SkillId          uint32         `json:"skillId"`          // 技能Id
	Name             string         `json:"name"`             // 技能名称
	Type             uint32         `json:"type"`             // 类型: 1=主动 2=被动
	TargetType       uint32         `json:"targetType"`       // 目标类型: 1=敌人 2=自己 3=队友
	CoolDown         uint32         `json:"coolDown"`         // 冷却时间(毫秒)
	ManaCost         uint32         `json:"manaCost"`         // 魔法消耗
	Damage           uint32         `json:"damage"`           // 伤害值
	DamageType       uint32         `json:"damageType"`       // 伤害类型: 1=物理 2=魔法
	Range            uint32         `json:"range"`            // 攻击范围
	Effects          []*SkillEffect `json:"effects"`          // 技能效果
	Description      string         `json:"description"`      // 描述
	LevelRequirement uint32         `json:"levelRequirement"` // 学习等级要求（0表示无要求）
	MaxLevel         uint32         `json:"maxLevel"`         // 最大等级（默认10）

}

// SkillEffect 技能效果
type SkillEffect struct {
	Type     uint32  `json:"type"`     // 效果类型: 1=伤害 2=治疗 3=加Buff 4=减Buff
	Value    uint32  `json:"value"`    // 效果值
	Duration uint32  `json:"duration"` // 持续时间(毫秒)
	Rate     float32 `json:"rate"`     // 触发概率 (0-1)
	DelayMs  uint32  `json:"delayMs"`  // 生效延迟（毫秒）
}
