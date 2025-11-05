/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc:
**/

package internel

import (
	"context"
	"fmt"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/routine"
	"sync"
)

// ConnectionHandler 连接处理器
type ConnectionHandler struct {
	conn          IConnection
	session       *Session
	sessionMgr    *SessionManager
	gsConn        IGameServerConnector
	authenticator IAuthenticator
	config        *Config
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewConnectionHandler 创建连接处理器
func NewConnectionHandler(
	conn IConnection,
	sessionMgr *SessionManager,
	gsConn IGameServerConnector,
	authenticator IAuthenticator,
	config *Config,
) *ConnectionHandler {
	return &ConnectionHandler{
		conn:          conn,
		sessionMgr:    sessionMgr,
		gsConn:        gsConn,
		authenticator: authenticator,
		config:        config,
		stopChan:      make(chan struct{}),
	}
}

// Handle 处理连接
func (ch *ConnectionHandler) Handle(ctx context.Context) error {
	// 创建会话
	session, err := ch.sessionMgr.CreateSession(ch.conn)
	if err != nil {
		ch.conn.Close()
		return fmt.Errorf("create session failed: %w", err)
	}

	ch.session = session
	log.Infof("new connection: sessionID=%s, addr=%s, type=%d",
		session.ID, session.Addr.String(), session.ConnType)

	// 启动读写协程
	ch.wg.Add(2)
	routine.GoV2(func() error {
		ch.readLoop(ctx)
		return nil
	})
	routine.GoV2(func() error {
		ch.writeLoop(ctx)
		return nil
	})

	// 等待停止
	<-ch.stopChan
	ch.wg.Wait()

	// 清理
	err = ch.sessionMgr.CloseSession(session.ID)
	if err != nil {
		log.Errorf("err:%v", err)
	}
	err = ch.conn.Close()
	if err != nil {
		log.Errorf("err:%v", err)
	}

	log.Infof("connection closed: sessionID=%s", session.ID)
	return nil
}

// readLoop 读取循环
func (ch *ConnectionHandler) readLoop(ctx context.Context) {
	defer ch.wg.Done()
	defer ch.stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ch.stopChan:
			return
		default:
		}

		// 读取消息
		data, err := ch.conn.Read(ctx)
		if err != nil {
			log.Infof("read error: sessionID=%s, err=%v", ch.session.ID, err)
			return
		}

		// 更新活跃时间
		ch.sessionMgr.UpdateActivity(ch.session.ID)

		// 简单地转发所有消息到GameServer

		// 转发到GameServer
		msg := &FramedMessage{
			SessionID: ch.session.ID,
			Payload:   data,
		}

		if err := ch.gsConn.ForwardMessage(ctx, msg); err != nil {
			log.Warnf("forward message failed: sessionID=%s, err=%v", ch.session.ID, err)
			// 不中断连接,继续处理
		}
	}
}

// writeLoop 写入循环
func (ch *ConnectionHandler) writeLoop(ctx context.Context) {
	defer ch.wg.Done()
	defer ch.stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ch.stopChan:
			return
		case data, ok := <-ch.session.SendChan:
			if !ok {
				// 通道已关闭
				return
			}

			if err := ch.conn.Write(data); err != nil {
				log.Infof("write error: sessionID=%s, err=%v", ch.session.ID, err)
				return
			}
		}
	}
}

// stop 停止处理器
func (ch *ConnectionHandler) stop() {
	select {
	case <-ch.stopChan:
		// 已经关闭
	default:
		close(ch.stopChan)
	}
}
