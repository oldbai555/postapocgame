# 游戏服务器架构文档

## 服务架构概览

本项目采用微服务架构，包含三个核心服务：

```
┌─────────────┐
│   Client    │ (游戏客户端)
└──────┬──────┘
       │ TCP/WebSocket
       │
┌──────▼─────────────────────────────────────────┐
│            Gateway (网关服务)                    │
│  - 管理客户端连接 (TCP/WebSocket)                │
│  - 会话管理 (SessionManager)                    │
│  - 消息转发 (Client ↔ GameServer)              │
└──────┬─────────────────────────────────────────┘
       │ TCP (ForwardMessage)
       │
┌──────▼─────────────────────────────────────────┐
│          GameServer (游戏服务器)                 │
│  - 玩家业务逻辑处理                              │
│  - Actor模式处理玩家消息                         │
│  - 连接DungeonServer进行副本操作                │
└──────┬─────────────────────────────────────────┘
       │ TCP (RPC)
       │
┌──────▼─────────────────────────────────────────┐
│        DungeonServer (副本服务器)                │
│  - 副本/战斗逻辑处理                             │
│  - Actor模式处理副本消息                         │
│  - 实体系统 (Entity, Buff, Fight等)            │
└────────────────────────────────────────────────┘
```

## 服务间通信方式

### 1. Client ↔ Gateway

**协议**: TCP 或 WebSocket

**消息流向**:
- **客户端 → 网关**: 客户端发送业务消息
- **网关 → 客户端**: 网关转发GameServer的响应消息

**连接管理**:
- Gateway维护客户端Session
- 支持TCP和WebSocket两种连接方式
- 每个连接对应一个Session，Session有唯一ID

### 2. Gateway ↔ GameServer

**协议**: TCP (长连接)

**消息类型**:
- **ForwardMessage**: Gateway转发客户端消息到GameServer
- **ClientMessage**: GameServer发送消息给客户端（通过Gateway转发）
- **SessionEvent**: 会话事件（新建/关闭）

**连接特点**:
- Gateway主动连接GameServer
- 支持自动重连
- 心跳保活机制

### 3. GameServer ↔ DungeonServer

**协议**: TCP (长连接，支持多连接池)

**消息类型**:
- **RPCRequest**: GameServer调用DungeonServer的RPC方法
- **RPCResponse**: DungeonServer返回RPC响应
- **ClientMessage**: DungeonServer需要发送消息给客户端时（通过GameServer转发）

**连接特点**:
- GameServer主动连接DungeonServer
- 支持按服务类型(srvType)建立多个连接池
- 支持自动重连和心跳

## 通信流程

### 客户端消息上行流程

```
Client                    Gateway                    GameServer
  │                          │                            │
  │──[ClientMessage]────────>│                            │
  │                          │──[ForwardMessage]─────────>│
  │                          │                            │──[Actor处理]
  │                          │                            │
  │                          │<──[ClientMessage]──────────│
  │<──[ClientMessage]────────│                            │
```

**详细步骤**:
1. 客户端发送消息到Gateway（TCP/WebSocket）
2. Gateway创建/获取Session，封装为ForwardMessage
3. Gateway通过TCP连接转发到GameServer
4. GameServer解码消息，发送到对应的PlayerActor处理
5. GameServer处理完成后，通过Gateway转发响应给客户端

### 副本操作流程 (GameServer → DungeonServer)

```
GameServer              DungeonServer
    │                        │
    │──[RPCRequest]─────────>│
    │   (MsgId, SessionId,   │
    │    Data)               │
    │                        │──[Actor处理]
    │                        │
    │<──[RPCResponse]────────│
    │   (RequestId, Code,    │
    │    Data)               │
```

**详细步骤**:
1. GameServer构造RPCRequest（包含RequestId, SessionId, MsgId, Data）
2. 通过TCP连接发送到DungeonServer
3. DungeonServer解码并路由到对应的RPC Handler
4. Handler处理完成后构造RPCResponse返回
5. GameServer接收响应并处理

### 副本消息下行流程 (DungeonServer → Client)

```
DungeonServer          GameServer              Gateway              Client
    │                      │                      │                    │
    │──[ClientMessage]────>│                      │                    │
    │   (SessionId,        │──[ForwardMessage]───>│                    │
    │    Payload)          │                      │                    │
    │                      │                      │──[ClientMessage]──>│
```

**详细步骤**:
1. DungeonServer需要通知客户端时，发送ClientMessage到GameServer
2. GameServer封装为ForwardMessage转发到Gateway
3. Gateway根据SessionId找到对应Session
4. Gateway通过客户端连接发送消息

## 消息结构

### 基础消息格式 (所有TCP通信)

```
┌─────────────────────────────────────────┐
│  4字节长度  │  1字节类型  │  Payload    │
└─────────────────────────────────────────┘
```

- **长度字段**: uint32 (LittleEndian)，表示后续总长度（1字节类型 + Payload长度）
- **类型字段**: byte，消息类型枚举
- **Payload**: 可变长度，根据消息类型不同而不同

### 消息类型枚举

| 类型值 | 名称 | 说明 |
|--------|------|------|
| 0x01 | MsgTypeSessionEvent | 会话事件（新建/关闭） |
| 0x02 | MsgTypeClient | 客户端消息 |
| 0x03 | MsgTypeRPCRequest | RPC请求 |
| 0x04 | MsgTypeRPCResponse | RPC响应 |
| 0x05 | MsgTypeHandshake | 握手消息 |
| 0x06 | MsgTypeHeartbeat | 心跳消息 |

### 客户端消息结构 (ClientMessage)

**编码格式**:
```
┌──────────────┬──────────┐
│ msgId(2字节) │   data   │
└──────────────┴──────────┘
```

- **msgId**: uint16，业务消息ID（定义在proto/csproto中）
- **data**: []byte，protobuf序列化的业务数据

**使用场景**:
- Client → Gateway → GameServer (上行)
- GameServer → Gateway → Client (下行)
- DungeonServer → GameServer → Gateway → Client (副本消息下行)

### 转发消息结构 (ForwardMessage)

**编码格式**:
```
┌──────────────────┬──────────────┬──────────┐
│ sessionIdLen(2)  │  sessionId   │ payload  │
└──────────────────┴──────────────┴──────────┘
```

- **sessionIdLen**: uint16，SessionId字符串长度
- **sessionId**: string，会话ID
- **payload**: []byte，客户端消息的完整数据（包含msgId+data）

**使用场景**:
- Gateway → GameServer: 转发客户端消息
- GameServer → Gateway: 转发给客户端的消息

### RPC请求结构 (RPCRequest)

**编码格式**:
```
┌──────────────┬──────────────┬──────────────┬──────────┬──────────┐
│ requestId(4) │ sessionIdLen │  sessionId   │ msgId(2) │   data   │
└──────────────┴──────────────┴──────────────┴──────────┴──────────┘
```

- **requestId**: uint32，请求ID（用于匹配响应）
- **sessionIdLen**: uint16，SessionId长度
- **sessionId**: string，会话ID
- **msgId**: uint16，RPC方法ID
- **data**: []byte，RPC参数数据

**使用场景**:
- GameServer → DungeonServer: 调用副本RPC方法

### RPC响应结构 (RPCResponse)

**编码格式**:
```
┌──────────────┬──────────┬──────────┐
│ requestId(4) │ code(4)  │   data   │
└──────────────┴──────────┴──────────┘
```

- **requestId**: uint32，对应的请求ID
- **code**: int32，错误码（0表示成功）
- **data**: []byte，响应数据

**使用场景**:
- DungeonServer → GameServer: 返回RPC调用结果

### 会话事件结构 (SessionEvent)

**编码格式**:
```
┌──────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ eventType(1) │ sessionIdLen │  sessionId   │  userIdLen   │   userId     │
└──────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

- **eventType**: byte，事件类型（0x00=新建, 0x01=关闭）
- **sessionIdLen**: uint16
- **sessionId**: string
- **userIdLen**: uint16
- **userId**: string

**使用场景**:
- Gateway → GameServer: 通知会话创建/关闭

## 各服务工作方式

### Gateway (网关服务)

**职责**:
- 作为客户端和GameServer之间的中间层
- 管理客户端连接和会话
- 消息转发和路由

**核心组件**:
```
GatewayServer
├── SessionManager      # 会话管理器
│   ├── 创建/销毁Session
│   ├── 会话超时清理
│   └── Session路由表
├── ClientHandler       # 客户端消息处理器
│   ├── 接收客户端消息
│   ├── 创建Session
│   └── 转发到GameServer
├── GameServerConnector # GameServer连接器
│   ├── TCP长连接
│   ├── 自动重连
│   └── 消息收发
└── TCP/WebSocket Server # 客户端接入层
    ├── TCP服务器
    └── WebSocket服务器
```

**工作流程**:
1. **启动阶段**:
   - 初始化SessionManager
   - 连接GameServer
   - 启动TCP/WebSocket服务器

2. **客户端连接**:
   - 接受客户端连接（TCP/WebSocket）
   - 创建Session（生成唯一SessionId）
   - 通知GameServer会话创建事件

3. **消息处理**:
   - 接收客户端消息 → 封装为ForwardMessage → 转发到GameServer
   - 接收GameServer消息 → 根据SessionId路由 → 发送到对应客户端

4. **会话管理**:
   - 定期清理超时会话
   - 客户端断开时通知GameServer

### GameServer (游戏服务器)

**职责**:
- 处理玩家业务逻辑
- 管理玩家数据
- 协调DungeonServer进行副本操作

**核心组件**:
```
GameServer
├── PlayerActor         # 玩家Actor系统
│   ├── ModePerKey模式  # 每个玩家一个Actor
│   ├── 消息邮箱
│   └── 业务Handler注册
├── GatewayLink         # Gateway连接处理
│   ├── 接收Gateway消息
│   ├── 会话管理
│   └── 消息路由到Actor
├── DungeonServerLink   # DungeonServer连接
│   ├── 连接池管理
│   ├── RPC调用
│   └── 响应处理
└── TCP Server          # 接收Gateway连接
```

**工作流程**:
1. **启动阶段**:
   - 初始化PlayerActor系统
   - 启动TCP服务器（等待Gateway连接）
   - 连接DungeonServer（按srvType建立连接池）

2. **消息处理 (Actor模式)**:
   ```
   客户端消息 → Gateway → GameServer → PlayerActor
                                    ↓
                              根据SessionId路由
                                    ↓
                           找到对应的PlayerActor
                                    ↓
                           发送到Actor邮箱
                                    ↓
                           Actor异步处理
   ```

3. **玩家Actor**:
   - 每个玩家对应一个Actor实例
   - Actor按SessionId进行消息路由
   - 支持注册多个业务Handler（如登录、创建角色、进入游戏等）

4. **RPC调用DungeonServer**:
   - 构造RPCRequest
   - 根据srvType选择连接
   - 发送请求并等待响应

### DungeonServer (副本服务器)

**职责**:
- 处理副本/战斗逻辑
- 管理副本实体（玩家、怪物等）
- 处理战斗、Buff、技能等系统

**核心组件**:
```
DungeonServer
├── DungeonActor        # 副本Actor系统
│   └── ModeSingle模式  # 全局单例Actor
├── GameServerLink      # GameServer连接处理
│   ├── 接收GameServer RPC
│   ├── 连接管理
│   └── 消息路由
├── EntitySystem        # 实体系统
│   ├── EntityManager   # 实体管理
│   ├── AttrSys         # 属性系统
│   ├── FightSys        # 战斗系统
│   ├── BuffSys         # Buff系统
│   └── AOI             # 视野系统
├── FuBenManager        # 副本管理器
│   ├── 副本创建/销毁
│   └── 定期清理
└── TCP Server          # 接收GameServer连接
```

**工作流程**:
1. **启动阶段**:
   - 初始化DungeonActor（单例模式）
   - 启动TCP服务器（等待GameServer连接）
   - 初始化副本管理器

2. **RPC处理**:
   ```
   GameServer RPC → DungeonServer → DungeonActor
                                      ↓
                                路由到RPC Handler
                                      ↓
                                处理副本逻辑
                                      ↓
                                返回RPCResponse
   ```

3. **实体系统**:
   - 管理副本中的所有实体（玩家、怪物、NPC等）
   - 每个实体有唯一句柄(hdl)
   - 支持属性、战斗、Buff等子系统

4. **副本管理**:
   - 支持创建多个副本实例
   - 定期清理空副本
   - 实体进入/离开副本

## 技术特性

### 1. Actor模式
- **GameServer**: ModePerKey - 每个玩家一个Actor，保证玩家消息串行处理
- **DungeonServer**: ModeSingle - 全局单例Actor，处理所有副本消息

### 2. 连接管理
- 支持自动重连
- 心跳保活机制
- 连接池管理（GameServer ↔ DungeonServer）

### 3. 消息路由
- 基于SessionId进行消息路由
- Gateway维护Session到连接的映射
- GameServer维护Session到PlayerActor的映射

### 4. 错误处理
- 统一的错误码体系（protocol.ErrorCode）
- 错误自动记录调用位置
- 支持错误码扩展

### 5. 高可用设计
- 非阻塞消息发送
- 优雅关闭机制
- 资源自动清理
- 防止内存泄露

## 配置文件

### Gateway配置 (gateway.json)
```json
{
  "gameServerAddr": "127.0.0.1:8001",
  "tcp_addr": ":8080",
  "ws_addr": ":8081",
  "ws_path": "/ws",
  "maxSessions": 10000,
  "sessionBufferSize": 256,
  "sessionTimeout": "5m"
}
```

### GameServer配置 (gamesrv.json)
```json
{
  "app_id": 1,
  "platform_id": 1,
  "srv_id": 1,
  "tcp_addr": ":8001",
  "gateway_allow_ips": ["127.0.0.1"],
  "actor_mode": 1,
  "actor_pool_size": 1000,
  "actor_mailbox_size": 1000,
  "dungeon_server_addr_map": {
    "1": "127.0.0.1:9001"
  }
}
```

### DungeonServer配置 (dungeonsrv.json)
```json
{
  "app_id": 2,
  "platform_id": 1,
  "srv_id": 1,
  "tcp_addr": ":9001",
  "game_server_allow_ips": ["127.0.0.1"]
}
```

## 开发指南

### 添加新的客户端消息处理

1. 在 `proto/csproto/` 中定义新的消息结构
2. 在 `server/service/gameserver/internel/playeractor/entity/player_network.go` 中注册Handler:
```go
clientprotocol.Register(uint16(protocol.C2SProtocol_XXX), handleXXX)
```

### 添加新的RPC方法

1. 在 `proto/csproto/rpc.proto` 中定义RPC消息
2. 在GameServer中注册RPC Handler:
```go
dungeonserverlink.RegisterRPCHandler(msgId, handler)
```
3. 在DungeonServer中实现RPC Handler:
```go
dshare.RegisterHandler(msgId, handler)
```

### 错误处理

统一使用 `customerr.NewErrorByCode` 创建错误:
```go
return customerr.NewErrorByCode(int32(protocol.ErrorCode_XXX), "error message")
```

## 部署说明

1. **启动顺序**: DungeonServer → GameServer → Gateway
2. **端口要求**: 
   - Gateway: TCP 8080, WebSocket 8081
   - GameServer: TCP 8001
   - DungeonServer: TCP 9001
3. **依赖关系**: Gateway依赖GameServer，GameServer依赖DungeonServer

