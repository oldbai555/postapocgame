package entity

import (
	"fmt"
	"postapocgame/server/internal/custom_id"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/iface"
	entitysystem2 "postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"time"
)

// PlayerRole 玩家角色
type PlayerRole struct {
	// 基础信息
	SessionId string             `json:"sessionId"`
	RoleInfo  *protocol.RoleInfo `json:"roleInfo"`

	// 重连相关
	ReconnectKey string    `json:"reconnectKey"`
	IsOnline     bool      `json:"isOnline"`
	DisconnectAt time.Time `json:"disconnectAt"`

	// 事件总线
	eventBus *event.Bus

	// 系统管理器
	sysMgr *entitysystem2.SysMgr
}

// NewPlayerRole 创建玩家角色
func NewPlayerRole(sessionId string, roleInfo *protocol.RoleInfo) *PlayerRole {
	pr := &PlayerRole{
		SessionId:    sessionId,
		RoleInfo:     roleInfo,
		IsOnline:     true,
		eventBus:     event.NewEventBus(),
		ReconnectKey: generateReconnectKey(sessionId, roleInfo.RoleId),
	}

	// 创建系统管理器
	pr.sysMgr = entitysystem2.NewSysMgr(pr)

	return pr
}

// OnLogin 登录回调
func (pr *PlayerRole) OnLogin() error {
	log.Infof("[PlayerRole] OnLogin: RoleId=%d, SessionId=%s", pr.RoleInfo.RoleId, pr.SessionId)

	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 下发重连密钥
	if err := pr.sendReconnectKey(); err != nil {
		log.Errorf("Send reconnect key failed: %v", err)
	}

	pr.sysMgr.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleLogin()
	})

	return nil
}

// OnLogout 登出回调
func (pr *PlayerRole) OnLogout() error {
	log.Infof("[PlayerRole] OnLogout: RoleId=%d", pr.RoleInfo.RoleId)

	pr.IsOnline = false
	return nil
}

// OnReconnect 重连回调
func (pr *PlayerRole) OnReconnect(newSessionId string) error {
	log.Infof("[PlayerRole] OnReconnect: RoleId=%d, OldSession=%s, NewSession=%s",
		pr.RoleInfo.RoleId, pr.SessionId, newSessionId)

	pr.SessionId = newSessionId
	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 下发重连密钥
	if err := pr.sendReconnectKey(); err != nil {
		log.Errorf("Send reconnect key failed: %v", err)
	}

	// 调用系统管理器的重连方法
	return pr.sysMgr.OnReconnect()
}

// OnDisconnect 断线回调
func (pr *PlayerRole) OnDisconnect() {
	log.Infof("[PlayerRole] OnDisconnect: RoleId=%d", pr.RoleInfo.RoleId)

	pr.IsOnline = false
	pr.DisconnectAt = time.Now()
}

// Close 关闭回调（3分钟超时或主动登出）
func (pr *PlayerRole) Close() error {
	log.Infof("[PlayerRole] Close: RoleId=%d", pr.RoleInfo.RoleId)

	// 调用登出
	err := pr.OnLogout()
	if err != nil {
		log.Errorf("err:%v", err)
	}

	// 调用系统管理器的关闭方法
	err = pr.sysMgr.OnClose()
	if err != nil {
		log.Errorf("err:%v", err)
	}

	// 可以在这里保存数据到数据库
	// TODO: Save to database

	return nil
}

// GiveAwards 发放奖励
func (pr *PlayerRole) GiveAwards(awards []protocol.Item) error {
	for _, item := range awards {
		switch item.Type {
		case protocol.ItemTypeMoney:
			// 货币加入MoneySys
			moneySys := pr.sysMgr.GetSystem(custom_id.SysMoney)
			if moneySys != nil {
				if ms, ok := moneySys.(*entitysystem2.MoneySys); ok {
					if err := ms.AddMoney(item.ItemId, item.Count); err != nil {
						return customerr.Wrap(err)
					}
				}
			}
		default:
			// 其他道具加入BagSys
			bagSys := pr.sysMgr.GetSystem(custom_id.SysBag)
			if bagSys != nil {
				if bs, ok := bagSys.(*entitysystem2.BagSys); ok {
					if err := bs.AddItem(item); err != nil {
						return fmt.Errorf("add item to bag failed: %w", err)
					}
				}
			}
		}
	}
	return nil
}

// Consume 消耗道具
func (pr *PlayerRole) Consume(items []protocol.Item) error {
	// 先检查是否足够
	for _, item := range items {
		switch item.Type {
		case protocol.ItemTypeMoney:
			moneySys := pr.sysMgr.GetSystem(custom_id.SysMoney)
			if moneySys != nil {
				if ms, ok := moneySys.(*entitysystem2.MoneySys); ok {
					if !ms.HasEnough(item.ItemId, item.Count) {
						return fmt.Errorf("money not enough: itemId=%d", item.ItemId)
					}
				}
			}
		default:
			bagSys := pr.sysMgr.GetSystem(custom_id.SysBag)
			if bagSys != nil {
				if bs, ok := bagSys.(*entitysystem2.BagSys); ok {
					if !bs.HasEnough(item.ItemId, item.Count) {
						return fmt.Errorf("item not enough: itemId=%d", item.ItemId)
					}
				}
			}
		}
	}

	// 执行消耗
	for _, item := range items {
		switch item.Type {
		case protocol.ItemTypeMoney:
			moneySys := pr.sysMgr.GetSystem(custom_id.SysMoney)
			if moneySys != nil {
				if ms, ok := moneySys.(*entitysystem2.MoneySys); ok {
					ms.ConsumeMoney(item.ItemId, item.Count)
				}
			}
		default:
			bagSys := pr.sysMgr.GetSystem(custom_id.SysBag)
			if bagSys != nil {
				if bs, ok := bagSys.(*entitysystem2.BagSys); ok {
					bs.ConsumeItem(item.ItemId, item.Count)
				}
			}
		}
	}

	return nil
}

// AddExp 增加经验
func (pr *PlayerRole) AddExp(exp uint64) {
	levelSys := pr.sysMgr.GetSystem(custom_id.SysLevel)
	if levelSys != nil {
		if ls, ok := levelSys.(*entitysystem2.LevelSys); ok {
			ls.AddExp(exp)
		}
	}
}

func (pr *PlayerRole) GetPlayerRoleInfo() *protocol.RoleInfo {
	return pr.RoleInfo
}

func (pr *PlayerRole) GetPlayerRoleId() uint64 {
	return pr.GetPlayerRoleInfo().RoleId
}

func (pr *PlayerRole) GetSessionId() string {
	return pr.SessionId
}

func (pr *PlayerRole) GetReconnectKey() string {
	return pr.ReconnectKey
}

func (pr *PlayerRole) GetSystem(sysId custom_id.SystemId) iface.ISystem {
	return pr.sysMgr.GetSystem(sysId)
}

// SendMessage 发送消息给客户端
func (pr *PlayerRole) SendMessage(protoIdH uint16, protoIdL uint16, data []byte) error {
	protoId := protoIdH<<8 | protoIdL
	return gatewaylink.SendToSession(pr.SessionId, protoId, data)
}

// sendReconnectKey 下发重连密钥
func (pr *PlayerRole) sendReconnectKey() error {
	resp := &protocol.LoginSuccessResponse{
		ReconnectKey: pr.ReconnectKey,
		RoleInfo:     pr.RoleInfo,
	}

	data, err := tool.JsonMarshal(resp)
	if err != nil {
		return customerr.Wrap(err)
	}

	return pr.SendMessage(1, 13, data)
}

func init() {

}
