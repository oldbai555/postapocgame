package monitor

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// DefaultSlowQueryThreshold 默认慢查询阈值（毫秒）
	DefaultSlowQueryThreshold = 1000
)

// SlowQueryMonitor 慢查询监控
type SlowQueryMonitor struct {
	threshold int64 // 慢查询阈值（毫秒）
}

// NewSlowQueryMonitor 创建慢查询监控实例
func NewSlowQueryMonitor(threshold int64) *SlowQueryMonitor {
	if threshold <= 0 {
		threshold = DefaultSlowQueryThreshold
	}
	return &SlowQueryMonitor{
		threshold: threshold,
	}
}

// RecordSlowQuery 记录慢查询
func (m *SlowQueryMonitor) RecordSlowQuery(ctx context.Context, query string, duration time.Duration, err error) {
	if duration.Milliseconds() < m.threshold {
		return
	}

	// 记录慢查询日志
	logx.WithContext(ctx).Errorf("慢查询检测: duration=%dms, threshold=%dms, query=%s, error=%v",
		duration.Milliseconds(), m.threshold, query, err)
}

// RecordQuery 记录查询（自动判断是否为慢查询）
func (m *SlowQueryMonitor) RecordQuery(ctx context.Context, query string, fn func() error) error {
	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	// 如果是慢查询，记录日志
	if duration.Milliseconds() >= m.threshold {
		m.RecordSlowQuery(ctx, query, duration, err)
	}

	return err
}
