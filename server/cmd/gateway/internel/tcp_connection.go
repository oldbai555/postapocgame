/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

var (
	ErrFrameTooLarge = errors.New("frame too large")
	ErrInvalidFrame  = errors.New("invalid frame")
)

// TCPConnection TCP连接实现
type TCPConnection struct {
	conn         net.Conn
	maxFrameSize int
}

// NewTCPConnection 创建TCP连接
func NewTCPConnection(conn net.Conn, maxFrameSize int) *TCPConnection {
	return &TCPConnection{
		conn:         conn,
		maxFrameSize: maxFrameSize,
	}
}

// Read 读取帧数据
// 帧格式: [4字节长度][消息体]
func (tc *TCPConnection) Read(ctx context.Context) ([]byte, error) {
	// 读取帧头(4字节长度)
	header := make([]byte, 4)
	if _, err := io.ReadFull(tc.conn, header); err != nil {
		return nil, err
	}

	// 解析长度
	length := binary.BigEndian.Uint32(header)
	if length > uint32(tc.maxFrameSize) {
		return nil, ErrFrameTooLarge
	}

	if length == 0 {
		return nil, ErrInvalidFrame
	}

	// 读取消息体
	body := make([]byte, length)
	if _, err := io.ReadFull(tc.conn, body); err != nil {
		return nil, err
	}

	return body, nil
}

// Write 写入帧数据
func (tc *TCPConnection) Write(data []byte) error {
	if len(data) > tc.maxFrameSize {
		return ErrFrameTooLarge
	}

	// 构造帧: [4字节长度][消息体]
	frame := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(frame[0:4], uint32(len(data)))
	copy(frame[4:], data)

	_, err := tc.conn.Write(frame)
	return err
}

// Close 关闭连接
func (tc *TCPConnection) Close() error {
	return tc.conn.Close()
}

// RemoteAddr 远程地址
func (tc *TCPConnection) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

// Type 连接类型
func (tc *TCPConnection) Type() ConnType {
	return ConnTypeTCP
}
