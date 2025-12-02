package engine

import (
	"postapocgame/server/service/gameserver/internel/core/iface"
	"sync"

	"google.golang.org/protobuf/proto"
)

// MessageCallbackFunc 玩家消息回调函数
type MessageCallbackFunc func(owner iface.IPlayerRole, msg proto.Message) error

// MessagePb3FactoryFunc 用于创建消息对应的 proto 结构
type MessagePb3FactoryFunc func() proto.Message

var (
	messageRegistryMu sync.RWMutex
	messageCallbacks  = make(map[int32]MessageCallbackFunc)
	messageFactories  = make(map[int32]MessagePb3FactoryFunc)
)

// RegisterMessageCallback 注册玩家消息回调；传入 nil 表示移除
func RegisterMessageCallback(msgType int32, callback MessageCallbackFunc) {
	messageRegistryMu.Lock()
	defer messageRegistryMu.Unlock()

	if callback == nil {
		delete(messageCallbacks, msgType)
		return
	}
	messageCallbacks[msgType] = callback
}

// RegisterMessagePb3Factory 注册玩家消息 proto 工厂；传入 nil 表示移除
func RegisterMessagePb3Factory(msgType int32, factory MessagePb3FactoryFunc) {
	messageRegistryMu.Lock()
	defer messageRegistryMu.Unlock()

	if factory == nil {
		delete(messageFactories, msgType)
		return
	}
	messageFactories[msgType] = factory
}

// GetMessageCallback 获取注册的回调
func GetMessageCallback(msgType int32) MessageCallbackFunc {
	messageRegistryMu.RLock()
	defer messageRegistryMu.RUnlock()
	return messageCallbacks[msgType]
}

// GetMessagePb3 获取 proto 工厂实例
func GetMessagePb3(msgType int32) proto.Message {
	messageRegistryMu.RLock()
	factory := messageFactories[msgType]
	messageRegistryMu.RUnlock()
	if factory == nil {
		return nil
	}
	return factory()
}
