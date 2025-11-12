/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package entity

import (
	"postapocgame/server/internal/attrdef"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

// RoleEntity 角色实体
type RoleEntity struct {
	*BaseEntity
	sessionId string
	roleInfo  *protocol.PlayerSimpleData
}

// NewRoleEntity 创建角色实体
func NewRoleEntity(sessionId string, roleInfo *protocol.PlayerSimpleData) *RoleEntity {
	entity := &RoleEntity{
		BaseEntity: NewBaseEntity(roleInfo.RoleId, uint32(protocol.EntityType_EtRole)),
		sessionId:  sessionId,
		roleInfo:   roleInfo,
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

func (r *RoleEntity) initSkills() {
	// TODO: 从配置或数据库读取已学技能
	// 临时给予一些初始技能
	r.GetFightSys().LearnSkill(1001, 1)
	r.GetFightSys().LearnSkill(1002, 1)
}

func (r *RoleEntity) SendMessage(protoId uint16, data []byte) error {
	return gameserverlink.SendToClient(r.sessionId, protoId, data)
}

func (r *RoleEntity) SendJsonMessage(protoId uint16, v interface{}) error {
	bytes, err := tool.JsonMarshal(v)
	if err != nil {
		return customerr.Wrap(err)
	}
	return r.SendMessage(protoId, bytes)
}
