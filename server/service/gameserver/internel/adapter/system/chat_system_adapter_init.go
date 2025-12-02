package system

import (
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/core/iface"
)

func init() {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysChat), func() iface.ISystem {
		return NewChatSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
