package system

import (
	"context"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/app/playeractor/entitysystem"
	"postapocgame/server/service/gameserver/internel/gshare"
	"postapocgame/server/service/gameserver/internel/iface"
)

type MessageSystemAdapter struct {
	*BaseSystemAdapter
	owner iface.IPlayerRole
}

const (
	// MaxPlayerActorMessages 每个玩家最多保存的消息数量
	MaxPlayerActorMessages = 1000
	// MessageExpireDays 消息过期天数（超过此天数的消息将被清理）
	MessageExpireDays = 7
)

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

func (ms *MessageSystemAdapter) ensureOwner(ctx context.Context) iface.IPlayerRole {
	if ms.owner != nil {
		return ms.owner
	}
	playerRole, err := gshare.GetPlayerRoleFromContext(ctx)
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

// OnNewDay 新的一天时清理过期消息
func (ms *MessageSystemAdapter) OnNewDay(ctx context.Context) {
	if ms.owner == nil {
		return
	}
	// 清理超过7天的消息
	deletedCount, err := database.DeleteExpiredPlayerActorMessages(MessageExpireDays)
	if err != nil {
		log.Errorf("MessageSys.OnNewDay: delete expired messages failed err=%v", err)
	} else if deletedCount > 0 {
		log.Infof("MessageSys.OnNewDay: deleted %d expired messages (older than %d days)", deletedCount, MessageExpireDays)
	}
}

// RunOne 每帧调用，检查消息数量限制
func (ms *MessageSystemAdapter) RunOne(ctx context.Context) {
	if ms.owner == nil {
		return
	}
	// 检查消息数量限制
	count, err := database.GetPlayerActorMessageCount(ms.owner.GetPlayerRoleId())
	if err != nil {
		log.Errorf("MessageSys.RunOne: get message count failed role=%d err=%v", ms.owner.GetPlayerRoleId(), err)
		return
	}
	if count > MaxPlayerActorMessages {
		// 删除最旧的消息，保留最新的 MaxPlayerActorMessages 条
		deletedCount, err := database.DeleteOldestPlayerActorMessages(ms.owner.GetPlayerRoleId(), MaxPlayerActorMessages)
		if err != nil {
			log.Errorf("MessageSys.RunOne: delete oldest messages failed role=%d err=%v", ms.owner.GetPlayerRoleId(), err)
		} else if deletedCount > 0 {
			log.Warnf("MessageSys.RunOne: deleted %d oldest messages for role=%d (exceeded limit %d)", deletedCount, ms.owner.GetPlayerRoleId(), MaxPlayerActorMessages)
		}
	}
}

// 确保实现 ISystem 接口
var _ iface.ISystem = (*MessageSystemAdapter)(nil)

func init() {
	entitysystem.RegisterSystemFactory(uint32(protocol.SystemId_SysMessage), func() iface.ISystem {
		return NewMessageSystemAdapter()
	})
}
