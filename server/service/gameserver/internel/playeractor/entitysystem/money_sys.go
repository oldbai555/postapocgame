package entitysystem

import (
	"context"
	"fmt"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
)

var (
	ErrMoneyNotEnough   = fmt.Errorf("money not enough")
	ErrUnknownMoneyType = fmt.Errorf("unknown money type")
)

// MoneySys 货币系统
type MoneySys struct {
	*BaseSystem
	gold    uint64
	diamond uint64
	coin    uint64
}

// NewMoneySys 创建货币系统
func NewMoneySys(role iface.IPlayerRole) *MoneySys {
	sys := &MoneySys{
		BaseSystem: NewBaseSystem(custom_id.SysMoney, role),
		gold:       1000,
		diamond:    100,
		coin:       10000,
	}
	return sys
}

// OnRoleLogin 角色登录时下发货币数据
func (s *MoneySys) OnRoleLogin() {
	return
}

// SendData 下发货币数据
func (s *MoneySys) SendData() error {
	data := &protocol.MoneyData{
		Gold:    s.gold,
		Diamond: s.diamond,
		Coin:    s.coin,
	}
	jsonData, _ := tool.JsonMarshal(data)
	return s.role.SendMessage(protocol.S2C_MoneyData, jsonData)
}

// AddMoney 增加货币
func (s *MoneySys) AddMoney(itemID uint32, count uint32) error {
	// 简化：1=金币 2=钻石 3=铜币
	switch itemID {
	case 1:
		s.gold += uint64(count)
		s.role.Publish(gevent.OnGoldChange, s.gold, int64(count))
	case 2:
		s.diamond += uint64(count)
		s.role.Publish(gevent.OnDiamondChange, s.diamond, int64(count))
	case 3:
		s.coin += uint64(count)
		s.role.Publish(gevent.OnCoinChange, s.coin, int64(count))
	default:
		return ErrUnknownMoneyType
	}

	// 发布通用货币变化事件
	s.role.Publish(gevent.OnMoneyChange, itemID, count)

	return s.SendData()
}

// ConsumeMoney 消耗货币
func (s *MoneySys) ConsumeMoney(itemID uint32, count uint32) error {
	// 先检查是否足够
	if !s.HasEnough(itemID, count) {
		return ErrMoneyNotEnough
	}

	switch itemID {
	case 1:
		s.gold -= uint64(count)
		s.role.Publish(gevent.OnGoldChange, s.gold, -int64(count))
	case 2:
		s.diamond -= uint64(count)
		s.role.Publish(gevent.OnDiamondChange, s.diamond, -int64(count))
	case 3:
		s.coin -= uint64(count)
		s.role.Publish(gevent.OnCoinChange, s.coin, -int64(count))
	default:
		return ErrUnknownMoneyType
	}

	// 发布通用货币变化事件
	s.role.Publish(gevent.OnMoneyChange, itemID, -int64(count))

	return s.SendData()
}

// HasEnough 检查是否足够
func (s *MoneySys) HasEnough(itemID uint32, count uint32) bool {
	switch itemID {
	case 1:
		return s.gold >= uint64(count)
	case 2:
		return s.diamond >= uint64(count)
	case 3:
		return s.coin >= uint64(count)
	}
	return false
}

// GetGold 获取金币
func (s *MoneySys) GetGold() uint64 {
	return s.gold
}

// GetDiamond 获取钻石
func (s *MoneySys) GetDiamond() uint64 {
	return s.diamond
}

// GetCoin 获取铜币
func (s *MoneySys) GetCoin() uint64 {
	return s.coin
}

// 注册系统工厂
func init() {
	RegisterSystemFactory(custom_id.SysMoney, func(role iface.IPlayerRole) iface.ISystem {
		return NewMoneySys(role)
	})

	// 注册玩家级别的事件处理器
	gevent.SubscribePlayerEvent(gevent.OnMoneyChange, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			itemID, _ := ev.Data[0].(uint32)
			change, _ := ev.Data[1].(int64)
			log.Infof("[MoneySys Event] Money changed: itemID=%d, change=%d (source: %s)", itemID, change, ev.Source)
		}
	})

	gevent.SubscribePlayerEvent(gevent.OnGoldChange, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			newGold, _ := ev.Data[0].(uint64)
			change, _ := ev.Data[1].(int64)
			log.Debugf("[MoneySys Event] Gold changed: new=%d, change=%d (source: %s)", newGold, change, ev.Source)
		}
	})

	gevent.SubscribePlayerEvent(gevent.OnDiamondChange, func(ctx context.Context, ev *event.Event) {
		if len(ev.Data) >= 2 {
			newDiamond, _ := ev.Data[0].(uint64)
			change, _ := ev.Data[1].(int64)
			log.Debugf("[MoneySys Event] Diamond changed: new=%d, change=%d (source: %s)", newDiamond, change, ev.Source)
		}
	})
}
