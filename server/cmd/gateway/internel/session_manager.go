/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"errors"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"sync"
	"time"
)

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionClosed     = errors.New("session already closed")
	ErrMaxSessionReached = errors.New("max sessions reached")
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	config   *Config
	gsConn   IGameServerConnector
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewSessionManager 创建会话管理器
func NewSessionManager(cfg *Config, gsConn IGameServerConnector) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		config:   cfg,
		gsConn:   gsConn,
		stopChan: make(chan struct{}),
	}
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(conn IConnection) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.sessions) >= sm.config.MaxSessions {
		return nil, ErrMaxSessionReached
	}

	sessionID := tool.GenUUID()
	now := time.Now()

	session := &Session{
		ID:         sessionID,
		Addr:       conn.RemoteAddr(),
		ConnType:   conn.Type(),
		State:      SessionStateConnected,
		SendChan:   make(chan []byte, sm.config.SessionBufferSize),
		CreatedAt:  now,
		LastActive: now,
	}

	sm.sessions[sessionID] = session

	// 通知GameServer新会话
	event := &SessionEvent{
		Type:      SessionEventNew,
		SessionID: sessionID,
		Timestamp: now,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sm.gsConn.NotifySessionEvent(ctx, event); err != nil {
			log.Infof("notify session event failed: %v", err)
		}
	}()

	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, ok := sm.sessions[sessionID]
	return session, ok
}

// AuthSession 认证会话
func (sm *SessionManager) AuthSession(sessionID, userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	session.State = SessionStateAuthed
	session.UserID = userID
	session.LastActive = time.Now()

	// 通知GameServer会话认证
	event := &SessionEvent{
		Type:      SessionEventAuth,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sm.gsConn.NotifySessionEvent(ctx, event); err != nil {
			log.Infof("notify auth event failed: %v", err)
		}
	}()

	return nil
}

// CloseSession 关闭会话
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	session, ok := sm.sessions[sessionID]
	if !ok {
		sm.mu.Unlock()
		return ErrSessionNotFound
	}

	if session.State == SessionStateClosed {
		sm.mu.Unlock()
		return ErrSessionClosed
	}

	session.State = SessionStateClosed
	close(session.SendChan)
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	// 通知GameServer会话关闭
	event := &SessionEvent{
		Type:      SessionEventClose,
		SessionID: sessionID,
		UserID:    session.UserID,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sm.gsConn.NotifySessionEvent(ctx, event); err != nil {
		log.Infof("notify close event failed: %v", err)
	}

	return nil
}

// UpdateActivity 更新会话活跃时间
func (sm *SessionManager) UpdateActivity(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, ok := sm.sessions[sessionID]; ok {
		session.LastActive = time.Now()
	}
}

// StartCleanup 启动会话清理协程
func (sm *SessionManager) StartCleanup(ctx context.Context) {
	sm.wg.Add(1)
	go func() {
		defer sm.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sm.stopChan:
				return
			case <-ticker.C:
				sm.cleanupTimeoutSessions()
			}
		}
	}()
}

// cleanupTimeoutSessions 清理超时会话
func (sm *SessionManager) cleanupTimeoutSessions() {
	now := time.Now()
	var toClose []string

	sm.mu.RLock()
	for id, session := range sm.sessions {
		if now.Sub(session.LastActive) > sm.config.SessionTimeout {
			toClose = append(toClose, id)
		}
	}
	sm.mu.RUnlock()

	for _, id := range toClose {
		log.Infof("closing timeout session: %s", id)
		err := sm.CloseSession(id)
		if err != nil {
			log.Errorf("CloseSession err: %s", err)
		}
	}
}

// Stop 停止会话管理器
func (sm *SessionManager) Stop() {
	close(sm.stopChan)
	sm.wg.Wait()

	// 关闭所有会话
	sm.mu.Lock()
	for id := range sm.sessions {
		err := sm.CloseSession(id)
		if err != nil {
			log.Errorf("CloseSession err: %s", err)
		}
	}
	sm.mu.Unlock()
}

// GetSessionCount 获取会话数量
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
