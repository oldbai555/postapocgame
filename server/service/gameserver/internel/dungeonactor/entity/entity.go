package entity

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/dungeonactor/entitysystem"
	"postapocgame/server/service/gameserver/internel/dungeonactor/iface"
	"time"
)

// BaseEntity 实体基类
type BaseEntity struct {
	hdl        uint64 // 全局唯一句柄
	Id         uint64 // 实体Id(玩家Id/怪物Id)
	entityType uint32
	position   argsdef.Position
	sceneId    uint32
	fuBenId    uint32

	// 使用新的属性系统替代旧的attr
	attrSys  *entitysystem.AttrSys
	aoiSys   *entitysystem.AOISys
	fightSys *entitysystem.FightSys
	moveSys  *entitysystem.MoveSys

	// 状态标记
	stateFlags uint64

	name string
}

const (
	stateFlagDead         = uint64(1) << uint(protocol.EntityStateFlag_EntityStateFlagDead)
	stateFlagInvincible   = uint64(1) << uint(protocol.EntityStateFlag_EntityStateFlagInvincible)
	stateFlagCannotAttack = uint64(1) << uint(protocol.EntityStateFlag_EntityStateFlagCannotAttack)
	stateFlagCannotMove   = uint64(1) << uint(protocol.EntityStateFlag_EntityStateFlagCannotMove)
)

// NewBaseEntity 创建基础实体
func NewBaseEntity(Id uint64, entityType uint32) *BaseEntity {
	entity := &BaseEntity{
		hdl:        entitymgr.CreateEntityHandle(entityType),
		Id:         Id,
		entityType: entityType,
		position:   argsdef.Position{X: 0, Y: 0},
	}

	entity.fightSys = entitysystem.NewFightSys()
	entity.fightSys.SetEntity(entity) // 设置实体引用
	entity.attrSys = entitysystem.NewAttrSys(entity)
	entity.aoiSys = entitysystem.NewAOISys(entity)
	entity.moveSys = entitysystem.NewMoveSys(entity)

	return entity
}

func (e *BaseEntity) GetHdl() uint64 {
	return e.hdl
}

func (e *BaseEntity) GetId() uint64 {
	return e.Id
}

func (e *BaseEntity) GetEntityType() uint32 {
	return e.entityType
}

func (e *BaseEntity) GetLevel() uint32 {
	level := e.attrSys.GetAttrValue(attrdef.Level)
	return uint32(level)
}

func (e *BaseEntity) GetPosition() *argsdef.Position {
	return &argsdef.Position{X: e.position.X, Y: e.position.Y}
}

func (e *BaseEntity) SetPosition(x, y uint32) {
	e.position.X = x
	e.position.Y = y
}

func (e *BaseEntity) GetSceneId() uint32 {
	return e.sceneId
}

func (e *BaseEntity) SetSceneId(sceneId uint32) {
	e.sceneId = sceneId
}

func (e *BaseEntity) GetFuBenId() uint32 {
	return e.fuBenId
}

func (e *BaseEntity) SetFuBenId(fuBenId uint32) {
	e.fuBenId = fuBenId
}

// GetAttrSys 获取属性系统
func (e *BaseEntity) GetAttrSys() iface.IAttrSys {
	return e.attrSys
}

func (e *BaseEntity) GetMoveSys() iface.IMoveSys {
	return e.moveSys
}
func (e *BaseEntity) GetHP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.HP)
}

func (e *BaseEntity) SetHP(hp int64) {
	e.GetAttrSys().SetAttrValue(attrdef.HP, hp)
}
func (e *BaseEntity) GetMaxHP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.MaxHP)
}

func (e *BaseEntity) GetMP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.MP)
}

func (e *BaseEntity) SetMP(hp int64) {
	e.GetAttrSys().SetAttrValue(attrdef.MP, hp)
}

func (e *BaseEntity) GetMaxMP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.MaxMP)
}

func (e *BaseEntity) IsDead() bool {
	return tool.IsSetBit64(e.stateFlags, stateFlagDead)
}

func (e *BaseEntity) IsInvincible() bool {
	return tool.IsSetBit64(e.stateFlags, stateFlagInvincible)
}

func (e *BaseEntity) CanBeAttacked() bool {
	return tool.IsSetBit64(e.stateFlags, stateFlagDead) && tool.IsSetBit64(e.stateFlags, stateFlagInvincible)
}

func (e *BaseEntity) GetStateFlags() uint64 {
	return e.stateFlags
}

func (e *BaseEntity) AddState(state uint32) {

}

func (e *BaseEntity) RemoveState(stateId uint32) {

}
func (e *BaseEntity) HasState(stateId uint32) bool {
	return false
}

func (e *BaseEntity) RunOne(now time.Time) {
	if e.attrSys != nil {
		e.attrSys.RunOne()
	}
	if e.fightSys != nil {
		e.fightSys.RunOne(now)
	}
}

func (e *BaseEntity) GetFightSys() iface.IFightSys {
	return e.fightSys
}

func (e *BaseEntity) OnAttacked(attacker iface.IEntity, damage int64) {
	currentHP := e.GetHP()
	if damage >= currentHP {
		e.SetHP(0)
		e.OnDie(attacker)
	} else {
		e.SetHP(currentHP - damage)
	}
}

func (e *BaseEntity) OnDie(killer iface.IEntity) {
	tool.SetBit64(e.stateFlags, stateFlagDead)
}

func (e *BaseEntity) GetAOISys() iface.IAOISys {
	return e.aoiSys
}

func (e *BaseEntity) OnEnterScene() {
	// 子类重写
}

func (e *BaseEntity) OnLeaveScene() {
	// 子类重写
}

func (e *BaseEntity) OnMove(newX, newY uint32) {
	oldPos := e.GetPosition()
	e.SetPosition(newX, newY)

	// 更新AOI
	if e.aoiSys != nil {
		e.aoiSys.OnMove(oldPos, &argsdef.Position{X: newX, Y: newY})
	}
}

func (e *BaseEntity) SendMessage(protoId uint16, data []byte) error {
	return nil
}

func (e *BaseEntity) SendProtoMessage(protoId uint16, v proto.Message) error {
	return nil
}

func (e *BaseEntity) Close() error {
	// 清理其他资源
	return nil
}

func (e *BaseEntity) GetName() string {
	if e.name == "" {
		return fmt.Sprintf("Entity[%d][%d]", e.GetEntityType(), e.GetId())
	}
	return e.name
}

func (e *BaseEntity) SetName(name string) {
	e.name = name
}

func (e *BaseEntity) buildAttrMap() map[uint32]int64 {
	attrSys := e.GetAttrSys()
	attrTypes := []uint32{
		attrdef.HP,
		attrdef.MaxHP,
		attrdef.MP,
		attrdef.MaxMP,
	}
	attrs := make(map[uint32]int64, len(attrTypes))
	for _, attrType := range attrTypes {
		value := attrSys.GetAttrValue(attrType)
		attrs[attrType] = value
	}
	return attrs
}

func (e *BaseEntity) BuildProtoEntitySt() *protocol.EntitySt {
	pos := e.GetPosition()
	if pos == nil {
		pos = &argsdef.Position{}
	}
	return &protocol.EntitySt{
		Hdl:        e.GetHdl(),
		Id:         e.GetId(),
		Et:         e.GetEntityType(),
		PosX:       pos.X,
		PosY:       pos.Y,
		SceneId:    e.GetSceneId(),
		FbId:       e.GetFuBenId(),
		Level:      e.GetLevel(),
		ShowName:   e.GetName(),
		Attrs:      e.buildAttrMap(),
		StateFlags: e.GetStateFlags(),
	}
}
