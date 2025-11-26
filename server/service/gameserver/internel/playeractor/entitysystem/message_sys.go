package entitysystem

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
	"postapocgame/server/service/gameserver/internel/engine"
	"postapocgame/server/service/gameserver/internel/iface"
)

// MessageSys 玩家消息系统：负责加载并回放离线期间积累的消息
type MessageSys struct {
	*BaseSystem
	owner iface.IPlayerRole
}

// NewMessageSys 创建消息系统
func NewMessageSys() *MessageSys {
	return &MessageSys{
		BaseSystem: NewBaseSystem(uint32(protocol.SystemId_SysMessage)),
	}
}

// GetMessageSys 便捷方法：从上下文中获取消息系统
func GetMessageSys(ctx context.Context) *MessageSys {
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("MessageSys.GetMessageSys: %v", err)
		return nil
	}
	system := playerRole.GetSystem(uint32(protocol.SystemId_SysMessage))
	if system == nil {
		return nil
	}
	msgSys, _ := system.(*MessageSys)
	return msgSys
}

func (ms *MessageSys) OnInit(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

func (ms *MessageSys) OnRoleLogin(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

func (ms *MessageSys) OnRoleReconnect(ctx context.Context) {
	if ms.ensureOwner(ctx) == nil {
		return
	}
	ms.loadMsgFromDB(0)
}

func (ms *MessageSys) ensureOwner(ctx context.Context) iface.IPlayerRole {
	if ms.owner != nil {
		return ms.owner
	}
	playerRole, err := GetIPlayerRoleByContext(ctx)
	if err != nil {
		log.Errorf("MessageSys.ensureOwner: %v", err)
		return nil
	}
	ms.owner = playerRole
	return playerRole
}

func (ms *MessageSys) loadMsgFromDB(afterMsgId uint64) {
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

func (ms *MessageSys) onLoadMsgFromDB(messages []*database.PlayerActorMessage) {
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

func (ms *MessageSys) processMessage(msg *database.PlayerActorMessage) bool {
	if err := DispatchPlayerMessage(ms.owner, msg.MsgType, msg.MsgData); err != nil {
		log.Errorf("MessageSys.processMessage: callback failed msgId=%d type=%d err=%v", msg.ID, msg.MsgType, err)
		return false
	}
	return true
}

// DispatchPlayerMessage 用于（在线或离线加载时）分发玩家消息
func DispatchPlayerMessage(owner iface.IPlayerRole, msgType int32, msgData []byte) error {
	if owner == nil {
		return fmt.Errorf("DispatchPlayerMessage: owner is nil")
	}

	var payload proto.Message
	if pb := engine.GetMessagePb3(msgType); pb != nil {
		if len(msgData) > 0 {
			if err := proto.Unmarshal(msgData, pb); err != nil {
				return err
			}
		}
		payload = pb
	} else if len(msgData) > 0 {
		return fmt.Errorf("no proto factory registered for msgType=%d", msgType)
	}

	callback := engine.GetMessageCallback(msgType)
	if callback == nil {
		log.Warnf("DispatchPlayerMessage: no callback registered msgType=%d role=%d", msgType, owner.GetPlayerRoleId())
		return nil
	}

	return callback(owner, payload)
}

func init() {
	RegisterSystemFactory(uint32(protocol.SystemId_SysMessage), func() iface.ISystem {
		return NewMessageSys()
	})
}
