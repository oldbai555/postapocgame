package manager

import (
	"context"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/iface"
	"sync"
)

// PlayerRoleManager 玩家角色管理器
type PlayerRoleManager struct {
	mu      sync.RWMutex
	roleMgr map[uint64]iface.IPlayerRole // roleId -> PlayerRole
}

var (
	once              sync.Once
	playerRoleManager *PlayerRoleManager
)

// GetPlayerRoleManager 获取全局玩家角色管理器
func GetPlayerRoleManager() *PlayerRoleManager {
	once.Do(func() {
		playerRoleManager = &PlayerRoleManager{
			roleMgr: make(map[uint64]iface.IPlayerRole),
		}
	})
	return playerRoleManager
}

// Add 添加玩家角色
func (m *PlayerRoleManager) Add(playerRole iface.IPlayerRole) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.roleMgr[playerRole.GetPlayerRoleId()] = playerRole
}

// Remove 移除玩家角色
func (m *PlayerRoleManager) Remove(playerRoleId uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if playerRole, ok := m.roleMgr[playerRoleId]; ok {
		delete(m.roleMgr, playerRole.GetPlayerRoleId())
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

func (m *PlayerRoleManager) GetBySession(sessionId string) iface.IPlayerRole {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, role := range m.roleMgr {
		if role.GetSessionId() == sessionId {
			return role
		}
	}
	return nil
}

// FlushAndSave 遍历所有在线角色并同步保存数据，用于优雅停服
func (m *PlayerRoleManager) FlushAndSave(ctx context.Context) {
	roles := m.GetAll()
	total := len(roles)
	if total == 0 {
		return
	}

	log.Infof("FlushAndSave start: total=%d", total)
	for idx, role := range roles {
		if role == nil {
			continue
		}

		if ctx != nil {
			select {
			case <-ctx.Done():
				log.Warnf("FlushAndSave cancelled after %d/%d roles: %v", idx, total, ctx.Err())
				return
			default:
			}
		}

		if err := role.SaveToDB(); err != nil {
			log.Errorf("FlushAndSave roleId=%d failed: %v", role.GetPlayerRoleId(), err)
		}
	}
	log.Infof("FlushAndSave completed: total=%d", total)
}
