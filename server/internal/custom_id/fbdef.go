/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package custom_id

// FuBenType 副本类型
type FuBenType uint32

const (
	FuBenTypePermanent   FuBenType = 1 // 常驻副本
	FuBenTypeTimed       FuBenType = 2 // 限时副本
	FuBenTypeTimedSingle FuBenType = 3 // 限时单人副本
	FuBenTypeTimedMulti  FuBenType = 4 // 限时多人副本
)

// FuBenState 副本状态
type FuBenState uint32

const (
	FuBenStateNormal  FuBenState = 1 // 正常
	FuBenStateClosing FuBenState = 2 // 关闭中
	FuBenStateClosed  FuBenState = 3 // 已关闭
)
