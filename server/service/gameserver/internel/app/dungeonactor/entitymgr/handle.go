/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package entitymgr

import "math"

var (
	magicMap = make(map[uint32]uint16)
	idxMap   = make(map[uint32]uint32)
)

func CreateEntityHandle(et uint32) uint64 {
	if _, ok := magicMap[et]; !ok {
		magicMap[et] = 1
	}
	if _, ok := idxMap[et]; !ok {
		idxMap[et] = 1
	}

	// 检查是否需要重置 idxMap（当达到最大值时）
	if idxMap[et] == math.MaxUint32 {
		idxMap[et] = 1
		// 增加 magic 值，如果达到最大值则重置
		if magicMap[et] == math.MaxUint16 {
			magicMap[et] = 1
		} else {
			magicMap[et]++
		}
	} else {
		idxMap[et]++
	}

	idx := idxMap[et]

	// 正确的组合方式: [48位idx][8位et][8位magic]
	return uint64(idx) | (uint64(et) << 48) | (uint64(magicMap[et]) << 56)
}
func init() {}
