package entitysystem

import (
	"postapocgame/server/internal/custom_id"
	protocol2 "postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/iface"
)

// LevelSys 等级系统
type LevelSys struct {
	*BaseSystem
	level uint32
	exp   uint64
}

// NewLevelSys 创建等级系统
func NewLevelSys(role iface.IPlayerRole) *LevelSys {
	sys := &LevelSys{
		BaseSystem: NewBaseSystem(custom_id.SysLevel, role),
		level:      role.GetPlayerRoleInfo().Level,
		exp:        0,
	}
	return sys
}

// OnRoleLogin 角色登录时下发等级数据
func (s *LevelSys) OnRoleLogin() {
	return
}

// SendData 下发等级数据
func (s *LevelSys) SendData() error {
	data := &protocol2.LevelData{
		Level: s.level,
		Exp:   s.exp,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(1, 6, jsonData)
}

// AddExp 增加经验
func (s *LevelSys) AddExp(exp uint64) {
	oldLevel := s.level
	s.exp += exp

	// 简化升级逻辑：每1000经验升1级
	for s.exp >= 1000 {
		s.exp -= 1000
		s.level++
	}

	// 如果升级了，发布升级事件
	if s.level > oldLevel {

	}

	s.SendData()
}

// GetLevel 获取当前等级
func (s *LevelSys) GetLevel() uint32 {
	return s.level
}

// GetExp 获取当前经验
func (s *LevelSys) GetExp() uint64 {
	return s.exp
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(custom_id.SysLevel, func(role iface.IPlayerRole) iface.ISystem {
		return NewLevelSys(role)
	})
}
