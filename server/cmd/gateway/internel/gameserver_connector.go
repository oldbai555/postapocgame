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
	"fmt"
	"io"
	"net"
	"postapocgame/server/pkg/log"
	"sync"
	"time"
)

var (
	ErrNotConnected = errors.New("not connected to game server")
)

// GameServerConnector GameServer连接器实现(TCP长连接)
type GameServerConnector struct {
	addr      string
	conn      net.Conn
	config    *Config
	mu        sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	recvChan  chan *FramedMessage
	connected bool
}

// NewGameServerConnector 创建GameServer连接器
func NewGameServerConnector(addr string, config *Config) *GameServerConnector {
	return &GameServerConnector{
		addr:      addr,
		config:    config,
		stopChan:  make(chan struct{}),
		recvChan:  make(chan *FramedMessage, 1024),
		connected: false,
	}
}

// Connect 连接到GameServer
func (gsc *GameServerConnector) Connect(ctx context.Context, addr string) error {
	gsc.mu.Lock()
	defer gsc.mu.Unlock()

	if gsc.connected {
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to game server failed: %w", err)
	}

	gsc.conn = conn
	gsc.addr = addr
	gsc.connected = true

	// 启动接收协程
	gsc.wg.Add(1)
	go gsc.receiveLoop(ctx)

	log.Infof("connected to game server: %s", addr)
	return nil
}

// NotifySessionEvent 通知会话事件
func (gsc *GameServerConnector) NotifySessionEvent(ctx context.Context, event *SessionEvent) error {
	gsc.mu.RLock()
	defer gsc.mu.RUnlock()

	if !gsc.connected {
		return ErrNotConnected
	}

	// 编码事件
	// 帧格式: [1字节类型=0x01表示事件][事件数据]
	// 事件数据格式: [1字节事件类型][sessionID长度(2字节)][sessionID][userID长度(2字节)][userID]

	sessionIDBytes := []byte(event.SessionID)
	userIDBytes := []byte(event.UserID)

	size := 1 + 1 + 2 + len(sessionIDBytes) + 2 + len(userIDBytes)
	data := make([]byte, size)

	offset := 0
	data[offset] = 0x01 // 消息类型: 事件
	offset++
	data[offset] = byte(event.Type) // 事件类型
	offset++
	binary.BigEndian.PutUint16(data[offset:], uint16(len(sessionIDBytes)))
	offset += 2
	copy(data[offset:], sessionIDBytes)
	offset += len(sessionIDBytes)
	binary.BigEndian.PutUint16(data[offset:], uint16(len(userIDBytes)))
	offset += 2
	copy(data[offset:], userIDBytes)

	return gsc.writeFrame(data)
}

// ForwardMessage 转发消息
func (gsc *GameServerConnector) ForwardMessage(ctx context.Context, msg *FramedMessage) error {
	gsc.mu.RLock()
	defer gsc.mu.RUnlock()

	if !gsc.connected {
		return ErrNotConnected
	}

	// 编码消息
	// 帧格式: [1字节类型=0x02表示消息][sessionID长度(2字节)][sessionID][payload]

	sessionIDBytes := []byte(msg.SessionID)
	size := 1 + 2 + len(sessionIDBytes) + len(msg.Payload)
	data := make([]byte, size)

	offset := 0
	data[offset] = 0x02 // 消息类型: 转发消息
	offset++
	binary.BigEndian.PutUint16(data[offset:], uint16(len(sessionIDBytes)))
	offset += 2
	copy(data[offset:], sessionIDBytes)
	offset += len(sessionIDBytes)
	copy(data[offset:], msg.Payload)

	return gsc.writeFrame(data)
}

// ReceiveMessage 接收来自GameServer的消息
func (gsc *GameServerConnector) ReceiveMessage(ctx context.Context) (*FramedMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-gsc.recvChan:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	}
}

// receiveLoop 接收循环
func (gsc *GameServerConnector) receiveLoop(ctx context.Context) {
	defer gsc.wg.Done()
	defer gsc.disconnect()

	for {
		select {
		case <-ctx.Done():
			return
		case <-gsc.stopChan:
			return
		default:
		}

		data, err := gsc.readFrame()
		if err != nil {
			log.Infof("read frame from game server failed: %v", err)
			return
		}

		// 解析消息
		if len(data) < 3 {
			log.Infof("invalid frame from game server")
			continue
		}

		msgType := data[0]
		if msgType != 0x02 { // 只处理转发消息
			continue
		}

		// 解析sessionID和payload
		offset := 1
		sessionIDLen := binary.BigEndian.Uint16(data[offset:])
		offset += 2
		if offset+int(sessionIDLen) > len(data) {
			log.Infof("invalid frame: sessionID overflow")
			continue
		}

		sessionID := string(data[offset : offset+int(sessionIDLen)])
		offset += int(sessionIDLen)
		payload := data[offset:]

		msg := &FramedMessage{
			SessionID: sessionID,
			Payload:   payload,
		}

		select {
		case gsc.recvChan <- msg:
		default:
			log.Infof("receive channel full, drop message")
		}
	}
}

// writeFrame 写入帧
func (gsc *GameServerConnector) writeFrame(data []byte) error {
	// 帧格式: [4字节长度][消息体]
	frame := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(frame[0:4], uint32(len(data)))
	copy(frame[4:], data)

	_, err := gsc.conn.Write(frame)
	return err
}

// readFrame 读取帧
func (gsc *GameServerConnector) readFrame() ([]byte, error) {
	// 读取帧头
	header := make([]byte, 4)
	if _, err := io.ReadFull(gsc.conn, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header)
	if length > uint32(gsc.config.MaxFrameSize) {
		return nil, ErrFrameTooLarge
	}

	// 读取消息体
	body := make([]byte, length)
	if _, err := io.ReadFull(gsc.conn, body); err != nil {
		return nil, err
	}

	return body, nil
}

// disconnect 断开连接
func (gsc *GameServerConnector) disconnect() {
	gsc.mu.Lock()
	defer gsc.mu.Unlock()

	if !gsc.connected {
		return
	}

	if gsc.conn != nil {
		gsc.conn.Close()
		gsc.conn = nil
	}

	gsc.connected = false
	close(gsc.recvChan)

	log.Infof("disconnected from game server: %s", gsc.addr)
}

// Close 关闭连接
func (gsc *GameServerConnector) Close() error {
	close(gsc.stopChan)
	gsc.wg.Wait()
	gsc.disconnect()
	return nil
}
