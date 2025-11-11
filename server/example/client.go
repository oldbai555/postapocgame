package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"postapocgame/server/internal/network"
	"postapocgame/server/internal/protocol"
	"postapocgame/server/pkg/customerr"
	"postapocgame/server/pkg/log"
	"postapocgame/server/pkg/tool"
	"time"
)

const GatewayAddr = "0.0.0.0:1011"

type Client struct {
	tcpClient    *network.TCPClient
	reconnectKey string
	codec        *network.Codec
}

func NewClient() *Client {
	config := &network.TCPClientConfig{
		ConnectTimeout:  5 * time.Second,
		EnableReconnect: true,
		ReconnectConfig: network.DefaultReconnectConfig(),
	}
	handler := &MessageHandler{
		codec: network.DefaultCodec(),
	}
	client := &Client{
		tcpClient: network.NewTCPClient(config, handler),
		codec:     network.DefaultCodec(),
	}
	return client
}

// Start 连接服务器
func (c *Client) Start(ctx context.Context) error {
	fmt.Printf("🔌 正在连接到网关 %s...\n", GatewayAddr)
	if err := c.tcpClient.Connect(ctx, GatewayAddr); err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	fmt.Println("✅ 成功连接到网关!")
	return nil
}

// SendMessage 发送消息
func (c *Client) SendMessage(msgId uint16, data []byte) error {
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

func (c *Client) QueryRoles() error {
	fmt.Println("\n[1] 查询角色列表中...")
	if err := c.SendMessage(protocol.C2S_QueryRoles, []byte{}); err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() {
	_ = c.tcpClient.Close()
}

type MessageHandler struct {
	codec *network.Codec
}

func (h *MessageHandler) HandleMessage(_ context.Context, conn network.IConnection, msg *network.Message) error {
	codec := network.DefaultCodec()

	// 解出 ClientMessage
	clientMsg, err := codec.DecodeClientMessage(msg.Payload)
	if err != nil {
		fmt.Printf("❌ 解析消息失败: %v\n", err)
		return err
	}

	msgId := clientMsg.MsgId
	data := clientMsg.Data

	// 根据不同消息类型进行分类展示
	switch msgId {
	case protocol.S2C_Error:
		var errResp protocol.ErrorResponse
		if err := tool.JsonUnmarshal(data, &errResp); err == nil {
			fmt.Printf("\n⚠️ 服务器错误: %s\n> ", errResp.ErrMsg)
		}
	case protocol.S2C_RoleList:
		var resp protocol.RoleListResponse
		err := tool.JsonUnmarshal(data, &resp)
		if err != nil {
			return customerr.Wrap(err)
		}
		for i, role := range resp.Roles {
			fmt.Printf(" [%d] 角色ID: %d, 名字: %s, 职业: %d, 等级: %d\n", i+1, role.RoleId, role.Name, role.Job, role.Level)
		}
		fmt.Printf("🎮 进入游戏: RoleID=%d\n", 10001)
		req := protocol.SelectRoleRequest{RoleId: 10001}
		reqData, err := tool.JsonMarshal(req)
		if err != nil {
			return customerr.Wrap(err)
		}
		bytes, err := h.codec.EncodeClientMessageWithJSON(protocol.C2S_EnterGame, reqData)
		if err != nil {
			return customerr.Wrap(err)
		}
		if err := conn.SendMessage(&network.Message{
			Type:    network.MsgTypeClient,
			Payload: bytes,
		}); err != nil {
			return err
		}
	default:
		fmt.Printf("\n📨 未知消息: MsgID=%d, Len=%d\n> ", msgId, len(data))
	}
	return nil
}

func main() {
	log.InitLogger(log.WithAppName("example_client"), log.WithScreen(true), log.WithPath(tool.GetCurDir()+"log"))
	defer log.Flush()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("===============================")
	fmt.Println("   游戏客户端测试程序 (使用TCPClient)")
	fmt.Println("===============================")

	client := NewClient()
	defer client.Close()

	if err := client.Start(ctx); err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	time.Sleep(300 * time.Millisecond)

	err := client.QueryRoles()
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}
	fmt.Println("\n✅ 测试完成，按 Enter 退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
