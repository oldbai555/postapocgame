/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"github.com/gorilla/websocket"
	"net"
)

// WebSocketConnection WebSocket连接实现
type WebSocketConnection struct {
	conn         *websocket.Conn
	maxFrameSize int
}

// NewWebSocketConnection 创建WebSocket连接
func NewWebSocketConnection(conn *websocket.Conn, maxFrameSize int) *WebSocketConnection {
	return &WebSocketConnection{
		conn:         conn,
		maxFrameSize: maxFrameSize,
	}
}

// Read 读取数据
func (wc *WebSocketConnection) Read(ctx context.Context) ([]byte, error) {
	messageType, data, err := wc.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// 只处理二进制消息
	if messageType != websocket.BinaryMessage {
		return nil, ErrInvalidFrame
	}

	if len(data) > wc.maxFrameSize {
		return nil, ErrFrameTooLarge
	}

	return data, nil
}

// Write 写入数据
func (wc *WebSocketConnection) Write(data []byte) error {
	if len(data) > wc.maxFrameSize {
		return ErrFrameTooLarge
	}

	return wc.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close 关闭连接
func (wc *WebSocketConnection) Close() error {
	return wc.conn.Close()
}

// RemoteAddr 远程地址
func (wc *WebSocketConnection) RemoteAddr() net.Addr {
	return wc.conn.RemoteAddr()
}

// Type 连接类型
func (wc *WebSocketConnection) Type() ConnType {
	return ConnTypeWebSocket
}
