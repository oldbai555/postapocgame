package entity

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/etsystem"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
)

// BaseEntity 实体基类
type BaseEntity struct {
	mu         sync.RWMutex
	hdl        uint64 // 全局唯一句柄
	Id         uint64 // 实体Id(玩家Id/怪物Id)
	entityType uint32
	position   argsdef.Position
	sceneId    uint32
	fuBenId    uint32

	// 使用新的属性系统替代旧的attr
	AttrSys  *etsystem.AttrSys
	aoi      *etsystem.AOI
	FightSys *etsystem.FightSys
	BuffSys  *etsystem.BuffSys

	// 状态标记
	isDead       bool
	isInvincible bool // 无敌状态
	cannotAttack bool // 无法攻击状态
	cannotMove   bool // 无法移动状态
}

// NewBaseEntity 创建基础实体
func NewBaseEntity(Id uint64, entityType uint32) *BaseEntity {
	entity := &BaseEntity{
		hdl:        entitymgr.CreateEntityHandle(entityType),
		Id:         Id,
		entityType: entityType,
		position:   argsdef.Position{X: 0, Y: 0},

		isDead:       false,
		isInvincible: false,
		cannotAttack: false,
		cannotMove:   false,
	}

	entity.FightSys = etsystem.NewFightSys()
	entity.BuffSys = etsystem.NewBuffSystem()
	entity.AttrSys = etsystem.NewAttrSys()

	// 创建AOI
	entity.aoi = etsystem.NewAOI(entity)

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
	level := e.AttrSys.GetAttrValue(attrdef.AttrLevel)
	return uint32(level)
}

func (e *BaseEntity) GetPosition() *argsdef.Position {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return &argsdef.Position{X: e.position.X, Y: e.position.Y}
}

func (e *BaseEntity) SetPosition(x, y uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.position.X = x
	e.position.Y = y
}

func (e *BaseEntity) GetSceneId() uint32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sceneId
}

func (e *BaseEntity) SetSceneId(sceneId uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sceneId = sceneId
}

func (e *BaseEntity) GetFuBenId() uint32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.fuBenId
}

func (e *BaseEntity) SetFuBenId(fuBenId uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.fuBenId = fuBenId
}

// GetAttrSys 获取属性系统
func (e *BaseEntity) GetAttrSys() iface.IAttrSys {
	return e.AttrSys
}
func (e *BaseEntity) GetHP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.AttrHP)
}

func (e *BaseEntity) SetHP(hp int64) {
	e.GetAttrSys().SetAttrValue(attrdef.AttrHP, hp)
}
func (e *BaseEntity) GetMaxHP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.AttrMaxHP)
}

func (e *BaseEntity) GetMP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.AttrMP)
}

func (e *BaseEntity) SetMP(hp int64) {
	e.GetAttrSys().SetAttrValue(attrdef.AttrMP, hp)
}

func (e *BaseEntity) GetMaxMP() int64 {
	return e.GetAttrSys().GetAttrValue(attrdef.AttrMaxMP)
}

func (e *BaseEntity) IsDead() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isDead
}

func (e *BaseEntity) IsInvincible() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isInvincible
}

func (e *BaseEntity) SetInvincible(invincible bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.isInvincible = invincible
}

func (e *BaseEntity) CannotAttack() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cannotAttack
}

func (e *BaseEntity) SetCannotAttack(cannot bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cannotAttack = cannot
}

func (e *BaseEntity) CanBeAttacked() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return !e.isDead && !e.isInvincible
}

func (e *BaseEntity) GetFightSys() iface.IFightSys {
	return e.FightSys
}

func (e *BaseEntity) GetBuffSys() iface.IBuffSys {
	return e.BuffSys
}

func (e *BaseEntity) OnAttacked(attacker iface.IEntity, damage int64) {
	currentHP := e.GetHP()
	if damage >= currentHP {
		e.SetHP(0)
		e.OnDie(attacker)
	} else {
		e.SetHP(currentHP - damage)
	}

	// TODO: 广播血量变化
}

func (e *BaseEntity) OnDie(killer iface.IEntity) {
	e.mu.Lock()
	e.isDead = true
	e.mu.Unlock()

	// TODO: 处理死亡逻辑
	// TODO: 广播死亡消息
}

func (e *BaseEntity) GetAOISys() iface.IAOISys {
	return e.aoi
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
	if e.aoi != nil {
		e.aoi.OnMove(oldPos, &argsdef.Position{X: newX, Y: newY})
	}
}

func (e *BaseEntity) SendMessage(protoId uint16, data []byte) error {
	return nil
}
