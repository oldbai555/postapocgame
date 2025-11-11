package actor

import (
	"context"
	"fmt"
	"time"

	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
)

// MailboxFullStrategy 邮箱满策略
type MailboxFullStrategy int

const (
	StrategyBlock  MailboxFullStrategy = iota // 阻塞(默认)
	StrategyDrop                              // 丢弃新消息
	StrategyReject                            // 拒绝(返回错误)
)

// Actor Actor实现
type Actor struct {
	mode     Mode
	handler  IActorMsgHandler
	mailbox  chan *Message
	stopChan chan struct{}
	wg       sync.WaitGroup

	// 邮箱满策略
	fullStrategy MailboxFullStrategy
	// 监控
	monitor *monitoredActor
}

// NewActor 创建Actor
func NewActor(mode Mode, handler IActorMsgHandler) *Actor {
	return NewActorWithStrategy(mode, handler, StrategyBlock)
}

// NewActorWithStrategy 创建Actor(指定邮箱满策略)
func NewActorWithStrategy(mode Mode, handler IActorMsgHandler, strategy MailboxFullStrategy) *Actor {
	mailboxSize := 2000

	a := &Actor{
		mode:         mode,
		handler:      handler,
		mailbox:      make(chan *Message, mailboxSize),
		stopChan:     make(chan struct{}),
		fullStrategy: strategy,
	}

	// 注册到监控
	a.monitor = GetActorMonitor().Register(mode, mailboxSize)

	return a
}

// Start 启动Actor
func (a *Actor) Start(ctx context.Context) error {
	if a.handler == nil {
		return fmt.Errorf("actor handler is nil, mode:%d", a.mode)
	}
	a.wg.Add(1)
	routine.GoV2(func() error {
		a.run(ctx)
		return nil
	})
	log.Infof("actor started, mode:%d", a.mode)
	return nil
}

// Stop 停止Actor
func (a *Actor) Stop(ctx context.Context) error {
	close(a.stopChan)
	a.wg.Wait()
	log.Infof("actor stopped, mode:%d", a.mode)
	return nil
}

// Send 发送消息到Actor(根据策略处理邮箱满)
func (a *Actor) Send(msg *Message) error {
	switch a.fullStrategy {
	case StrategyBlock:
		// 阻塞发送
		select {
		case a.mailbox <- msg:
			if a.monitor != nil {
				a.monitor.incMailboxSize()
			}
			return nil
		case <-time.After(3 * time.Second):
			// 超时告警
			log.Warnf("actor mailbox full (blocking), mode:%d, waited 3s", a.mode)
			a.mailbox <- msg // 继续阻塞
			if a.monitor != nil {
				a.monitor.incMailboxSize()
			}
			return nil
		}

	case StrategyDrop:
		// 丢弃新消息
		select {
		case a.mailbox <- msg:
			if a.monitor != nil {
				a.monitor.incMailboxSize()
			}
			return nil
		default:
			log.Warnf("actor mailbox full (dropped), mode:%d", a.mode)
			if a.monitor != nil {
				a.monitor.incDropped()
			}
			return nil // 不返回错误,静默丢弃
		}

	case StrategyReject:
		// 拒绝
		select {
		case a.mailbox <- msg:
			if a.monitor != nil {
				a.monitor.incMailboxSize()
			}
			return nil
		default:
			if a.monitor != nil {
				a.monitor.incDropped()
			}
			return fmt.Errorf("actor mailbox full, mode:%d", a.mode)
		}

	default:
		return fmt.Errorf("unknown mailbox full strategy: %d", a.fullStrategy)
	}
}

// Call 调用Actor(同步,等待响应)
func (a *Actor) Call(ctx context.Context, msg *Message) (*Message, error) {
	// 创建响应通道
	replyChan := make(chan *Message, 1)
	msg.ReplyTo = replyChan

	// 发送消息
	if err := a.Send(msg); err != nil {
		return nil, err
	}

	// 等待响应(带超时)
	timeout := 3 * time.Second
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("actor call timeout, mode:%d", a.mode)
	case reply := <-replyChan:
		return reply, nil
	}
}

func (a *Actor) GetMode() Mode {
	return a.mode
}

// run Actor运行循环(单线程串行处理)
func (a *Actor) run(ctx context.Context) {
	defer a.wg.Done()

	for {
		routine.Run(a.handler.Loop)
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case msg := <-a.mailbox:
			// 监控
			if a.monitor != nil {
				a.monitor.decMailboxSize()
			}

			start := time.Now()
			err := a.handler.HandleActorMessage(msg)
			elapsed := time.Since(start).Milliseconds()

			// 记录处理时间
			if a.monitor != nil {
				a.monitor.addProcessTime(elapsed)
				a.monitor.incProcessed()
			}

			if err != nil {
				log.Errorf("actor handle message failed: mode:%d, msgId:%d, err:%v", a.mode, msg.MsgId, err)

				if a.monitor != nil {
					a.monitor.incFailed()
				}

				// 如果有回复通道,发送错误响应
				if msg.ReplyTo != nil {
					errorReply := &Message{
						MsgId: protocol.S2C_Error,
						Data:  []byte(err.Error()),
					}
					select {
					case msg.ReplyTo <- errorReply:
					default:
					}
				}
			}
		}
	}
}
