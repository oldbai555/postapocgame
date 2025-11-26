/**
 * @Author: zjj
 * @Date: 2025/11/26
 * @Desc:
**/

package gameserverlink

import (
	"context"
	"errors"
	"postapocgame/server/internal/network"
	"postapocgame/server/pkg/log"
	"sync"
)

// GameMessageHandler 消息处理器
type GameMessageHandler struct {
	codec    *network.Codec
	recvChan chan *network.ForwardMessage
	mu       sync.Mutex
	closed   bool
}

func NewGameMessageHandler() *GameMessageHandler {
	return &GameMessageHandler{
		codec:    network.DefaultCodec(),
		recvChan: make(chan *network.ForwardMessage, 1024),
	}
}

func (h *GameMessageHandler) HandleMessage(_ context.Context, _ network.IConnection, msg *network.Message) error {
	if msg.Type != network.MsgTypeClient {
		log.Debugf("ignore message type: %d", msg.Type)
		return nil
	}

	forwardMsg, err := h.codec.DecodeForwardMessage(msg.Payload)
	if err != nil {
		log.Errorf("DecodeForwardMessage failed: %v", err)
		return err
	}

	// 非阻塞发送到接收通道
	select {
	case h.recvChan <- forwardMsg:
		return nil
	default:
		log.Warnf("receive channel full, drop message for session: %s", forwardMsg.SessionId)
		network.PutForwardMessage(forwardMsg)
		return nil
	}
}

func (h *GameMessageHandler) ReceiveMessage(ctx context.Context) (*network.ForwardMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-h.recvChan:
		if !ok {
			return nil, errors.New("receive channel closed")
		}
		return msg, nil
	}
}

func (h *GameMessageHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.closed {
		close(h.recvChan)
		h.closed = true
	}
}
