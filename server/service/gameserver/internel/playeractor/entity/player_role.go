package entity

import (
	"context"
	"postapocgame/server/internal/event"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"postapocgame/server/service/gameserver/internel/gatewaylink"
	"postapocgame/server/service/gameserver/internel/gevent"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"time"
)

// PlayerRole 玩家角色
type PlayerRole struct {
	// 基础信息
	SessionId string                   `json:"sessionId"`
	RoleData  *protocol.PlayerRoleData `json:"roleInfo"`

	// 重连相关
	ReconnectKey string    `json:"reconnectKey"`
	IsOnline     bool      `json:"isOnline"`
	DisconnectAt time.Time `json:"disconnectAt"`

	// 事件总线（每个玩家独立的事件总线）
	eventBus *event.Bus

	// 系统管理器
	sysMgr *entitysystem.SysMgr
}

// NewPlayerRole 创建玩家角色
func NewPlayerRole(sessionId string, roleInfo *protocol.PlayerRoleData) *PlayerRole {
	pr := &PlayerRole{
		SessionId:    sessionId,
		RoleData:     roleInfo,
		IsOnline:     true,
		ReconnectKey: generateReconnectKey(sessionId, roleInfo.RoleId),
		// 从全局模板克隆独立的事件总线
		eventBus: gevent.ClonePlayerEventBus(),
	}

	// 创建系统管理器
	pr.sysMgr = entitysystem.NewSysMgr(pr)

	return pr
}

// OnLogin 登录回调
func (pr *PlayerRole) OnLogin() error {
	log.Infof("[PlayerRole] OnLogin: RoleId=%d, SessionId=%s", pr.RoleData.RoleId, pr.SessionId)

	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 下发重连密钥
	if err := pr.sendReconnectKey(); err != nil {
		log.Errorf("Send reconnect key failed: %v", err)
	}

	// 🔧 先调用所有系统的 OnOpen（确保初始化完成）
	pr.sysMgr.EachOpenSystem(func(system iface.ISystem) {
		system.OnOpen()
	})

	// 发布玩家登录事件
	pr.Publish(gevent.OnPlayerLogin)

	// 🔧 再调用 OnRoleLogin（此时所有系统已准备就绪）
	pr.sysMgr.EachOpenSystem(func(system iface.ISystem) {
		system.OnRoleLogin()
	})

	return nil
}

// OnLogout 登出回调
func (pr *PlayerRole) OnLogout() error {
	log.Infof("[PlayerRole] OnLogout: RoleId=%d", pr.RoleData.RoleId)

	pr.IsOnline = false

	// 发布玩家登出事件
	pr.Publish(gevent.OnPlayerLogout)

	return nil
}

// OnReconnect 重连回调
func (pr *PlayerRole) OnReconnect(newSessionId string) error {
	log.Infof("[PlayerRole] OnReconnect: RoleId=%d, OldSession=%s, NewSession=%s",
		pr.RoleData.RoleId, pr.SessionId, newSessionId)

	pr.SessionId = newSessionId
	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 下发重连密钥
	if err := pr.sendReconnectKey(); err != nil {
		log.Errorf("Send reconnect key failed: %v", err)
	}

	// 发布玩家重连事件
	pr.Publish(gevent.OnPlayerReconnect)

	// 调用系统管理器的重连方法
	return nil
}

// OnDisconnect 断线回调
func (pr *PlayerRole) OnDisconnect() {
	log.Infof("[PlayerRole] OnDisconnect: RoleId=%d", pr.RoleData.RoleId)

	pr.IsOnline = false
	pr.DisconnectAt = time.Now()
}

// Close 关闭回调（3分钟超时或主动登出）
func (pr *PlayerRole) Close() error {
	log.Infof("[PlayerRole] Close: RoleId=%d", pr.RoleData.RoleId)

	// 调用登出
	err := pr.OnLogout()
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return nil
}

func (pr *PlayerRole) GetPlayerRoleData() *protocol.PlayerRoleData {
	return pr.RoleData
}

func (pr *PlayerRole) GetPlayerRoleId() uint64 {
	return pr.GetPlayerRoleData().RoleId
}

func (pr *PlayerRole) GetSessionId() string {
	return pr.SessionId
}

func (pr *PlayerRole) GetReconnectKey() string {
	return pr.ReconnectKey
}

func (pr *PlayerRole) GetSystem(sysId uint32) iface.ISystem {
	return pr.sysMgr.GetSystem(sysId)
}

// SendMessage 发送消息给客户端
func (pr *PlayerRole) SendMessageHL(protoIdH uint16, protoIdL uint16, data []byte) error {
	protoId := protoIdH<<8 | protoIdL
	return pr.SendMessage(protoId, data)
}
func (pr *PlayerRole) SendMessage(protoId uint16, data []byte) error {
	return gatewaylink.SendToSession(pr.SessionId, protoId, data)
}

// sendReconnectKey 下发重连密钥
func (pr *PlayerRole) sendReconnectKey() error {
	resp := &protocol.S2CReconnectKeyReq{
		ReconnectKey: pr.ReconnectKey,
	}

	data, err := tool.JsonMarshal(resp)
	if err != nil {
		return customerr.Wrap(err)
	}

	return pr.SendMessage(uint16(protocol.S2CProtocol_S2CReconnectKey), data)
}

// Publish 发布事件（在当前玩家的事件总线上）
func (pr *PlayerRole) Publish(typ event.Type, args ...interface{}) {
	ev := event.NewEvent(typ, args...)
	ctx := context.Background()
	context.WithValue(ctx, "playerRoleId", pr.GetPlayerRoleId())
	pr.eventBus.Publish(ctx, ev)
	return
}

func init() {

}
