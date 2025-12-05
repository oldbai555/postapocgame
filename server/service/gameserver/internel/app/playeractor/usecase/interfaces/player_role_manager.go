package interfaces

import (
	"context"
	"postapocgame/server/service/gameserver/internel/iface"
)

// IPlayerRoleManager 玩家角色管理器接口（用于依赖注入和单元测试）
type IPlayerRoleManager interface {
	// Add 添加玩家角色
	Add(playerRole iface.IPlayerRole)

	// Remove 移除玩家角色
	Remove(playerRoleId uint64)

	// Get 通过 RoleID 获取玩家角色
	Get(playerRoleId uint64) (iface.IPlayerRole, bool)

	// GetAll 获取所有玩家角色
	GetAll() []iface.IPlayerRole

	// GetBySession 通过 SessionID 获取玩家角色（O(1) 查找）
	GetBySession(sessionId string) iface.IPlayerRole

	// UpdateSession 更新角色的 SessionID 索引（用于重连等场景）
	UpdateSession(roleId uint64, oldSessionId, newSessionId string)

	// FlushAndSave 遍历所有在线角色并同步保存数据，用于优雅停服
	// ctx: 上下文，用于超时控制；batchSize: 每批处理的角色数量，0 表示不限制
	FlushAndSave(ctx context.Context, batchSize int) error
}
