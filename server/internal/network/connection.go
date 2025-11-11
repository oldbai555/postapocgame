package network

import (
	"context"
	"net"
	"sync"
)

// TCPConnection TCP连接实现
type TCPConnection struct {
	conn net.Conn
	meta interface{}
	mu   sync.RWMutex
}

// NewTCPConnection 创建TCP连接
func NewTCPConnection(conn net.Conn) *TCPConnection {
	return &TCPConnection{
		conn: conn,
	}
}

// SendMessage 发送消息
func (tc *TCPConnection) SendMessage(msg *Message) error {
	data, err := defaultCompressionCodec.EncodeMessageWithCompression(msg)
	if err != nil {
		return err
	}
	defer PutBuffer(data)
	_, err = tc.conn.Write(data)
	return err
}

// ReceiveMessage 接收消息
func (tc *TCPConnection) ReceiveMessage(_ context.Context) (*Message, error) {
	msg, err := defaultCompressionCodec.DecodeMessageWithCompression(tc.conn)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Close 关闭连接
func (tc *TCPConnection) Close() error {
	return tc.conn.Close()
}

// RemoteAddr 远程地址
func (tc *TCPConnection) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

// GetMeta 获取连接元数据
func (tc *TCPConnection) GetMeta() interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.meta
}

// SetMeta 设置连接元数据
func (tc *TCPConnection) SetMeta(meta interface{}) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.meta = meta
}
