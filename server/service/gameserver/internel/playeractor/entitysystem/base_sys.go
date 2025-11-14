package entitysystem

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BaseSystem 系统基类
type BaseSystem struct {
	sysID  uint32
	opened bool
}

func NewBaseSystem(sysID uint32) *BaseSystem {
	return &BaseSystem{
		sysID:  sysID,
		opened: true, // 默认开启
	}
}

func (bs *BaseSystem) GetId() uint32 {
	return bs.sysID
}

func (bs *BaseSystem) OnInit(context.Context) {}

func (bs *BaseSystem) OnOpen(context.Context) {
}

func (bs *BaseSystem) OnRoleLogin(context.Context) {

}

func (bs *BaseSystem) OnRoleReconnect(context.Context) {
}

func (bs *BaseSystem) OnRoleLogout(context.Context) {
}

func (bs *BaseSystem) OnRoleClose(context.Context) {
}

func (bs *BaseSystem) OnNewHour(context.Context) {}
func (bs *BaseSystem) OnNewDay(context.Context)  {}
func (bs *BaseSystem) OnNewWeek(context.Context) {}
func (bs *BaseSystem) OnNewMonth(context.Context) {
}
func (bs *BaseSystem) OnNewYear(context.Context) {}

func (bs *BaseSystem) IsOpened() bool {
	return bs.opened
}

func (bs *BaseSystem) SetOpened(opened bool) {
	bs.opened = opened
}

func GetIPlayerRoleByContext(ctx context.Context) (iface.IPlayerRole, error) {
	value := ctx.Value(gshare.ContextKeyRole)
	if value == nil {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "not found %s value", gshare.ContextKeyRole)
	}
	iPlayerRole, ok := value.(iface.IPlayerRole)
	if !ok {
		return nil, customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "not convert to IPlayerRole")
	}
	return iPlayerRole, nil
}
