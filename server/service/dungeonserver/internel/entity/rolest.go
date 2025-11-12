/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

// RoleEntity 角色实体
type RoleEntity struct {
	*BaseEntity
	sessionId     string
	roleInfo      *protocol.PlayerSimpleData
	learnedSkills map[uint32]bool // 已学习的技能
}

// NewRoleEntity 创建角色实体
func NewRoleEntity(sessionId string, roleInfo *protocol.PlayerSimpleData) *RoleEntity {
	entity := &RoleEntity{
		BaseEntity:    NewBaseEntity(roleInfo.RoleId, uint32(protocol.EntityType_EtRole)),
		sessionId:     sessionId,
		roleInfo:      roleInfo,
		learnedSkills: make(map[uint32]bool),
	}

	// 根据角色等级设置属性
	entity.GetAttrSys().SetAttrValue(attrdef.AttrMaxHP, attrdef.AttrValue(1000+roleInfo.Level*100))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrHP, attrdef.AttrValue(1000+roleInfo.Level*100))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrAttack, attrdef.AttrValue(100+roleInfo.Level*10))
	entity.GetAttrSys().SetAttrValue(attrdef.AttrDefense, attrdef.AttrValue(50+roleInfo.Level*5))

	// 初始化技能
	entity.initSkills()

	return entity
}

func (r *RoleEntity) GetSessionId() string {
	return r.sessionId
}

func (r *RoleEntity) HasLearnedSkill(skillId uint32) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.learnedSkills[skillId]
}

func (r *RoleEntity) LearnSkill(skillId uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.learnedSkills[skillId] = true
}

func (r *RoleEntity) initSkills() {
	// TODO: 从配置或数据库读取已学技能
	// 临时给予一些初始技能
	r.learnedSkills[1001] = true // 火球术
	r.learnedSkills[1002] = true // 冰箭
}

func (r *RoleEntity) SendMessage(protoId uint16, data []byte) error {
	return gameserverlink.SendToClient(r.sessionId, protoId, data)
}
