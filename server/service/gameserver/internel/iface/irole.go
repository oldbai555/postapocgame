/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/custom_id"
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

	GiveAwards(awards []protocol.Item) error
	Consume(items []protocol.Item) error

	AddExp(exp uint64)

	GetPlayerRoleId() uint64
	GetReconnectKey() string
	GetSessionId() string
	GetPlayerRoleInfo() *protocol.RoleInfo
	GetSystem(sysId custom_id.SystemId) ISystem
}

type IPlayerEvent interface {
	Publish(typ event.Type, args ...interface{})
}
