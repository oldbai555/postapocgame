/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"math/rand"
	icalc "postapocgame/server/internal/attrcalc"
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/jsonconf"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitymgr"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/entitysystem"
	dungeonattrcalc "postapocgame/server/service/gameserver/internel/app/dungeonactor/entitysystem/attrcalc"
	"postapocgame/server/service/gameserver/internel/app/dungeonactor/iface"
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

	entity.SetName(cfg.Name)
	entity.initAttributes(cfg)
	entity.initSkills(cfg.SkillIds)
	entity.aiSys = entitysystem.NewAISys(entity, cfg)
	entity.GetAttrSys().SetAttrValue(attrdef.AttrLevel, attrdef.AttrValue(cfg.Level))

	// 计算并应用怪物属性（通过属性计算器）
	entity.ResetProperty()

	// 设置初始HP/MP为最大值
	if entity.GetAttrSys().GetAttrValue(attrdef.AttrHP) == 0 {
		maxHP := entity.GetAttrSys().GetAttrValue(attrdef.AttrMaxHP)
		if maxHP > 0 {
			entity.GetAttrSys().SetAttrValue(attrdef.AttrHP, maxHP)
		}
	}
	if entity.GetAttrSys().GetAttrValue(attrdef.AttrMP) == 0 {
		maxMP := entity.GetAttrSys().GetAttrValue(attrdef.AttrMaxMP)
		if maxMP > 0 {
			entity.GetAttrSys().SetAttrValue(attrdef.AttrMP, maxMP)
		}
	}

	// 标记属性系统初始化完成（允许广播属性）
	entity.GetAttrSys().SetInitFinish()

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

// ResetProperty 重置怪物属性（重新计算基础属性并触发属性汇总）
func (m *MonsterEntity) ResetProperty() {
	// 1. 重新计算怪物基础属性
	m.GetAttrSys().ResetSysAttr(uint32(protocol.SaAttrSys_MonsterBaseProperty))
	// 2. 触发属性汇总、转换、百分比加成和广播
	m.GetAttrSys().ResetProperty()
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

	// 注意：怪物基础属性现在通过属性计算器计算（monsterBaseProperty）
	// 这里只设置一些必要的初始值，具体属性由 ResetSysAttr 触发计算
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
			ownerRoleId = roleEntity.GetId()
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

// monsterBaseProperty 怪物基础属性计算
func monsterBaseProperty(owner iface.IEntity, calc *icalc.FightAttrCalc) {
	monster, ok := owner.(*MonsterEntity)
	if !ok {
		return
	}

	cfg := monster.monsterCfg
	if cfg == nil {
		return
	}

	// 从怪物配置表读取基础属性
	if cfg.HP > 0 {
		calc.AddValue(attrdef.AttrMaxHP, attrdef.AttrValue(cfg.HP))
	}
	if cfg.MP > 0 {
		calc.AddValue(attrdef.AttrMaxMP, attrdef.AttrValue(cfg.MP))
	}
	if cfg.Attack > 0 {
		calc.AddValue(attrdef.AttrAttack, attrdef.AttrValue(cfg.Attack))
	}
	if cfg.Defense > 0 {
		calc.AddValue(attrdef.AttrDefense, attrdef.AttrValue(cfg.Defense))
	}

	// 移动速度：如果配置中有则使用配置值，否则使用默认值
	moveSpeed := cfg.Speed
	if moveSpeed == 0 {
		moveSpeed = 20 // 默认值
	}
	calc.AddValue(attrdef.AttrMoveSpeed, attrdef.AttrValue(moveSpeed))
}

func init() {
	// 注册怪物基础属性计算器
	dungeonattrcalc.RegIncAttrCalcFn(uint32(protocol.SaAttrSys_MonsterBaseProperty), monsterBaseProperty)
}
