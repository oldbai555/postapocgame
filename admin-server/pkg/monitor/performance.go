package monitor

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// PerformanceMonitor 接口性能监控
type PerformanceMonitor struct {
	slowThreshold int64 // 慢接口阈值（毫秒）
}

// NewPerformanceMonitor 创建接口性能监控实例
func NewPerformanceMonitor(slowThreshold int64) *PerformanceMonitor {
	if slowThreshold <= 0 {
		slowThreshold = 2000 // 默认 2 秒
	}
	return &PerformanceMonitor{
		slowThreshold: slowThreshold,
	}
}

// RecordAPICall 记录接口调用
func (m *PerformanceMonitor) RecordAPICall(ctx context.Context, method, path string, duration time.Duration, statusCode int, err error) {
	durationMs := duration.Milliseconds()

	// 记录接口调用日志
	if err != nil {
		logx.WithContext(ctx).Errorf("接口调用失败: method=%s, path=%s, duration=%dms, statusCode=%d, error=%v",
			method, path, durationMs, statusCode, err)
	} else if durationMs >= m.slowThreshold {
		// 慢接口警告
		logx.WithContext(ctx).Errorf("慢接口检测: method=%s, path=%s, duration=%dms, threshold=%dms, statusCode=%d",
			method, path, durationMs, m.slowThreshold, statusCode)
	} else {
		// 正常接口调用（可选，避免日志过多）
		// logx.WithContext(ctx).Infof("接口调用: method=%s, path=%s, duration=%dms, statusCode=%d",
		// 	method, path, durationMs, statusCode)
	}
}

// RecordAPICallWithFunc 记录接口调用（使用函数包装）
func (m *PerformanceMonitor) RecordAPICallWithFunc(ctx context.Context, method, path string, fn func() (int, error)) (int, error) {
	startTime := time.Now()
	statusCode, err := fn()
	duration := time.Since(startTime)

	m.RecordAPICall(ctx, method, path, duration, statusCode, err)

	return statusCode, err
}
