package actor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
)

// ActorMetrics Actor指标
type ActorMetrics struct {
	Mode          Mode
	MailboxSize   int32
	MailboxCap    int
	ProcessedMsgs int64
	FailedMsgs    int64
	DroppedMsgs   int64 // 邮箱满丢弃的消息数
	AvgProcessMs  int64 // 平均处理时间(毫秒)
}

// ActorMonitor Actor监控器
type ActorMonitor struct {
	actors map[Mode]*monitoredActor
	mu     sync.RWMutex

	stopChan chan struct{}
	wg       sync.WaitGroup
}

type monitoredActor struct {
	mode           Mode
	mailboxSize    *atomic.Int32
	mailboxCap     int
	processedMsgs  *atomic.Int64
	failedMsgs     *atomic.Int64
	droppedMsgs    *atomic.Int64
	totalProcessMs *atomic.Int64
}

var (
	globalMonitor *ActorMonitor
	monitorOnce   sync.Once
)

// GetActorMonitor 获取全局监控器
func GetActorMonitor() *ActorMonitor {
	monitorOnce.Do(func() {
		globalMonitor = &ActorMonitor{
			actors:   make(map[Mode]*monitoredActor),
			stopChan: make(chan struct{}),
		}
	})
	return globalMonitor
}

// Register 注册Actor
func (am *ActorMonitor) Register(mode Mode, mailboxCap int) *monitoredActor {
	am.mu.Lock()
	defer am.mu.Unlock()

	ma := &monitoredActor{
		mode:           mode,
		mailboxSize:    &atomic.Int32{},
		mailboxCap:     mailboxCap,
		processedMsgs:  &atomic.Int64{},
		failedMsgs:     &atomic.Int64{},
		droppedMsgs:    &atomic.Int64{},
		totalProcessMs: &atomic.Int64{},
	}

	am.actors[mode] = ma
	return ma
}

// Start 启动监控
func (am *ActorMonitor) Start(ctx context.Context, interval time.Duration) {
	am.wg.Add(1)
	routine.GoV2(func() error {
		am.monitorLoop(ctx, interval)
		return nil
	})
}

// monitorLoop 监控循环
func (am *ActorMonitor) monitorLoop(ctx context.Context, interval time.Duration) {
	defer am.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			am.logMetrics()
			return
		case <-am.stopChan:
			am.logMetrics()
			return
		case <-ticker.C:
			am.logMetrics()
		}
	}
}

// logMetrics 记录指标
func (am *ActorMonitor) logMetrics() {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for mode := range am.actors {
		metrics := am.getMetrics(mode)

		// 计算邮箱使用率
		usage := float64(metrics.MailboxSize) / float64(metrics.MailboxCap) * 100

		log.Infof("[ActorMonitor] Mode=%d, Mailbox=%d/%d(%.1f%%), Processed=%d, Failed=%d, Dropped=%d, AvgMs=%d",
			mode, metrics.MailboxSize, metrics.MailboxCap, usage,
			metrics.ProcessedMsgs, metrics.FailedMsgs, metrics.DroppedMsgs, metrics.AvgProcessMs)

		// 告警
		if usage > 80 {
			log.Warnf("[ActorMonitor] Actor mode=%d mailbox usage high: %.1f%%", mode, usage)
		}

		if metrics.DroppedMsgs > 0 {
			log.Errorf("[ActorMonitor] Actor mode=%d dropped %d messages!", mode, metrics.DroppedMsgs)
		}
	}
}

// GetMetrics 获取指标
func (am *ActorMonitor) GetMetrics(mode Mode) *ActorMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.getMetrics(mode)
}

func (am *ActorMonitor) getMetrics(mode Mode) *ActorMetrics {
	ma, ok := am.actors[mode]
	if !ok {
		return nil
	}

	processed := ma.processedMsgs.Load()
	avgMs := int64(0)
	if processed > 0 {
		avgMs = ma.totalProcessMs.Load() / processed
	}

	return &ActorMetrics{
		Mode:          mode,
		MailboxSize:   ma.mailboxSize.Load(),
		MailboxCap:    ma.mailboxCap,
		ProcessedMsgs: processed,
		FailedMsgs:    ma.failedMsgs.Load(),
		DroppedMsgs:   ma.droppedMsgs.Load(),
		AvgProcessMs:  avgMs,
	}
}

// Stop 停止监控
func (am *ActorMonitor) Stop() {
	close(am.stopChan)
	am.wg.Wait()
}

// 以下方法由Actor调用

func (ma *monitoredActor) incMailboxSize() {
	ma.mailboxSize.Add(1)
}

func (ma *monitoredActor) decMailboxSize() {
	ma.mailboxSize.Add(-1)
}

func (ma *monitoredActor) incProcessed() {
	ma.processedMsgs.Add(1)
}

func (ma *monitoredActor) incFailed() {
	ma.failedMsgs.Add(1)
}

func (ma *monitoredActor) incDropped() {
	ma.droppedMsgs.Add(1)
}

func (ma *monitoredActor) addProcessTime(ms int64) {
	ma.totalProcessMs.Add(ms)
}
