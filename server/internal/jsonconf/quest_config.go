/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package jsonconf

// QuestConfig 任务配置
type QuestConfig struct {
	QuestId     uint32        `json:"questId"`     // 任务Id
	Name        string        `json:"name"`        // 任务名称
	Type        uint32        `json:"type"`        // 任务类型: 1=主线 2=支线 3=日常 4=周常
	Description string        `json:"description"` // 任务描述
	Level       uint32        `json:"level"`       // 需求等级
	PreQuests   []uint32      `json:"preQuests"`   // 前置任务
	NextQuests  []uint32      `json:"nextQuests"`  // 后续任务（任务链）
	Targets     []QuestTarget `json:"targets"`     // 任务目标数组 [type, ids, count]
	Rewards     []ItemSt      `json:"rewards"`     // 任务奖励
	ExpReward   uint64        `json:"expReward"`   // 经验奖励
	AutoTrack   bool          `json:"autoTrack"`   // 是否自动追踪
	MaxCount    uint32        `json:"maxCount"`    // 每日/周任务可完成次数（0表示不限）
	ActivePoint uint32        `json:"activePoint"` // 完成后奖励的活跃点
}

// QuestTarget 任务目标
// type: 任务类型（1=和NPC对话，2=学习任意技能，3=击杀任意怪物）
// ids: 条件数组，是一个整形数组（对于type=1，ids配置npcId；对于type=2和3，ids可以为空）
// count: 达标数量
type QuestTarget struct {
	Type  uint32   `json:"type"`  // 任务类型
	Ids   []uint32 `json:"ids"`   // 条件数组（整形数组）
	Count uint32   `json:"count"` // 达标数量
}
