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

### 客户端消息转发流程 (GameServer → DungeonServer)

```
Client                    GameServer                    DungeonServer
  │                          │                                │
  │──[ClientMessage]────────>│                                │
  │   (协议未在GameServer)     │                                │
  │                          │ [检查协议管理器]                 │
  │                          │ [确定转发的srvType]              │
  │                          │──[RPCRequest]─────────────────>│
  │                          │   (sessionId, msgId, payload)   │
  │                          │                                │──[处理]
```

**新增: 智能协议转发流程**:
1. GameServer接收客户端消息
2. 优先检查是否能在GameServer处理（使用clientprotocol.GetFunc）
3. 如果GameServer无法处理，检查协议管理器：
   - 检查是否是DungeonServer的协议
   - 如果是独有协议(srvType指定)，直接转发到该srvType
   - 如果是通用协议，根据玩家当前所在的DungeonServer类型转发
4. 发送RPC请求到DungeonServer处理
5. DungeonServer处理并通过ClientMessage下行


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
- **新增**: 管理DungeonServer的协议注册

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
│   ├── 响应处理
│   └── ✨ProtocolManager # 协议注册管理器
├── TCP Server          # 接收Gateway连接
```

**新增: ProtocolManager 组件**:
- 存储DungeonServer注册的所有协议
- 按srvType组织协议信息
- 支持通用协议和独有协议的区分
- 提供协议查询和路由功能

**工作流程**:
1. **启动阶段**:
   - 初始化PlayerActor系统
   - 启动TCP服务器（等待Gateway连接）
   - 初始化ProtocolManager
   - 连接DungeonServer（按srvType建立连接池）
   - 注册协议注册RPC Handler

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

3. **新增: 智能协议转发**:
   ```
   客户端消息 → handleDoNetWorkMsg
           ↓
   [检查GameServer协议]
           ↓ (无法处理)
   [检查ProtocolManager] ← 是否DungeonServer协议?
           ↓
   [判断协议类型]
   ├─ 独有协议 → 转发到指定srvType
   └─ 通用协议 → 根据玩家当前srvType转发
   ```

4. **玩家Actor**:
   - 每个玩家对应一个Actor实例
   - Actor按SessionId进行消息路由
   - 支持注册多个业务Handler
   - 记录玩家当前所在的DungeonServer类型(DungeonSrvType)

5. **RPC调用DungeonServer**:
   - 构造RPCRequest
   - 根据srvType选择连接
   - 发送请求并等待响应

6. **协议注册处理**:
   - 接收DungeonServer的D2GRegisterProtocols RPC
   - 通过ProtocolManager存储协议信息
   - 支持自动协议注册和注销


### DungeonServer (副本服务器)

**职责**:
- 处理副本/战斗逻辑
- 管理副本实体（玩家、怪物等）
- 处理战斗、Buff、技能等系统
- **新增**: 向GameServer注册客户端协议

**核心组件**:
```
DungeonServer
├── DungeonActor        # 副本Actor系统
│   └── ModeSingle模式  # 全局单例Actor
├── GameServerLink      # GameServer连接处理
│   ├── 接收GameServer RPC
│   ├── 连接管理
│   ├── 消息路由
│   └── ✨ProtocolRegistration # 协议注册管理
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

**新增: 协议注册机制**:
- 自动向GameServer注册所有客户端协议
- 支持标记通用协议和独有协议
- 连接建立时注册，断开时注销
- 支持srvType标识的多种服务器类型

**工作流程**:
1. **启动阶段**:
   - 从config读取srv_type
   - 初始化DungeonActor（单例模式）
   - 启动TCP服务器（等待GameServer连接）
   - 初始化副本管理器
   - 注册ClientProtocol中的所有协议

2. **连接GameServer**:
   - 与GameServer建立TCP连接
   - 通过gameserverlink.SetDungeonSrvType()设置自己的srvType
   - 第一次接收GameServer消息时触发协议注册
   - 通过RPC D2GRegisterProtocols向GameServer注册

3. **协议注册过程**:
   - 调用gameserverlink.TryRegisterProtocols()
   - 获取clientprotocol.GetRegisteredProtocols()的所有协议
   - 可配置通用协议和独有协议的分类
   - 发送D2GRegisterProtocolsReq给GameServer
   - GameServer使用ProtocolManager存储信息

4. **RPC处理**:
   ```
   GameServer RPC → DungeonServer → DungeonActor
                                      ↓
                                路由到RPC Handler
                                      ↓
                                处理副本逻辑
                                      ↓
                                返回RPCResponse
   ```

5. **实体系统**:
   - 管理副本中的所有实体（玩家、怪物、NPC等）
   - 每个实体有唯一句柄(hdl)
   - 支持属性、战斗、Buff等子系统

6. **副本管理**:
   - 支持创建多个副本实例
   - 定期清理空副本
   - 实体进入/离开副本

7. **优雅关闭**:
   - 服务器关闭时调用gameserverlink.UnregisterProtocols()
   - 向GameServer发送D2GUnregisterProtocols
   - GameServer清理对应的协议注册信息

## 技术特性

### 1. Actor模式
- **GameServer**: ModePerKey - 每个玩家一个Actor，保证玩家消息串行处理
- **DungeonServer**: ModeSingle - 全局单例Actor，处理所有副本消息

### 2. 连接管理
- 支持自动重连
- 心跳保活机制
- 连接池管理（GameServer ↔ DungeonServer，按srvType分类）

### 3. 消息路由
- 基于SessionId进行消息路由
- Gateway维护Session到连接的映射
- GameServer维护Session到PlayerActor的映射
- ProtocolManager支持多srvType的协议路由

### 4. ✨新增: 动态协议注册和转发
- **自动注册**: DungeonServer启动后自动向GameServer注册协议
- **智能转发**: GameServer根据协议属性和玩家状态智能转发
- **协议分类**: 支持通用协议和独有协议，避免协议重复注册
- **热扩展**: 扩展新的DungeonServer(srvType)无需修改GameServer代码
- **自动清理**: 连接断开时自动清理协议注册

### 5. 错误处理
- 统一的错误码体系（protocol.ErrorCode）
- 错误自动记录调用位置
- 支持错误码扩展

### 6. 高可用设计
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
    "3": "127.0.0.1:9001",
    "4": "127.0.0.1:9002"
  }
}
```

### DungeonServer配置 (dungeonsrv.json)
```json
{
  "srv_type": 3,
  "tcp_addr": ":9001",
  "actor_mailbox_size": 1000
}
```

**配置说明**:
- `srv_type`: DungeonServer的类型标识(3=副本服务器, 4=跨服服务器等)
- `dungeon_server_addr_map`: GameServer连接DungeonServer的地址映射，按srvType索引

## 开发指南

### 添加新的客户端消息处理

1. 在 `proto/csproto/` 中定义新的消息结构
2. 在 `server/service/gameserver/internel/playeractor/entity/player_network.go` 中注册Handler:
```go
clientprotocol.Register(uint16(protocol.C2SProtocol_XXX), handleXXX)
```

### 协议注册和转发机制

#### 工作流程

1. **DungeonServer启动时**:
   - 设置自己的srvType（如3或4）
   - 与GameServer建立连接
   - 通过RPC向GameServer注册所有客户端协议
   - 可标记为通用协议或独有协议

2. **GameServer接收协议注册**:
   - 使用ProtocolManager存储DungeonServer的协议信息
   - 按srvType组织协议
   - 支持通用协议(多个srvType共享)和独有协议(特定srvType)

3. **处理客户端消息**:
   - 优先在GameServer处理（检查clientprotocol）
   - 如无法处理，检查是否是DungeonServer的协议
   - 独有协议直接转发到指定srvType
   - 通用协议根据玩家当前所在srvType转发

#### 代码示例

**DungeonServer配置**:
```go
// main.go中设置srvType
gameserverlink.SetDungeonSrvType(serverConfig.SrvType)

// 自动在连接建立后注册协议
// DungeonServer会调用RegisterProtocolsToGameServer注册所有协议
```

**GameServer处理消息**:
```go
// player_network.go中的handleDoNetWorkMsg已实现智能转发
// 流程:
// 1. 检查GameServer是否能处理 → clientprotocol.GetFunc()
// 2. 检查是否DungeonServer协议 → protocolMgr.IsDungeonProtocol()
// 3. 判断转发规则 → protocolMgr.GetSrvTypeForProtocol()
// 4. 根据协议类型转发:
//    - 独有协议: 转发到指定srvType
//    - 通用协议: 转发到玩家当前所在srvType
```

**玩家DungeonServer类型**:
```go
// PlayerRole中记录玩家所在的DungeonServer
playerRole.SetDungeonSrvType(uint8(protocol.SrvType_SrvTypeDungeonServer))

// 转发通用协议时查询
targetSrvType := pr.GetDungeonSrvType()
```

#### 协议分类

- **通用协议**: 多个DungeonServer共享(如移动、释放技能等)
  - 使用isCommon=true标记
  - GameServer转发时根据玩家当前srvType决定目标
  - 只在GameServer注册一份

- **独有协议**: 特定DungeonServer独有(如跨服特定操作)
  - 使用isCommon=false标记，指定srvType
  - GameServer转发时直接发往指定srvType
  - 不同srvType可有相同msgId但不同含义

#### 扩展新srvType

1. **配置新的DungeonServer**:
```json
{
  "srv_type": 4,
  "tcp_addr": ":9002"
}
```

2. **GameServer配置关键的连接地址**:
```json
{
  "dungeon_server_addr_map": {
    "3": "127.0.0.1:9001",
    "4": "127.0.0.1:9002"
  }
}
```

3. **DungeonServer启动时自动注册协议**:
   - 无需修改GameServer代码
   - 协议通过RPC D2GRpcProtocol_D2GRegisterProtocols注册
   - 可混合通用和独有协议

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

