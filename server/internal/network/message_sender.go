package network

import (
	"encoding/json"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
)

type IMessageSender interface {
	SetConn(conn IConnection)
	SendToClient(sessionId string, msgId uint16, data []byte) error
	ForwardClientMsg(sessionId string, payload []byte) error
	SendRPCRequest(req *RPCRequest) error
	SendRPCResponse(resp *RPCResponse) error

	// 扩展功能
	SendToClientJSON(sessionId string, msgId uint16, v interface{}) error
}

// BaseMessageSender 基础消息发送器
type BaseMessageSender struct {
	conn  IConnection
	codec *Codec
}

// NewBaseMessageSender 创建基础消息发送器
func NewBaseMessageSender(conn IConnection) *BaseMessageSender {
	return &BaseMessageSender{
		conn:  conn,
		codec: DefaultCodec(),
	}
}

func NewBaseMessageSenderWithCodec(conn IConnection, codec *Codec) *BaseMessageSender {
	return &BaseMessageSender{
		conn:  conn,
		codec: codec,
	}
}

func (s *BaseMessageSender) SetConn(conn IConnection) {
	s.conn = conn
}

// SendToClient 发送消息给客户端
func (s *BaseMessageSender) SendToClient(sessionId string, msgId uint16, data []byte) error {
	if s.conn == nil {
		return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "conn is nil")
	}

	// 1. 编码客户端消息
	clientMessage := GetClientMessage()
	clientMessage.MsgId = msgId
	clientMessage.Data = data
	defer PutClientMessage(clientMessage)
	clientMsgBuf, err := s.codec.EncodeClientMessage(&ClientMessage{
		MsgId: msgId,
		Data:  data,
	})
	if err != nil {
		return customerr.Wrap(err, int32(protocol.ErrorCode_Internal_Error))
	}
	defer PutBuffer(clientMsgBuf)

	// 2. 编码转发消息
	forwardMessage := GetForwardMessage()
	forwardMessage.SessionId = sessionId
	forwardMessage.Payload = clientMsgBuf
	defer PutForwardMessage(forwardMessage)
	fwdBuf := s.codec.EncodeForwardMessage(forwardMessage)
	defer PutBuffer(fwdBuf)

	// 3. 编码通用消息
	message := GetMessage()
	message.Type = MsgTypeClient
	message.Payload = fwdBuf
	defer PutMessage(message)

	// 发送
	return s.conn.SendMessage(message)
}

// SendToClientJSON 发送JSON消息给客户端
func (s *BaseMessageSender) SendToClientJSON(sessionId string, msgId uint16, v interface{}) error {
	var data []byte
	var err error

	if d, ok := v.([]byte); ok {
		data = d
	} else {
		data, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}

	return s.SendToClient(sessionId, msgId, data)
}

// ForwardClientMsg 转发客户端消息
func (s *BaseMessageSender) ForwardClientMsg(sessionId string, payload []byte) error {
	// 编码转发消息
	forwardMessage := GetForwardMessage()
	forwardMessage.SessionId = sessionId
	forwardMessage.Payload = payload
	defer PutForwardMessage(forwardMessage)
	fwdBuf := s.codec.EncodeForwardMessage(forwardMessage)
	defer PutBuffer(fwdBuf)

	// 编码通用消息
	message := GetMessage()
	message.Type = MsgTypeClient
	message.Payload = fwdBuf
	defer PutMessage(message)

	return s.conn.SendMessage(message)
}

// SendRPCRequest 发送RPC请求
func (s *BaseMessageSender) SendRPCRequest(req *RPCRequest) error {
	// 编码RPC请求
	rpcBuf := s.codec.EncodeRPCRequest(req)
	defer PutBuffer(rpcBuf)

	// 编码通用消息
	message := GetMessage()
	message.Type = MsgTypeRPCRequest
	message.Payload = rpcBuf
	defer PutMessage(message)

	// 发送
	return s.conn.SendMessage(message)
}

// SendRPCResponse 发送RPC响应
func (s *BaseMessageSender) SendRPCResponse(resp *RPCResponse) error {
	// 编码RPC响应
	rpcBuf := s.codec.EncodeRPCResponse(resp)
	defer PutBuffer(rpcBuf)

	// 编码通用消息
	message := GetMessage()
	message.Type = MsgTypeRPCResponse
	message.Payload = rpcBuf
	defer PutMessage(message)

	// 发送
	return s.conn.SendMessage(message)
}
