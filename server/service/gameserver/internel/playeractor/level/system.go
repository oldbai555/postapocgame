package level

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/deps"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/playeractor/sysbase"
)

var _ iface.ISystem = (*SystemAdapter)(nil)

type SystemAdapter struct {
	*sysbase.BaseSystem
	rt *deps.Runtime
}

// NewLevelSystemAdapter 创建等级系统适配器
func NewLevelSystemAdapter(rt *deps.Runtime) *SystemAdapter {
	return &SystemAdapter{
		BaseSystem: sysbase.NewBaseSystem(uint32(protocol.SystemId_SysLevel)),
		rt:         rt,
	}
}

// OnInit 系统初始化
func (a *SystemAdapter) OnInit(ctx context.Context) {

}

// AddExp 添加经验值（对外接口，供其他系统调用）
func (a *SystemAdapter) AddExp(ctx context.Context, exp uint64) error {
	return nil
}

// GetLevel 获取当前等级
func (a *SystemAdapter) GetLevel(ctx context.Context) (uint32, error) {
	return 0, nil
}

// GetExp 获取当前经验
func (a *SystemAdapter) GetExp(ctx context.Context) (int64, error) {
	return 0, nil
}

// GetLevelSys 获取等级系统
func GetLevelSys(ctx context.Context) *SystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysLevel))
	if system == nil {
		log.Errorf("not found system [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	sys, ok := system.(*SystemAdapter)
	if !ok {
		log.Errorf("invalid system type for [%v]", protocol.SystemId_SysLevel)
		return nil
	}
	if sys == nil || !sys.IsOpened() {
		log.Errorf("get player role system [%v] error", protocol.SystemId_SysLevel)
		return nil
	}
	return sys
}

// RegisterSystemFactory 注册等级系统工厂（由 register.RegisterAll 调用）
func RegisterSystemFactory(rt *deps.Runtime) {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysLevel), func() iface.ISystem {
		return NewLevelSystemAdapter(rt)
	})
}
