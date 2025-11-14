/**
 * @Author: zjj
 * @Date: 2025/01/XX
 * @Desc: 每日任务配置
**/

package jsonconf

// DailyQuestConfig 每日任务配置
type DailyQuestConfig struct {
	QuestId     uint32        `json:"questId"`     // 任务ID
	Name        string        `json:"name"`        // 任务名称
	Description string        `json:"description"` // 任务描述
	Level       uint32        `json:"level"`       // 需求等级
	Targets     []QuestTarget `json:"targets"`     // 任务目标数组
	Rewards     []ItemSt      `json:"rewards"`     // 任务奖励
	ExpReward   uint64        `json:"expReward"`   // 经验奖励
	MaxCount    uint32        `json:"maxCount"`    // 每日最大完成次数（0表示无限制）
}
