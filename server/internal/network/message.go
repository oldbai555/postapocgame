/**
 * @Author: zjj
 * @Date: 2025/11/6
 * @Desc:
**/

package network

// Message 通用消息结构
type Message struct {
	Type    MessageType // 消息类型
	Payload []byte      // 消息体
}

func (x *Message) Reset() {
	*x = Message{}
}

// SessionEvent 会话事件
type SessionEvent struct {
	EventType SessionEventType
	SessionId string
	UserId    string
}

// ForwardMessage 转发消息
type ForwardMessage struct {
	SessionId string
	Payload   []byte
}

func (x *ForwardMessage) Reset() {
	x.SessionId = ""
	if x.Payload != nil {
		x.Payload = x.Payload[:0]
	}
}

// RPCRequest RPC请求
type RPCRequest struct {
	RequestId uint32 // 请求ID
	SessionId string // sessionId 空表示为系统消息，否则就是玩家消息
	MsgId     uint16 // 消息ID
	Data      []byte // 消息数据
}

func (x *RPCRequest) Reset() {
	*x = RPCRequest{}
}

// RPCResponse RPC响应
type RPCResponse struct {
	RequestId uint32 // 请求ID
	Code      int32  // 响应码(0表示成功)
	Data      []byte // 响应数据
}

func (x *RPCResponse) Reset() {
	*x = RPCResponse{}
}

// ClientMessage 客户端消息
type ClientMessage struct {
	MsgId uint16
	Data  []byte
}

func (x *ClientMessage) Reset() {
	*x = ClientMessage{}
}
