## 1. 目的

为 `server/service/gameserver/internel/app/playeractor` 提供一份可逐步执行的瘦身方案，减少过度分层，同时保持 Clean Architecture 与 Actor 约束。

## 2. 约束

- Controller 只做协议入口、系统开关检查、调用 UseCase，回包由 Presenter 完成。
- SystemAdapter 只保留生命周期/事件胶水与少量与 Actor 运行模型强绑定的状态，不写业务规则。
- UseCase 承载业务规则，依赖接口，不感知 Actor/网络。
- 不破坏 Actor 单线程与 servertime 约束，不改消息枚举/对外协议。

## 3. 执行步骤（建议顺序）

1) 移除纯脚手架层
   - 删除 `di/container.go`，在 `adapter.go` 或各系统 init 中直接注入依赖。
   - 若 `adapter/context/context_helper.go` 仅做上下文取值，合并到 `gshare` 或 `PlayerHandler`，移除该包。
   - 删除 `usecaseadapter/*`（consume/reward 等），接口统一留在 `usecase/interfaces`，实现放回对应 UseCase。

2) Controller 双文件合并
   - 将各 `*_controller_init.go` 注册逻辑并入 `*_controller.go`，统一通过 `adapter/router/protocol_router_controller.go` 注册。
   - 去掉成对的 init/impl 文件，减少 20+ 文件。

3) Gateway/Repository 精简
   - 以 `usecase/interfaces` 为唯一接口声明，删除 `domain/repository` 中重复接口层（若无差异）。
   - 合并 `network_gateway.go` 与 `session_gateway.go` 为 `client_gateway.go`，集中客户端通信能力。
   - 保留 `public_actor_gateway.go` 与 `dungeon_server_gateway.go`，将纯 Actor 工具函数下沉到 `core/gshare`，避免再包一层。

4) SystemAdapter 瘦身
   - 逐个 SystemAdapter 排查，删除 if/for 业务逻辑，全部下沉到对应 UseCase；仅保留生命周期/事件到 UseCase 的调度。
   - 去除空事件订阅或仅转发的薄封装，直接在 SystemAdapter 调用 UseCase。
   - 统一生命周期签名：`OnInit/OnEnterGame/RunOne/OnNewDay/OnNewWeek/OnLogout`，在文件头注明“只做调度”。

5) 实体层与遗留清理
   - `entitysystem` 若仅做系统注册，保留注册入口，删除冗余调度逻辑或重复状态字段，状态迁移到 UseCase/PlayerRole。
   - 将 `entity/player_network.go` 中的遗留协议入口迁移到对应 Controller，避免 Actor 直收协议。

6) Handler/Adapter 精简
   - `adapter.go` + `handler.go` 保留 Actor 管理与 `DoRunOne` 调度；提取 `Loop` 的上下文构造为小函数，删除无用克隆/包装。
   - 若 `adapter/event/event_adapter.go` 仅转发到事件总线，则在调用处直接使用事件发布器，移除该适配层。

7) 校验与文档
   - 每完成一批（如 Controller 合并、Gateway 合并、SystemAdapter 瘦身），补充/更新 `docs/服务端开发进度文档.md` 与 full 版，记录已完成功能与关键入口。
   - 运行 `go test ./server/service/gameserver/internel/app/playeractor/...`（或现有脚本）及 lint，确保瘦身不破坏行为。

## 4. 预期结果

- 目录收敛为：`adapter/{controller,gateway,system,presenter,router}`、`usecase`（含 interfaces）、`entity`、`domain/model`；无 `di`、无 `usecaseadapter`、无双文件 Controller/空适配器。
- SystemAdapter 文件数显著下降，内容仅保留“何时调用哪个 UseCase”的调度；Controller 文件数约对半减少。

## 5. 落地顺序与检查表（带文件指引）

1) **删脚手架 / 单点注入**
   - 删 `di/container.go`；若需初始化，改在 `adapter.go` 或各系统的 `init()` 中直接装配。
   - 检查 `adapter/context/context_helper.go`：若仅封装 ctx 值，合并到 `core/gshare` 或 `PlayerHandler`。
   - 删除 `adapter/usecaseadapter/*`，接口留在 `usecase/interfaces`，实现放回对应 UseCase。

2) **Controller 合并（减少 ~50% 文件）**
   - 将 `adapter/controller/*_controller_init.go` 合并入同名 `.go`，在 `adapter/router/protocol_router_controller.go` 统一注册。
   - 顺带检查是否有 Controller 直写业务逻辑，立即下沉到 UseCase。

3) **Gateway/Repository 精简**
   - 以 `usecase/interfaces` 为唯一接口声明，移除 `domain/repository` 中重复接口（若完全一致）。
   - 合并 `adapter/gateway/network_gateway.go` + `session_gateway.go` 为 `client_gateway.go`；校对调用方 import。
   - 将纯 Actor 工具函数下沉到 `core/gshare`，减少 gateway 再包一层。

4) **SystemAdapter 排毒**
   - 逐个文件搜索业务逻辑：奖励/校验/堆叠/次数/冷却/条件等立即迁移到对应 UseCase。
   - 清理空事件订阅或仅转发包装；保留生命周期/事件 → UseCase 调度；统一签名与头注释。
   - 状态字段若可迁移到 `PlayerRole` 或 UseCase 内部，迁移后删除适配层冗余字段。

5) **实体层/遗留入口清理**
   - `entitysystem/sys_mgr.go`、`system_registry.go` 仅保留注册；删除重复调度或状态。
   - 将 `entity/player_network.go` 残留协议迁到对应 Controller，保证协议入口唯一。

6) **Handler/事件适配简化**
   - `adapter.go`/`handler.go`：保留 Actor 管理 + `DoRunOne`，提炼 ctx 构造小函数，删除无用克隆。
   - 若 `adapter/event/event_adapter.go` 只是转发事件，直接在调用处用事件发布器，移除该层。

7) **验证与文档同步**
   - 每完成一批，运行 `go test ./server/service/gameserver/internel/app/playeractor/...` 与 lint。
   - 在《docs/服务端开发进度文档.md》记录“已完成功能”与“关键代码位置”；必要时同步 full 版。

## 6. UseCase 初始化判空收敛（EnsureXData）

目标：将 UseCase 内重复的 nil 判空与初始化收敛到 `PlayerRole` 的单点懒加载 Getter，保证返回值永远非 nil，UseCase 不再写 `if data == nil { ... }`。

统一策略
- 在 `entity/player_role.go` 为各数据块新增 `EnsureXData()`（或 `GetXData()`），内部判空并初始化默认结构，返回非空引用。
- 初始化函数只做数据结构/默认值填充，不夹杂业务校验；业务校验继续留在 UseCase。
- 调用保持 Actor 单线程，不新增并发。

模块清单与替换点
- Bag：`EnsureBagData()`；替换 `usecase/bag/{add_item,add_item_tx,remove_item,remove_item_tx,has_item}.go` 判空。
- Money：`EnsureMoneyData()`；替换 `usecase/money/{init_money_data,add_money,consume_money,update_balance_tx}.go` 判空。
- Level：`EnsureLevelData()`；替换 `usecase/level/{init_level_data,add_exp,level_up}.go` 判空。
- Equip：`EnsureEquipData()`；替换 `usecase/equip/{init_equip_data,equip_item,unequip_item}.go` 判空。
- Skill：`EnsureSkillData()`；替换 `usecase/skill/{init_skill_data,learn_skill,upgrade_skill}.go` 判空（视实际文件）。
- ItemUse：`EnsureItemUseData()`；替换 `usecase/item_use/{init_item_use_data,use_item}.go` 判空。
- Quest：`EnsureQuestData()`；替换 `usecase/quest/{init_quest_data,accept_quest,submit_quest,refresh_quest_type}.go` 判空。
- Fuben：`EnsureDungeonData()`；替换 `usecase/fuben/{init_dungeon_data,get_dungeon_record,enter_dungeon,settle_dungeon}.go` 判空。
- 其他：Mail/Message/Shop/Recycle 如有持久字段，可增加对应 `EnsureXData` 并替换判空。

测试与文档
- 为每个 `EnsureXData` 补最小单测：nil -> 初始化；非 nil -> 返回同一引用不覆盖。
- 回归核心 UseCase 单测（加物品、扣钱、升级、学技能、接任务、进副本）。
- 完成一批后在《docs/服务端开发进度文档.md》“已完成功能”注明“UseCase 初始化判空收敛为 EnsureXData”，并在“关键代码位置”标注 `entity/player_role.go`。

## 7. `player_role.go` 去耦与收敛方案

目标：简化实体层依赖，移除 `PlayerRole` 对事件总线等框架设施的直接耦合，并把数据初始化集中到实体层的 Ensure 方法。

调整点
- 去耦 event bus
  - 移除 `PlayerRole` 字段 `eventBus *event.Bus` 及构造时的 `gevent.ClonePlayerEventBus()`。
  - 事件发布改由 UseCase 通过 `EventPublisher` 接口完成；事件订阅（如等级变化→排行榜刷新）改放到 SystemAdapter 或专门的事件装配处。
  - `OnLogin`、`OnLevelUp` 等调用 `pr.Publish(...)` 的位置，改为 UseCase/Adapter 触发，或通过注入的接口调用。
- 初始化收敛
  - 在 `entity/player_role.go` 增加/强化 `EnsureXData()`（Bag/Money/Level/Equip/Skill/ItemUse/Quest/Fuben…）作为唯一数据入口，构造时不再显式做各系统初始化判空。
  - 构造函数只负责加载 DB 数据与基础字段，业务数据的 lazy init 由各 EnsureXData 执行。
- 依赖最小化
  - 继续保留 `attrCalculator` 注入，但避免在构造函数中订阅事件；属性变更触发的推送/排行榜更新由 SystemAdapter/UseCase 驱动。
  - 检查并移除对 `di.Container` 等的直接引用，如有使用改为通过 Adapter/UseCase 注入。

落地步骤
1) 删除 `PlayerRole` 中 eventBus 字段与相关 import/初始化；将 `Publish`/订阅逻辑迁到 SystemAdapter/UseCase 层。
2) 在 `entity/player_role.go` 内集中实现/调用 `EnsureXData()`，构造函数仅保留 DB 加载与基础字段设置。
3) 更新调用方：UseCase/Adapter 获取数据一律调用 EnsureXData；清理重复判空。
4) 校验：回归登录/升级/加物品等核心链路；确保事件发布改由 UseCase/Adapter 触发后行为一致。
5) 文档同步：在 `docs/服务端开发进度文档.md` 标注“PlayerRole 去耦事件总线、初始化收敛到 EnsureXData”。

## 8. 事件总线使用策略（gevent）

结论：不要使用单例 `gevent` 作为全局共享 EventBus，也不要在 `PlayerRole` 内克隆/持有；应由 SystemAdapter/UseCase 通过注入的 `EventPublisher`/订阅装配来使用，按 Actor 隔离。

原因与风险
- PlayerActor 是单线程，但有多个 PlayerActor 并发运行；全局共享 EventBus 会被多个 goroutine 同时访问，`event.Bus` 并非并发安全，存在数据竞争风险。
- 玩家事件语义应隔离（玩家 A 的事件不应默认影响玩家 B）；按 Actor 维度持有/装配可避免跨玩家污染。

推荐做法
- 在 `gevent` 中保留“事件总线工厂/构造函数”，按需创建玩家级或系统级总线，不使用全局单例共享实例。
- 事件发布：在 UseCase 中通过接口（`EventPublisher`）调用，由外部注入实现，避免实体直接持有 EventBus。
- 事件订阅：在 SystemAdapter 或专门的事件装配处注册，生命周期随 Actor/系统启动与销毁。
- 如果后续需要全局事件（极少数场景），应提供专门的并发安全实现，并明确跨 Actor 的语义与锁/队列策略。

## 9. EventPublisher 注入式事件发布方案（可执行指引）

目标：UseCase 通过接口发布事件，由外部注入实现；实体层不再持有/感知 EventBus，事件总线按 Actor 隔离。

接口与位置
- 接口声明：`internel/app/playeractor/service/interfaces/event.go`
  ```go
  type EventPublisher interface {
      Publish(ctx context.Context, topic string, payload any) error
  }
  ```
- 实现：`internel/app/playeractor/event/event_publisher.go`
  - 构造入参：玩家级事件总线（由 `gevent.NewPlayerEventBus()` 等工厂创建）。
  - 方法：`Publish` 内部直接调用 bus.Publish，记录错误，不 panic；如需异步，保持在 Actor 单线程队列。

装配与注入
- 在 PlayerActor 启动/系统 init 处创建玩家级 bus（不使用全局单例），如：
  - `bus := gevent.NewPlayerEventBus()`
  - `publisher := eventadapter.NewEventPublisher(bus)`
- 将 `publisher` 注入各需要事件的 UseCase 构造函数；禁止 UseCase 去拉容器或全局变量。
- 事件订阅放在 SystemAdapter 或专门的事件装配处，随 Actor 生命周期管理。

迁移步骤
- 删除 `PlayerRole` 中对 eventBus 的持有与 Publish 方法，将订阅/发布迁移到 Adapter/UseCase。
- UseCase 内的事件触发改为调用 `eventPublisher.Publish(ctx, topic, payload)`。
- gevent 仅提供工厂/构造函数，不暴露全局共享实例；如需跨 Actor 全局事件，另行提供并发安全实现并标明语义。

测试与文档
- 回归登录/升级/任务/副本等触发事件的路径，确保事件仍被发布且无数据竞争。
- 在 `docs/服务端开发进度文档.md` 标注：事件发布改为 EventPublisher 注入，PlayerRole 不再持有 EventBus；事件总线按 Actor 创建。

## 10. 2025-12-10 代码评审结论与瘦身动作

本轮审查聚焦 `internel/app/playeractor`，主要发现的过度分层点与对应瘦身动作如下：

- 过度抽象的容器与上下文包装
  - `di/container.go` 维护全局单例，UseCase 构造时通过 Container 拉依赖，违背“构造注入 + 无全局状态”原则。
  - `adapter/context/context_helper.go` 仅做 Context 取值/日志包装，可合并到 `core/gshare` 或直接由 Controller/SystemAdapter 读取。
  - 行动：删除 DI 容器与 context 包，改为在 `adapter.go`/各 SystemAdapter 构造函数显式传入依赖；`gshare.ContextKey*` 保留，取值直接使用标准库。

- 事件发布仍耦合 PlayerRole 与全局模板
  - `adapter/event/event_adapter.go` 通过 `playerRole.Publish` 和 `gevent.SubscribePlayerEvent` 全局模板发布/订阅，继续让实体持有事件总线。
  - 行动：按本文件第 8/9 节执行，创建玩家级 EventBus 并注入 EventPublisher，移除 PlayerRole.Publish 调用与 gevent 全局模板订阅。

- Controller 双文件 + 遗留入口未收敛
  - 各 `*_controller.go` 搭配 `*_controller_init.go`，且 `entity/player_network.go` 仍直接注册 `C2SEnterGame/C2SQueryRank`、PlayerActorMsg 处理。
  - 行动：合并 init/impl，注册集中到 `adapter/router/protocol_router_controller.go`；将 `handleEnterGame/handleQueryRank/handlePlayerMessageMsg` 等迁入对应 Controller，`player_network.go` 仅保留极少必要胶水或删除。

- Gateway/Repository 重复包装
  - `adapter/gateway/network_gateway.go` 与 `session_gateway.go` 仅薄封装 `gatewaylink`，`domain/repository` 与 `usecase/interfaces` 存在重复接口定义。
  - 行动：合并为 `client_gateway.go`（含 Session/Send），接口以 `usecase/interfaces` 为唯一声明，删除重复仓储接口层，直接在构造处注入实现。

- UseCaseAdapter 直接回调实体，破坏依赖倒置
  - `adapter/usecaseadapter/{consume,reward}_use_case_adapter.go` 透传到 `PlayerRole.CheckConsume/GrantRewards`，让 UseCase 依赖实体。
  - 行动：删除 usecaseadapter 包，将消耗/奖励逻辑收敛回对应 UseCase（bag/money/reward），对外暴露接口由 UseCase 层实现。

- SystemAdapter 胶水外溢
  - 多数 SystemAdapter 通过 `di.Container` 获取 Gateway/UseCase；部分维护状态（如背包索引）而非纯调度。
  - 行动：在各 `*_system_adapter.go` 构造函数显式注入依赖；状态类需求若需保留（如背包索引），明确注释“仅缓存/调度”，其余逻辑下沉 UseCase。统一生命周期函数签名与头部注释。

执行顺序建议（可与既有步骤对齐）
1) 移除 DI 与 context 包，重写构造路径（adapter.go + SystemAdapter 构造函数）。
2) 事件发布/订阅重构，完成 EventPublisher 注入，删除 PlayerRole.Publish 依赖。 
3) Controller 合并与协议入口收敛，清理 `player_network.go` 遗留处理器。
4) Gateway/Repository 收敛：合并 client gateway，接口只保留 `usecase/interfaces`。
5) 删除 usecaseadapter 包，补齐对应 UseCase 实现；同步 SystemAdapter 注入新接口。
6) SystemAdapter 全量巡检：剥离业务逻辑与多余状态，保留生命周期调度与必要缓存。

验收标准
- 目录收敛：移除 `di/`、`adapter/context/`、`adapter/usecaseadapter/`，Controller 不再成对 init/impl。
- UseCase 构造与依赖注入无全局单例；事件发布按 Actor 隔离；协议入口仅在 Controller 层。
- lint/go test 通过，核心链路（登录/进副本/加物品/扣钱）回归正常。

## 11. 快速落地与回归清单

执行优先级（建议顺序）
- P0：删 `di/` 与 `adapter/context/`，重连依赖注入路径；事件层改为按 Actor 注入 EventPublisher。
- P0：Controller 合并注册，迁移 `player_network.go` 遗留协议入口。
- P1：合并 client gateway，删除 usecaseadapter 包，接口以 `usecase/interfaces` 为唯一声明。
- P1：SystemAdapter 全量巡检，剥离业务逻辑，保留生命周期调度与必要缓存。
- P2：整理 `entitysystem` 仅做注册；补充 EnsureXData 判空收敛的调用面检查。

最小回归（每批改动后跑一遍）
- `go test ./server/service/gameserver/internel/app/playeractor/...`
- 手动链路：注册→登录→创角→进入游戏→加物品/扣钱→进本结算→技能学习/升级→任务接受/提交。
- 内部消息：DungeonActor 进出本、属性同步、玩家离线消息回放。
- 协议注册：确认 `protocol_router_controller` 覆盖全部 C2S；PlayerActorMsg 与 DungeonActorMsg 均能正常路由。