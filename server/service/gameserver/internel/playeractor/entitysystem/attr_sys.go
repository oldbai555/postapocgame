/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package entitysystem

import (
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/iface"
)

// AttrSys 属性系统
type AttrSys struct {
	*BaseSystem
	attr *protocol.AttrData
}

func NewAttrSys(role iface.IPlayerRole) *AttrSys {
	return &AttrSys{
		BaseSystem: NewBaseSystem(custom_id.SysAttr, role),
		attr: &protocol.AttrData{
			HP:      1000,
			MP:      500,
			Attack:  100,
			Defense: 50,
			Speed:   10,
		},
	}
}

func (s *AttrSys) OnRoleLogin() {
	return
}

func (s *AttrSys) SendData() error {
	jsonData, _ := tool.JsonMarshal(s.attr)
	return s.role.SendMessage(protocol.S2C_AttrData, jsonData)
}

func (s *AttrSys) AddHP(hp uint32) {
	s.attr.HP += hp
	s.SendData()
}

func (s *AttrSys) AddAttack(attack uint32) {
	s.attr.Attack += attack
	s.SendData()
}

func init() {
	RegisterSystemFactory(custom_id.SysAttr, func(role iface.IPlayerRole) iface.ISystem {
		return NewAttrSys(role)
	})
}
