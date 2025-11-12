/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package iface

import (
	"postapocgame/server/internal/argsdef"
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

	SendMessage(protoId uint16, data []byte) error
	SendJsonMessage(protoId uint16, v interface{}) error

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

	// 战斗相关
	OnAttacked(attacker IEntity, damage int64)
	OnDie(killer IEntity)
	IsInvincible() bool
	CanBeAttacked() bool

	// AOI相关

	OnEnterScene()
	OnLeaveScene()
	OnMove(newX, newY uint32)
}
