/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package iface

import (
	"context"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
)

type IPlayerRole interface {
	IPlayerSiDataRepository
	IPlayerEvent

	WithContext(parentCtx context.Context) context.Context

	Close() error
	OnDisconnect()
	OnLogin() error
	OnLogout() error

	SendMessage(protoId uint16, data []byte) error
	SendProtoMessage(protoId uint16, v proto.Message) error

	GetPlayerRoleId() uint64
	GetReconnectKey() string
	GetSessionId() string
	GetBinaryData() *protocol.PlayerRoleBinaryData

	GetJob() uint32                                  // 获取职业ID
	GetPlayerSimpleData() *protocol.PlayerSimpleData // 获取角色信息

	GetSysMgr() ISystemMgr
	GetSystem(sysId uint32) ISystem
	GetSysStatus(sysId uint32) bool
	GetSysStatusData() map[uint32]uint32
	SetSysStatus(sysId uint32, isOpen bool)

	CallDungeonActor(ctx context.Context, msgId uint16, data []byte) error

	SaveToDB() error
	RunOne()
	OnNewHour(ctx context.Context)
	OnNewDay(ctx context.Context)
	OnNewWeek(ctx context.Context)
	OnNewMonth(ctx context.Context)
	OnNewYear(ctx context.Context)
}

type IPlayerEvent interface {
	Publish(typ event.Type, args ...interface{})
}

type IPlayerSiDataRepository interface {
	GetLevelData() *protocol.SiLevelData
	GetSkillData() *protocol.SiSkillData
}
