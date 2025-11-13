package clientnet

import (
	"context"
	"errors"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"postapocgame/server/pkg/tool"
	"sync"
	"time"

	"postapocgame/server/pkg/customerr"
)

var (
	ErrMaxSessionReached = errors.New("max sessions reached")
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex

	gsConn   IGameServerConnector
	stopChan chan struct{}
	wg       sync.WaitGroup

	sessionBufferSize int           // 每个会话的发送缓冲区大小
	maxSessions       int           // 最大会话数
	sessionTimeout    time.Duration // 会话超时时间
}

// NewSessionManager 创建会话管理器
func NewSessionManager(maxSessions int, sessionBufferSize int, sessionTimeout time.Duration, gsConn IGameServerConnector) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		gsConn:   gsConn,
		stopChan: make(chan struct{}),

		sessionBufferSize: sessionBufferSize,
		maxSessions:       maxSessions,
		sessionTimeout:    sessionTimeout,
	}
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(conn IConnection) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.maxSessions > 0 && len(sm.sessions) >= sm.maxSessions {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "max session reached")
	}

	sessionID := tool.GenUUID()
	now := time.Now()

	session := &Session{
		Id:         sessionID,
		Addr:       conn.RemoteAddr(),
		ConnType:   conn.Type(),
		State:      SessionStateConnected,
		SendChan:   make(chan []byte, sm.sessionBufferSize),
		stopChan:   make(chan struct{}), // 🔧 初始化停止信号
		CreatedAt:  now,
		LastActive: now,
	}

	sm.sessions[sessionID] = session

	// 通知 GameServer
	ev := &network.SessionEvent{
		EventType: network.SessionEventNew,
		SessionId: sessionID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sm.gsConn.NotifySessionEvent(ctx, ev); err != nil {
		log.Errorf("notify session event failed: %v", err)
		return nil, customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}
	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[sessionID]
	return session, ok
}

// CloseSession closes session safely and is fully idempotent
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	session, ok := sm.sessions[sessionID]
	if !ok {
		sm.mu.Unlock()
		// 幂等：如果不存在，视为已关闭成功
		return nil
	}

	// 如果已经标记为 closed，直接返回（幂等）
	if session.State == SessionStateClosed {
		sm.mu.Unlock()
		return nil
	}

	session.State = SessionStateClosed
	// 安全关闭 channel（SafeClose 内部用 sync.Once）
	session.SafeClose()
	// 从 map 中删除
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	// 通知 GameServer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sm.gsConn.NotifySessionEvent(ctx, &network.SessionEvent{
		EventType: network.SessionEventClose,
		SessionId: sessionID,
		UserId:    session.UserId,
	}); err != nil {
		log.Errorf("notify close event failed: %v", err)
	}
	return nil
}

// UpdateActivity 更新会话活跃时间
func (sm *SessionManager) UpdateActivity(sessionId string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, ok := sm.sessions[sessionId]; ok {
		session.LastActive = time.Now()
	}
}

// StartCleanup 启动会话清理协程
func (sm *SessionManager) StartCleanup(ctx context.Context) {
	interval := 30 * time.Second
	sm.wg.Add(1)
	routine.GoV2(func() error {
		defer sm.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-sm.stopChan:
				return nil
			case <-ticker.C:
				sm.cleanupTimeoutSessions()
			}
		}
	})
}

// cleanupTimeoutSessions 清理超时会话
func (sm *SessionManager) cleanupTimeoutSessions() {
	now := time.Now()
	var toClose []string

	sm.mu.RLock()
	for id, session := range sm.sessions {
		if sm.sessionTimeout > 0 && now.Sub(session.LastActive) > sm.sessionTimeout {
			toClose = append(toClose, id)
		}
	}
	sm.mu.RUnlock()

	for _, id := range toClose {
		log.Infof("closing timeout session: %s", id)
		err := sm.CloseSession(id)
		if err != nil {
			log.Errorf("CloseSession %d err:%v", id, err)
		}

	}
}

// Stop 停止会话管理器
func (sm *SessionManager) Stop() {
	// 先告诉 cleanup goroutine 停止
	close(sm.stopChan)
	// 等待 cleanup 等后台 goroutine 退出
	sm.wg.Wait()

	// 复制当前 session id 列表（避免在持锁时做繁重操作）
	sm.mu.RLock()
	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	sm.mu.RUnlock()

	// 逐个安全关闭
	for _, id := range ids {
		if err := sm.CloseSession(id); err != nil {
			log.Errorf("CloseSession err: %s", err)
		}
	}
}

// GetSessionCount 获取会话数量
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
