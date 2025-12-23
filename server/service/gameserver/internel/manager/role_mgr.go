package manager

import (
	"context"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"sync"
)

// PlayerRoleManager 玩家角色管理器
type PlayerRoleManager struct {
	mu           sync.RWMutex
	roleMgr      map[uint64]iface.IPlayerRole // roleId -> PlayerRole
	sessionIndex map[string]uint64            // sessionId -> roleId (用于 O(1) 查找)
}

var (
	once              sync.Once
	playerRoleManager *PlayerRoleManager
)

// GetPlayerRoleManager 获取全局玩家角色管理器
func GetPlayerRoleManager() *PlayerRoleManager {
	once.Do(func() {
		playerRoleManager = &PlayerRoleManager{
			roleMgr:      make(map[uint64]iface.IPlayerRole),
			sessionIndex: make(map[string]uint64),
		}
	})
	if playerRoleManager == nil {
		panic("playerRoleManager not initialized")
	}
	return playerRoleManager
}

// Add 添加玩家角色
func (m *PlayerRoleManager) Add(playerRole iface.IPlayerRole) {
	m.mu.Lock()
	defer m.mu.Unlock()

	roleId := playerRole.GetPlayerRoleId()
	sessionId := playerRole.GetSessionId()

	// 如果已存在相同 roleId 的角色，先清理旧的 sessionIndex 条目
	if oldRole, exists := m.roleMgr[roleId]; exists {
		oldSessionId := oldRole.GetSessionId()
		if oldSessionId != "" && oldSessionId != sessionId {
			delete(m.sessionIndex, oldSessionId)
		}
	}

	m.roleMgr[roleId] = playerRole
	// 同步维护 sessionIndex
	if sessionId != "" {
		m.sessionIndex[sessionId] = roleId
	}
}

// Remove 移除玩家角色
func (m *PlayerRoleManager) Remove(playerRoleId uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if playerRole, ok := m.roleMgr[playerRoleId]; ok {
		// 同步删除 sessionIndex 中的条目
		sessionId := playerRole.GetSessionId()
		if sessionId != "" {
			delete(m.sessionIndex, sessionId)
		}
		delete(m.roleMgr, playerRoleId)
	}
}

// Get 通过SessionID获取玩家角色
func (m *PlayerRoleManager) Get(playerRoleId uint64) (iface.IPlayerRole, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	playerRole, ok := m.roleMgr[playerRoleId]
	return playerRole, ok
}

// GetAll 获取所有玩家角色
func (m *PlayerRoleManager) GetAll() []iface.IPlayerRole {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := make([]iface.IPlayerRole, 0, len(m.roleMgr))
	for _, playerRole := range m.roleMgr {
		roles = append(roles, playerRole)
	}
	return roles
}

// UpdateSession 更新角色的 SessionID 索引（用于重连等场景）
func (m *PlayerRoleManager) UpdateSession(roleId uint64, oldSessionId, newSessionId string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 删除旧的 sessionIndex 条目
	if oldSessionId != "" {
		delete(m.sessionIndex, oldSessionId)
	}

	// 添加新的 sessionIndex 条目
	if newSessionId != "" {
		m.sessionIndex[newSessionId] = roleId
	}
}

// GetBySession 通过 SessionID 获取玩家角色（O(1) 查找）
func (m *PlayerRoleManager) GetBySession(sessionId string) iface.IPlayerRole {
	if sessionId == "" {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// 使用 sessionIndex 进行 O(1) 查找
	roleId, ok := m.sessionIndex[sessionId]
	if !ok {
		return nil
	}

	// 通过 roleId 获取角色
	playerRole, ok := m.roleMgr[roleId]
	if !ok {
		// sessionIndex 与 roleMgr 不一致，清理无效的索引
		// 注意：这里不能直接删除，因为我们在读锁中，需要延迟清理
		// 实际使用中，这种情况应该很少发生，如果发生说明有并发问题
		return nil
	}

	// 验证 sessionId 是否匹配（防止索引不一致）
	if playerRole.GetSessionId() != sessionId {
		return nil
	}

	return playerRole
}

// FlushAndSave 遍历所有在线角色并同步保存数据，用于优雅停服
// ctx: 上下文，用于超时控制；batchSize: 每批处理的角色数量，0 表示不限制
func (m *PlayerRoleManager) FlushAndSave(ctx context.Context, batchSize int) error {
	// 先获取所有角色列表（避免长时间持有锁）
	roles := m.GetAll()
	total := len(roles)
	if total == 0 {
		return nil
	}

	log.Infof("FlushAndSave start: total=%d, batchSize=%d", total, batchSize)

	// 如果指定了批次大小，分批处理
	if batchSize > 0 {
		return m.flushAndSaveBatched(ctx, roles, batchSize, total)
	}

	// 否则逐个处理（保持原有行为）
	return m.flushAndSaveSequential(ctx, roles, total)
}

// flushAndSaveSequential 顺序处理所有角色
func (m *PlayerRoleManager) flushAndSaveSequential(ctx context.Context, roles []iface.IPlayerRole, total int) error {
	for idx, role := range roles {
		if role == nil {
			continue
		}

		// 检查上下文是否已取消
		if ctx != nil {
			select {
			case <-ctx.Done():
				log.Warnf("FlushAndSave cancelled after %d/%d roles: %v", idx, total, ctx.Err())
				return ctx.Err()
			default:
			}
		}

		if err := role.SaveToDB(); err != nil {
			log.Errorf("FlushAndSave roleId=%d failed: %v", role.GetPlayerRoleId(), err)
			// 继续处理其他角色，不因单个失败而中断
		}
	}
	log.Infof("FlushAndSave completed: total=%d", total)
	return nil
}

// flushAndSaveBatched 分批处理角色，避免长时间持有锁
func (m *PlayerRoleManager) flushAndSaveBatched(ctx context.Context, roles []iface.IPlayerRole, batchSize, total int) error {
	processed := 0
	for i := 0; i < len(roles); i += batchSize {
		// 检查上下文是否已取消
		if ctx != nil {
			select {
			case <-ctx.Done():
				log.Warnf("FlushAndSave cancelled after %d/%d roles: %v", processed, total, ctx.Err())
				return ctx.Err()
			default:
			}
		}

		// 确定当前批次的结束位置
		end := i + batchSize
		if end > len(roles) {
			end = len(roles)
		}

		// 处理当前批次
		batch := roles[i:end]
		for _, role := range batch {
			if role == nil {
				continue
			}

			if err := role.SaveToDB(); err != nil {
				log.Errorf("FlushAndSave roleId=%d failed: %v", role.GetPlayerRoleId(), err)
				// 继续处理其他角色
			}
			processed++
		}

		log.Debugf("FlushAndSave progress: %d/%d roles processed", processed, total)
	}

	log.Infof("FlushAndSave completed: total=%d, processed=%d", total, processed)
	return nil
}
