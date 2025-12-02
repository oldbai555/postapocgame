package interfaces

import "context"

// TimeCallbackUseCase 时间回调用例接口（Use Case 层，可选）
type TimeCallbackUseCase interface {
	// OnNewHour 新小时回调
	OnNewHour(ctx context.Context) error

	// OnNewDay 新天回调
	OnNewDay(ctx context.Context) error

	// OnNewWeek 新周回调
	OnNewWeek(ctx context.Context) error

	// OnNewMonth 新月回调
	OnNewMonth(ctx context.Context) error

	// OnNewYear 新年回调
	OnNewYear(ctx context.Context) error
}
