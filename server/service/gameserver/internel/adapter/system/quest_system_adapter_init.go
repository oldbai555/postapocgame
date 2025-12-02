package system

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysQuest), func() iface.ISystem {
		return NewQuestSystemAdapter()
	})

	// 每日/每周刷新由 DailyActivity 或时间事件统一驱动，这里暂不直接订阅 gevent
}
