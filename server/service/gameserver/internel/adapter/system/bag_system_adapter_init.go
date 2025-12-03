package system

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

// 注册系统工厂和协议
func init() {
	// 注册系统适配器工厂
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSystemAdapter()
	})

	// 注意：背包系统的事件（OnItemAdd/OnItemRemove/OnBagExpand）由 UseCase 层发布，
	// 其他系统如需响应这些事件，应在各自的 UseCase 中订阅，不在 SystemAdapter 层处理

	// 协议注册由 controller 包负责，避免系统与控制器之间的循环依赖
}
