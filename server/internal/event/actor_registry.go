package event

import (
	"sync"
)

// Mailbox actor 的邮箱类型
type Mailbox chan *Event

// ActorRegistry 维护所有 actor 的邮箱，支持全局广播
type ActorRegistry struct {
	mu        sync.RWMutex
	mailboxes map[string]Mailbox // actorID -> mailbox
	closed    bool
}

// NewActorRegistry 创建 actor 注册表
func NewActorRegistry() *ActorRegistry {
	return &ActorRegistry{
		mailboxes: make(map[string]Mailbox),
	}
}

// Register 注册 actor 邮箱
// 返回 unregister 函数用于注销
func (r *ActorRegistry) Register(actorID string, mailbox Mailbox) (unregister func()) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return func() {}
	}

	r.mailboxes[actorID] = mailbox
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.mailboxes, actorID)
	}
}

// Broadcast 广播事件到所有 actor
// 使用非阻塞策略：当邮箱满时丢弃（避免慢消费者卡住发布者）
func (r *ActorRegistry) Broadcast(event *Event) {
	r.mu.RLock()
	// 复制一份 mailboxes 列表（避免持锁时间过长）
	mailboxes := make([]Mailbox, 0, len(r.mailboxes))
	for _, m := range r.mailboxes {
		mailboxes = append(mailboxes, m)
	}
	r.mu.RUnlock()

	for _, mailbox := range mailboxes {
		select {
		case mailbox <- event:
			// 发送成功
		default:
			// 邮箱满，丢弃事件（生产环境应该记录 metrics）
			// 可以改为阻塞或其他策略
		}
	}
}

// BroadcastToActors 广播事件到指定的 actors
func (r *ActorRegistry) BroadcastToActors(event *Event, actorIDs []string) {
	r.mu.RLock()
	mailboxes := make([]Mailbox, 0, len(actorIDs))
	for _, id := range actorIDs {
		if mailbox, ok := r.mailboxes[id]; ok {
			mailboxes = append(mailboxes, mailbox)
		}
	}
	r.mu.RUnlock()

	for _, mailbox := range mailboxes {
		select {
		case mailbox <- event:
		default:
			// 丢弃
		}
	}
}

// SendToActor 发送事件到指定 actor
func (r *ActorRegistry) SendToActor(actorID string, event *Event) bool {
	r.mu.RLock()
	mailbox, ok := r.mailboxes[actorID]
	r.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case mailbox <- event:
		return true
	default:
		return false
	}
}

// GetActorCount 获取 actor 数量
func (r *ActorRegistry) GetActorCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.mailboxes)
}

// Close 关闭注册表
func (r *ActorRegistry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true
	r.mailboxes = make(map[string]Mailbox)
}

// HasActor 检查 actor 是否存在
func (r *ActorRegistry) HasActor(actorID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.mailboxes[actorID]
	return ok
}

// GetAllActorIDs 获取所有 actor IDs
func (r *ActorRegistry) GetAllActorIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.mailboxes))
	for id := range r.mailboxes {
		ids = append(ids, id)
	}
	return ids
}
