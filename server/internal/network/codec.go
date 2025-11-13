package network

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrFrameTooLarge  = errors.New("frame too large")
	DefaultByteOrder  = binary.LittleEndian
	defaultCodec      = NewCodec()
)

const (
	MaxFrameSize = 1024 * 1024 // 1MB
)

// Codec 编解码器
type Codec struct {
	byteOrder    binary.ByteOrder
	maxFrameSize int
}

// NewCodec 创建编解码器
func NewCodec() *Codec {
	return &Codec{
		byteOrder:    DefaultByteOrder,
		maxFrameSize: MaxFrameSize,
	}
}

// DefaultCodec 获取默认编解码器
func DefaultCodec() *Codec {
	return defaultCodec
}

// ============================================
// 1. 通用消息编解码（使用内存池）
// 格式: [4字节长度][消息类型(1字节)][消息体]
// ============================================

// EncodeMessage 编码消息（使用内存池）
func (c *Codec) EncodeMessage(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}

	payloadLen := len(msg.Payload)
	totalLen := 1 + payloadLen // 1字节类型 + payload

	if totalLen > c.maxFrameSize {
		return nil, ErrFrameTooLarge
	}

	// 从内存池获取缓冲区
	frameSize := 4 + totalLen
	buf := GetBuffer(frameSize)

	// 写入长度
	c.byteOrder.PutUint32(buf[:4], uint32(totalLen))

	// 写入类型
	buf[4] = msg.Type

	// 写入payload
	copy(buf[5:], msg.Payload)

	// 返回数据（调用方负责归还）
	return buf, nil
}

// DecodeMessage 解码消息
func (c *Codec) DecodeMessage(reader io.Reader) (*Message, error) {
	// 读取帧头
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, err
	}

	totalLen := c.byteOrder.Uint32(header)
	if totalLen > uint32(c.maxFrameSize) {
		return nil, ErrFrameTooLarge
	}

	if totalLen < 1 {
		return nil, ErrInvalidMessage
	}

	// 读取消息体
	body := make([]byte, totalLen)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, err
	}

	return &Message{
		Type:    body[0],
		Payload: body[1:],
	}, nil
}

// ============================================
// 2. 客户端消息编解码
// 格式: [msgId(2字节)][data]
// ============================================

// EncodeClientMessage 编码客户端消息（使用内存池）
func (c *Codec) EncodeClientMessage(msg *ClientMessage) ([]byte, error) {
	size := 2 + len(msg.Data)
	buf := GetBuffer(size)

	c.byteOrder.PutUint16(buf[:2], msg.MsgId)
	copy(buf[2:], msg.Data)

	return buf, nil
}

// DecodeClientMessage 解码客户端消息（从payload）
func (c *Codec) DecodeClientMessage(payload []byte) (*ClientMessage, error) {
	if len(payload) < 2 {
		return nil, ErrInvalidMessage
	}

	return &ClientMessage{
		MsgId: c.byteOrder.Uint16(payload[:2]),
		Data:  payload[2:],
	}, nil
}

// ============================================
// 3. 转发消息编解码
// 格式: [sessionIdLen(2字节)][sessionId][payload]
// ============================================

// EncodeForwardMessage 编码转发消息（使用内存池）
func (c *Codec) EncodeForwardMessage(msg *ForwardMessage) []byte {
	sessionIdBytes := []byte(msg.SessionId)
	size := 2 + len(sessionIdBytes) + len(msg.Payload)

	buf := GetBuffer(size)

	offset := 0
	c.byteOrder.PutUint16(buf[offset:], uint16(len(sessionIdBytes)))
	offset += 2

	copy(buf[offset:], sessionIdBytes)
	offset += len(sessionIdBytes)

	copy(buf[offset:], msg.Payload)

	return buf
}

// DecodeForwardMessage 解码转发消息
func (c *Codec) DecodeForwardMessage(data []byte) (*ForwardMessage, error) {
	if len(data) < 2 {
		return nil, ErrInvalidMessage
	}

	offset := 0
	sessionIdLen := c.byteOrder.Uint16(data[offset:])
	offset += 2

	if offset+int(sessionIdLen) > len(data) {
		return nil, ErrInvalidMessage
	}

	sessionId := string(data[offset : offset+int(sessionIdLen)])
	offset += int(sessionIdLen)

	return &ForwardMessage{
		SessionId: sessionId,
		Payload:   data[offset:],
	}, nil
}

// ============================================
// 4. RPC消息编解码
// ============================================

// EncodeRPCRequest 编码RPC请求（使用内存池）
func (c *Codec) EncodeRPCRequest(req *RPCRequest) []byte {
	sessionIdBytes := []byte(req.SessionId)
	size := 4 + 2 + len(sessionIdBytes) + 2 + len(req.Data)

	buf := GetBuffer(size)

	offset := 0
	c.byteOrder.PutUint32(buf[offset:], req.RequestId)
	offset += 4

	c.byteOrder.PutUint16(buf[offset:], uint16(len(sessionIdBytes)))
	offset += 2

	copy(buf[offset:], sessionIdBytes)
	offset += len(sessionIdBytes)

	c.byteOrder.PutUint16(buf[offset:], req.MsgId)
	offset += 2

	copy(buf[offset:], req.Data)

	return buf
}

// DecodeRPCRequest 解码RPC请求
func (c *Codec) DecodeRPCRequest(data []byte) (*RPCRequest, error) {
	if len(data) < 8 {
		return nil, ErrInvalidMessage
	}

	offset := 0
	requestId := c.byteOrder.Uint32(data[offset:])
	offset += 4

	sessionIdLen := c.byteOrder.Uint16(data[offset:])
	offset += 2

	if offset+int(sessionIdLen)+2 > len(data) {
		return nil, ErrInvalidMessage
	}

	sessionId := string(data[offset : offset+int(sessionIdLen)])
	offset += int(sessionIdLen)

	msgId := c.byteOrder.Uint16(data[offset:])
	offset += 2

	return &RPCRequest{
		RequestId: requestId,
		SessionId: sessionId,
		MsgId:     msgId,
		Data:      data[offset:],
	}, nil
}

// EncodeRPCResponse 编码RPC响应（使用内存池）
func (c *Codec) EncodeRPCResponse(resp *RPCResponse) []byte {
	size := 4 + 4 + len(resp.Data)
	buf := GetBuffer(size)

	offset := 0
	c.byteOrder.PutUint32(buf[offset:], resp.RequestId)
	offset += 4

	c.byteOrder.PutUint32(buf[offset:], uint32(resp.Code))
	offset += 4

	copy(buf[offset:], resp.Data)

	return buf
}

// DecodeRPCResponse 解码RPC响应
func (c *Codec) DecodeRPCResponse(data []byte) (*RPCResponse, error) {
	if len(data) < 8 {
		return nil, ErrInvalidMessage
	}

	offset := 0
	requestId := c.byteOrder.Uint32(data[offset:])
	offset += 4

	code := int32(c.byteOrder.Uint32(data[offset:]))
	offset += 4

	return &RPCResponse{
		RequestId: requestId,
		Code:      code,
		Data:      data[offset:],
	}, nil
}

// ============================================
// 5. 会话事件编解码
// ============================================

// EncodeSessionEvent 编码会话事件（使用内存池）
func (c *Codec) EncodeSessionEvent(event *SessionEvent) []byte {
	sessionIdBytes := []byte(event.SessionId)
	userIdBytes := []byte(event.UserId)

	size := 1 + 2 + len(sessionIdBytes) + 2 + len(userIdBytes)
	buf := GetBuffer(size)

	offset := 0
	buf[offset] = byte(event.EventType)
	offset++

	c.byteOrder.PutUint16(buf[offset:], uint16(len(sessionIdBytes)))
	offset += 2

	copy(buf[offset:], sessionIdBytes)
	offset += len(sessionIdBytes)

	c.byteOrder.PutUint16(buf[offset:], uint16(len(userIdBytes)))
	offset += 2

	copy(buf[offset:], userIdBytes)

	return buf
}

// DecodeSessionEvent 解码会话事件
func (c *Codec) DecodeSessionEvent(data []byte) (*SessionEvent, error) {
	if len(data) < 5 {
		return nil, ErrInvalidMessage
	}

	offset := 0
	eventType := SessionEventType(data[offset])
	offset++

	sessionIdLen := c.byteOrder.Uint16(data[offset:])
	offset += 2

	if offset+int(sessionIdLen)+2 > len(data) {
		return nil, ErrInvalidMessage
	}

	sessionId := string(data[offset : offset+int(sessionIdLen)])
	offset += int(sessionIdLen)

	userIdLen := c.byteOrder.Uint16(data[offset:])
	offset += 2

	if offset+int(userIdLen) > len(data) {
		return nil, ErrInvalidMessage
	}

	userId := string(data[offset : offset+int(userIdLen)])

	return &SessionEvent{
		EventType: eventType,
		SessionId: sessionId,
		UserId:    userId,
	}, nil
}

// ============================================
// 7. JSON序列化辅助（使用内存池）
// ============================================

// EncodeClientMessageWithJSON 编码客户端消息（JSON序列化）
func (c *Codec) EncodeClientMessageWithJSON(msgId uint16, v interface{}) ([]byte, error) {
	var data []byte
	var err error

	if d, ok := v.([]byte); ok {
		data = d
	} else {
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	return c.EncodeClientMessage(&ClientMessage{
		MsgId: msgId,
		Data:  data,
	})
}
