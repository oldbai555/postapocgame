/**
 * @Author: zjj
 * @Date: 2025/11/12
 * @Desc:
**/

package jsonconf

type ItemSt struct {
	ItemId uint32 `json:"itemId"` // 道具Id
	Type   uint32 `json:"type"`   // 道具类型
	Count  uint32 `json:"count"`  // 数量
}
