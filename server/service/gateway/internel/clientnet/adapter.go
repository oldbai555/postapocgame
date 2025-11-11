/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package clientnet

import (
	"context"
	"net"
	"postapocgame/server/internal/network"
)

type ConnectionAdapter struct {
	conn network.IConnection
}

func (ca *ConnectionAdapter) Read(ctx context.Context) ([]byte, error) {
	msg, err := ca.conn.ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Payload, nil
}

func (ca *ConnectionAdapter) Write(data []byte) error {
	message := network.GetMessage()
	message.Type = network.MsgTypeClient
	message.Payload = data
	defer network.PutMessage(message)
	return ca.conn.SendMessage(message)
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
