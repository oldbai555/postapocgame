package event

import (
	"context"
	"sync"
	"time"

	"postapocgame/server/pkg/log"
)

// ISystem 系统接口
type ISystem interface {
	// Init 初始化系统（注册事件处理器）
	Init(mgr *SystemMgr)
	// Name 系统名称
	Name() string
	// OnClose 系统关闭时回调（可选）
	OnClose()
}

// SystemFactory 系统工厂函数
type SystemFactory func() ISystem

// SystemMgr 系统管理器（代表一个 actor）
type SystemMgr struct {
	id       string             // actor ID
	localBus *Bus               // 本地事件总线
	mailbox  Mailbox            // 邮箱
	systems  map[string]ISystem // 系统实例

	actorCtx context.Context
	cancel   context.CancelFunc

	unregister func() // 注销函数
	wg         sync.WaitGroup
}

var (
	// 全局模板（单例）
	globalBusTemplate *Bus
	localBusTemplate  *Bus

	// actor 注册表
	actorRegistry *ActorRegistry

	// 系统工厂注册表
	systemFactoriesMu sync.RWMutex
	systemFactories   []SystemFactory

	// 初始化一次
	eventInitOnce sync.Once
)

// initEventSystem 初始化事件系统
func initEventSystem() {
	eventInitOnce.Do(func() {
		globalBusTemplate = NewEventBus()
		localBusTemplate = NewEventBus()
		actorRegistry = NewActorRegistry()
	})
}

// RegisterSystemFactory 注册系统工厂（在 init() 中调用）
func RegisterSystemFactory(factory SystemFactory) {
	initEventSystem()

	systemFactoriesMu.Lock()
	systemFactories = append(systemFactories, factory)
	systemFactoriesMu.Unlock()

	// 将工厂注册到模板的 registry（用于克隆时重放）
	localBusTemplate.registry = append(localBusTemplate.registry, func(b *Bus) {
		// no-op: factory-based systems will be instantiated at actor create time
	})
	globalBusTemplate.registry = append(globalBusTemplate.registry, func(b *Bus) {
		// no-op placeholder
	})
}

// NewSystemMgr 创建系统管理器（actor）
func NewSystemMgr(actorID string, mailboxSize int) *SystemMgr {
	initEventSystem()

	ctx, cancel := context.WithCancel(context.Background())
	mgr := &SystemMgr{
		id:       actorID,
		actorCtx: ctx,
		cancel:   cancel,
		mailbox:  make(Mailbox, mailboxSize),
		systems:  make(map[string]ISystem),
	}

	// 克隆 localBus（通过 registry 重放）
	mgr.localBus = localBusTemplate.CloneByReplay()

	// 实例化系统并让它们注册事件处理器
	systemFactoriesMu.RLock()
	factories := append([]SystemFactory(nil), systemFactories...)
	systemFactoriesMu.RUnlock()

	for _, factory := range factories {
		sys := factory()
		sys.Init(mgr)
		mgr.systems[sys.Name()] = sys
	}

	// 注册到 actor registry
	unreg := actorRegistry.Register(actorID, mgr.mailbox)
	mgr.unregister = unreg

	// 启动邮箱处理协程
	mgr.wg.Add(1)
	go mgr.mailboxLoop()

	log.Infof("[EventSystem] Actor created: actorID=%s, systems=%d", actorID, len(mgr.systems))

	return mgr
}

// mailboxLoop 邮箱处理循环（保证顺序性）
func (m *SystemMgr) mailboxLoop() {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.mailbox:
			// 创建带超时的 context
			ctx, cancel := context.WithTimeout(m.actorCtx, 5*time.Second)

			// 在 actor 的协程中同步执行所有 handlers（保证顺序性）
			if err := m.localBus.Publish(ctx, event); err != nil {
				log.Errorf("[EventSystem] Actor %s handle event error: type=%d, err=%v",
					m.id, event.Type, err)
			}

			cancel()

		case <-m.actorCtx.Done():
			// 退出前尝试处理剩余事件（可选）
			m.drainMailbox()
			if m.unregister != nil {
				m.unregister()
			}
			return
		}
	}
}

// drainMailbox 排空邮箱（快速处理剩余事件）
func (m *SystemMgr) drainMailbox() {
	for {
		select {
		case event := <-m.mailbox:
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_ = m.localBus.Publish(ctx, event)
			cancel()
		default:
			return
		}
	}
}

// PublishLocal 发布本地事件（直接在 localBus 上发布）
// 通常在 actor 内部调用
func (m *SystemMgr) PublishLocal(event *Event) error {
	ctx, cancel := context.WithTimeout(m.actorCtx, 3*time.Second)
	defer cancel()

	return m.localBus.Publish(ctx, event)
}

// Subscribe 在 localBus 上订阅事件
func (m *SystemMgr) Subscribe(eventType Type, priority int, handler Handler) {
	m.localBus.Subscribe(eventType, priority, handler)
}

// GetSystem 获取系统实例
func (m *SystemMgr) GetSystem(name string) ISystem {
	return m.systems[name]
}

// GetActorID 获取 actor ID
func (m *SystemMgr) GetActorID() string {
	return m.id
}

// GetLocalBus 获取本地事件总线
func (m *SystemMgr) GetLocalBus() *Bus {
	return m.localBus
}

// Close 关闭系统管理器
func (m *SystemMgr) Close() {
	log.Infof("[EventSystem] Closing actor: actorID=%s", m.id)

	// 通知所有系统关闭
	for _, sys := range m.systems {
		sys.OnClose()
	}

	// 取消 context
	m.cancel()

	// 等待邮箱处理完成
	m.wg.Wait()

	log.Infof("[EventSystem] Actor closed: actorID=%s", m.id)
}

// IsRunning 检查 actor 是否运行中
func (m *SystemMgr) IsRunning() bool {
	select {
	case <-m.actorCtx.Done():
		return false
	default:
		return true
	}
}
