package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LevelSys 等级系统
type LevelSys struct {
	*BaseSystem
}

// NewLevelSys 创建等级系统
func NewLevelSys() *LevelSys {
	sys := &LevelSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysLevel)),
	}
	return sys
}

func GetLevelSys(ctx context.Context) *LevelSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysLevel))
	if system == nil {
		log.Errorf("not found system [%v] error:%v", protocol.SystemId_SysLevel, err)
		return nil
	}
	sys := system.(*LevelSys)
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error:%v", protocol.SystemId_SysLevel, err)
		return nil
	}
	return sys
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSys()
	})
	gevent.SubscribePlayerEventH(gevent.OnPlayerLevelUp, func(ctx context.Context, ev *event.Event) {})
	gevent.SubscribePlayerEventH(gevent.OnPlayerExpChange, func(ctx context.Context, ev *event.Event) {})
}
