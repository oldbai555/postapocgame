package system

import "context"

// GetRecycleSys 获取回收系统适配器（保持接口一致性）
func GetRecycleSys(ctx context.Context) *RecycleSystemAdapter {
	_ = ctx // 当前回收系统无状态，仅保留参数以保持接口一致
	return getRecycleSysInstance()
}
