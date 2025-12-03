## 1. 文档目的与阅读顺序

- **文档目的**：梳理 `server/service/gameserver` 的目录结构与主要调用关系，说明它是如何按照 Clean Architecture（清洁架构）原则组织依赖与职责的，方便新人或后续维护者快速上手。
- **适用读者**：需要在 GameServer 侧新增/重构业务、排查线上问题、或评审架构设计的开发者。
- **推荐阅读顺序**：
  1. **第 2 章 目录总览**：先对整体目录有全局认识；
  2. **第 3 章 分层与依赖方向**：理解 Clean Architecture 在本项目中的落地方式；
  3. **第 4 章 核心调用链路**：了解请求从 Gateway → GameServer → DungeonServer 的完整流向；
  4. **第 5 章 与 Clean Architecture 的对应关系**：需要做新系统/重构时重点参考。

> 建议在阅读本文件前，先通读《`docs/服务端开发进度文档.md`》第 0/7 章及《`docs/gameserver_CleanArchitecture重构文档.md`》第 1/3/4 章，本文件视其为前置知识，不再重复理念性内容。

---

## 2. 目录总览（逻辑分层视角）

以下仅列出与 Clean Architecture 强相关的目录，并按“内层 → 外层”的依赖方向自下而上说明：

```text
server/service/gameserver
├── main.go
├── requires.go
└── internel
    ├── domain/          # 领域实体与领域服务（Entities）
    ├── usecase/         # 业务用例与用例接口（Use Cases）
    ├── adapter/         # 控制器 / Presenter / Gateway / System Adapter / Context（Interface Adapters）
    ├── infrastructure/  # DungeonServer / Gateway / 事件 等基础设施封装（Frameworks & Drivers）
    ├── app/             # PlayerActor / PublicActor / Engine / Manager 等应用层组合
    ├── core/            # gshare 与通用接口定义（横切基础设施）
    ├── di/              # 依赖注入容器
    └── ...
```

各子目录职责简述：

- **`domain/`**：只关心“游戏世界中的概念”，如背包、邮件等；不依赖具体网络/数据库实现。
- **`usecase/`**：定义并实现具体业务用例（AddItem、BuyItem、EnterDungeon、SendChat 等），通过接口与外部世界交互。
- **`adapter/`**：承上启下：
  - `controller/`：处理 C2S/S2C 与内部 RPC 请求，调用 Use Case；
  - `presenter/`：把用例输出转换为协议消息；
  - `gateway/`：实现 Repository / Gateway 接口，调用 `database`、`gatewaylink`、`dungeonserverlink` 等；
  - `system/`：把 PlayerActor 生命周期（OnInit/RunOne/OnLogout 等）适配到新架构；
  - `context/`：封装每次用例调用所需的上下文（RoleID、SessionID、logger 等）。
- **`infrastructure/`**：对 Actor 框架外的“其他服务”进行适配，如 DungeonServer RPC、Gateway 会话映射、事件总线。
- **`app/`**：组合 Actor、Adapter 与 Use Case，形成真正运行的 GameServer 应用（PlayerActor/PublicActor/Engine）。
- **`core/`**：`gshare` 和通用接口 `iface`，抽象出 GameServer 公共能力。
- **`di/`**：提供依赖注入容器，将接口与具体实现绑定。

---

## 3. 分层与依赖方向

### 3.1 逻辑分层与 Go 包依赖

总体分层与《`docs/gameserver_CleanArchitecture重构文档.md`》第 3 章一致，这里结合当前目录做一次“落地化”对应：

```text
Entities（领域层）
  └── internel/domain

Use Cases（用例层）
  └── internel/usecase

Interface Adapters（接口适配层）
  └── internel/adapter
        ├── controller
        ├── presenter
        ├── gateway
        ├── system
        └── context

Frameworks & Drivers（框架与驱动层）
  └── internel/infrastructure
  └── internel/app
  └── internel/core
  └── server/internal/*（Actor / database / network / servertime 等）
```

**依赖方向约束（编译期）**：

- `domain` **不依赖** `usecase/adapter/infrastructure/app`；
- `usecase` **只能依赖**：
  - `domain`
  - `usecase/interfaces`（同层接口定义）
  - 标准库
- `adapter` **可以依赖**：
  - `domain`（读取/写入领域对象）
  - `usecase`（调用用例）
  - `usecase/interfaces`（实现端）
  - `infrastructure` / `core` / `server/internal/*`（通过 Gateway/Adapter 与框架对接）
- `infrastructure`、`app`、`core` 在物理结构上处于最外层，反向依赖内层仅通过接口，不会让内层出现“向外层 import”。

换句话说：**业务规则（Use Case + Domain）只依赖抽象，所有具体实现都在 Adapter / Infrastructure / App 层完成**，从而满足 Clean Architecture 的依赖规则。

### 3.2 目录级别依赖示意

```text
domain              ←  usecase           ←  adapter              ←  app
  ↑                       ↑                   ↑                      ↑
  └──── server/internal/protocol  ←  presenter/gateway  ←  infrastructure/core
```

- `presenter` 依赖协议定义（`proto/csproto/*.proto` 与生成的 `server/internal/protocol/*.pb.go`），但不把协议类型泄漏到 Use Case 内部；
- `gateway` 实现 `usecase/interfaces` 与 `domain/repository` 中的接口，并使用 `database` / `gatewaylink` / `dungeonserverlink` 等外部设施；
- `app` 把 `adapter/system` 注册到 PlayerActor / PublicActor 上，实现 **“系统生命周期 → 用例调用”** 的桥接。

---

## 4. 各子目录职责与调用关系

### 4.1 `internel/domain` – 领域模型

- **位置**：`internel/domain/*`
- **代表文件**：
  - `bag/`：背包相关领域对象与纯业务操作；
  - `chat/chat.go`：聊天领域模型（消息、频道、频率限制相关值对象）；
  - `mail/mail.go` 等；
  - `repository/player_repository.go`：领域侧的玩家数据访问接口。

**依赖特点**：

- 只依赖标准库与 `server/internal/protocol` 中的基础类型（如 `PlayerRoleBinaryData`），**不感知 Actor、网络、数据库等细节**。
- 所有需要持久化或跨进程通信的能力，均通过接口（如 `PlayerRepository`、领域层定义的服务接口）对外暴露。

### 4.2 `internel/usecase` – 业务用例

- **位置**：`internel/usecase/*`
- **代表文件/目录**：
  - `bag/*`：`add_item.go`、`remove_item.go`、`has_item.go`；
  - `money/*`：`add_money.go`、`consume_money.go`；
  - `equip/*`、`attr/*`、`skill/*`、`quest/*`、`fuben/*`、`shop/*`、`recycle/*`、`item_use/*` 等；
  - `chat/*`、`mail/*` 等；
  - `interfaces/*`：`config.go`、`dungeon.go`、`public_actor_gateway.go`、`level.go`、`consume.go` 等接口定义。

**核心调用关系**：

- 每个用例包都通过构造函数注入依赖，例如 `PlayerRepository`、`ConfigManager`、`DungeonServerGateway`、`PublicActorGateway` 等接口；
- 用例内部只做业务决策：检查前置条件、修改领域对象、调用接口方法（例如发奖励、发消息）；
- 不直接触碰：
  - 网络发送（不 import `gatewaylink`）；
  - Actor 上下文（不 import PlayerActor 类型）；
  - 数据库实现（不 import `database` 包）。

这保证了 **Use Case 层可以通过 Mock 接口独立单元测试**，是 Clean Architecture 的关键收益之一。

### 4.3 `internel/adapter/controller` – 协议与入口控制器

- **位置**：`internel/adapter/controller/*`
- **代表文件**：
  - `*_controller.go`：如 `bag_controller.go`、`money_controller.go`、`skill_controller.go`、`quest_controller.go`、`fuben_controller.go`、`shop_controller.go`、`recycle_controller.go`；
  - `*_controller_init.go`：对应协议/RPC 注册入口；
  - `chat_controller.go`；
  - `item_use_controller.go`、`mail_controller.go`（如有）等；
  - `protocol_router_controller.go`：通用协议路由控制器。

**职责与调用链**：

1. **协议信息进入点**  
   - GameServer 收到 Gateway 转发的 `C2S` 协议后，先进入 `player_network`，再由 `protocol_router_controller` 解析并分发到具体业务控制器。
2. **控制器内部逻辑**：
   - 从上下文中读取 `RoleID`、`SessionID` 等信息（由 `adapter/context` 提供）；
   - 把协议参数转换为 Use Case 所需的输入 DTO；
   - 调用对应 Use Case（如 `bag.AddItem`、`fuben.EnterDungeon` 等）；
   - 调用对应 `presenter` 构建响应，并通过 `NetworkGateway` 发送给客户端或上游服务。
3. **RPC / 转发逻辑**：
   - 与 DungeonServer 的 RPC 往来统一通过 `DungeonServerGateway` 适配，而不是在控制器中直接依赖 `dungeonserverlink`。

> 小结：Controller 是 **“协议世界” 与 **“业务世界（Use Case）”** 之间的唯一入口，确保协议变更不会直接影响 Use Case 的接口设计。

### 4.4 `internel/adapter/presenter` – 响应构建器

- **位置**：`internel/adapter/presenter/*`
- **代表文件**：`bag_presenter.go`、`money_presenter.go`、`equip_presenter.go`、`skill_presenter.go`、`quest_presenter.go`、`fuben_presenter.go`、`shop_presenter.go`、`recycle_presenter.go`、`chat_presenter.go`、`item_use_presenter.go` 等。

**职责**：

- 将 Use Case 的结果（领域对象、简单结构体）转换为 `S2C*` 或内部 RPC 协议；
- 统一从 `NetworkGateway` 获取发送能力，而非直接操作底层连接；
- 集中管理下行协议格式，减轻 Use Case 对协议演化的感知。

这样 Use Case 即使改变内部实现或返回结构，只要 Presenter 保持契约不变，客户端就能稳态运行。

### 4.5 `internel/adapter/gateway` – 基础设施适配器

- **位置**：`internel/adapter/gateway/*`
- **代表文件**：
  - `player_gateway.go`：实现 `PlayerRepository` 接口，负责从 `PlayerRoleBinaryData` 读取/写入玩家数据；
  - `network_gateway.go`：封装对 Gateway 会话的网络发送（C2S/S2C）；
  - `public_actor_gateway.go`：封装向 `PublicActor` 发送消息；
  - `dungeon_server_gateway.go`：封装 GameServer ↔ DungeonServer 的 RPC 调用与协议注册；
  - `config_gateway.go`：实现 `ConfigManager` 接口，从 `jsonconf` 读取配置；
  - `blacklist_repository.go` 等：实现社交等系统的仓储接口。

**调用方向**：

- **实现端**：实现 `usecase/interfaces` 与 `domain/repository` 中定义的接口；
- **被谁调用**：主要被 Use Case、System Adapter、Controller/Presenter 使用；
- **向下依赖**：
  - `server/internal/database`：进行真实的 DB 操作；
  - `internel/infrastructure/gatewaylink`、`dungeonserverlink`：进行网络转发；
  - `server/internal/event`、`jsonconf` 等。

通过 Gateway，业务用例“看到”的只是接口；具体用什么 DB/消息队列/网络协议，都可以在 Adapter 层替换，而不需要改动 Use Case。

### 4.6 `internel/adapter/system` – 系统生命周期适配层

- **位置**：`internel/adapter/system/*`
- **代表含义**：
  - 每一个旧的 `entitysystem/*_sys.go`，在完成 Clean Architecture 重构后，都有一个同名或相近名称的 System Adapter 目录；
  - 示例：`adapter/system/bag/`、`.../money/`、`.../equip/`、`.../attr/`、`.../skill/`、`.../quest/`、`.../fuben/`、`.../item_use/`、`.../shop/`、`.../recycle/` 等。

**职责**：

- 负责把 PlayerActor 的生命周期事件（Init/RunOne/OnNewDay/OnLogout）转化为对 Use Case 的调用；
- 维护系统本地缓存（如属性脏标记、加成缓存等），但不直接持久化数据；
- 完成与 Actor/事件总线的粘合：在 `init()` 中注册系统工厂、订阅事件等。

**关键点**：System Adapter **只向内依赖 Use Case 与 Domain**，向外只依赖 Actor 框架提供的上下文接口，**不直接与 Gateway / DungeonServer 等外部设施打交道**（这类调用都通过 Use Case 接口间接完成）。

### 4.7 `internel/adapter/context` – 上下文辅助

- **位置**：`internel/adapter/context/context_helper.go`
- **作用**：
  - 封装在 PlayerActor/PublicActor 中运行 Use Case 时所需的上下文（如角色 ID、Session ID、日志请求器、DI 容器引用等）；
  - 提供统一入口（例如 `GetBagSys(ctx)`、`GetLevelSys(ctx)`）获取系统适配器；
  - 避免在业务代码中到处显式地通过全局变量或单例查找依赖。

---

## 5. `internel/infrastructure` 与 `internel/app` – 框架层组合

### 5.1 `internel/infrastructure` – 外部服务与事件封装

- **位置**：`internel/infrastructure/*`
- **子目录**：
  - `dungeonserverlink/`：`dungeon_cli.go`、`handler.go`、`protocol_manager.go`，封装与 DungeonServer 的 RPC 通信；
  - `gatewaylink/`：`handler.go`、`sender.go`，封装与 Gateway 的 Session 映射与消息转发；
  - `gevent/`：事件总线封装，定义 GameServer 内部事件枚举与发布/订阅入口。

**与 Adapter 的关系**：

- `adapter/gateway/dungeon_server_gateway.go` 会持有 `infrastructure/dungeonserverlink` 的引用，对 Use Case 暴露统一的 `DungeonServerGateway` 接口；
- `adapter/gateway/network_gateway.go` 通过 `infrastructure/gatewaylink` 发送消息给客户端；
- `adapter/event/event_adapter.go` 与 `infrastructure/gevent` 协同工作，实现 Use Case 侧的事件发布与订阅能力。

### 5.2 `internel/app` – PlayerActor / PublicActor 与 Engine

- **位置**：`internel/app/*`
- **子目录**：
  - `engine/`：`config.go`、`server.go`、`message_registry.go`，负责 GameServer 启动、配置加载与 Actor 引擎初始化；
  - `manager/`：`role_mgr.go` 等，负责玩家管理与关服刷盘；
  - `playeractor/`：`adapter.go`、`handler.go`、`entity/*`、`entitysystem/*`（仅保留未迁移或极少量过渡代码）；
  - `publicactor/`：`public_role.go` 及各类 `public_role_*` 文件，承载 PublicActor 的全局社交/经济逻辑。

**调用关系（高层视角）**：

1. `main.go` → `engine/server.go`：初始化配置、日志、Actor 系统。
2. Engine 创建 `PlayerActor` 与 `PublicActor`：
   - PlayerActor 在初始化时会通过 `adapter/system` 注册好的工厂创建各个系统适配器；
   - PublicActor 通过自身的 handler 与 `PublicActorGateway`、`OfflineDataManager` 集成。
3. 请求流向：
   - **客户端 C2S**：Client → Gateway → GameServer → `player_network` → `protocol_router_controller` → 某业务 `*_controller` → Use Case → Presenter → `NetworkGateway` → Gateway → Client；
   - **GameServer ↔ DungeonServer**：Use Case 通过 `DungeonServerGateway` 调用 RPC，底层由 `dungeonserverlink` 实现；
   - **GameServer ↔ PublicActor**：Use Case 或 System Adapter 通过 `PublicActorGateway` 发送消息，由 PublicActor 在 `publicactor/*` 中处理。

---

## 6. `internel/core` 与 `di` – 公共接口与依赖注入

### 6.1 `internel/core`

- **位置**：`internel/core/*`
- **主要内容**：
  - `gshare/*`：封装游戏服共享能力，如 `srv.go` 中的开服信息、`log_helper.go` 中的日志上下文辅助、`message_sender.go` 中的玩家消息发送工具；
  - `iface/*`：`irole.go`、`iserver.go`、`isystem.go`，定义 PlayerRole、系统接口等抽象。

**在 Clean Architecture 中的定位**：

- 提供一个“横切”的基础层，被 Use Case/Adapter/App 共同依赖；
- 通过接口抽象避免上层直接依赖具体实现（例如 PlayerRole 的具体 struct）。

### 6.2 `internel/di`

- **位置**：`internel/di/container.go`
- **职责**：
  - 创建并维护依赖注入容器，将接口类型（如 `PlayerRepository`、`ConfigManager`、`DungeonServerGateway` 等）与具体 Adapter 实现绑定；
  - 为 Controller/System Adapter 提供一种“按需获取依赖”的机制，减少手工 new 与全局变量。

**与 Clean Architecture 的关系**：

- DI 容器本身属于框架层实现，但它通过注册接口 → 实现的方式，保证了 **内层（Use Case/Domain）只面向接口编程**；
- 编译依赖仍然遵循内向原则：Use Case 不 import DI 容器，只在外层初始化时被“喂入”具体实现。

---

## 7. 如何满足 Clean Architecture 规范（总结视图）

结合以上目录与调用关系，GameServer 目前通过以下几点满足 Clean Architecture 的核心要求：

- **依赖倒置**：  
  - 内层（Domain/Use Case）定义 `Repository`、`Gateway`、`ConfigManager`、`PublicActorGateway`、`DungeonServerGateway` 等接口；
  - 外层（Adapter/Infrastructure/App）实现这些接口，并在 DI 容器中完成绑定。

- **依赖方向内聚**：  
  - `domain` 不依赖任何外层代码；`usecase` 只依赖 `domain` 与接口定义；  
  - `adapter` 负责所有协议处理、数据转换、外部系统调用；  
  - `infrastructure/app` 组合 Actor 框架、网络与 Adapter，不反向侵入内层。

- **用例可测试性**：  
  - Use Case 层只依赖接口，适合使用 Mock 做单元测试；  
  - Controller/Presenter 属于集成层，可以通过端到端或组件级测试验证。

- **避免循环依赖**：  
  - 不同系统间（如 Bag/Equip/Attr/Fuben/Quest/Shop/Recycle 等）通过 `usecase/interfaces` 中定义的 UseCase 接口交互，而不是直接相互 import；  
  - 通过 `usecaseadapter`（如 `bag_use_case_adapter.go`、`level_use_case_adapter.go` 等）打破编译期环依赖。

---

## 8. 后续扩展与阅读建议

- **新增系统时的自检清单**（与 `docs/gameserver_CleanArchitecture重构文档.md` 0 章保持一致）：
  - [ ] 是否在 `domain/` 中定义了清晰的领域模型与必要的 Repository 接口？
  - [ ] 是否在 `usecase/` 中实现了业务用例，且只依赖接口与领域模型？
  - [ ] 是否在 `adapter/gateway` 中实现了所有外部依赖（数据库/网络/配置等）的适配？
  - [ ] 是否在 `adapter/controller` 与 `adapter/presenter` 中处理了协议编解码与响应构建？
  - [ ] 是否为 PlayerActor/PublicActor 提供了对应的 `adapter/system` 实现，并在 `*_init.go` 中注册？
  - [ ] 是否避免了跨系统的直接依赖，通过 `usecase/interfaces` + `usecaseadapter` 进行解耦？

- **推荐配套文档**：
  - 《`docs/服务端开发进度文档.md`》：了解当前重构进度、已完成功能与注意事项；
  - 《`docs/gameserver_CleanArchitecture重构文档.md`》：理解更细致的分层设计和重构路线；
  - 《`docs/统一数据访问和网络发送验证.md`》：确认 PlayerRepository/NetworkGateway 等关键抽象的一致性；
  - 《`docs/gameserver_adapter_system演进规划.md`》：针对 `adapter/system` 的优化演进路线与分阶段复选框清单，适合在精简 SystemAdapter 时按步骤推进；
  - 各系统的专用文档（如属性、离线数据管理器、玩家消息系统等），用于深入理解局部设计。


