/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package entitysystem

import (
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/service/gameserver/internel/iface"
)

// OfflineMsgSys 离线消息系统
type OfflineMsgSys struct {
	*BaseSystem
	messages []string
}

func NewOfflineMsgSys(role iface.IPlayerRole) *OfflineMsgSys {
	return &OfflineMsgSys{
		BaseSystem: NewBaseSystem(custom_id.SysOfflineMsg, role),
		messages:   make([]string, 0),
	}
}

func (s *OfflineMsgSys) OnRoleLogin() {
	return
}

func (s *OfflineMsgSys) ProcessOfflineMessages() error {
	// TODO: 从数据库读取离线消息
	// 测试阶段：模拟一条充值离线消息
	if len(s.messages) > 0 {
		// 处理离线消息
		for _, msg := range s.messages {
			// TODO: 根据消息类型处理
			_ = msg
		}
		s.messages = s.messages[:0]
	}

	return nil
}

func (s *OfflineMsgSys) AddMessage(msg string) {
	s.messages = append(s.messages, msg)
}

func init() {
	RegisterSystemFactory(custom_id.SysOfflineMsg, func(role iface.IPlayerRole) iface.ISystem {
		return NewOfflineMsgSys(role)
	})
}
