/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"net"
	"time"
)

// ConnType 连接类型
type ConnType int

const (
	ConnTypeTCP ConnType = iota
	ConnTypeWebSocket
)

// SessionState 会话状态
type SessionState int

const (
	SessionStateConnected SessionState = iota // 已连接
	SessionStateAuthed                        // 已认证
	SessionStateClosed                        // 已关闭
)

// Session 统一的会话抽象
type Session struct {
	ID         string       // 会话ID
	Addr       net.Addr     // 客户端地址
	ConnType   ConnType     // 连接类型
	State      SessionState // 会话状态
	UserID     string       // 用户ID(认证后设置)
	SendChan   chan []byte  // 发送消息通道
	CreatedAt  time.Time    // 创建时间
	LastActive time.Time    // 最后活跃时间
}

// SessionEventType 会话事件类型
type SessionEventType int

const (
	SessionEventNew   SessionEventType = iota // 新建会话
	SessionEventAuth                          // 会话认证
	SessionEventClose                         // 会话关闭
)

// SessionEvent 会话事件
type SessionEvent struct {
	Type      SessionEventType
	SessionID string
	UserID    string
	Timestamp time.Time
}

// FramedMessage 帧消息(包含会话ID和消息体)
type FramedMessage struct {
	SessionID string
	Payload   []byte
}

// IConnection 网络连接接口
type IConnection interface {
	// Read 读取数据
	Read(ctx context.Context) ([]byte, error)
	// Write 写入数据
	Write(data []byte) error
	// Close 关闭连接
	Close() error
	// RemoteAddr 远程地址
	RemoteAddr() net.Addr
	// Type 连接类型
	Type() ConnType
}

// IAuthenticator 认证器接口
type IAuthenticator interface {
	// Authenticate 执行认证
	Authenticate(ctx context.Context, token string) (userID string, err error)
}

// IGameServerConnector GameServer连接器接口
type IGameServerConnector interface {
	// Connect 连接到GameServer
	Connect(ctx context.Context, addr string) error
	// NotifySessionEvent 通知会话事件
	NotifySessionEvent(ctx context.Context, event *SessionEvent) error
	// ForwardMessage 转发消息
	ForwardMessage(ctx context.Context, msg *FramedMessage) error
	// ReceiveMessage 接收来自GameServer的消息
	ReceiveMessage(ctx context.Context) (*FramedMessage, error)
	// Close 关闭连接
	Close() error
}

// IGateway Gateway接口
type IGateway interface {
	// Start 启动网关
	Start(ctx context.Context) error
	// Stop 停止网关
	Stop(ctx context.Context) error
	// GetSession 获取会话
	GetSession(sessionID string) (*Session, bool)
	// CloseSession 关闭会话
	CloseSession(sessionID string) error
}
