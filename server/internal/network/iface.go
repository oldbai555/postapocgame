/**
 * @Author: zjj
 * @Date: 2025/11/10
 * @Desc:
**/

package network

import (
	"context"
	"net"
)

type INetworkMessageHandler interface {
	// HandleMessage 处理原始消息（不需要解码）
	HandleMessage(ctx context.Context, conn IConnection, msg *Message) error
}

// IConnection 连接接口
type IConnection interface {
	// SendMessage 发送消息
	SendMessage(msg *Message) error
	// ReceiveMessage 接收消息
	ReceiveMessage(ctx context.Context) (*Message, error)
	// Close 关闭连接
	Close() error
	// RemoteAddr 远程地址
	RemoteAddr() net.Addr
	// GetMeta 获取连接元数据
	GetMeta() interface{}
	// SetMeta 设置连接元数据
	SetMeta(meta interface{})
}

// ITCPServer TCP服务器接口
type ITCPServer interface {
	// Start 启动服务器
	Start(ctx context.Context) error
	// Stop 停止服务器
	Stop(ctx context.Context) error
}

// ITCPClient TCP客户端接口
type ITCPClient interface {
	Connect(ctx context.Context, addr string) error
	SendMessage(msg *Message) error
	Close() error
	GetConnection() IConnection
	IsConnected() bool
}
