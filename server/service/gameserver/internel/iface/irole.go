/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
)

type IPlayerRole interface {
	IPlayerEvent

	Close() error
	OnReconnect(newSessionId string) error
	OnDisconnect()
	OnLogin() error
	OnLogout() error

	SendMessage(protoId uint16, data []byte) error
	SendMessageHL(protoIdH uint16, protoIdL uint16, data []byte) error

	GetPlayerRoleId() uint64
	GetReconnectKey() string
	GetSessionId() string
	GetPlayerRoleData() *protocol.PlayerRoleData
	GetSystem(sysId uint32) ISystem
}

type IPlayerEvent interface {
	Publish(typ event.Type, args ...interface{})
}
