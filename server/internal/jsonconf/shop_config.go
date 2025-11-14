/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 商城配置
**/

package jsonconf

// ShopConfig 商城配置
type ShopConfig struct {
	ShopId     uint32 `json:"shopId"`     // 商城ID
	ItemId     uint32 `json:"itemId"`     // 物品ID
	ConsumeId  uint32 `json:"consumeId"`  // 消耗配置ID
	RewardId   uint32 `json:"rewardId"`   // 奖励配置ID
	LimitCount uint32 `json:"limitCount"` // 限购数量（0表示不限购）
	LimitType  uint32 `json:"limitType"`  // 限购类型: 1=每日限购 2=每周限购 3=永久限购
}
