/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package clientnet

import (
	"net"
	"postapocgame/server/internal/network"
)

type ConnectionAdapter struct {
	conn network.IConnection
}

func (ca *ConnectionAdapter) Close() error {
	return ca.conn.Close()
}

func (ca *ConnectionAdapter) RemoteAddr() net.Addr {
	return ca.conn.RemoteAddr()
}

func (ca *ConnectionAdapter) Type() ConnType {
	// 根据连接类型判断
	switch ca.conn.(type) {
	case *network.WebSocketConnection:
		return ConnTypeWebSocket
	case *network.TCPConnection:
		return ConnTypeTCP
	default:
		return ConnTypeTCP
	}
}
