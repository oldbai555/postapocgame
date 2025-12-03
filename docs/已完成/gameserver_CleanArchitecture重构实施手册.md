# GameServer Clean Architecture 重构实施手册

## 0. 文档目的与阅读须知

- **文档定位**：面向 `server/service/gameserver` 的 Clean Architecture 重构工作指引，聚焦代码目录梳理、阶段路线、检查清单与关键文件，确保多人协作时有统一的执行标准。  
- **适用角色**：重构负责人、子系统 Owner、Code Review 参与者、测试与运维同学。  
- **阅读顺序**：建议先通读第 1~3 章掌握架构背景，再依据自身负责的子系统查阅第 4~6 章的任务清单。  
- **更新规则**：当阶段任务完成或新增约束时务必同步 `docs/服务端开发进度文档.md` 的第 4/6/7/8 章，并在本文追加变更记录。  
- **命名约定**：文中 `[路径]` 均指向 `server/service/gameserver/internel/` 下的相对目录，除非特别说明。

---

## 1. 当前架构速览

### 1.1 运行拓扑（概括）

```
Gateway (Session) → GameServer PlayerActor (per player) / PublicActor (global)
                           ↘ DungeonServerGateway (Async RPC) ↔ DungeonServer
```

- PlayerActor：一玩家一 Actor，所有玩家系统 RunOne/事件均在主循环串行执行。  
- PublicActor：承载社交/全局经济，提供在线映射与异步广播。  
- DungeonServerGateway：统一封装 GameServer ↔ DungeonServer 的 RPC/协议注册。  
- GatewayLink：负责 Session 映射与消息下发。

### 1.2 目录-分层映射

| 层级 | 目录/文件 | 说明 |
| --- | --- | --- |
| Entities/Domain | `domain/{bag,level,friend,guild,...}` | 领域模型、值对象、纯业务规则 |
| Use Cases | `usecase/{system}/`、`usecase/interfaces/*` | 业务用例、接口契约、跨系统依赖抽象 |
| Interface Adapters | `adapter/controller/*`、`adapter/presenter/*`、`adapter/system/*`、`adapter/gateway/*`、`adapter/context/context_helper.go` | 协议入口、回包装配、生命周期胶水、外部资源适配 |
| Frameworks & Drivers | `app/*`、`core/*`、`infrastructure/*`、`main.go` | Actor 框架、网络、RPC、事件、DI 容器、进程入口 |
| DI & 运行支撑 | `di/container.go`、`adapter/context/context_helper.go` | 提供系统实例、UseCase 依赖注入、上下文 Helper |
| Legacy/待清理 | `playeractor/entitysystem/*`（仅保留 AntiCheat/GMSys/MessageSys 等） | 重构完成后需逐步删除或改造成 Adapter |

### 1.3 关键执行链路

1. **协议入口**：`adapter/controller/*_controller_init.go` 在 OnSrvStart 时注册协议 → Controller 解析请求 → 调用 UseCase。  
2. **数据访问**：UseCase 只依赖 `usecase/interfaces` 中的 Repository/Gateway。默认通过 `adapter/gateway/{player,network,config,...}` 实现。  
3. **SystemAdapter**：负责 Actor 生命周期（Init/Login/RunOne/OnNewDay/OnNewWeek）调度，内部禁止落入业务规则。  
4. **事件**：`adapter/event/event_adapter.go` 统一把 gevent 的玩家事件映射给 SystemAdapter/UseCase。  
5. **存储**：玩家数据通过 `PlayerGateway` 操作 `PlayerRoleBinaryData`，公共数据通过 PublicActor 自己的 DAO。  
6. **跨服务 RPC**：全部通过 `adapter/gateway/dungeon_server_gateway.go` 统一注册与发送。

---

## 2. Clean Architecture 分层与依赖约束

1. **依赖方向**：允许内层依赖内层，同层互斥；外层通过接口适配进入内层。  
2. **接口定义位置**：全部放在 `usecase/interfaces`，包括 Repository、Gateway、EventPublisher、ConfigManager、时间服务等。  
3. **DI 容器**：`di/container.go` 负责注册所有 Adapter/UseCase/SystemAdapter 工厂，`adapter/context` 暴露 `GetXxx(ctx)` Helper，Actor 系统通过 `BaseSystemAdapter.Resolve()` 获取实例。  
4. **协作规则**：  
   - Controller ↔ UseCase：只传输 DTO/值对象；错误码由 Presenter 统一处理。  
   - UseCase ↔ Repository/Gateway：禁止直接触碰 Actor/网络/数据库。  
   - SystemAdapter ↔ UseCase：SystemAdapter 只负责在合适的时机调度 UseCase，禁止编写校验/奖励等业务逻辑。  
   - PublicActor ↔ PlayerActor：统一经 `PublicActorGateway`/`PlayerRole.sendPublicActorMessage`。  
5. **测试策略**：UseCase 层使用接口 Mock；Controller/Presenter 使用 in-memory Adapter；SystemAdapter 侧以脚本或 integration test 验证生命周期调度。  
6. **编解码规范**：Proto 变更需同步 `proto/genproto.sh`，Controller 禁止重复解码同一消息。

---

## 3. 重构阶段路线（建议按序执行）

> ✅ 表示已完成，⏳ 表示进行中，☐ 表示待做；请在 `docs/服务端开发进度文档.md` 的 6.1 小节同步状态。

### 阶段 0：基线确认（进行中）
- [⏳] ☑ 审核 `playeractor/entitysystem` 中剩余模块（AntiCheat、MessageSys、GM Tools）及其依赖范围。  
- [⏳] ☑ 整理 `usecase/interfaces` 中未使用的接口，确认是否删除或回收。  
- [✅] ☑ 建立统一日志、时间、配置访问规范（参考第 7 章）。

### 阶段 1：基础设施 & DI 整体化
- [✅] PlayerGateway、NetworkGateway、ConfigGateway、PublicActorGateway、DungeonServerGateway、EventAdapter。  
- [✅] BaseSystemAdapter + Context Helper。  
- [☐] 审核 `adapter/gateway/player_gateway.go` 是否完整覆盖剩余玩家字段（特别是 GM/MessageSys 仍在访问的字段）；缺失字段需通过接口补齐。  
- [☐] 为 Gateway 提供统一的健康检查（如连接状态、自检命令），输出到 `docs/SystemAdapter验证清单.md`。

### 阶段 2：核心成长/经济
- [✅] Bag/Money/Level/Equip/Attr Clean Architecture 化。  
- [⏳] Attr Calculator 抽象与 DungeonServer 同步链路验证，整理于 `docs/属性系统重构文档.md`。  
- [☐] 汇总 Bag/Money/Equip 的并发/一致性断言用例，确保 `AddItemTx`/`UpdateBalanceTx` 在 Actor 单线程下无共享引用副作用。

### 阶段 3：玩法系统
- [✅] ItemUse/Skill/Fuben/Quest/Shop/Recycle Controllers + UseCases + SystemAdapter。  
- [☐] Quest 与 DailyActivity 解耦后，需要把活跃度事件改成 Hook（若后续恢复日常系统）。  
- [☐] ItemUse HP/MP 同步至 DungeonServer（通过事件或 Gateway）——见服务端开发进度文档 6.1 TODO。

### 阶段 4：社交经济
- [✅] Friend/Guild/Chat/Auction Clean Architecture 化。  
- [☐] PublicActor 离线快照：为各社交用例补齐 OfflineData 读取/写入的自动化检查。  
- [☐] BlacklistRepository 扩展审计字段，写入 `database/blacklist.go`。

### 阶段 5：辅助系统
- [✅] MailSys Clean Architecture；Vip/DailyActivity 移除。  
- [⏳] MessageSys：生命周期已迁移到 Adapter，需要完成阶段四"控制台/监控与过期策略"。  
- [☐] AntiCheatSys：待拆分 Domain + UseCase + SystemAdapter，补齐频率检测/封禁用例。

### 阶段 6：Legacy 清理与防退化
- [✅] ProtocolRouterController 接管所有 C2S → DungeonServer 转发。  
- [☐] 在 `playeractor/entitysystem` 下保留的文件逐一增添 deprecate 注释与替代方案链接。  
- [☐] 引入 `go vet` 自定义检查或脚本，确保没有新的 `entitysystem/*` 引用被添加。  
- [⏳] SystemAdapter 防退化机制：继续扩展 Code Review 清单与自动化脚本（lint grep）。
- [⏳] 按《`docs/gameserver_兼容代码清理规划.md`》执行 AntiCheat/MessageSys/GM Tools 等兼容层迁移与删除。

### 阶段 7：测试、监控、文档
- [⏳] UseCase 单测补齐（Equip/Fuben/Quest/Shop/Recycle/Skill/ItemUse 等）。  
- [☐] Controller/Presenter 集成脚本（自动化冒烟流程：登录→开背包→副本→聊天）。  
- [☐] `docs/服务端开发进度文档.md` + `docs/gameserver_CleanArchitecture重构文档.md` 同步新的约束、关键位点。  
- [☐] 为 Gateway/SystemAdapter 增加 Prometheus/日志指标（调度耗时、失败次数）。

---

## 4. 子系统梳理（按域拆解）

### 4.1 玩家生命周期与存储
- **入口**：`adapter/system/*_system_adapter.go` 内的 `OnInit/OnRoleLogin/OnRoleLogout/OnNewDay/RunOne`。  
- **数据源**：`PlayerGateway` 提供 `GetBinaryData()` 共享引用；落库由 `PlayerRole` 的存盘周期统一处理。  
- **任务**：
  - [☐] 在 `PlayerGateway` 层统一封装 `MarkDirty(SystemId)`，替代各系统临时字段。  
  - [☐] 补齐 `PlayerRoleBinaryData` 的 proto 注释 & 迁移脚本。  
  - [☐] `PlayerGateway` 应提供事务式 API（例如 `WithMutableBag(func(*BagData) error)`）以减少重复的深拷贝代码。

### 4.2 经济/成长系统（Bag/Money/Level/Equip/Attr）
- **领域对象**：`domain/bag/item.go`、`domain/money/currency.go`、`domain/level/level.go`、`domain/equip/equip.go`、`domain/attr/value.go`。  
- **UseCase**：`usecase/{bag,money,level,equip,attr}` 提供 Add/Consume/Calc 等操作。  
- **控制器**：`adapter/controller/{bag,money,equip,skill}_controller.go` + Presenter。  
- **SystemAdapter**：对应目录下的 helper/init/go 文件提供 `GetXxxSys(ctx)`。  
- **任务**：
  - [☐] Attr 系统需补齐 `CalculateAllAttrs` 的并发测试（Mock DungeonServerGateway）。  
  - [☐] Money/Bag Equip 之间的 UseCase 接口梳理：将重复的校验逻辑移动到 `usecase/interfaces`.  
  - [☐] 统一 Presenter 推送（背包/货币变化走 `adapter/presenter/push_helpers.go`）。

### 4.3 玩法系统（ItemUse/Skill/Fuben/Quest/Shop/Recycle）
- **核心文件**：  
  - ItemUse：`usecase/item_use/use_item.go`、`adapter/controller/item_use_controller.go`、`adapter/system/item_use/*`。  
  - Skill：`usecase/skill/{learn_skill,upgrade_skill}.go`、`adapter/controller/skill_controller.go`、`adapter/system/skill/*`。  
  - Fuben：`usecase/fuben/{enter_dungeon,settle_dungeon}.go`、`adapter/controller/fuben_controller.go`、`adapter/system/fuben/*`。  
  - Quest：`usecase/quest/*`、`adapter/controller/quest_controller.go`、`adapter/system/quest/*`。  
  - Shop：`usecase/shop/buy_item.go`、`adapter/controller/shop_controller.go`。  
  - Recycle：`usecase/recycle/recycle_item.go`、`adapter/controller/recycle_controller.go`。  
- **任务**：
  - [☐] ItemUse HP/MP 同步（在 Presenter 或 SystemAdapter 中调用 DungeonServerGateway）。  
  - [☐] Fuben 奖励链路：确保 RewardUseCase/Presenter 覆盖所有奖励类型。  
  - [☐] Quest 日/周任务刷新 → 创建独立 UseCase `RefreshQuestBucketUseCase` 并在测试中模拟 OnNewDay/OnNewWeek。  
  - [☐] Shop purchaseCounters 是否需要持久化，若需要则由 PlayerGateway 承担。

### 4.4 社交/公共系统（Friend/Guild/Chat/Auction/PublicActor）
- **公共入口**：`adapter/controller/{friend,guild,chat,auction}_controller.go`；全部走 `PublicActorGateway`。  
- **PublicActor**：`app/publicactor/*.go` + `publicactor/offlinedata/*`；仍然是单 Actor。  
- **任务**：
  - [☐] 为每个社交 UseCase 添加 OfflineData 缓存策略（只读/写时机）。  
  - [☐] Friend 黑名单接口→ `adapter/gateway/blacklist_repository.go` 引入审计日志。  
  - [☐] Auction 公共订单与玩家背包的原子性验证（引入事务或两阶段提交）。

### 4.5 辅助/框架系统（GM/AntiCheat/MessageSys/Timesync）
- **GM**：`adapter/system/gm/*` + `adapter/controller/gm_controller.go`；GM tools 仍留在 `adapter/system/gm/gm_tools.go`。  
- **AntiCheat**：仍在 `playeractor/entitysystem/anti_cheat_sys.go`（待 Clean 化）。  
- **MessageSys**：`adapter/system/message_system_adapter.go` + `engine/message_registry.go`。  
- **Timesync**：`internel/timesync/*`。  
- **任务**：
  - [☐] AntiCheat Domain/UseCase 拆解，定义 `usecase/anti_cheat`。  
  - [☐] MessageSys 阶段四：监控指标、过期清理配置化。  
  - [☐] GM 权限校验 → Controller 层统一接入权限/审计（参考进度文档 7.4）。

---

## 5. 跨模块能力与注意事项

1. **DungeonServerGateway**：  
   - `RegisterProtocols`、`RegisterRPCHandler`、`AsyncCall` 是唯一入口。  
   - 新增 RPC 时需在本文 & 进度文档登记调用方/上下文策略。  
2. **PublicActorGateway**：  
   - 提供 `SendMessage(ctx, msgId, payload)`、登录/离线通知，禁止直接调用 `gshare.SendPublicMessageAsync`。  
3. **ProtocolRouterController**：  
   - 任何需要透传至 DungeonServer 的客户端协议必须在 `adapter/router/protocol_router_controller.go` 注册，禁止散落在 Controller 内。  
4. **事件系统**：  
   - `adapter/event/event_adapter.go` 会把 gevent 的 OnNewDay/OnNewWeek/OnSrvStart 等事件分发给 SystemAdapter。  
5. **Servertime**：  
   - 统一通过 `server/internal/servertime`，禁止 `time.Now()`；若需要 mock 时间，使用接口注入。  
6. **日志上下文**：  
   - `core/gshare/log_helper.go` 提供 requester 注入示例；Controller/UseCase 打日志时需带 RoleId/SessionId。  
7. **配置访问**：  
   - 所有配置读取通过 `ConfigGateway`，支持热更新/缓存；UseCase 内禁止直接访问 `jsonconf`。  
8. **数据一致性**：  
   - PlayerGateway 返回的是共享引用，修改时务必调用 `MarkDirty`；禁止复制 `BinaryData` 导致存盘失效。  
9. **接口命名**：  
   - UseCase 接口统一为 `XxxUseCase`、Repository 为 `XxxRepository`、Gateway 为 `XxxGateway`；SystemAdapter helper 为 `GetXxxSys(ctx)`。

---

## 6. 测试与验收

| 维度 | 验收项 | 工具/脚本 |
| --- | --- | --- |
| UseCase 单测 | Bag Add/Remove、Money Tx、Fuben Enter/Settle、Quest Submit、Skill Upgrade 等 | `go test ./server/service/gameserver/internel/usecase/...` |
| Controller 集成 | `go test` + `server/example` 脚本：`register→login→enter→bag→shop→quest` | `server/example/internal/script/*.go` |
| SystemAdapter | 自定义脚本触发 OnInit/OnRunOne/OnNewDay，验证 UseCase 调度顺序 | `docs/SystemAdapter验证清单.md` |
| RPC 回路 | 使用本地 `dungeonserver.exe` + `gameserver.exe`，配合 `server/example` 执行副本进入/结算 | `scripts/run_dungeon_and_game.ps1`（可自建） |
| 性能基准 | Actor RunOne 耗时、PlayerGateway 存盘延迟 | 建议集成 `pprof` 或自定义统计 |
| 文档同步 | 每次需求完成后更新 `docs/服务端开发进度文档.md` 4/6/7/8 章 + 本手册变更记录 | 手工校对 |

---

## 7. 关键文件速查表

| 能力 | 文件 | 说明 |
| --- | --- | --- |
| DI 容器 | `di/container.go` | 注册 Adapter、UseCase、SystemAdapter |
| 玩家上下文 | `adapter/context/context_helper.go` | 暴露 `GetXxxSys(ctx)` |
| Controller 注册 | `adapter/controller/*_controller_init.go` | OnSrvStart 时统一注册协议 |
| Protocol Router | `adapter/router/protocol_router_controller.go` | C2S→DungeonServer 转发 |
| PlayerGateway | `adapter/gateway/player_gateway.go` | 操作 `PlayerRoleBinaryData` |
| NetworkGateway | `adapter/gateway/network_gateway.go` | 下发 S2C/广播 |
| PublicActorGateway | `adapter/gateway/public_actor_gateway.go` | 公共 Actor 交互 |
| DungeonServerGateway | `adapter/gateway/dungeon_server_gateway.go` | RPC/协议注册 |
| SystemAdapter 基类 | `adapter/system/base_system_adapter.go` | 生命周期/防退化约束 |
| Message Registry | `engine/message_registry.go` | 玩家消息回放注册中心 |
| Player Actor | `app/playeractor/*` | 玩家 Actor 主循环 |
| Public Actor | `app/publicactor/*` | 社交/全局数据 |
| Legacy EntitySystem | `playeractor/entitysystem/*` | 待清理模块列表 |

---

## 8. 变更记录

| 日期 | 内容 |
| --- | --- |
| 2025-12-03 | 首版：补齐 Clean Architecture 分层映射、阶段路线、子系统任务清单与测试/验收指引 |

---

> **执行提醒**：重构过程中务必保持 Actor 单线程特性、统一时间源、Proto 编译要求及文档同步流程。若发现新的架构约束或跨模块依赖，请第一时间更新本手册与 `docs/服务端开发进度文档.md`。

