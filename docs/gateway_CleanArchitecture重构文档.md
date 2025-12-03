# Gateway Clean Architecture 重构文档

更新时间：2025-01-XX  
责任人：开发团队

## 1. 文档目的

本文档旨在将 `server/service/gateway` 按照 Clean Architecture（清洁架构）原则进行重构，实现业务逻辑与框架解耦，提高代码可测试性、可维护性和可扩展性。

## 2. 当前架构现状与问题分析

### 2.0 代码现状快速梳理（2025-12-03）

- `main.go`：进程入口，加载 `gateway.json`，构造 `engine.GatewayServer` 并托管生命周期。
- `internel/engine`（进程级生命周期 + 接入）：
  - `config.go`：`Config` 结构体（`GameServerAddr/TCPAddr/WSAddr/WSPath/SessionBufferSize/MaxSessions/SessionTimeout/MaxFrameSize`），负责默认值填充与合法性校验（地址、Session/Frame 上限等），从 `gateway.json` 读取配置。
  - `server.go`：`GatewayServer` 进程级对象，内部持有 `SessionManager` + `IGameServerConnector`（`gameserverlink.GameClient`）+ `network.ITCPServer` + `network.WSServer`，负责：
    - 启动时连接 GameServer、启动 Session 清理协程、拉起 TCP/WS 监听。
    - 循环拉取 GameServer 回包（`ReceiveGsMessage`）→ 定位 Session → 投递到 `Session.SendChan`。
    - 统一 Stop：停止 TCP/WS、停止 SessionManager、关闭 GameClient。
- `internel/clientnet`（当前的“Session / Handler / GameServer 适配”集中地）：
  - `iface.go`：定义 `IConnection`（面向 Gateway 自身的抽象）与 `IGameServerConnector`（对 GameServer 的抽象），以及连接/会话状态枚举。
  - `adapter.go`：`ConnectionAdapter` 将 `network.IConnection` 适配为 `IConnection`，是当前唯一的“网络适配”位置。
  - `session.go`：Gateway 内部 `Session` 实体（会话 ID / 地址 / 连接类型 / 状态 / UserId / SendChan / stopChan / 时间戳），内置 `SafeClose/Stop` 做通道幂等关闭，保证 `SendChan/stopChan` 只会被关闭一次。
  - `session_mgr.go`：`SessionManager` 管理所有会话（创建/查找/关闭/统计），负责：
    - 控制最大会话数（`MaxSessions`）、发送缓冲区大小（`SessionBufferSize`）、超时时间（`SessionTimeout`）。
    - 通过 `servertime.Now()` 维护活跃时间，定期清理超时会话（`StartCleanup/cleanupTimeoutSessions`）。
    - 通过 `IGameServerConnector.NotifySessionEvent` 向 GameServer 通知 Session 新建/关闭。
  - `handler.go`：`ClientHandler` 作为 TCP/WS 消息入口与当前的“控制器 + 发送管线”：
    - 在 `HandleMessage` 中直接读取 `network.Message`，为 `MsgTypeClient` 时获取/创建 `Session`、更新活跃时间并调用 `IGameServerConnector.ForwardClientMsg`。
    - 维护 `map[network.IConnection]*Session`，负责为每个会话起一个发送协程（`handleSend`），从 `Session.SendChan` 读取并通过 `network.IConnection.SendMessage` 下发给客户端；内部通过 `stopChan` + 连续失败计数（`maxConsecutiveFailures`）判断连接是否需要关闭。
- `internel/gameserverlink`（GameServer TCP 客户端与消息编解码）：
  - `game_cli.go`：`GameClient` 封装与 GameServer 的 TCP 客户端，负责编码 SessionEvent/ForwardMessage 并发送。
  - `msg_handler.go`：`GameMessageHandler` 解码来自 GameServer 的 `MsgTypeClient`，转换为 `ForwardMessage` 放入内部 `recvChan`，供 `GatewayServer.dispatchGameServerMessages` 轮询。
  - 当前 `GameClient` 只通过 `network.ITCPClient` + `network.Codec` 发送/接收消息，不感知 Session 具体实现。

### 2.1 依赖方向混乱

**问题描述：**
- 业务逻辑层直接依赖 `network`、`gameserverlink` 等框架层
- Session 管理、消息转发、连接处理等逻辑混在一起
- 内层（业务逻辑）依赖外层（框架），违反了依赖倒置原则

**典型示例（摘自当前实现）：**

- `clientnet.ClientHandler.HandleMessage` 同时：
  - 直接依赖 `network.IConnection` 与 `network.Message`。
  - 负责“获取/创建 Session + 更新活跃时间 + 调用 GameServerConnector.ForwardClientMsg`”，既做网络适配又做业务决策。
- `clientnet.SessionManager.CreateSession` 在创建 Session 时直接构造 `network.SessionEvent` 并调用 `IGameServerConnector.NotifySessionEvent`，Repository 与 RPC 责任耦合在一起。

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

### 2.5 现有代码到 Clean Architecture 分层的映射与缺口

> 本小节是“从当前实现到目标架构”的导航，用来回答“这几份 Go 文件在 Clean Architecture 里应该长成什么样”以及“现在还缺什么”。

- **Entities（domain 层，目标）**
  - 目标实体：`domain.Session`、`domain.Message`、`domain.ConnType`、`domain.SessionState` 等。
  - 当前落点：`clientnet/session.go` + `clientnet/iface.go`（结构体/枚举定义 + 并发控制）。
  - 缺口：尚未有 `internel/domain/*` 目录与独立的纯业务实体，`Session` 同时承担了“并发/通道管理 + 业务含义”（状态、超时规则）。
- **Repositories（domain/repository 层，目标）**
  - 目标接口：`SessionRepository`（Create/Get/Update/Delete/GetAll/Count）。
  - 当前落点：`SessionManager.sessions map[string]*Session` 直接由 `SessionManager` 管理。
  - 缺口：没有独立的 Repository 接口与实现类型，`SessionManager` 同时是“仓库 + 业务规则 + GameServer 通知者”。
- **Use Cases（usecase 层，目标）**
  - 目标用例：`CreateSession/CloseSession/UpdateActivity/ForwardToGameServer/ForwardToClient/CleanupTimeoutSessions` 等，负责：
    - 会话数量上限校验、超时规则、错误码选择。
    - 是否需要通知 GameServer、是否记录日志/事件。
  - 当前落点：分散在 `SessionManager`（Create/Close/UpdateActivity/cleanupTimeoutSessions）、`GatewayServer.dispatchGameServerMessages`、`ClientHandler.HandleMessage/handleSend`。
  - 缺口：`internel/usecase/*` 目录尚未创建，业务规则直接写在“管理器/Handler/Server”中，无法单测。
- **Interface Adapters（adapter/controller & adapter/gateway 层，目标）**
  - 目标 Controller：`ClientMessageController`（从 `network.Message` 中解析出 domain.Message + SessionID）、`GameServerMessageController`（从 GameServer 回包构建 domain.Message 并路由）。
  - 目标 Gateway：
    - `NetworkGateway`：负责 `network.IConnection` ↔ SessionID 绑定、向客户端发送二进制。
    - `GameServerGateway`：基于现有 `gameserverlink.GameClient` 实现 `GameServerRPC` 接口。
    - `SessionRepositoryImpl`：基于内存 map 封装 Repository 实现。
  - 当前落点：
    - `clientnet.ClientHandler` 既是“Controller”（处理 `network.Message`），又是“Gateway”（持有 `map[network.IConnection]*Session`）与“用例调用者”（直接调 `ForwardClientMsg`）。
    - `gameserverlink.GameClient` 直接实现了 `IGameServerConnector`，但该接口定义在 `clientnet/iface.go`。
  - 缺口：
    - `internel/adapter/controller/*`、`internel/adapter/gateway/*` 目录尚未创建。
    - `NetworkGateway` 与 `GameServerRPC` 接口目前混在 `clientnet.IGameServerConnector` 与 `ClientHandler` 中。
- **Frameworks & Drivers（infrastructure 层，目标）**
  - 目标：在 `internel/infrastructure/network/*` 下包装 `server/internal/network` 的 TCP/WS Server/Client，实现“启动/停止逻辑 + 统一 Handler 接口”。
  - 当前落点：
    - `engine.startTCPServer/startWSServer` 直接依赖 `network.NewTCPServer/NewWSServer` 并将 `ClientHandler` 作为 `NetworkMessageHandler`。
    - WebSocket 安全配置仍为开发模式：`AllowedIPs=nil`，`HandshakeEnable=false`，`CheckOrigin=func() bool { return true }`。
  - 缺口：
    - `internel/infrastructure/*` 目录尚未创建。
    - WebSocket/TCP 接入安全、限流、监控等能力尚未抽象为可配置的 UseCase/Adapter，只在 `Config` 中提供了连接/会话上限的基础参数。

> 结论（现状）：当前 gateway 已具备完整的“接入 → Session 生命周期 → GameServer 转发 → 回包下发”功能，但所有职责都集中在 `engine/clientnet/gameserverlink` 三个包内，尚未按照 Clean Architecture 拆出 `domain/usecase/adapter/infrastructure/di` 等层次。后续重构以“平移逻辑到 UseCase + Adapter”，再逐步瘦身 `SessionManager/ClientHandler/GatewayServer` 为主。

### 2.6 2025-12-03 本轮梳理结论

- ✅ 已创建第一批 Clean Architecture 目录与骨架代码：
  - `internel/domain/session.go`、`internel/domain/message.go`：沉淀 Session/Message 纯业务实体与 `ConnType/SessionState` 值对象。
  - `internel/domain/repository/session_repository.go`：抽象 `SessionRepository`，供 UseCase 注入。
  - `internel/usecase/interfaces/{gameserver_rpc,event_publisher,network_gateway}.go`：定义 GameServer 通知、事件发布、客户端回包三个关键信道。
  - `internel/usecase/session/create_session.go`、`internel/usecase/message/forward_to_gameserver.go`：编写 CreateSession / ForwardToGameServer 用例骨架，带默认 `Clock/IDGenerator`，支持注入。
- ⚠️ 仍未落地的关键缺口：
  - **Repository 实现**：`clientnet.SessionManager` 仍将 `map` 与并发控制、GameServer 通知耦合，需要拆分出实现 `SessionRepository` 的结构体，再由 UseCase 驱动通知流程。
  - **Controllers / Gateways**：`clientnet.ClientHandler` 依旧承担入口、网关、用例 orchestration，多数 TODO 已在 2.5 中列出，此轮新增代码尚未接入 Handler。
  - **依赖反转**：`engine.GatewayServer` 仍直接 new `SessionManager` 和 `gameserverlink.GameClient`，未通过接口组合。后续需引入 DI（或最简工厂）把 UseCase 注入到 Handler。
  - **安全与可观测**：WebSocket 握手、Origin/IP 校验、限流（连接级/消息级）仍缺；`handleSend` 只打印日志，未反馈 UseCase，也无 metrics。
  - **上下文传递**：`ClientHandler.HandleMessage` 使用 `context.Background()` 转发，导致请求级 trace 丢失；`SessionManager.CreateSession`/`CloseSession` 对 GameServer 的通知没有超时/重试策略。
- 🧭 建议的下一步：
  1. 以 `SessionManager` 为切入点，实现 `repository.SessionRepository` 适配层，让 UseCase 成为唯一入口。
  2. 编写 `adapter/controller/client_message_controller.go`，负责 `network.Message → domain.Message` 转换并注入 UseCase。
  3. 在 `engine` 层引入装配函数（`builder`/`wire`），把 `GatewayServer` 依赖换成接口，便于单测。
  4. 引入 `infrastructure/network` 包承载 TCP/WS server 配置和安全策略，避免 `engine` 直接操作底层网络细节。

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

### 3.2 目录结构设计（目标状态）

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
│   │   ├── controller/           # 消息控制器（面向 network / GameServer 回包）
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

### 3.3 Controller / SystemAdapter / UseCase 职责边界（在本项目中的统一约定）

> 本小节既约束 GameServer / DungeonServer，也约束 Gateway 重构后的分层行为，保证三类组件职责一致且不破坏 Clean Architecture 依赖方向。这里的约定需要和 `docs/服务端开发进度文档.md` 的 4.3 小节保持完全一致。

- **Controller（适配协议 / 驱动 UseCase）**
  - 面向外部协议与框架类型：在 GameServer 中处理 `ClientMessage` / Actor 上下文，在 Gateway 中处理 `network.Message` / `IConnection`。
  - 负责“入口级逻辑”（**只在入口做决策，不持久化状态**）：
    - 解析与基础参数校验（协议字段、必要 ID、Session 是否存在等）。
    - **框架层校验**：系统是否开启（通过 SystemAdapter 或进程级 Feature 开关）、权限/限流、上游依赖是否就绪（例如 Gateway 到 GameServer 的连接状态）。
    - 从上下文中提取 `SessionID/RoleID` 等，再拼装成 domain/usecase 所需的输入。
  - 只依赖 UseCase 层暴露的接口（或具体用例类型），不直接操作 Repository / Entity，也不感知底层网络实现细节。

- **SystemAdapter（Actor 侧系统适配 / 生命周期管理）**
  - 主要存在于 GameServer/DungeonServer：挂在 Actor 上，负责系统的生命周期（Init/Login/RunOne/NewDay）、事件订阅与路由。
  - 持有与 Actor 运行模型强相关的运行时状态：例如系统是否已解锁、是否处于冷却或维护期等。
  - **对外暴露系统开启/关闭视图（只读）**：
    - GameServer 侧通过 helper（如 `GetBagSys(ctx)`）或统一的 `SystemValidator` 暴露“系统是否存在 / 是否开启”，供 Controller 调用（详细见 `docs/已完成/SystemAdapter系统开启检查优化方案.md`）。
    - Gateway 没有玩家级 SystemAdapter，但可以在进程级维护类似“子系统开关”（例如是否开启 TCP 接入、是否允许客户端消息转发），对外暴露为只读配置接口（由 Controller / Adapter 查询）。
  - 不下沉业务规则（数值、掉落、具体玩法、转发策略等）的判断，这些仍由 UseCase 层负责。

- **UseCase（纯业务用例 / 不感知系统开关）**
  - 只依赖：
    - `internel/domain/*` 中的实体/值对象。
    - `internel/domain/repository` 与 `internel/usecase/interfaces` 中定义的抽象接口（Repository / RPC / Gateway / PresenterAdapter 等）。
  - 负责业务规则本身：会话数量上限、超时与重试策略、消息路由规则、背包/货币/副本结算等。
  - **不感知也不查询“系统是否开启”**：
    - UseCase 假定被调用时“对应系统已经就绪”，是否允许调用由外层 Controller / SystemAdapter / 进程级适配层决定。
    - UseCase 不依赖 SystemAdapter 或任何“开关配置”类型，也不持有 System 开关状态，避免依赖方向从内层倒向外层。

- **在当前 Gateway 代码中的映射（便于对照重构）**
  - 现状：
    - `clientnet.ClientHandler` 同时承担 “Controller + NetworkGateway + ForwardMessage 用例调度” 三个角色。
    - `clientnet.SessionManager` 同时承担 “SessionRepository + Session 用例 + 通知 GameServer” 三个角色。
    - `engine.GatewayServer` 同时承担 “基础设施启动/停止 + GameServer 消息循环调度（相当于一个进程级 SystemAdapter）”。
  - 目标：
    - 将 Session/转发规则迁移到 `internel/usecase/*`，让 `SessionManager` 剩下 Repository + 并发控制。
    - 将 `ClientHandler` 瘦身为真正的 Controller，只做网络消息 → UseCase 调用；连接映射与系统开关检查通过 `NetworkGateway` / `GatewayFeatureConfig` 等接口实现。
    - 将 `GatewayServer.dispatchGameServerMessages` 中的路由/丢弃策略迁移到 UseCase 中，`GatewayServer` 本身只保留“进程级 SystemAdapter + 基础设施装配”的角色。

### 3.4 系统开关检查策略（GameServer / DungeonServer / Gateway 统一方案）

#### 3.4.1 设计原则

- **不破坏 Clean Architecture 依赖方向**
  - System 开关状态（“是否开启/维护中/未解锁”）属于框架与运行时管理的范畴，位于 Adapter / SystemAdapter / Infrastructure 一侧，而不是 UseCase 内部状态。
  - UseCase 层只通过接口调用外部依赖，不反向查询 SystemAdapter，也不持有具体 System 对象。

- **“能否调用 UseCase”由外层决策**
  - 所有“系统是否开启”的判断都在 **Controller 层或 SystemAdapter 层** 完成，结论只有两种：
    - 不允许：直接返回错误或构造失败响应，不调用 UseCase。
    - 允许：继续构造输入参数，调用 UseCase 执行业务逻辑。
  - 这样 UseCase 可以保持“被调用即表示系统就绪”的简单前提，易于测试和复用。

#### 3.4.2 GameServer / DungeonServer：沿用 SystemAdapter 方案

- **Controller 层统一做系统开启检查（推荐）**
  - 参考 `docs/已完成/SystemAdapter系统开启检查优化方案.md`，Controller 在调用 UseCase 前先通过 SystemAdapter helper（如 `system.GetBagSys(ctx)`）或 `SystemValidator.CheckSystemEnabled` 检查系统状态：
    - 系统不存在 / 未开启：直接通过 Presenter 返回统一错误码与提示文案（例如“背包系统未开启”），不执行 UseCase。
    - 系统已开启：继续提取 `RoleID/SessionID` 等信息并调用对应 UseCase。
  - 对一些内部/定时任务（没有显式 Controller）：
    - 由 SystemAdapter 在定时驱动 UseCase 前自行检查自身开关状态（例如停服维护时不再驱动某些结算逻辑），仍然遵守“UseCase 不主动感知开关”的前提。

- **职责拆分回顾**
  - SystemAdapter：管理系统状态（含开启/关闭）和 Actor 生命周期，对内驱动 UseCase，对外提供“是否开启”的只读视图。
  - Controller：在“协议入口”处消费 SystemAdapter 提供的视图，决定是否放行到 UseCase。
  - UseCase：完全不包含“某系统是否开启”的判断逻辑，只关心规则本身和输入参数是否满足业务约束。

#### 3.4.3 Gateway：在 Controller / Adapter 层实现“子系统开关”

> Gateway 没有玩家级 SystemAdapter，但同样存在“系统级开关”的需求，例如：是否允许新的 TCP 会话接入、是否仍然允许客户端发消息给 GameServer、是否仅允许白名单 IP 访问等。

- **Gateway 中“系统开关”的划分建议**
  - **接入系统开关**：是否启动 TCP/WS 监听、是否允许新连接接入（例如临时只保留已有连接）。
  - **转发系统开关**：是否允许客户端消息转发到 GameServer、是否允许 GameServer 回包下发给客户端。
  - **安全策略开关**：黑名单/白名单、握手 Token 校验、Origin 校验是否启用。

- **落地方式一：Controller 层前置校验（推荐，与 GameServer 一致）**
  - 在 `internel/adapter/controller/client_message_controller.go` 中，增加对 Gateway 子系统开关的前置检查：
    - 通过注入的接口读取只读运行时配置，例如：
      - `GatewayFeatureConfig.CanAcceptClientMessage()`
      - `GatewayFeatureConfig.CanForwardToGameServer()`
      - `GatewayFeatureConfig.IsIpAllowed(remoteAddr)` 等。
    - 不满足条件时：
      - 直接丢弃消息或返回统一错误码（由 Gateway 自己定义），不调用 UseCase（如 `ForwardToGameServerUseCase`）。
  - 示例（伪代码）：

  ```go
  // client_message_controller.go（示意）
  func (c *ClientMessageController) HandleMessage(ctx context.Context, conn network.IConnection, msg *network.Message) error {
      if !c.featureCfg.CanAcceptClientMessage() {
          // 记录日志 + 计数器，直接返回，不调用 UseCase
          return nil
      }
      if !c.featureCfg.IsIpAllowed(conn.RemoteAddr().String()) {
          // 可以按需做黑名单计数/审计，这里同样不触达 UseCase
          return nil
      }
      // 1. 解析 Session / 构造 domain.Message
      // 2. 调用 ForwardToGameServerUseCase
  }
  ```

- **子系统开关配置接口（示意）**
  - 为避免 UseCase 反向依赖配置，开关接口应定义在 Adapter 层，并仅被 Controller / 进程级适配层持有：

  ```go
  // adapter/gateway/feature_config.go（示意）
  package gateway

  type GatewayFeatureConfig interface {
      CanAcceptClientMessage() bool
      CanForwardToGameServer() bool
      IsIpAllowed(remoteAddr string) bool
  }
  ```

  - 具体实现可以直接读取 `engine.Config` + 热更新配置（如 `gateway_feature.json`），但这些实现细节只存在于 Adapter/Infrastructure 层，对 UseCase 完全不可见。

- **落地方式二：进程级“GatewaySystemAdapter”做守门（适用于没有显式 Controller 的场景）**
  - 对于由 `engine.GatewayServer` 直接驱动的流程（例如 GameServer 回包循环 `dispatchGameServerMessages`），可抽象出进程级的“GatewaySystemAdapter”：
    - 对接当前 `SessionManager` / `GameClient` 等运行时组件。
    - 在调用 UseCase（例如 `ForwardToClientUseCase`）之前，检查转发系统是否开启、GameServer 连接状态是否健康。
  - 该适配层依然属于外圈（Adapter / Infrastructure），只向内注入 UseCase 所需依赖，不反向暴露给 UseCase。

- **与依赖方向的关系**
  - Gateway 的 UseCase（`CreateSession/ForwardToGameServer/ForwardToClient` 等）：
    - 只依赖 `SessionRepository`、`GameServerRPC`、`NetworkGateway` 等接口。
    - 不依赖 `GatewayServer`、`SessionManager`、`GameClient` 或任何具体“开关配置”类型。
  - 开关状态（包括是否允许接入/转发）全部封装在：
    - Controller 层：基于配置/开关决定是否调用 UseCase。
    - 进程级适配层：在定时/后台任务场景中决定是否驱动 UseCase。

#### 3.4.4 小结：Controller / SystemAdapter / UseCase 与系统开关的依赖关系

- **统一依赖方向**
  - System 开关（无论是 GameServer 的玩法系统，还是 Gateway 的“接入/转发/安全策略”子系统）都位于 Adapter / SystemAdapter / Infrastructure 外圈，由这些组件维护状态并对外提供只读视图或接口。
  - UseCase 永远向外依赖抽象接口（Repository / RPC / Gateway / PresenterAdapter），**不反向依赖任何“开关持有者”**。

- **统一调用原则**
  - **Controller / SystemAdapter / 进程级适配层** 负责决定“是否调用 UseCase”：  
    - 若系统未开启或当前请求不被允许（限流、灰度、维护等），在外圈直接返回错误或丢弃，不触达 UseCase。  
    - 若系统允许调用，则构造好 UseCase 的输入参数，交给 UseCase 处理纯业务逻辑。
  - Gateway 落地时，可简单理解为：  
    - “是否允许接入/转发？”——由 Controller / `GatewayFeatureConfig` / `GatewaySystemAdapter` 回答；  
    - “接入/转发以后具体做什么？”——由 UseCase 回答。

> 实施 Gateway 重构时，若出现“在 UseCase 里想拿配置或判断是否开启某个功能”的冲动，优先回到本小节检查：  
> - 能否在 Controller 层提前做掉？  
> - 或者能否抽成进程级适配层（类 SystemAdapter）做一次性决策？  
> 一旦 UseCase 需要感知“系统是否开启”，基本可以认为依赖方向已经开始反转，需要重新设计。

### 3.5 Gateway 系统开关落地步骤（与现有代码映射）

> 结合 `server/service/gateway` 现状，以下步骤可直接落地，保证“Controller / SystemAdapter / UseCase”各司其职，同时在入口层完成系统开关检查。

1. **声明运行时开关接口（Adapter 层）**  
   - 新建 `internel/adapter/gateway/feature_config.go`：  
     - 定义 `GatewayFeatureConfig`（`CanAcceptNewSession/CanForwardToGameServer/CanForwardToClient/IsIpAllowed`）。  
     - 依赖来源：`engine.Config` + 热更新配置（例如 `gateway_feature.json`），仅在 Adapter 层解析，向外暴露只读方法。  
     - 可附带 `Subscribe(func(GatewayFeatureSnapshot))`，供 Web 控制台/运维系统刷新；UseCase 不需要该能力。

2. **Controller 层前置检查**  
   - `adapter/controller/client_message_controller.go`：构造函数注入 `GatewayFeatureConfig`。  
   - `HandleMessage` 流程：  
     1. `if !featureCfg.CanAcceptNewSession()`：直接关闭连接或返回维护中提示。  
     2. `if !featureCfg.IsIpAllowed(conn.RemoteAddr().String())`：打日志 + 计数，丢弃消息。  
     3. `if !featureCfg.CanForwardToGameServer()`：允许 Session 存续，但拒绝转发（可回包“维护中”消息）。  
     4. 以上任一失败都**不调用 UseCase**，确保 UseCase 不感知开关。  
   - Controller 仍只依赖注入的接口（`CreateSessionUseCase`、`ForwardToGameServerUseCase`、`GatewayFeatureConfig`、`NetworkGateway`），不直接读取 `engine.Config`。

3. **进程级 SystemAdapter 守门（无显式 Controller 的链路）**  
   - `engine.GatewayServer.dispatchGameServerMessages` 等内部循环改为委托 `GatewaySystemAdapter`：  
     - `GatewaySystemAdapter` 位于 `internel/adapter/system`（与 GameServer 风格一致），管理进程级状态（GameClient 连接、Feature 开关快照、后台 goroutine）。  
     - 在驱动 `ForwardToClientUseCase` 前调用 `featureCfg.CanForwardToClient()`；当 GameServer 断连时可统一短路。  
     - SystemAdapter 可定期刷新 Feature 配置，并通过事件通道通知 Controller（例如广播“转发关闭”以便 Controller 主动断开连接）。  
   - 这样 Gateway 也具备“SystemAdapter 决定是否驱动 UseCase”的通用模式。

4. **依赖注入层集中装配**  
   - `internel/di/container.go` 增加 `featureCfg GatewayFeatureConfig` 字段，在构造函数中：  
     - `featureCfg := gateway.NewFeatureConfig(configProvider, hotReloadSource)`  
     - `clientMessageController := controller.NewClientMessageController(..., featureCfg)`  
     - `gatewaySystemAdapter := system.NewGatewaySystemAdapter(featureCfg, forwardToClientUseCase, gameServerMessageSubscriber)`  
   - `engine.GatewayServer` 只依赖接口（`ClientMessageController` / `GatewaySystemAdapter`），保持对内层的单向依赖。

5. **测试与校验**  
   - **Controller 层单测**：覆盖“关闭会话/转发开关”“IP 黑名单”等分支，确认 UseCase 未被调用（可用 Mock UseCase 断言）。  
   - **SystemAdapter 层单测**：模拟 GameServer 断线 + 开关关闭场景，确保 `ForwardToClientUseCase` 不被触发。  
   - **E2E 验证**：通过 example 客户端模拟三个阶段（正常 → 转发关闭 → 接入关闭），观察日志与错误码，验证依赖方向保持正确。

> 通过上述步骤，Gateway 可以在 **Controller / SystemAdapter 层完成“系统是否开启”的判断**，同时让 UseCase 继续承担“Session / 消息转发的纯业务规则”，实现与 GameServer/DungeonServer 一致的 Clean Architecture 约束。


## 4. 重构方案

### 4.1 阶段一：Entities 层重构

**目标：** 提取纯业务实体，移除所有框架依赖

#### 4.1.1 创建 Domain 实体（由现有 `clientnet.Session` 抽象升级）

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

#### 4.1.2 定义 Repository 接口（替代当前 `SessionManager` 里直接持有 map）

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

### 4.2 阶段二：Use Cases 层重构（Session / 转发用例拆分）

**目标：** 实现业务用例，依赖 Entities 和 Repository 接口

#### 4.2.1 创建 Use Case（映射现有流程）

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

### 4.3 阶段三：Interface Adapters 层重构（面向现有 `clientnet` / `gameserverlink` 的适配）

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

#### 4.3.3 Session Repository 实现（迁移 `SessionManager.sessions`）

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

## 5. 重构步骤（按现有代码的落地路线）

### 5.1 阶段一：基础结构搭建 + 现状抽象（建议先完成）

1. **创建目录结构**
   - 创建 `internel/domain/`、`internel/usecase/`、`internel/adapter/`、`internel/infrastructure/` 以及 `internel/di/`。
   - 将现有 `clientnet` / `gameserverlink` 视为“过渡 Adapter”，后续逐步迁移能力。
2. **提取 Domain/Repository 草稿**
   - 从 `clientnet/session.go` 抽象出 `domain.Session/SessionState/ConnType`，保持与现有字段语义一致。
   - 在 `internel/domain/repository/` 下定义 `SessionRepository` 等接口，并用内存实现包装当前 `SessionManager.sessions` 的 map。
3. **用例骨架**
   - 在 `internel/usecase/session` 中创建 `CreateSession/CloseSession/UpdateActivity` 的空实现，只保留参数/返回值与 TODO 注释。
   - 在 `internel/usecase/message` 中创建 `ForwardToGameServer` 用例空实现，用接口形式依赖 GameServerRPC。
4. **最小接线验证**
   - 不改动现有逻辑，仅在 `NewGatewayServer` 中初始化 DI 容器并实例化 UseCase/Repository（暂不使用），确保编译通过，为后续迁移留出接入点。

### 5.2 阶段二：核心链路迁移（Session 管理 + 消息转发）

1. **重构 Session 管理**
   - 将 `SessionManager.CreateSession/CloseSession/UpdateActivity/cleanupTimeoutSessions` 的业务逻辑迁移到 UseCase 层，通过 `SessionRepository` + `GameServerRPC` 完成持久化与 RPC 通知。
   - `SessionManager` 精简为基础设施组件，仅维护 `map[sessionID]*Session` 与 goroutine 生命周期，对外通过 Repository 接口暴露。
2. **重构消息转发**
   - 将 `ClientHandler.HandleMessage` 中“获取/创建 Session + 更新活跃时间 + 调用 ForwardClientMsg”的流程拆分为 UseCase 调用：
     - Controller 负责把 `network.Message` 转换为 `domain.Message`，并获取/绑定 `SessionID`。
     - UseCase 负责验证 Session 状态、更新活跃时间、调用 `GameServerRPC.ForwardMessage`。
   - 对 GameServer 回包链路做同样处理：为 `dispatchGameServerMessages` 引入 UseCase（例如 `ForwardToClientUseCase`），由 UseCase 决定是否丢弃消息和如何记日志。
3. **统一 NetworkGateway 抽象**
   - 引入 `NetworkGateway` 接口，把现在分散在 `ClientHandler` / `GatewayServer.startTCPServer` / `startWSServer` 中的连接操作统一封装，方便未来扩展监控/限流。

### 5.3 阶段三：安全与运维能力收口

1. **接入全局安全约束**
   - 按 `docs/服务端开发进度文档.md` 7.4 / 6.2 中“网关 / 副本接入安全加固”的约束，为 WebSocket 接入增加 IP 白名单、Origin 校验和握手 Token 校验，将策略（白名单、允许 Origin、签名算法）通过 UseCase 或 ConfigGateway 暴露。
   - 为 TCP/WS 接入层加上基础限流（Session 数量、防秒连/秒断、消息频率）的 UseCase 与测试用计数器。
2. **监控与日志**
   - 将当前散落在 `GatewayServer` / `SessionManager` / `ClientHandler` 中的日志统一抽象为 Gateway 侧的 Logger 接口，与全局 `IRequester` 约束对齐（方便打 SessionId / RemoteAddr）。
   - 增加基础指标收集：当前 Session 数、断线原因统计、被丢弃消息计数等。
3. **清理与文档**
   - 删除已迁移职责的旧实现（`clientnet.SessionManager` 中的业务逻辑、直接依赖 GameServer 的方法等），保留最小壳。
   - 完成 UseCase/Controller 层的单测与端到端联调脚本，并在本文件与 `docs/服务端开发进度文档_full.md` 中更新实现说明。

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

### 6.2 在 GatewayServer 中使用（与现有 `engine.GatewayServer` 集成）

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

