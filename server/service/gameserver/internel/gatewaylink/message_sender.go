package gatewaylink

import (
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

func SendToSessionJSON(sessionId string, msgId uint16, v interface{}) error {
	return GetMessageSender().SendToClientJSON(sessionId, msgId, v)
}

func ForwardClientMsg(sessionId string, payload []byte) error {
	return GetMessageSender().ForwardClientMsg(sessionId, payload)
}
