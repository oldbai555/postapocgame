/**
 * @Author: zjj
 * @Date: 2025/01/XX
 * @Desc: 背包配置
**/

package jsonconf

// BagConfig 背包配置
type BagConfig struct {
	BagType uint32 `json:"bagType"` // 背包类型（1=主背包，2=仓库等）
	Size    uint32 `json:"size"`    // 背包默认容量
}
