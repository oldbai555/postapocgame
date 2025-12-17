package router

import (
	"context"
	"sync"

	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
)

// ProtocolHandler 统一的 C2S 协议处理函数签名
type ProtocolHandler func(ctx context.Context, msg *network.ClientMessage) error

var (
	protocolHandlers   = make(map[uint16]ProtocolHandler)
	protocolHandlersMu sync.RWMutex
)

// RegisterProtocolHandler 在 ProtocolRouter 层注册 C2S 协议入口
func RegisterProtocolHandler(protoID uint16, handler ProtocolHandler) {
	if handler == nil {
		log.Warnf("[protocol-router] skip registering proto=%d with nil handler", protoID)
		return
	}

	protocolHandlersMu.Lock()
	defer protocolHandlersMu.Unlock()
	if _, exists := protocolHandlers[protoID]; exists {
		log.Stackf("[protocol-router] proto=%d already registered", protoID)
		return
	}
	protocolHandlers[protoID] = handler
}

func getProtocolHandler(protoID uint16) ProtocolHandler {
	protocolHandlersMu.RLock()
	defer protocolHandlersMu.RUnlock()
	return protocolHandlers[protoID]
}
