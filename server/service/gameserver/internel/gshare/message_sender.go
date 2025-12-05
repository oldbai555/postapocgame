package gshare

import (
	"context"
	"postapocgame/server/service/gameserver/internel/app/manager"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/database"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/log"
)

// SendPlayerActorMessage 将消息发送给玩家Actor（在线即刻处理，离线入库）
func SendPlayerActorMessage(roleId uint64, msgType int32, msg proto.Message) error {
	var msgBytes []byte
	var err error
	if msg != nil {
		msgBytes, err = proto.Marshal(msg)
		if err != nil {
			return err
		}
	}

	role := manager.GetPlayerRole(roleId)
	if role != nil {
		payload := &protocol.AddActorMessageMsg{
			RoleId:  roleId,
			MsgType: msgType,
			MsgData: msgBytes,
		}
		data, err := proto.Marshal(payload)
		if err != nil {
			return err
		}

		ctx := role.WithContext(context.TODO())
		ctx = context.WithValue(ctx, ContextKeySession, role.GetSessionId())
		actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdPlayerMessageMsg), data)
		if err := SendMessageAsync(role.GetSessionId(), actorMsg); err == nil {
			return nil
		} else {
			log.Warnf("SendPlayerActorMessage: send async failed role=%d err=%v, fallback to DB", roleId, err)
		}
	}

	return database.SavePlayerActorMessage(roleId, msgType, msgBytes)
}
