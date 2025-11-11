package entitysystem

import (
	"postapocgame/server/internal/custom_id"
	protocol2 "postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/iface"
)

// VipSys VIP系统
type VipSys struct {
	*BaseSystem
	vipLevel uint32
	vipExp   uint64
}

func NewVipSys(role iface.IPlayerRole) *VipSys {
	return &VipSys{
		BaseSystem: NewBaseSystem(custom_id.SysVip, role),
		vipLevel:   0,
		vipExp:     0,
	}
}

func (s *VipSys) OnRoleLogin() {
	return
}

func (s *VipSys) SendData() error {
	data := &protocol2.VipData{
		VipLevel: s.vipLevel,
		VipExp:   s.vipExp,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(1, 8, jsonData)
}

func (s *VipSys) AddVipExp(exp uint64) {
	s.vipExp += exp
	// TODO: 检查VIP升级
	s.SendData()
}

func init() {
	RegisterSystemFactory(custom_id.SysVip, func(role iface.IPlayerRole) iface.ISystem {
		return NewVipSys(role)
	})
}
