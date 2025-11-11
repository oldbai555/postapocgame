/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package clientnet

import (
	"context"
	"net"
	"postapocgame/server/internal/network"
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

// IConnection 网络连接接口
type IConnection interface {
	Close() error
	RemoteAddr() net.Addr
	Type() ConnType
}

// IGameServerConnector GameServer连接器接口
type IGameServerConnector interface {
	Connect(ctx context.Context, addr string) error
	NotifySessionEvent(ctx context.Context, event *network.SessionEvent) error
	ForwardClientMsg(ctx context.Context, fwdMsg *network.ForwardMessage) error
	ReceiveGsMessage(ctx context.Context) (*network.ForwardMessage, error)
	Close() error
}
