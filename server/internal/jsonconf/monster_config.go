/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// MonsterConfig 怪物配置
type MonsterConfig struct {
	MonsterId uint32        `json:"monsterId"` // 怪物Id
	Name      string        `json:"name"`      // 怪物名称
	Level     uint32        `json:"level"`     // 等级
	Type      uint32        `json:"type"`      // 类型: 1=普通 2=精英 3=BOSS
	HP        uint32        `json:"hp"`        // 生命值
	MP        uint32        `json:"mp"`        // 魔法值
	Attack    uint32        `json:"attack"`    // 攻击力
	Defense   uint32        `json:"defense"`   // 防御力
	Speed     uint32        `json:"speed"`     // 速度
	SkillIds  []uint32      `json:"skillIds"`  // 技能Id列表
	DropItems []MonsterDrop `json:"dropItems"` // 掉落物品
	ExpReward uint64        `json:"expReward"` // 经验奖励
}

// MonsterDrop 怪物掉落
type MonsterDrop struct {
	ItemId   uint32  `json:"itemId"`   // 道具Id
	DropRate float32 `json:"dropRate"` // 掉落概率 (0-1)
	MinCount uint32  `json:"minCount"` // 最小数量
	MaxCount uint32  `json:"maxCount"` // 最大数量
}
