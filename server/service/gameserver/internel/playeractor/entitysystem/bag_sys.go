package entitysystem

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

// BagSys 背包系统
type BagSys struct {
	*BaseSystem
}

// NewBagSys 创建背包系统
func NewBagSys() *BagSys {
	sys := &BagSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysBag)),
	}
	return sys
}

// OnRoleLogin 角色登录时下发背包数据
func (s *BagSys) OnRoleLogin() {
	return
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysBag), func() iface.ISystem {
		return NewBagSys()
	})

	// 注册玩家级别的事件处理器
	gevent.SubscribePlayerEvent(gevent.OnItemAdd, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			itemID, _ := ev.Data[0].(uint32)
			count, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Item added: itemID=%d, count=%d (source: %s)", itemID, count, ev.Source)

			// 这里可以处理道具添加后的逻辑，比如：
			// 1. 检查是否完成任务目标
			// 2. 触发成就系统
			// 3. 发送道具获得通知
		}
	})

	gevent.SubscribePlayerEvent(gevent.OnItemRemove, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			itemID, _ := ev.Data[0].(uint32)
			count, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Item removed: itemID=%d, count=%d (source: %s)", itemID, count, ev.Source)
		}
	})

	gevent.SubscribePlayerEventL(gevent.OnBagExpand, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			newCapacity, _ := ev.Data[0].(uint32)
			added, _ := ev.Data[1].(uint32)
			log.Infof("[BagSys Event] Bag expanded: newCapacity=%d, added=%d (source: %s)", newCapacity, added, ev.Source)
		}
	})
}
