/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package iface

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
)

type IPlayerRole interface {
	IPlayerEvent

	WithContext(parentCtx context.Context) context.Context

	Close() error
	OnReconnect(newSessionId string) error
	OnDisconnect()
	OnLogin() error
	OnLogout() error

	SendMessage(protoId uint16, data []byte) error
	SendJsonMessage(protoId uint16, v interface{}) error

	GetPlayerRoleId() uint64
	GetReconnectKey() string
	GetSessionId() string
	GetBinaryData() *protocol.PlayerRoleBinaryData

	GetSysMgr() ISystemMgr
	GetSystem(sysId uint32) ISystem
	GetSysStatus(sysId uint32) bool
	GetSysStatusData() map[uint32]uint32
	SetSysStatus(sysId uint32, isOpen bool)
}

type IPlayerEvent interface {
	Publish(typ event.Type, args ...interface{})
}
