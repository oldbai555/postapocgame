/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc: 物品回收配置
**/

package jsonconf

// ItemRecycleConfig 物品回收配置
type ItemRecycleConfig struct {
	ItemId uint32        `json:"itemId"` // 物品ID
	Awards []*ItemAmount `json:"awards"` // 回收奖励（物品列表）
}
