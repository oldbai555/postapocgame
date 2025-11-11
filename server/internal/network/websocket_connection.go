package network

import (
	"bytes"
	"context"
	"github.com/gorilla/websocket"
	"net"
	"sync"
)

// WebSocketConnection WebSocket连接实现
type WebSocketConnection struct {
	conn  *websocket.Conn
	codec *Codec
	meta  interface{}
	mu    sync.RWMutex
}

// NewWebSocketConnection 创建WebSocket连接
func NewWebSocketConnection(conn *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		conn:  conn,
		codec: DefaultCodec(),
	}
}

// SendMessage 发送消息
func (wc *WebSocketConnection) SendMessage(msg *Message) error {
	data, err := wc.codec.EncodeMessage(msg)
	if err != nil {
		return err
	}
	return wc.conn.WriteMessage(websocket.BinaryMessage, data)
}

// ReceiveMessage 接收消息
func (wc *WebSocketConnection) ReceiveMessage(ctx context.Context) (*Message, error) {
	messageType, data, err := wc.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// 只处理二进制消息
	if messageType != websocket.BinaryMessage {
		return nil, ErrInvalidMessage
	}

	// 使用 bytes.Buffer 包装数据以便 DecodeMessage 读取
	buf := bytes.NewBuffer(data)
	return wc.codec.DecodeMessage(buf)
}

// Close 关闭连接
func (wc *WebSocketConnection) Close() error {
	return wc.conn.Close()
}

// RemoteAddr 远程地址
func (wc *WebSocketConnection) RemoteAddr() net.Addr {
	return wc.conn.RemoteAddr()
}

// GetMeta 获取连接元数据
func (wc *WebSocketConnection) GetMeta() interface{} {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.meta
}

// SetMeta 设置连接元数据
func (wc *WebSocketConnection) SetMeta(meta interface{}) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.meta = meta
}
