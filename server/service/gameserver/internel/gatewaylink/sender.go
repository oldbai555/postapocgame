package gatewaylink

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"sync"
)

var (
	globalSender network.IMessageSender
	senderOnce   sync.Once
)

// GetMessageSender 获取全局消息发送器
func GetMessageSender() network.IMessageSender {
	senderOnce.Do(func() {
		globalSender = network.NewBaseMessageSender(nil)
	})
	return globalSender
}

func SendToSession(sessionId string, msgId uint16, data []byte) error {
	return GetMessageSender().SendToClient(sessionId, msgId, data)
}

func SendToSessionProto(sessionId string, msgId uint16, message proto.Message) error {
	return GetMessageSender().SendToClientProto(sessionId, msgId, message)
}

func ForwardClientMsg(sessionId string, payload []byte) error {
	return GetMessageSender().ForwardClientMsg(sessionId, payload)
}
