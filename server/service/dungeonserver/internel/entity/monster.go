/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"math/rand"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/dungeonserver/internel/entitymgr"
	"postapocgame/server/service/dungeonserver/internel/entitysystem"
	"postapocgame/server/service/dungeonserver/internel/iface"
	"time"
)

const fallbackMonsterSkillID = entitysystem.DefaultMonsterSkillID

// MonsterEntity 怪物实体
type MonsterEntity struct {
	*BaseEntity
	monsterCfg *jsonconf.MonsterConfig
	name       string
	aiSys      *entitysystem.AISys
}

// NewMonsterEntity 创建怪物实体
func NewMonsterEntity(cfg *jsonconf.MonsterConfig) *MonsterEntity {
	entity := &MonsterEntity{
		BaseEntity: NewBaseEntity(uint64(cfg.MonsterId), uint32(protocol.EntityType_EtMonster)),
		monsterCfg: cfg,
		name:       cfg.Name,
	}

	entity.initAttributes(cfg)
	entity.initSkills(cfg.SkillIds)
	entity.aiSys = entitysystem.NewAISys(entity, cfg)

	return entity
}

func (m *MonsterEntity) GetMonsterId() uint32 {
	if m.monsterCfg == nil {
		return 0
	}
	return m.monsterCfg.MonsterId
}

func (m *MonsterEntity) GetName() string {
	return m.name
}

func (m *MonsterEntity) GetLevel() uint32 {
	if m.monsterCfg == nil {
		return 0
	}
	return m.monsterCfg.Level
}

func (m *MonsterEntity) OnDie(killer iface.IEntity) {
	m.BaseEntity.OnDie(killer)
	if m.aiSys != nil {
		m.aiSys.NotifyDie()
	}

	// 生成掉落物
	m.generateDrops(killer)
}

func (m *MonsterEntity) OnEnterScene() {
	if m.aiSys != nil {
		pos := m.GetPosition()
		m.aiSys.SetHomePosition(pos)
	}
}

func (m *MonsterEntity) OnLeaveScene() {
	// 预留：可在此重置AI状态
}

func (m *MonsterEntity) RunOne(now time.Time) {
	m.BaseEntity.RunOne(now)
	if m.aiSys != nil {
		m.aiSys.RunOne(now)
	}
}

func (m *MonsterEntity) initAttributes(cfg *jsonconf.MonsterConfig) {
	attr := m.GetAttrSys()

	// 先应用等级属性表（高等级覆盖低等级）
	levelAttrs := jsonconf.GetConfigManager().GetLevelAttrs(cfg.Level)
	for attrType, attrValue := range levelAttrs {
		attr.SetAttrValue(attrdef.AttrType(attrType), attrdef.AttrValue(attrValue))
	}

	// 设置等级
	attr.SetAttrValue(attrdef.AttrLevel, attrdef.AttrValue(cfg.Level))

	// 怪物配置中的属性值会覆盖等级属性（怪物特定属性优先）
	if cfg.HP > 0 {
		attr.SetAttrValue(attrdef.AttrMaxHP, attrdef.AttrValue(cfg.HP))
		attr.SetAttrValue(attrdef.AttrHP, attrdef.AttrValue(cfg.HP))
	}
	if cfg.MP > 0 {
		attr.SetAttrValue(attrdef.AttrMaxMP, attrdef.AttrValue(cfg.MP))
		attr.SetAttrValue(attrdef.AttrMP, attrdef.AttrValue(cfg.MP))
	}
	if cfg.Attack > 0 {
		attr.SetAttrValue(attrdef.AttrAttack, attrdef.AttrValue(cfg.Attack))
	}
	if cfg.Defense > 0 {
		attr.SetAttrValue(attrdef.AttrDefense, attrdef.AttrValue(cfg.Defense))
	}

	// 移动速度：如果配置中有则使用配置值，否则使用等级属性或默认值
	moveSpeed := cfg.Speed
	if moveSpeed == 0 {
		// 如果等级属性中有移动速度，使用等级属性的值
		moveSpeedValue := attr.GetAttrValue(attrdef.AttrMoveSpeed)
		if moveSpeedValue > 0 {
			moveSpeed = uint32(moveSpeedValue)
		} else {
			moveSpeed = 20 // 默认值
		}
	}
	attr.SetAttrValue(attrdef.AttrMoveSpeed, attrdef.AttrValue(moveSpeed))

	// 如果配置中没有设置HP/MP的当前值，则设置为最大值
	if attr.GetAttrValue(attrdef.AttrHP) == 0 {
		maxHP := attr.GetAttrValue(attrdef.AttrMaxHP)
		if maxHP > 0 {
			attr.SetAttrValue(attrdef.AttrHP, maxHP)
		}
	}
	if attr.GetAttrValue(attrdef.AttrMP) == 0 {
		maxMP := attr.GetAttrValue(attrdef.AttrMaxMP)
		if maxMP > 0 {
			attr.SetAttrValue(attrdef.AttrMP, maxMP)
		}
	}
}

func (m *MonsterEntity) initSkills(skillIds []uint32) {
	fightSys, ok := m.GetFightSys().(*entitysystem.FightSys)
	if !ok {
		return
	}
	if len(skillIds) == 0 {
		if err := fightSys.LearnSkill(fallbackMonsterSkillID, 1); err != nil {
			log.Warnf("monster learn default skill failed: %v", err)
		}
		return
	}
	for _, skillID := range skillIds {
		if err := fightSys.LearnSkill(skillID, 1); err != nil {
			log.Warnf("monster learn skill %d failed: %v", skillID, err)
		}
	}
}

// generateDrops 生成掉落物
func (m *MonsterEntity) generateDrops(killer iface.IEntity) {
	if m.monsterCfg == nil || len(m.monsterCfg.DropItems) == 0 {
		return
	}

	// 获取怪物位置
	pos := m.GetPosition()
	x, y := pos.X, pos.Y

	// 获取归属者信息
	var ownerHdl uint64
	var ownerRoleId uint64
	if killer != nil {
		ownerHdl = killer.GetHdl()
		// 如果是角色实体，获取角色ID
		if roleEntity, ok := killer.(*RoleEntity); ok {
			ownerRoleId = roleEntity.GetRoleId()
		}
	}

	// 获取场景
	scene, ok := entitymgr.GetEntityMgr().GetSceneByHandle(m.GetHdl())
	if !ok || scene == nil {
		log.Warnf("Monster %d: scene not found when generating drops", m.GetHdl())
		return
	}

	// 遍历掉落配置，根据概率生成掉落物
	for _, drop := range m.monsterCfg.DropItems {
		// 根据掉落概率决定是否掉落
		if rand.Float32() > drop.DropRate {
			continue
		}

		// 计算掉落数量
		count := drop.MinCount
		if drop.MaxCount > drop.MinCount {
			count = drop.MinCount + uint32(rand.Intn(int(drop.MaxCount-drop.MinCount+1)))
		}

		// 创建掉落物实体（传递存在时间配置）
		dropEntity := NewDropItemEntity(drop.ItemId, count, x, y, ownerHdl, ownerRoleId, drop.LifetimeSeconds)

		// 添加到场景（会自动注册到EntityMgr）
		if err := scene.AddEntity(dropEntity); err != nil {
			log.Errorf("Failed to add drop item to scene: %v", err)
			continue
		}

		log.Infof("Monster %d dropped item %d x%d at (%d, %d)", m.GetHdl(), drop.ItemId, count, x, y)
	}
}
