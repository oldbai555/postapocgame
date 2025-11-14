package entity

import (
	"postapocgame/server/internal/argsdef"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/entityhelper"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/entitysystem"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"sync"
	"time"
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
	attrSys      *entitysystem.AttrSys
	aoiSys       *entitysystem.AOISys
	fightSys     *entitysystem.FightSys
	buffSys      *entitysystem.BuffSys
	stateMachine *entitysystem.StateMachine // 状态机
	moveSys      *entitysystem.MoveSys

	// 状态标记
	stateFlags uint64
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
	entity.buffSys = entitysystem.NewBuffSystem(entity)
	entity.attrSys = entitysystem.NewAttrSys()
	entity.aoiSys = entitysystem.NewAOISys(entity)
	entity.stateMachine = entitysystem.NewStateMachine(entity) // 创建状态机
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
	level := e.attrSys.GetAttrValue(attrdef.AttrLevel)
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
	return e.attrSys
}

func (e *BaseEntity) GetMoveSys() iface.IMoveSys {
	return e.moveSys
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
	return e.stateFlags&stateFlagDead != 0
}

func (e *BaseEntity) IsInvincible() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateFlags&stateFlagInvincible != 0
}

func (e *BaseEntity) SetInvincible(invincible bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if invincible {
		e.stateFlags |= stateFlagInvincible
	} else {
		e.stateFlags &^= stateFlagInvincible
	}
}

func (e *BaseEntity) CannotAttack() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateFlags&stateFlagCannotAttack != 0
}

func (e *BaseEntity) SetCannotAttack(cannot bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cannot {
		e.stateFlags |= stateFlagCannotAttack
	} else {
		e.stateFlags &^= stateFlagCannotAttack
	}
}

func (e *BaseEntity) CannotMove() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateFlags&stateFlagCannotMove != 0
}

func (e *BaseEntity) SetCannotMove(cannot bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if cannot {
		e.stateFlags |= stateFlagCannotMove
	} else {
		e.stateFlags &^= stateFlagCannotMove
	}
}

func (e *BaseEntity) CanBeAttacked() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateFlags&stateFlagDead == 0 && e.stateFlags&stateFlagInvincible == 0
}

func (e *BaseEntity) GetStateFlags() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateFlags
}

func (e *BaseEntity) ApplyExtraState(stateId uint32, duration time.Duration) {
	if e.stateMachine != nil {
		e.stateMachine.AddExtraState(stateId, duration)
	}
}

func (e *BaseEntity) RemoveExtraState(stateId uint32) {
	if e.stateMachine != nil {
		e.stateMachine.RemoveExtraState(stateId)
	}
}

func (e *BaseEntity) RunOne(now time.Time) {
	if e.buffSys != nil {
		e.buffSys.RunOne(now)
	}
	if e.stateMachine != nil {
		e.stateMachine.Update()
	}
	if e.moveSys != nil {
		e.moveSys.RunOne(now)
	}
}

func (e *BaseEntity) GetFightSys() iface.IFightSys {
	return e.fightSys
}

func (e *BaseEntity) GetBuffSys() iface.IBuffSys {
	return e.buffSys
}

func (e *BaseEntity) OnAttacked(attacker iface.IEntity, damage int64) {
	currentHP := e.GetHP()
	if damage >= currentHP {
		e.SetHP(0)
		e.OnDie(attacker)
	} else {
		e.SetHP(currentHP - damage)

		// 受到攻击后进入硬直状态（如果伤害足够大）
		if damage > currentHP/10 && e.stateMachine != nil {
			// 伤害超过当前血量10%时进入硬直状态，持续0.5秒
			if e.stateMachine.CanChangeState(entitysystem.StateHardHit) {
				e.stateMachine.SetState(entitysystem.StateHardHit, 500*time.Millisecond)
			}
		}
	}

	// 广播血量变化给视野内的玩家
	e.broadcastHpChange()
}

func (e *BaseEntity) OnDie(killer iface.IEntity) {
	e.mu.Lock()
	e.stateFlags |= stateFlagDead
	e.mu.Unlock()

	// 广播死亡消息给视野内的玩家
	e.broadcastDeath(killer)
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

func (e *BaseEntity) SendJsonMessage(protoId uint16, v interface{}) error {
	return nil
}

func (e *BaseEntity) Close() error {
	// 清理其他资源
	return nil
}

// broadcastHpChange 广播血量变化给视野内的玩家
func (e *BaseEntity) broadcastHpChange() {
	if e.aoiSys == nil {
		return
	}

	// 获取视野内的所有实体
	visibleEntities := e.aoiSys.GetVisibleEntities()
	if len(visibleEntities) == 0 {
		return
	}

	// 构建实体快照（包含最新的HP等属性）
	entitySt := entityhelper.BuildEntitySnapshot(e)
	if entitySt == nil {
		return
	}

	// 只发送给玩家实体
	for _, target := range visibleEntities {
		if target.GetEntityType() == uint32(protocol.EntityType_EtRole) {
			// 使用S2CEntityAppear协议更新实体信息（包含属性变化）
			_ = target.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), &protocol.S2CEntityAppearReq{
				Entity: entitySt,
			})
		}
	}
}

// broadcastDeath 广播死亡消息给视野内的玩家
func (e *BaseEntity) broadcastDeath(killer iface.IEntity) {
	if e.aoiSys == nil {
		return
	}

	// 获取视野内的所有实体
	visibleEntities := e.aoiSys.GetVisibleEntities()
	if len(visibleEntities) == 0 {
		return
	}

	// 构建实体快照（包含死亡状态）
	entitySt := entityhelper.BuildEntitySnapshot(e)
	if entitySt == nil {
		return
	}

	// 只发送给玩家实体
	for _, target := range visibleEntities {
		if target.GetEntityType() == uint32(protocol.EntityType_EtRole) {
			// 使用S2CEntityAppear协议更新实体信息（StateFlags包含死亡标记）
			_ = target.SendJsonMessage(uint16(protocol.S2CProtocol_S2CEntityAppear), &protocol.S2CEntityAppearReq{
				Entity: entitySt,
			})
		}
	}
}
