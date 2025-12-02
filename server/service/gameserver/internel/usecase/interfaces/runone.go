package interfaces

import "context"

// RunOneUseCase RunOne 用例接口（Use Case 层，可选）
// 只有需要定期执行的系统才实现此接口
type RunOneUseCase interface {
	// RunOne 每帧调用（在 Actor 主线程中执行）
	RunOne(ctx context.Context) error
}
