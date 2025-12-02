package system

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	iface2 "postapocgame/server/service/gameserver/internel/core/iface"

	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	adaptercontext "postapocgame/server/service/gameserver/internel/adapter/context"
)

// MessageSystemAdapter 玩家消息系统：负责加载并回放离线期间积累的消息
type MessageSystemAdapter struct {
	*BaseSystemAdapter
	owner iface2.IPlayerRole
}

// NewMessageSystemAdapter 创建消息系统适配器
func NewMessageSystemAdapter() *MessageSystemAdapter {
	return &MessageSystemAdapter{
		BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysMessage)),
	}
}

// OnInit 系统初始化时加载离线消息
func (ms *MessageSystemAdapter) OnInit(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

// OnRoleLogin 登录时加载离线消息
func (ms *MessageSystemAdapter) OnRoleLogin(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

// OnRoleReconnect 重连时加载离线消息
func (ms *MessageSystemAdapter) OnRoleReconnect(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

func (ms *MessageSystemAdapter) ensureOwner(ctx context.Context) iface2.IPlayerRole {
	if ms.owner != nil {
		return ms.owner
	}
	playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
	if err != nil {
		log.Errorf("MessageSys.ensureOwner: %v", err)
		return nil
	}
	ms.owner = playerRole
	return playerRole
}

func (ms *MessageSystemAdapter) loadMsgFromDB(afterMsgId uint64) {
	if ms.owner == nil {
		log.Warnf("MessageSys.loadMsgFromDB: owner is nil")
		return
	}
	messages, err := database.LoadPlayerActorMessages(ms.owner.GetPlayerRoleId(), afterMsgId)
	if err != nil {
		log.Errorf("MessageSys.loadMsgFromDB: load failed role=%d err=%v", ms.owner.GetPlayerRoleId(), err)
		return
	}
	if len(messages) == 0 {
		return
	}
	ms.onLoadMsgFromDB(messages)
}

func (ms *MessageSystemAdapter) onLoadMsgFromDB(messages []*database.PlayerActorMessage) {
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if ms.processMessage(msg) {
			if err := database.DeletePlayerActorMessage(uint64(msg.ID)); err != nil {
				log.Warnf("MessageSys.onLoadMsgFromDB: delete msgId=%d failed err=%v", msg.ID, err)
			}
		}
	}
}

func (ms *MessageSystemAdapter) processMessage(msg *database.PlayerActorMessage) bool {
	if err := entitysystem.DispatchPlayerMessage(ms.owner, msg.MsgType, msg.MsgData); err != nil {
		log.Errorf("MessageSys.processMessage: callback failed msgId=%d type=%d err=%v", msg.ID, msg.MsgType, err)
		return false
	}
	return true
}

// 确保实现 ISystem 接口
var _ iface2.ISystem = (*MessageSystemAdapter)(nil)
