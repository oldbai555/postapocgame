package publicactor

import (
	"context"
	"postapocgame/server/service/gameserver/internel/gshare"

	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

func sendClientMessageViaPlayerActor(sessionId string, msgId uint16, data []byte) error {
	if sessionId == "" {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "session id is empty")
	}

	req := &protocol.PlayerActorMsgIdSendToClientReq{
		MsgId: uint32(msgId),
		Data:  data,
	}
	payload, err := proto.Marshal(req)
	if err != nil {
		return customerr.Wrap(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, gshare.ContextKeySession, sessionId)
	actorMsg := actor.NewBaseMessage(ctx, uint16(protocol.PlayerActorMsgId_PlayerActorMsgIdSendToClient), payload)
	if err := gshare.SendMessageAsync(sessionId, actorMsg); err != nil {
		return customerr.Wrap(err)
	}
	return nil
}

func sendClientProtoViaPlayerActor(sessionId string, msgId uint16, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return customerr.Wrap(err)
	}
	return sendClientMessageViaPlayerActor(sessionId, msgId, data)
}

func logSendFailure(sessionId string, msgId uint16, err error) {
	if err != nil {
		log.Warnf("send message to player failed: session=%s msgId=%d err=%v", sessionId, msgId, err)
	}
}
