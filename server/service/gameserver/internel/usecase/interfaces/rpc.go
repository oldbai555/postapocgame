package interfaces

import "context"

// ProtocolType DungeonServer 协议类型
type ProtocolType uint8

const (
	ProtocolTypeUnknown ProtocolType = 0
	ProtocolTypeCommon  ProtocolType = 1
	ProtocolTypeUnique  ProtocolType = 2
)

// ProtocolInfo DungeonServer 协议信息
type ProtocolInfo struct {
	ProtoID  uint16
	IsCommon bool
}

// DungeonServerGateway DungeonServer RPC 接口（Use Case 层定义）
type DungeonServerGateway interface {
	// AsyncCall 异步调用 DungeonServer
	AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error

	// RegisterRPCHandler 注册 RPC 处理器（用于接收 DungeonServer 回调）
	RegisterRPCHandler(msgId uint16, handler func(ctx context.Context, sessionId string, data []byte) error)

	// 协议路由相关方法
	IsDungeonProtocol(protoId uint16) bool
	GetSrvTypeForProtocol(protoId uint16) (srvType uint8, protocolType ProtocolType, ok bool)

	// 协议注册
	RegisterProtocols(srvType uint8, protocols []ProtocolInfo) error
	UnregisterProtocols(srvType uint8) error
}
