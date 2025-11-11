package network

// MessageType 消息类型
type MessageType = byte

const (
	MsgTypeSessionEvent MessageType = 0x01 // 会话事件
	MsgTypeClient       MessageType = 0x02 // 客户端消息
	MsgTypeRPCRequest   MessageType = 0x03 // RPC请求
	MsgTypeRPCResponse  MessageType = 0x04 // RPC响应
	MsgTypeHandshake    MessageType = 0x05 // 握手消息
	MsgTypeHeartbeat    MessageType = 0x06 // 心跳消息
)

// SessionEventType 会话事件类型
type SessionEventType byte

const (
	SessionEventNew   SessionEventType = 0x00
	SessionEventClose SessionEventType = 0x01
)
