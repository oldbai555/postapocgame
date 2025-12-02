package entitysystem

import (
	"fmt"
	"postapocgame/server/service/gameserver/internel/app/engine"
	"postapocgame/server/service/gameserver/internel/core/iface"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/pkg/log"
)

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
