package system

import (
	"context"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
	"postapocgame/server/service/gameserver/internel/core/iface"
	"postapocgame/server/service/gameserver/internel/usecase/interfaces"
	"time"
)

// ChatSystemAdapter 聊天系统适配器
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
	if _, err := adaptercontext.GetPlayerRoleFromContext(ctx); err != nil {
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
