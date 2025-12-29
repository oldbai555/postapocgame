package hub

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"postapocgame/admin-server/internal/model"
	"postapocgame/admin-server/internal/repository"
)

const (
	// 写超时时间
	writeWait = 10 * time.Second

	// 读超时时间（ping 间隔）
	pongWait = 60 * time.Second

	// ping 间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	Hub          *ChatHub
	Conn         *websocket.Conn
	Send         chan []byte
	UserID       uint64
	Username     string
	RoomID       string // 当前所在的聊天室ID
	ConnectionID string // WebSocket 连接 ID
}

// ChatHub 管理所有 WebSocket 连接和消息广播
type ChatHub struct {
	// 注册的客户端连接（按用户ID分组）
	clients map[uint64]*Client

	// 按房间ID分组的客户端连接
	rooms map[string]map[uint64]*Client

	// 广播消息到所有客户端
	broadcast chan []byte

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// Repository 用于数据库操作
	onlineUserRepo repository.ChatOnlineUserRepository

	mu sync.RWMutex
}

// Register 返回注册通道（供外部使用）
func (h *ChatHub) Register() chan<- *Client {
	return h.register
}

// NewChatHub 创建新的 ChatHub
func NewChatHub(onlineUserRepo repository.ChatOnlineUserRepository) *ChatHub {
	return &ChatHub{
		clients:        make(map[uint64]*Client),
		rooms:          make(map[string]map[uint64]*Client),
		broadcast:      make(chan []byte, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		onlineUserRepo: onlineUserRepo,
	}
}

// Run 启动 Hub，处理注册、注销和广播
func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			if client.RoomID != "" {
				if h.rooms[client.RoomID] == nil {
					h.rooms[client.RoomID] = make(map[uint64]*Client)
				}
				h.rooms[client.RoomID][client.UserID] = client
			}
			h.mu.Unlock()
			logx.Infof("客户端注册: UserID=%d, Username=%s, RoomID=%s, ConnectionID=%s", client.UserID, client.Username, client.RoomID, client.ConnectionID)

			// 保存在线用户到数据库
			if h.onlineUserRepo != nil {
				go func() {
					ctx := context.Background()
					now := time.Now().Unix()
					onlineUser := &model.ChatOnlineUser{
						UserId:       client.UserID,
						ConnectionId: client.ConnectionID,
						IpAddress:    "", // 可以从 client.Conn.RemoteAddr() 获取
						UserAgent:    "",
						LastActiveAt: now,
						CreatedAt:    now,
						UpdatedAt:    now,
					}
					if err := h.onlineUserRepo.Create(ctx, onlineUser); err != nil {
						logx.Errorf("保存在线用户失败: %v", err)
					}
				}()
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			if client.RoomID != "" && h.rooms[client.RoomID] != nil {
				delete(h.rooms[client.RoomID], client.UserID)
				if len(h.rooms[client.RoomID]) == 0 {
					delete(h.rooms, client.RoomID)
				}
			}
			h.mu.Unlock()
			logx.Infof("客户端注销: UserID=%d, Username=%s, ConnectionID=%s", client.UserID, client.Username, client.ConnectionID)

			// 从数据库删除在线用户
			if h.onlineUserRepo != nil && client.ConnectionID != "" {
				go func() {
					ctx := context.Background()
					if err := h.onlineUserRepo.DeleteByConnectionID(ctx, client.ConnectionID); err != nil {
						logx.Errorf("删除在线用户失败: %v", err)
					}
				}()
			}

		case message := <-h.broadcast:
			// 广播消息到所有客户端
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToRoom 向指定房间的所有客户端广播消息
func (h *ChatHub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[roomID]; ok {
		for _, client := range room {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client.UserID)
				delete(room, client.UserID)
			}
		}
	}
}

// SendToUser 向指定用户发送消息
func (h *ChatHub) SendToUser(userID uint64, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		select {
		case client.Send <- message:
			return true
		default:
			close(client.Send)
			delete(h.clients, client.UserID)
			return false
		}
	}
	return false
}

// GetOnlineUsers 获取所有在线用户ID列表
func (h *ChatHub) GetOnlineUsers() []uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uint64, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// IsUserOnline 检查用户是否在线
func (h *ChatHub) IsUserOnline(userID uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.clients[userID]
	return ok
}

// ReadPump 从 WebSocket 连接读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("WebSocket 读取错误: %v", err)
			}
			break
		}

		// 处理接收到的消息（可以在这里添加消息处理逻辑）
		logx.Infof("收到消息 from UserID=%d: %s", c.UserID, string(message))
	}
}

// WritePump 向 WebSocket 连接写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ChatMessage WebSocket 消息结构
type ChatMessage struct {
	Type      string `json:"type"`      // 消息类型：chat, task_progress, notification, system, join, leave, error
	FromID    uint64 `json:"fromId"`    // 发送者ID
	FromName  string `json:"fromName"`  // 发送者名称
	ToID      uint64 `json:"toId"`      // 接收者ID（0表示群聊）
	RoomID    string `json:"roomId"`    // 聊天室ID
	Content   string `json:"content"`   // 消息内容
	MessageID uint64 `json:"messageId"` // 消息ID
	CreatedAt string `json:"createdAt"` // 创建时间
	// 任务进度相关字段
	TaskID   string `json:"taskId,omitempty"`   // 任务ID
	TaskName string `json:"taskName,omitempty"` // 任务名称
	Progress int    `json:"progress,omitempty"` // 进度百分比
	Status   string `json:"status,omitempty"`   // 任务状态
	// 通知相关字段
	Title string `json:"title,omitempty"` // 通知标题
	Level string `json:"level,omitempty"` // 通知级别：info, success, warning, error
}

// BroadcastChatMessage 广播聊天消息
func (h *ChatHub) BroadcastChatMessage(msg *ChatMessage) error {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 优先判断私聊：ToID > 0 表示私聊消息
	if msg.ToID > 0 {
		// 私聊：发送给发送者和接收者
		// 发送给接收者
		h.SendToUser(msg.ToID, messageBytes)
		// 发送给发送者（如果发送者在线）
		if msg.FromID > 0 && msg.FromID != msg.ToID {
			h.SendToUser(msg.FromID, messageBytes)
		}
	} else if msg.RoomID != "" {
		// 群聊：向房间内所有用户广播
		h.BroadcastToRoom(msg.RoomID, messageBytes)
	} else {
		// 广播给所有人
		h.broadcast <- messageBytes
	}

	return nil
}
