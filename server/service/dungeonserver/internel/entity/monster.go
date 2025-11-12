/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/iface"
)

// MonsterEntity 怪物实体
type MonsterEntity struct {
	*BaseEntity
	monsterId  uint32
	monsterCfg interface{} // TODO: 从配置读取
	name       string
	level      uint32
}

// NewMonsterEntity 创建怪物实体
func NewMonsterEntity(hdl uint64, monsterId uint32, name string, level uint32) *MonsterEntity {
	entity := &MonsterEntity{
		BaseEntity: NewBaseEntity(uint64(monsterId), uint32(protocol.EntityType_EtMonster)),
		monsterId:  monsterId,
		name:       name,
		level:      level,
	}

	// TODO: 从配置读取怪物属性
	// 临时硬编码
	entity.GetAttrSys().SetAttrValue(attrdef.AttrMaxHP, attrdef.AttrValue(2000+level*200))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrHP, attrdef.AttrValue(2000+level*200))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrAttack, attrdef.AttrValue(80+level*8))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrDefense, attrdef.AttrValue(40+level*4))

	return entity
}

func (m *MonsterEntity) GetMonsterId() uint32 {
	return m.monsterId
}

func (m *MonsterEntity) GetName() string {
	return m.name
}

func (m *MonsterEntity) GetLevel() uint32 {
	return m.level
}

func (m *MonsterEntity) OnDie(killer iface.IEntity) {
	m.BaseEntity.OnDie(killer)

	// 怪物死亡额外处理
	// TODO: 掉落物品
	// TODO: 给击杀者经验
}
