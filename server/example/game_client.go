package main

import (
	"context"
	"fmt"
	"postapocgame/server/internal/actor"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
)

// GameClient 游戏客户端（使用Actor）
type GameClient struct {
	playerID    string
	tcpClient   network.ITCPClient
	codec       *network.Codec
	actorMgr    actor.IActorManager
	actorCtx    actor.IActorContext
	gatewayAddr string
}

func NewGameClient(playerID string, gatewayAddr string, actorMgr actor.IActorManager) *GameClient {
	return &GameClient{
		playerID:    playerID,
		codec:       network.DefaultCodec(),
		actorMgr:    actorMgr,
		gatewayAddr: gatewayAddr,
	}
}

// Start 连接服务器
func (c *GameClient) Start(ctx context.Context) error {
	// 创建网络消息处理器（转发到Actor）
	handler := &NetworkMessageHandler{
		client: c,
	}

	c.tcpClient = network.NewTCPClient(
		network.WithTCPClientOptionNetworkMessageHandler(handler),
		network.WithTCPClientOptionOnDisConn(func(conn network.IConnection) {
			log.Infof("dis connect gateway")
		}),
		network.WithTCPClientOptionOnConn(func(conn network.IConnection) {
			log.Warnf("connect gateway")
		}),
	)

	log.Infof("[%s] 🔌 正在连接到网关 %s...\n", c.playerID, c.gatewayAddr)
	if err := c.tcpClient.Connect(ctx, c.gatewayAddr); err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	// 获取或创建Actor
	actorCtx, err := c.actorMgr.GetOrCreateActor(c.playerID)
	if err != nil {
		return customerr.Wrap(err)
	}
	c.actorCtx = actorCtx

	// 设置客户端引用到Actor数据
	c.actorCtx.SetData(c)

	log.Infof("[%s] ✅ 成功连接到网关!\n", c.playerID)
	return nil
}

// SendMessage 发送消息
func (c *GameClient) SendMessage(msgId uint16, data []byte) error {
	bytes, err := c.codec.EncodeClientMessageWithJSON(msgId, data)
	if err != nil {
		return customerr.Wrap(err)
	}

	conn := c.tcpClient.GetConnection()
	if conn == nil {
		return fmt.Errorf("未连接到服务器")
	}

	return conn.SendMessage(&network.Message{
		Type:    network.MsgTypeClient,
		Payload: bytes,
	})
}

// QueryRoles 查询角色列表
func (c *GameClient) QueryRoles() error {
	log.Infof("[%s] 查询角色列表中...\n", c.playerID)
	if err := c.SendMessage(uint16(protocol.C2SProtocol_C2SQueryRoles), []byte{}); err != nil {
		return err
	}
	return nil
}

// Close 关闭客户端
func (c *GameClient) Close() {
	if c.tcpClient != nil {
		_ = c.tcpClient.Close()
	}
	if c.actorMgr != nil && c.playerID != "" {
		_ = c.actorMgr.RemoveActor(c.playerID)
	}
}

// GetPlayerID 获取玩家ID
func (c *GameClient) GetPlayerID() string {
	return c.playerID
}
