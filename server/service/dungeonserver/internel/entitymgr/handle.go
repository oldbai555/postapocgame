/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package entitymgr

import "math"

var (
	magicMap map[uint32]uint16
	idxMap   map[uint32]uint32
)

func CreateEntityHandle(et uint32) uint64 {
	if _, ok := magicMap[et]; !ok {
		magicMap[et] = 1
	}
	if _, ok := idxMap[et]; !ok {
		idxMap[et] = 1
	}
	if idxMap[et] >= math.MaxUint32 {
		idxMap[et] = 0
		if magicMap[et] >= math.MaxUint8 {
			magicMap[et] = 0
		}
		magicMap[et]++
	}
	idxMap[et]++
	return uint64(idxMap[et] | et<<8 | uint32(magicMap[et]))
}
func init() {
	idxMap = make(map[uint32]uint32)
	magicMap = make(map[uint32]uint16)
}
