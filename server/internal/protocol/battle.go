package protocol

// 战斗相关消息Id
const (
	C2S_AttackTarget uint16 = 2<<8 | 1 // 攻击目标
	C2S_UseSkill     uint16 = 2<<8 | 2 // 使用技能
	C2S_StopAttack   uint16 = 2<<8 | 3 // 停止攻击

	S2C_BattleStart  uint16 = 2<<8 | 1 // 战斗开始
	S2C_BattleResult uint16 = 2<<8 | 2 // 战斗结果
	S2C_DamageInfo   uint16 = 2<<8 | 3 // 伤害信息
	S2C_EntityHP     uint16 = 2<<8 | 4 // 实体血量更新
	S2C_BuffAdd      uint16 = 2<<8 | 5 // Buff添加
	S2C_BuffRemove   uint16 = 2<<8 | 6 // Buff移除
	S2C_EntityDie    uint16 = 2<<8 | 7 // 实体死亡
)

// AttackTargetRequest 攻击目标请求
type AttackTargetRequest struct {
	TargetEntityId uint64 `json:"targetEntityId"` // 目标实体Id
}

// UseSkillRequest 使用技能请求
type UseSkillRequest struct {
	SkillId        uint32 `json:"skillId"`        // 技能Id
	TargetEntityId uint64 `json:"targetEntityId"` // 目标实体Id
}

// StopAttackRequest 停止攻击请求
type StopAttackRequest struct {
}

// BattleStartNotify 战斗开始通知
type BattleStartNotify struct {
	AttackerEntityId uint64 `json:"attackerEntityId"` // 攻击者实体Id
	DefenderEntityId uint64 `json:"defenderEntityId"` // 防御者实体Id
}

// BattleResult 战斗结果
type BattleResult struct {
	IsWin     bool   `json:"isWin"`     // 是否胜利
	ExpGained uint64 `json:"expGained"` // 获得经验
	ItemsGot  []Item `json:"itemsGot"`  // 获得道具
}

// DamageInfo 伤害信息
type DamageInfo struct {
	AttackerEntityId uint64 `json:"attackerEntityId"` // 攻击者实体Id
	DefenderEntityId uint64 `json:"defenderEntityId"` // 防御者实体Id
	Damage           uint32 `json:"damage"`           // 伤害值
	DamageType       uint32 `json:"damageType"`       // 伤害类型: 1=普通 2=暴击 3=技能
	IsDodge          bool   `json:"isDodge"`          // 是否闪避
	SkillId          uint32 `json:"skillId"`          // 技能Id(如果是技能攻击)
}

// EntityHPNotify 实体血量更新通知
type EntityHPNotify struct {
	EntityId  uint64  `json:"entityId"`  // 实体Id
	CurrentHP uint32  `json:"currentHp"` // 当前血量
	MaxHP     uint32  `json:"maxHp"`     // 最大血量
	HPPercent float32 `json:"hpPercent"` // 血量百分比
}

// BuffAddNotify Buff添加通知
type BuffAddNotify struct {
	EntityId   uint64 `json:"entityId"`   // 实体Id
	BuffId     uint32 `json:"buffId"`     // Buff Id
	BuffName   string `json:"buffName"`   // Buff名称
	Duration   uint32 `json:"duration"`   // 持续时间(毫秒)
	StackCount uint32 `json:"stackCount"` // 叠加层数
}

// BuffRemoveNotify Buff移除通知
type BuffRemoveNotify struct {
	EntityId uint64 `json:"entityId"` // 实体Id
	BuffId   uint32 `json:"buffId"`   // Buff Id
}

// EntityDieNotify 实体死亡通知
type EntityDieNotify struct {
	EntityId uint64 `json:"entityId"` // 实体Id
	KillerId uint64 `json:"killerId"` // 击杀者Id
}

// MonsterInfo 怪物信息
type MonsterInfo struct {
	EntityId  uint64  `json:"entityId"`  // 实体Id
	MonsterId uint32  `json:"monsterId"` // 怪物配置Id
	Name      string  `json:"name"`      // 名称
	Level     uint32  `json:"level"`     // 等级
	HP        uint32  `json:"hp"`        // 当前血量
	MaxHP     uint32  `json:"maxHp"`     // 最大血量
	PosX      float32 `json:"posX"`      // X坐标
	PosY      float32 `json:"posY"`      // Y坐标
}
