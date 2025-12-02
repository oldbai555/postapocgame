# Gateway Clean Architecture 重构文档

更新时间：2025-01-XX  
责任人：开发团队

## 1. 文档目的

本文档旨在将 `server/service/gateway` 按照 Clean Architecture（清洁架构）原则进行重构，实现业务逻辑与框架解耦，提高代码可测试性、可维护性和可扩展性。

## 2. 当前架构问题分析

### 2.1 依赖方向混乱

**问题描述：**
- 业务逻辑层直接依赖 `network`、`gameserverlink` 等框架层
- Session 管理、消息转发、连接处理等逻辑混在一起
- 内层（业务逻辑）依赖外层（框架），违反了依赖倒置原则

**典型示例：**

```go
// clientnet/handler.go - 直接依赖网络和 GameServer 连接器
func (h *ClientHandler) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
    // ❌ 直接处理网络消息
    session := h.getOrCreateSession(conn)
    
    // ❌ 直接调用 GameServer 连接器
    return h.GsConnector.ForwardClientMsg(context.Background(), &network.ForwardMessage{
        SessionId: session.Id,
        Payload:   msg.Payload,
    })
}

// clientnet/session_mgr.go - 直接依赖网络和 GameServer 连接器
func (sm *SessionManager) CreateSession(conn IConnection) (*Session, error) {
    // ❌ 直接通知 GameServer
    if err := sm.gsConn.NotifySessionEvent(ctx, ev); err != nil {
        return nil, err
    }
}
```

### 2.2 业务逻辑与框架耦合

**问题描述：**
- Session 管理、消息转发、连接处理等业务逻辑混在框架代码中
- 业务逻辑无法独立测试，必须启动完整的网络服务和 GameServer 连接
- 系统之间通过直接调用而非接口交互

### 2.3 职责不清晰

**问题描述：**
- `ClientHandler` 既处理网络消息，又管理 Session，还转发消息
- `SessionManager` 既管理 Session，又通知 GameServer
- 没有明确的分层，职责混乱

### 2.4 接口适配层缺失

**问题描述：**
- 没有明确的 Adapter 层来适配网络框架
- 协议编解码、消息转换等逻辑混在业务代码中
- 无法轻松替换底层网络实现

## 3. Clean Architecture 分层设计

### 3.1 分层结构

```
┌─────────────────────────────────────────────────────────┐
│  Frameworks & Drivers (框架层)                          │
│  - 网络层 (network)                                      │
│  - TCP/WebSocket 服务器                                 │
│  - 消息编解码器                                          │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Interface Adapters (接口适配层)                         │
│  - Controllers: 消息处理器                               │
│  - Presenters: 消息构建器                               │
│  - Gateways: 网络适配器、GameServer 适配器               │
│  - Codec Adapters: 编解码适配器                         │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Use Cases (用例层)                                      │
│  - 业务用例: CreateSession, ForwardMessage, CloseSession 等 │
│  - 业务规则: Session 超时、限流、消息路由等              │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Entities (实体层)                                       │
│  - 业务实体: Session, Connection 等                      │
│  - 值对象: SessionID, Message 等                        │
└─────────────────────────────────────────────────────────┘
```

### 3.2 目录结构设计

```
server/service/gateway/
├── internel/
│   ├── domain/                    # Entities 层
│   │   ├── session.go             # Session 实体
│   │   ├── connection.go          # Connection 实体
│   │   ├── message.go             # Message 值对象
│   │   └── ...
│   │
│   ├── usecase/                   # Use Cases 层
│   │   ├── session/               # Session 用例
│   │   │   ├── create_session.go
│   │   │   ├── close_session.go
│   │   │   └── update_activity.go
│   │   ├── message/                # 消息转发用例
│   │   │   ├── forward_to_gameserver.go
│   │   │   ├── forward_to_client.go
│   │   │   └── route_message.go
│   │   └── ...
│   │
│   ├── adapter/                   # Interface Adapters 层
│   │   ├── controller/           # 消息控制器
│   │   │   ├── client_message_controller.go
│   │   │   └── gameserver_message_controller.go
│   │   ├── gateway/              # 网络和 GameServer 适配器
│   │   │   ├── network_gateway.go
│   │   │   ├── gameserver_gateway.go
│   │   │   └── codec_gateway.go
│   │   └── ...
│   │
│   ├── infrastructure/           # Frameworks & Drivers 层
│   │   ├── network/               # 网络适配
│   │   ├── tcp/                   # TCP 服务器适配
│   │   ├── websocket/             # WebSocket 服务器适配
│   │   └── ...
│   │
│   └── ... (保留现有目录用于过渡)
```

## 4. 重构方案

### 4.1 阶段一：Entities 层重构

**目标：** 提取纯业务实体，移除所有框架依赖

#### 4.1.1 创建 Domain 实体

**目录：** `internel/domain/`

**示例：Session 实体**

```go
// domain/session.go
package domain

import "time"

// Session 会话实体（纯业务对象，不依赖任何框架）
type Session struct {
    ID         string
    RemoteAddr string
    ConnType   ConnType
    State      SessionState
    UserID     string
    CreatedAt  time.Time
    LastActive time.Time
}

// IsActive 判断会话是否活跃
func (s *Session) IsActive() bool {
    return s.State == SessionStateConnected
}

// UpdateActivity 更新活跃时间
func (s *Session) UpdateActivity(now time.Time) {
    s.LastActive = now
}

// IsTimeout 判断是否超时
func (s *Session) IsTimeout(timeout time.Duration, now time.Time) bool {
    return now.Sub(s.LastActive) > timeout
}
```

**示例：Message 值对象**

```go
// domain/message.go
package domain

// Message 消息值对象
type Message struct {
    Type    MessageType
    Payload []byte
    SessionID string
}

// MessageType 消息类型
type MessageType uint8

const (
    MessageTypeClient MessageType = iota
    MessageTypeSessionEvent
)
```

#### 4.1.2 定义 Repository 接口

**目录：** `internel/domain/repository/`

```go
// domain/repository/session_repository.go
package repository

import "postapocgame/server/service/gateway/internel/domain"

// SessionRepository 会话数据访问接口（定义在 domain 层）
type SessionRepository interface {
    Create(session *domain.Session) error
    GetByID(sessionID string) (*domain.Session, error)
    Update(session *domain.Session) error
    Delete(sessionID string) error
    GetAll() ([]*domain.Session, error)
    Count() int
}
```

### 4.2 阶段二：Use Cases 层重构

**目标：** 实现业务用例，依赖 Entities 和 Repository 接口

#### 4.2.1 创建 Use Case

**目录：** `internel/usecase/`

**示例：CreateSession Use Case**

```go
// usecase/session/create_session.go
package session

import (
    "context"
    "postapocgame/server/service/gateway/internel/domain"
    "postapocgame/server/service/gateway/internel/domain/repository"
    "postapocgame/server/service/gateway/internel/usecase/interfaces"
)

// CreateSessionUseCase 创建会话用例
type CreateSessionUseCase struct {
    sessionRepo    repository.SessionRepository
    gameServerRPC  interfaces.GameServerRPC
    eventPublisher interfaces.EventPublisher
}

func NewCreateSessionUseCase(
    sessionRepo repository.SessionRepository,
    gameServerRPC interfaces.GameServerRPC,
    eventPublisher interfaces.EventPublisher,
) *CreateSessionUseCase {
    return &CreateSessionUseCase{
        sessionRepo:   sessionRepo,
        gameServerRPC: gameServerRPC,
        eventPublisher: eventPublisher,
    }
}

// Execute 执行创建会话用例
func (uc *CreateSessionUseCase) Execute(ctx context.Context, remoteAddr string, connType domain.ConnType) (*domain.Session, error) {
    // 1. 检查会话数量限制（业务规则）
    if uc.sessionRepo.Count() >= uc.maxSessions {
        return nil, ErrMaxSessionsReached
    }
    
    // 2. 创建会话实体（纯业务逻辑）
    session := &domain.Session{
        ID:         generateSessionID(),
        RemoteAddr: remoteAddr,
        ConnType:   connType,
        State:      domain.SessionStateConnected,
        CreatedAt:  getCurrentTime(),
        LastActive: getCurrentTime(),
    }
    
    // 3. 保存会话
    if err := uc.sessionRepo.Create(session); err != nil {
        return nil, err
    }
    
    // 4. 通知 GameServer（通过接口）
    if err := uc.gameServerRPC.NotifySessionCreated(ctx, session.ID); err != nil {
        // 如果通知失败，回滚会话创建
        uc.sessionRepo.Delete(session.ID)
        return nil, err
    }
    
    // 5. 发布事件
    uc.eventPublisher.PublishSessionCreated(ctx, session)
    
    return session, nil
}
```

**示例：ForwardMessage Use Case**

```go
// usecase/message/forward_to_gameserver.go
package message

import (
    "context"
    "postapocgame/server/service/gateway/internel/domain"
    "postapocgame/server/service/gateway/internel/domain/repository"
    "postapocgame/server/service/gateway/internel/usecase/interfaces"
)

// ForwardToGameServerUseCase 转发消息到 GameServer 用例
type ForwardToGameServerUseCase struct {
    sessionRepo   repository.SessionRepository
    gameServerRPC interfaces.GameServerRPC
}

func NewForwardToGameServerUseCase(
    sessionRepo repository.SessionRepository,
    gameServerRPC interfaces.GameServerRPC,
) *ForwardToGameServerUseCase {
    return &ForwardToGameServerUseCase{
        sessionRepo:   sessionRepo,
        gameServerRPC: gameServerRPC,
    }
}

// Execute 执行转发消息用例
func (uc *ForwardToGameServerUseCase) Execute(ctx context.Context, sessionID string, message *domain.Message) error {
    // 1. 验证会话存在
    session, err := uc.sessionRepo.GetByID(sessionID)
    if err != nil {
        return err
    }
    
    if !session.IsActive() {
        return ErrSessionNotActive
    }
    
    // 2. 更新会话活跃时间（业务规则）
    session.UpdateActivity(getCurrentTime())
    uc.sessionRepo.Update(session)
    
    // 3. 转发消息到 GameServer（通过接口）
    return uc.gameServerRPC.ForwardMessage(ctx, sessionID, message.Payload)
}
```

#### 4.2.2 定义 Use Case 依赖接口

**目录：** `internel/usecase/interfaces/`

```go
// usecase/interfaces/gameserver_rpc.go
package interfaces

import "context"

// GameServerRPC GameServer RPC 接口（Use Case 层定义）
type GameServerRPC interface {
    NotifySessionCreated(ctx context.Context, sessionID string) error
    NotifySessionClosed(ctx context.Context, sessionID string, userID string) error
    ForwardMessage(ctx context.Context, sessionID string, payload []byte) error
}
```

### 4.3 阶段三：Interface Adapters 层重构

**目标：** 实现消息处理、网络适配、GameServer 适配

#### 4.3.1 Controllers（消息控制器）

**目录：** `internel/adapter/controller/`

```go
// adapter/controller/client_message_controller.go
package controller

import (
    "context"
    "postapocgame/server/internal/network"
    "postapocgame/server/service/gateway/internel/adapter/gateway"
    "postapocgame/server/service/gateway/internel/domain"
    "postapocgame/server/service/gateway/internel/usecase/message"
    "postapocgame/server/service/gateway/internel/usecase/session"
)

// ClientMessageController 客户端消息控制器
type ClientMessageController struct {
    createSessionUseCase      *session.CreateSessionUseCase
    forwardToGameServerUseCase *message.ForwardToGameServerUseCase
    networkGateway            gateway.NetworkGateway
}

func NewClientMessageController(
    createSessionUseCase *session.CreateSessionUseCase,
    forwardToGameServerUseCase *message.ForwardToGameServerUseCase,
    networkGateway gateway.NetworkGateway,
) *ClientMessageController {
    return &ClientMessageController{
        createSessionUseCase:      createSessionUseCase,
        forwardToGameServerUseCase: forwardToGameServerUseCase,
        networkGateway:            networkGateway,
    }
}

// HandleMessage 处理客户端消息
func (c *ClientMessageController) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
    // 1. 获取或创建会话
    sessionID := c.networkGateway.GetSessionID(conn)
    if sessionID == "" {
        // 创建新会话
        connType := c.networkGateway.GetConnectionType(conn)
        remoteAddr := conn.RemoteAddr().String()
        session, err := c.createSessionUseCase.Execute(ctx, remoteAddr, connType)
        if err != nil {
            return err
        }
        sessionID = session.ID
        c.networkGateway.SetSessionID(conn, sessionID)
    }
    
    // 2. 转换为 Domain 对象
    domainMsg := &domain.Message{
        Type:      domain.MessageTypeClient,
        Payload:   msg.Payload,
        SessionID: sessionID,
    }
    
    // 3. 调用 Use Case 转发消息
    return c.forwardToGameServerUseCase.Execute(ctx, sessionID, domainMsg)
}
```

#### 4.3.2 Gateways（网络和 GameServer 适配器）

**目录：** `internel/adapter/gateway/`

```go
// adapter/gateway/gameserver_gateway.go
package gateway

import (
    "context"
    "postapocgame/server/service/gateway/internel/usecase/interfaces"
    "postapocgame/server/service/gateway/internel/gameserverlink"
)

// GameServerGateway GameServer 适配器（实现 Use Case 层的 GameServerRPC 接口）
type GameServerGateway struct {
    gameClient *gameserverlink.GameClient
}

func NewGameServerGateway(gameClient *gameserverlink.GameClient) interfaces.GameServerRPC {
    return &GameServerGateway{
        gameClient: gameClient,
    }
}

func (g *GameServerGateway) NotifySessionCreated(ctx context.Context, sessionID string) error {
    event := &network.SessionEvent{
        EventType: network.SessionEventNew,
        SessionId: sessionID,
    }
    return g.gameClient.NotifySessionEvent(ctx, event)
}

func (g *GameServerGateway) ForwardMessage(ctx context.Context, sessionID string, payload []byte) error {
    forwardMsg := &network.ForwardMessage{
        SessionId: sessionID,
        Payload:   payload,
    }
    return g.gameClient.ForwardClientMsg(ctx, forwardMsg)
}
```

```go
// adapter/gateway/network_gateway.go
package gateway

import (
    "postapocgame/server/internal/network"
)

// NetworkGateway 网络网关接口（Adapter 层定义）
type NetworkGateway interface {
    GetSessionID(conn network.IConnection) string
    SetSessionID(conn network.IConnection, sessionID string)
    GetConnectionType(conn network.IConnection) domain.ConnType
    SendToClient(conn network.IConnection, data []byte) error
}

// NetworkGatewayImpl 网络网关实现
type NetworkGatewayImpl struct {
    sessionMap map[network.IConnection]string
    mu         sync.RWMutex
}

func NewNetworkGateway() NetworkGateway {
    return &NetworkGatewayImpl{
        sessionMap: make(map[network.IConnection]string),
    }
}

func (g *NetworkGatewayImpl) GetSessionID(conn network.IConnection) string {
    g.mu.RLock()
    defer g.mu.RUnlock()
    return g.sessionMap[conn]
}

func (g *NetworkGatewayImpl) SetSessionID(conn network.IConnection, sessionID string) {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.sessionMap[conn] = sessionID
}
```

#### 4.3.3 Session Repository 实现

**目录：** `internel/adapter/gateway/`

```go
// adapter/gateway/session_repository.go
package gateway

import (
    "postapocgame/server/service/gateway/internel/domain"
    "postapocgame/server/service/gateway/internel/domain/repository"
    "sync"
)

// SessionRepositoryImpl 会话仓库实现（实现 domain 层的 Repository 接口）
type SessionRepositoryImpl struct {
    sessions map[string]*domain.Session
    mu       sync.RWMutex
}

func NewSessionRepository() repository.SessionRepository {
    return &SessionRepositoryImpl{
        sessions: make(map[string]*domain.Session),
    }
}

func (r *SessionRepositoryImpl) Create(session *domain.Session) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.sessions[session.ID] = session
    return nil
}

func (r *SessionRepositoryImpl) GetByID(sessionID string) (*domain.Session, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    session, ok := r.sessions[sessionID]
    if !ok {
        return nil, ErrSessionNotFound
    }
    return session, nil
}
```

### 4.4 阶段四：Infrastructure 层重构

**目标：** 封装框架调用，提供统一接口

#### 4.4.1 Network Infrastructure

**目录：** `internel/infrastructure/network/`

```go
// infrastructure/network/tcp_server.go
package network

import (
    "context"
    "postapocgame/server/internal/network"
    "postapocgame/server/service/gateway/internel/adapter/controller"
)

// TCPServerAdapter TCP 服务器适配器
type TCPServerAdapter struct {
    tcpServer network.ITCPServer
    controller *controller.ClientMessageController
}

func NewTCPServerAdapter(addr string, controller *controller.ClientMessageController) *TCPServerAdapter {
    return &TCPServerAdapter{
        controller: controller,
    }
}

func (a *TCPServerAdapter) Start(ctx context.Context) error {
    a.tcpServer = network.NewTCPServer(
        network.WithTCPServerOptionNetworkMessageHandler(a.controller),
        network.WithTCPServerOptionAddr(a.addr),
    )
    return a.tcpServer.Start(ctx)
}
```

## 5. 重构步骤

### 5.1 阶段一：基础结构搭建（1周）

1. **创建目录结构**
   - 创建 `domain/`、`usecase/`、`adapter/`、`infrastructure/` 目录
   - 定义基础接口（Repository、Gateway、RPC 等）

2. **迁移 Session 管理作为示例**
   - 创建 `domain/session.go`（提取 Session 实体）
   - 创建 `usecase/session/`（提取业务逻辑）
   - 创建 `adapter/gateway/session_repository.go`（数据访问实现）

3. **验证重构效果**
   - 确保功能正常
   - 编写单元测试（Use Case 层可独立测试）

### 5.2 阶段二：核心功能重构（2周）

1. **重构消息转发**
   - `ClientHandler` → `usecase/message/` + `adapter/controller/client_message_controller.go`
   - `GameMessageHandler` → `adapter/controller/gameserver_message_controller.go`

2. **统一网络适配**
   - 所有网络操作通过 NetworkGateway 接口
   - 实现 TCP 和 WebSocket 的统一适配

3. **统一 GameServer 通信**
   - 所有 GameServer 通信通过 GameServerRPC 接口
   - 实现 GameServerGateway 统一封装

### 5.3 阶段三：清理与优化（1周）

1. **移除旧代码**
   - 删除 `clientnet/` 中的旧实现
   - 清理直接依赖框架的代码

2. **完善测试**
   - 为 Use Case 层编写单元测试
   - 为 Controller 层编写集成测试

3. **文档更新**
   - 更新架构文档
   - 更新开发指南

## 6. 依赖注入设计

### 6.1 依赖注入容器

**目录：** `internel/di/container.go`

```go
// di/container.go
package di

import (
    "postapocgame/server/service/gateway/internel/adapter/controller"
    "postapocgame/server/service/gateway/internel/adapter/gateway"
    "postapocgame/server/service/gateway/internel/usecase/message"
    "postapocgame/server/service/gateway/internel/usecase/session"
)

// Container 依赖注入容器
type Container struct {
    // Repositories
    sessionRepo gateway.SessionRepository
    
    // Gateways
    networkGateway    gateway.NetworkGateway
    gameServerGateway gateway.GameServerGateway
    
    // Use Cases
    createSessionUseCase      *session.CreateSessionUseCase
    forwardToGameServerUseCase *message.ForwardToGameServerUseCase
    
    // Controllers
    clientMessageController *controller.ClientMessageController
}

func NewContainer() *Container {
    c := &Container{}
    
    // 初始化 Repositories
    c.sessionRepo = gateway.NewSessionRepository()
    
    // 初始化 Gateways
    c.networkGateway = gateway.NewNetworkGateway()
    c.gameServerGateway = gateway.NewGameServerGateway(...)
    
    // 初始化 Use Cases
    c.createSessionUseCase = session.NewCreateSessionUseCase(c.sessionRepo, c.gameServerGateway, ...)
    c.forwardToGameServerUseCase = message.NewForwardToGameServerUseCase(c.sessionRepo, c.gameServerGateway)
    
    // 初始化 Controllers
    c.clientMessageController = controller.NewClientMessageController(
        c.createSessionUseCase,
        c.forwardToGameServerUseCase,
        c.networkGateway,
    )
    
    return c
}
```

### 6.2 在 GatewayServer 中使用

```go
// engine/server.go
func NewGatewayServer(config *Config) (*GatewayServer, error) {
    container := di.NewContainer()
    
    return &GatewayServer{
        config:      config,
        container:   container,
        tcpServer:   infrastructure.NewTCPServerAdapter(config.TCPAddr, container.ClientMessageController),
        wsServer:    infrastructure.NewWSServerAdapter(config.WSAddr, container.ClientMessageController),
    }
}
```

## 7. 测试策略

### 7.1 Use Case 层单元测试

```go
// usecase/session/create_session_test.go
func TestCreateSessionUseCase_Execute(t *testing.T) {
    // Mock Repository
    mockRepo := &MockSessionRepository{}
    mockRPC := &MockGameServerRPC{}
    mockEventPub := &MockEventPublisher{}
    
    // 创建 Use Case
    uc := NewCreateSessionUseCase(mockRepo, mockRPC, mockEventPub)
    
    // 执行测试
    session, err := uc.Execute(ctx, "127.0.0.1:8080", domain.ConnTypeTCP)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, session)
    assert.True(t, mockRepo.CreateCalled)
    assert.True(t, mockRPC.NotifySessionCreatedCalled)
}
```

### 7.2 Controller 层集成测试

```go
// adapter/controller/client_message_controller_test.go
func TestClientMessageController_HandleMessage(t *testing.T) {
    // 使用真实 Repository（可以连接测试环境）
    sessionRepo := gateway.NewSessionRepository()
    // ...
    
    controller := NewClientMessageController(createSessionUseCase, forwardUseCase, networkGateway)
    
    // 执行测试
    err := controller.HandleMessage(ctx, conn, msg)
    
    // 验证结果
    assert.NoError(t, err)
}
```

## 8. 迁移检查清单

### 8.1 每个功能迁移检查项

- [ ] 创建 Domain 实体（移除框架依赖）
- [ ] 定义 Repository 接口
- [ ] 创建 Use Case（业务逻辑）
- [ ] 创建 Controller（消息处理）
- [ ] 实现 Gateway（网络和 GameServer 适配）
- [ ] 编写单元测试
- [ ] 验证功能正常
- [ ] 删除旧代码

### 8.2 整体检查项

- [ ] 所有业务逻辑已迁移
- [ ] 所有框架依赖已移除
- [ ] 依赖注入容器已配置
- [ ] 单元测试覆盖率 > 70%
- [ ] 集成测试通过
- [ ] 文档已更新

## 9. 注意事项

### 9.1 保持向后兼容

- 重构过程中保持旧代码可用
- 新代码与旧代码可以并存
- 逐步迁移，不一次性替换

### 9.2 性能考虑

- Gateway 是高频转发服务，必须保证性能
- 避免过度抽象导致性能下降
- 消息转发路径必须高效

### 9.3 并发安全

- Session 管理涉及并发访问
- 必须保证线程安全
- 使用适当的锁机制

### 9.4 连接管理

- 正确处理连接断开
- 清理相关资源
- 避免资源泄漏

### 9.5 限流和资源保护

- 实现会话数量限制
- 实现消息频率限制
- 防止资源耗尽

## 10. 参考资源

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Clean Architecture Example](https://github.com/bxcodec/go-clean-arch)
- [GameServer Clean Architecture 重构文档](./gameserver_CleanArchitecture重构文档.md)
- [DungeonServer Clean Architecture 重构文档](./dungeonserver_CleanArchitecture重构文档.md)
- [项目开发进度文档](./服务端开发进度文档.md)

## 11. 关键代码位置（重构后）

### 11.1 Domain 层
- `internel/domain/session.go` - Session 实体
- `internel/domain/message.go` - Message 值对象
- `internel/domain/repository/` - Repository 接口定义

### 11.2 Use Case 层
- `internel/usecase/session/` - Session 用例
- `internel/usecase/message/` - 消息转发用例
- `internel/usecase/interfaces/` - Use Case 依赖接口

### 11.3 Adapter 层
- `internel/adapter/controller/` - 消息控制器
- `internel/adapter/gateway/` - 网络和 GameServer 适配器

### 11.4 Infrastructure 层
- `internel/infrastructure/network/` - 网络适配
- `internel/infrastructure/tcp/` - TCP 服务器适配
- `internel/infrastructure/websocket/` - WebSocket 服务器适配

### 11.5 DI 容器
- `internel/di/container.go` - 依赖注入容器

---

**下一步行动：**
1. 评审本文档，确认重构方案
2. 创建基础目录结构
3. 选择第一个功能（建议 Session 管理）进行试点重构
4. 验证重构效果后，逐步迁移其他功能

