/**
 * @Author: zjj
 * @Date: 2025/01/XX
 * @Desc: 成就配置
**/

package jsonconf

// AchievementConfig 成就配置
type AchievementConfig struct {
	AchievementId uint32        `json:"achievementId"` // 成就ID
	Name          string        `json:"name"`          // 成就名称
	Description   string        `json:"description"`   // 成就描述
	Type          uint32        `json:"type"`          // 成就类型: 1=等级 2=任务 3=副本 4=战斗
	Level         uint32        `json:"level"`         // 需求等级（0表示无等级要求）
	Targets       []QuestTarget `json:"targets"`       // 成就目标数组
	Rewards       []ItemSt      `json:"rewards"`       // 成就奖励
	ExpReward     uint64        `json:"expReward"`     // 经验奖励
}
