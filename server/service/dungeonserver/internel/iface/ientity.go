/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/argsdef"
	"time"
)

// IEntity 实体接口
type IEntity interface {
	GetHdl() uint64
	GetId() uint64
	GetEntityType() uint32
	GetPosition() *argsdef.Position
	SetPosition(x, y uint32)
	GetSceneId() uint32
	SetSceneId(sceneId uint32)
	GetFuBenId() uint32
	SetFuBenId(fuBenId uint32)
	GetLevel() uint32
	GetStateFlags() uint64

	SendMessage(protoId uint16, data []byte) error
	SendProtoMessage(protoId uint16, v proto.Message) error

	// 属性相关
	GetHP() int64
	SetHP(hp int64)
	GetMaxHP() int64
	GetMP() int64
	SetMP(mp int64)
	GetMaxMP() int64
	IsDead() bool

	// 系统相关
	GetFightSys() IFightSys
	GetBuffSys() IBuffSys
	GetAOISys() IAOISys
	GetAttrSys() IAttrSys
	GetMoveSys() IMoveSys

	// 战斗相关
	OnAttacked(attacker IEntity, damage int64)
	OnDie(killer IEntity)
	IsInvincible() bool
	CanBeAttacked() bool
	ApplyExtraState(stateId uint32, duration time.Duration)
	RemoveExtraState(stateId uint32)

	// AOI相关

	OnEnterScene()
	OnLeaveScene()
	OnMove(newX, newY uint32)
	RunOne(now time.Time)
}
