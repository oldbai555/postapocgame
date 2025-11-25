/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// MonsterConfig 怪物配置
type MonsterConfig struct {
	MonsterId uint32          `json:"monsterId"` // 怪物Id
	Name      string          `json:"name"`      // 怪物名称
	Level     uint32          `json:"level"`     // 等级
	Type      uint32          `json:"type"`      // 类型: 1=普通 2=精英 3=BOSS
	HP        uint32          `json:"hp"`        // 生命值
	MP        uint32          `json:"mp"`        // 魔法值
	Attack    uint32          `json:"attack"`    // 攻击力
	Defense   uint32          `json:"defense"`   // 防御力
	Speed     uint32          `json:"speed"`     // 速度
	SkillIds  []uint32        `json:"skillIds"`  // 技能Id列表
	DropItems []MonsterDrop   `json:"dropItems"` // 掉落物品
	ExpReward uint64          `json:"expReward"` // 经验奖励
	AIConfig  MonsterAIConfig `json:"aiConfig"`  // AI参数
}

// MonsterDrop 怪物掉落
type MonsterDrop struct {
	ItemId          uint32  `json:"itemId"`          // 道具Id
	DropRate        float32 `json:"dropRate"`        // 掉落概率 (0-1)
	MinCount        uint32  `json:"minCount"`        // 最小数量
	MaxCount        uint32  `json:"maxCount"`        // 最大数量
	LifetimeSeconds uint32  `json:"lifetimeSeconds"` // 存在时间（秒），0表示使用默认60秒
}

// PathfindingType 寻路算法类型
type PathfindingType uint32

const (
	PathfindingTypeStraight PathfindingType = 1 // 直线寻路（最短直线，不绕障碍）
	PathfindingTypeAStar    PathfindingType = 2 // A*寻路（绕障碍，贴墙走）
)

// MonsterAIConfig 怪物AI参数
type MonsterAIConfig struct {
	PatrolRadius      uint32          `json:"patrolRadius"`      // 巡逻半径
	DetectRange       uint32          `json:"detectRange"`       // 侦测范围
	AttackRange       uint32          `json:"attackRange"`       // 攻击范围
	ResetDistance     uint32          `json:"resetDistance"`     // 超出此距离则回家
	ThinkIntervalMS   uint32          `json:"thinkInterval"`     // 决策间隔(ms)
	PatrolPathfinding PathfindingType `json:"patrolPathfinding"` // 巡逻时使用的寻路算法: 1=直线 2=A*
	ChasePathfinding  PathfindingType `json:"chasePathfinding"`  // 追击时使用的寻路算法: 1=直线 2=A*
}

// WithDefaults 填充默认值
func (cfg MonsterAIConfig) WithDefaults() MonsterAIConfig {
	if cfg.PatrolRadius == 0 {
		cfg.PatrolRadius = 120
	}
	if cfg.DetectRange == 0 {
		cfg.DetectRange = 320
	}
	if cfg.AttackRange == 0 {
		cfg.AttackRange = 40
	}
	if cfg.ResetDistance == 0 {
		cfg.ResetDistance = cfg.DetectRange * 2
	}
	if cfg.ThinkIntervalMS == 0 {
		cfg.ThinkIntervalMS = 500
	}
	if cfg.PatrolPathfinding == 0 {
		cfg.PatrolPathfinding = PathfindingTypeStraight // 默认巡逻用直线
	}
	if cfg.ChasePathfinding == 0 {
		cfg.ChasePathfinding = PathfindingTypeAStar // 默认追击用A*
	}
	return cfg
}
