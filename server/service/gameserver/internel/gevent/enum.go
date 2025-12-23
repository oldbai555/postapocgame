package gevent

import "postapocgame/server/internal/event"

// 服务器级别事件
const (
	OnSrvStart event.Type = iota + 1
	OnSrvStop
)

// 玩家级别事件（从1000开始，避免冲突）
const (
	// 玩家登录相关
	OnPlayerLogin event.Type = iota + 1000
	OnPlayerLogout
)
