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
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"postapocgame/server/service/gameserver/internel/playeractor/entitysystem"
	"time"
)

// PlayerRole 玩家角色
type PlayerRole struct {
	// 基础信息
	SessionId  string
	SimpleData *protocol.PlayerSimpleData
	BinaryData *protocol.PlayerRoleBinaryData

	// 重连相关
	ReconnectKey string
	IsOnline     bool
	DisconnectAt time.Time

	// 事件总线（每个玩家独立的事件总线）
	eventBus *event.Bus

	// 系统管理器
	sysMgr iface.ISystemMgr
}

// NewPlayerRole 创建玩家角色
func NewPlayerRole(sessionId string, roleInfo *protocol.PlayerSimpleData) *PlayerRole {
	pr := &PlayerRole{
		SessionId:    sessionId,
		SimpleData:   roleInfo,
		IsOnline:     true,
		ReconnectKey: generateReconnectKey(sessionId, roleInfo.RoleId),
		// 从全局模板克隆独立的事件总线
		eventBus: gevent.ClonePlayerEventBus(),
	}
	// 创建系统管理器
	pr.sysMgr = entitysystem.NewSysMgr()
	if pr.BinaryData == nil {
		pr.BinaryData = &protocol.PlayerRoleBinaryData{}
	}
	if pr.BinaryData.SysOpenStatus == nil {
		pr.BinaryData.SysOpenStatus = make(map[uint32]uint32)
	}
	err := pr.sysMgr.OnInit(pr.WithContext(nil))
	if err != nil {
		log.Errorf("sys mgr on init failed, err:%v", err)
		return nil
	}
	return pr
}

// OnLogin 登录回调
func (pr *PlayerRole) OnLogin() error {
	log.Infof("[PlayerRole] OnLogin: RoleId=%d, SessionId=%s", pr.SimpleData.RoleId, pr.SessionId)

	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 发布玩家登录事件
	pr.Publish(gevent.OnPlayerLogin)

	var resp protocol.S2CLoginSuccessReq
	resp.ReconnectKey = pr.ReconnectKey
	resp.RoleData = pr.SimpleData

	return pr.SendJsonMessage(uint16(protocol.S2CProtocol_S2CLoginSuccess), resp)
}

// OnLogout 登出回调
func (pr *PlayerRole) OnLogout() error {
	log.Infof("[PlayerRole] OnLogout: RoleId=%d", pr.SimpleData.RoleId)

	pr.IsOnline = false

	// 发布玩家登出事件
	pr.Publish(gevent.OnPlayerLogout)

	return nil
}

// OnReconnect 重连回调
func (pr *PlayerRole) OnReconnect(newSessionId string) error {
	log.Infof("[PlayerRole] OnReconnect: RoleId=%d, OldSession=%s, NewSession=%s",
		pr.SimpleData.RoleId, pr.SessionId, newSessionId)

	pr.SessionId = newSessionId
	pr.IsOnline = true
	pr.DisconnectAt = time.Time{}

	// 发布玩家重连事件
	pr.Publish(gevent.OnPlayerReconnect)

	var resp protocol.S2CReconnectSuccessReq
	resp.ReconnectKey = pr.ReconnectKey
	resp.RoleData = pr.SimpleData

	// 调用系统管理器的重连方法
	return pr.SendJsonMessage(uint16(protocol.S2CProtocol_S2CReconnectSuccess), resp)
}

// OnDisconnect 断线回调
func (pr *PlayerRole) OnDisconnect() {
	log.Infof("[PlayerRole] OnDisconnect: RoleId=%d", pr.SimpleData.RoleId)

	pr.IsOnline = false
	pr.DisconnectAt = time.Now()
}

// Close 关闭回调（3分钟超时或主动登出）
func (pr *PlayerRole) Close() error {
	log.Infof("[PlayerRole] Close: RoleId=%d", pr.SimpleData.RoleId)

	// 调用登出
	err := pr.OnLogout()
	if err != nil {
		log.Errorf("err:%v", err)
	}
	return nil
}

func (pr *PlayerRole) GetBinaryData() *protocol.PlayerRoleBinaryData {
	return pr.BinaryData
}

func (pr *PlayerRole) GetPlayerRoleId() uint64 {
	return pr.SimpleData.RoleId
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

func (pr *PlayerRole) SendMessage(protoId uint16, data []byte) error {
	return gatewaylink.SendToSession(pr.SessionId, protoId, data)
}

func (pr *PlayerRole) SendJsonMessage(protoId uint16, v interface{}) error {
	bytes, err := tool.JsonMarshal(v)
	if err != nil {
		return customerr.Wrap(err)
	}
	return pr.SendMessage(protoId, bytes)
}

func (pr *PlayerRole) Publish(typ event.Type, args ...interface{}) {
	ev := event.NewEvent(typ, args...)
	ctx := pr.WithContext(nil)
	pr.eventBus.Publish(ctx, ev)
	return
}

func (pr *PlayerRole) WithContext(parentCtx context.Context) context.Context {
	var ctx = parentCtx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, gshare.ContextKeyRole, pr)
	return ctx
}
func (pr *PlayerRole) GetSysStatus(sysId uint32) bool {
	idxInt := sysId / 32
	idxByte := sysId % 32

	flag := pr.GetBinaryData().SysOpenStatus[idxInt]

	return tool.IsSetBit(flag, idxByte)
}

func (pr *PlayerRole) GetSysStatusData() map[uint32]uint32 {
	return pr.GetBinaryData().SysOpenStatus
}

func (pr *PlayerRole) SetSysStatus(sysId uint32, isOpen bool) {
	idxInt := sysId / 32
	idxByte := sysId % 32

	binary := pr.GetBinaryData()
	if isOpen {
		binary.SysOpenStatus[idxInt] = tool.SetBit(binary.SysOpenStatus[idxInt], idxByte)
	} else {
		binary.SysOpenStatus[idxInt] = tool.ClearBit(binary.SysOpenStatus[idxInt], idxByte)
	}
}

func (pr *PlayerRole) GetSysMgr() iface.ISystemMgr {
	return pr.sysMgr
}
