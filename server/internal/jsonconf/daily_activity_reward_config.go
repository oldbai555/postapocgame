/**
 * @Author: zjj
 * @Date: 2025/11/17
 * @Desc: 日常活跃奖励配置
**/

package jsonconf

// DailyActivityRewardConfig 日常活跃奖励配置
// 对应配置文件：daily_activity_reward_config.json（可选）
// 示例字段：
// - rewardId: 奖励档位 ID，唯一标识一个活跃奖励
// - requiredPoint: 领取该奖励所需的活跃点数
// - rewards: 基础奖励列表（活跃奖励固定内容）
// - extraItems: 额外奖励列表（如活动加成、首日奖励等，可为空）
type DailyActivityRewardConfig struct {
	RewardId      uint32   `json:"rewardId"`      // 奖励档位 ID
	RequiredPoint uint32   `json:"requiredPoint"` // 所需活跃点
	Rewards       []ItemSt `json:"rewards"`       // 基础奖励
	ExtraItems    []ItemSt `json:"extraItems"`    // 额外奖励（可选）
}
