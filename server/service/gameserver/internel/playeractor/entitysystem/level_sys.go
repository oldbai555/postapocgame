package entitysystem

import (
	"context"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gevent"
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
	data := &protocol.LevelData{
		Level: s.level,
		Exp:   s.exp,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(protocol.S2C_LevelData, jsonData)
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

	// 发布经验变化事件
	s.role.Publish(gevent.OnPlayerExpChange, s.exp)

	// 如果升级了，发布升级事件
	if s.level > oldLevel {
		s.role.Publish(gevent.OnPlayerLevelUp, oldLevel, s.level)
		log.Infof("Player %d level up: %d -> %d", s.role.GetPlayerRoleId(), oldLevel, s.level)
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

	// 注册玩家级别的事件处理器（这些会被克隆到每个玩家）
	gevent.SubscribePlayerEventH(gevent.OnPlayerLevelUp, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			oldLevel, _ := ev.Data[0].(uint32)
			newLevel, _ := ev.Data[1].(uint32)
			log.Infof("[LevelSys Event] Player level up: %d -> %d (source: %s)", oldLevel, newLevel, ev.Source)

			// 这里可以处理升级后的逻辑，比如：
			// 1. 发放升级奖励
			// 2. 解锁新功能
			// 3. 发送升级通知
		}
	})

	gevent.SubscribePlayerEventH(gevent.OnPlayerExpChange, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 1 {
			exp, _ := ev.Data[0].(uint64)
			log.Debugf("[LevelSys Event] Player exp changed: %d (source: %s)", exp, ev.Source)
		}
	})
}
