package jsonconf

// RewardConfig 通用奖励配置
type RewardConfig struct {
	RewardId    uint32        `json:"rewardId"`
	Items       []*ItemAmount `json:"items"`
	Description string        `json:"description"`
}
