package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/app/playeractor/usecase/interfaces"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
	"time"
)

type ChatSystemAdapter struct {
	*BaseSystemAdapter
	lastChatTime time.Time
}

func NewChatSystemAdapter() *ChatSystemAdapter {
	return &ChatSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysChat)),
	}
}

// OnInit 初始化
func (a *ChatSystemAdapter) OnInit(ctx context.Context) {
	if _, err := gshare.GetPlayerRoleFromContext(ctx); err != nil {
		log.Errorf("chat sys OnInit get role err:%v", err)
	}
}

// CanSend 判断是否可发送
func (a *ChatSystemAdapter) CanSend(now time.Time, cooldown time.Duration) bool {
	return now.Sub(a.lastChatTime) >= cooldown
}

// MarkSent 记录发送时间
func (a *ChatSystemAdapter) MarkSent(now time.Time) {
	a.lastChatTime = now
}

var _ iface.ISystem = (*ChatSystemAdapter)(nil)
var _ interfaces.ChatRateLimiter = (*ChatSystemAdapter)(nil)

// GetChatSys 获取聊天系统
func GetChatSys(ctx context.Context) *ChatSystemAdapter {
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("get player role error:%v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysChat))
	if system == nil {
		return nil
	}
	chatSys, ok := system.(*ChatSystemAdapter)
	if !ok || !chatSys.IsOpened() {
		return nil
	}
	return chatSys
}

func init() {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysChat), func() iface.ISystem {
		return NewChatSystemAdapter()
	})

	// 协议注册由 controller 包负责
}
