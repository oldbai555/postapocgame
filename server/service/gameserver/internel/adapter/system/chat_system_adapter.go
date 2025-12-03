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
//
// 生命周期职责：
// - OnInit: 初始化聊天系统（暂无需特殊处理）
// - 其他生命周期: 暂未使用
//
// 业务逻辑：聊天消息发送、限频、敏感词过滤等规则均在 UseCase 层实现
// 状态管理：维护 lastChatTime 用于限频（限流逻辑属于框架状态管理，保留在适配层）
// 外部交互：统一通过 PublicActorGateway 与 PublicActor 通信
//
// ⚠️ 防退化机制：禁止在 SystemAdapter 中编写业务规则逻辑，只允许调用 UseCase 与管理生命周期
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
