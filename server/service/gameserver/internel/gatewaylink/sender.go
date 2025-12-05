package gatewaylink

import (
	"google.golang.org/protobuf/proto"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/customerr"
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
		if globalSender == nil {
			panic("gatewaylink: message sender init returned nil")
		}
	})
	return globalSender
}

func SendToSession(sessionId string, msgId uint16, data []byte) error {
	sender := GetMessageSender()
	if sender == nil {
		return customerr.NewError("message sender is nil")
	}
	return sender.SendToClient(sessionId, msgId, data)
}

func SendToSessionProto(sessionId string, msgId uint16, message proto.Message) error {
	sender := GetMessageSender()
	if sender == nil {
		return customerr.NewError("message sender is nil")
	}
	return sender.SendToClientProto(sessionId, msgId, message)
}

func ForwardClientMsg(sessionId string, payload []byte) error {
	sender := GetMessageSender()
	if sender == nil {
		return customerr.NewError("message sender is nil")
	}
	return sender.ForwardClientMsg(sessionId, payload)
}
