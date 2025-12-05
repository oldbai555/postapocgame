# DungeonActor 整合方案

更新时间：2025-12-03  
责任人：个人独立开发

---

## 1. 背景与目标

- **问题**：`server/service/dungeonserver` 作为独立进程，维护成本高，而当前阶段的诉求只是让战斗/副本在本地稳定跑通，支撑玩法开发与调试，并不需要维护远程战斗进程的上线/灰度/回滚能力。
- **目标**：在 `server/service/gameserver/internel/app` 内新增 `dungeonactor`，把原 DungeonServer 能力封装为 GameServer 内部的单 Actor，**长期只保留 InProcess DungeonActor，一次性下线独立 DungeonServer 能力**。
- **约束**：
  - `dungeonactor` 被视为 GameServer 的基础设施组件：进程启动后即常驻运行，不受玩法开关影响，所有“功能是否对玩家开放”的决策均在 `playeractor` Controller/SystemAdapter 层完成。
  - 不破坏 Clean Architecture 依赖方向（Controller → UseCase → Adapter/Gateway → Framework）。
  - UseCase 不感知“系统开关”等框架状态，仅处理领域规则。
  - “系统是否开启”的检查统一在 Controller 层（入口检查）和 SystemAdapter 层（生命周期调度）完成，并与 `docs/SystemAdapter系统开启检查优化方案.md`、`docs/服务端开发进度文档_full.md` 第 7.10/7.11 节保持一致。
  - `server/service/dungeonserver` 及其专用 RPC 协议会被物理删除；如未来真的需要战斗服横向拆分，将以当前 `dungeonactor` 代码为基础重新设计新的远程进程，而不是“恢复”旧 DungeonServer。

---

## 2. Controller / SystemAdapter / UseCase 职责（结合现有架构决策）

| 层级 | 本项目职责边界 | 关键位置 |
| ---- | -------------- | -------- |
| Controller | 处理协议入口（C2S/D2G/G2D），完成 **协议解析 + 上下文注入(Session/Role) + 权限/频控 + 系统开启检查**，然后把“已通过框架层校验的领域意图”交给 UseCase。禁止写业务循环或直接操作实体数据。系统未开启时，直接通过 Presenter 返回 `ErrorCode_System_NotEnabled`。 | `server/service/gameserver/internel/adapter/controller/*`、`docs/服务端开发进度文档_full.md` 第 7.10/7.11 节 |
| SystemAdapter | 将 UseCase 挂载到 Actor 生命周期：注册系统、订阅事件、在 `OnInit/OnLogin/OnReconnect/RunOne/OnNewDay` 等生命周期中**决定是否调用哪一个 UseCase**，并维护系统自身的“已开启/未开启”状态。可以读取 Actor 状态与系统开关配置，但只做“要不要调用 UseCase”的调度，不写具体业务分支。 | `server/service/gameserver/internel/adapter/system/*`、`docs/gameserver_adapter_system演进规划.md` |
| UseCase | 纯业务逻辑：只依赖 Repository/Gateway/Presenter **接口**；完全不关心 Actor、网络、协议、系统开关等框架状态。输入是经过 Controller/SystemAdapter 过滤后的领域意图，输出通过 Presenter/Gateway 对外表现。任何“系统是否可用/是否开启”的判断都在 UseCase 之外完成。 | `server/service/gameserver/internel/usecase/*` |

**依赖方向**：Controller → UseCase → Adapter（Repo/Gateway/Presenter） → Framework/Infra。SystemAdapter 与 Controller 同级，单向依赖 UseCase / Adapter，二者都不得依赖具体的 Actor/网络/数据库实现类型。

### 2.1 系统开启检查在三者中的位置

结合 `docs/SystemAdapter系统开启检查优化方案.md` 与 `docs/服务端开发进度文档_full.md` 第 7.10/7.11 节，本项目对“系统是否开启”的统一约定为：

1. **Controller 负责入口级检查**
   - 所有玩家侧业务协议（如 Bag/Fuben/Quest/Skill/Shop/Recycle 等），统一通过 SystemAdapter Helper（如 `GetBagSys(ctx)`、`GetFubenSys(ctx)`）获取系统实例。
   - 若 Helper 返回 `nil` 或 `sys.IsOpened()==false`，Controller 直接通过 Presenter 返回 `ErrorCode_System_NotEnabled`，**不再调用 UseCase**。
   - 这保证了 UseCase 不需要重复做“系统开关”判断，只要被调用，就意味着系统在当前玩家上下文中是开启的。
2. **SystemAdapter 负责生命周期阶段的检查**
   - 在 `OnInit/OnLogin/OnReconnect/RunOne/OnNewDay` 等生命周期方法中，SystemAdapter 首先检测自身是否处于“开启”状态（一般由 `SysMgr` 或内部开关字段维护）。
   - 若系统未开启，则该生命周期方法直接 `return`，不调用任何 UseCase；若需要“自恢复”，可以在 SystemAdapter 内部协调 `SysMgr` 重新开启系统，但**具体业务恢复逻辑仍在 UseCase 中实现**。
3. **UseCase 不感知系统开关**
   - UseCase 不读取也不维护“系统是否开启”的状态，只处理具体的业务规则（例如：副本进入条件校验、技能升级规则、掉落结算逻辑等）。
   - 如需根据“开服天数”“活动是否开放”等进行控制，应通过配置/领域规则在 UseCase 内实现，而不是读取 SystemAdapter 的开关字段。

---

## 3. 玩家侧系统可用性策略

> 不再引入额外配置或 `SystemToggleService`；判断逻辑全部围绕 PlayerActor 现有的 `SysMgr` 与 SystemAdapter 状态。

1. **SysMgr 管控生命周期**：`playeractor` 登录时 `SysMgr.OnInit/OnRoleLogin` 会遍历所有系统，调用 `system.SetOpened(true)` 并同步到 `playerRole.SetSysStatus`，保证 Actor 内部状态一致。若某系统初始化失败，可通过日志聚焦问题。
2. **Controller 入口校验**：继续通过 `system.GetXXXSys(ctx)` 获取具体系统适配器，若返回 `nil` 或 `sys.IsOpened()==false`，直接返回 `ErrorCode_System_NotEnabled`。这保证了 UseCase 只会在系统处于打开状态时被执行。
3. **SystemAdapter 调度**：在 `RunOne/OnNewDay/OnLogin` 等生命周期中，优先检查自身 `IsOpened()` 再决定是否调用 UseCase。必要时可调用 `SysMgr.CheckAllSysOpen(ctx)`，确保被动关闭的系统重新激活。
4. **例子**：
   - `adapter/controller/fuben_controller.go`：`handleEnterDungeon` 解析 → `GetFubenSys(ctx)` 判空 → 执行 UseCase；系统缺失时直接返回“副本系统未开启”。
   - `adapter/system/fuben/adapter.go`：`RunOne` 中先看 `IsOpened()`，关闭状态直接 return，不调用 `usecase/fuben.Tick()`。
   - `adapter/system/skill/adapter.go`：`OnInit` 时若 `SysMgr` 尚未标记开启，可等待 `SysMgr.CheckAllSysOpen` 或日志提醒；无需额外配置即可保证玩家侧行为。

---

## 4. DungeonServer → DungeonActor 整体方案

### 4.1 整体思路

1. **单体运行**：在 GameServer 进程内创建 `dungeonactor`（`actor.ModeSingle`），承担原 DungeonServer 的 RunLoop、场景管理和实体系统。
2. **跨 Actor 通信**：沿用 `DungeonServerGateway` 接口，但实现从原 IPC/RPC 切换为“同进程消息队列”（Actor 间消息投递）。Controller/UseCase 对网路/进程无感。
3. **目录规划**：
   ```
   server/service/gameserver/internel/app/
       playeractor/
       publicactor/
       dungeonactor/            <-- 新增
           engine/              (原 dungeonserver/internel/engine)
           entitysystem/        (移动/战斗/AI/属性等)
           scene/
           adapter/             (对 GameServer Adapter 的封装)
   ```
4. **对外暴露**：`dungeonactor` 暴露 `RegisterDungeonActor(engine actor.Engine)` 供 `cmd/gameserver/main.go` 启动；所有 `DungeonServerGateway` 调用在 `InProcess` 模式下改为投递到该 Actor。

### 4.2 分阶段迁移

| 阶段 | 目标 | 关键动作 |
| ---- | ---- | -------- |
| Phase 0：准备 | 建立基础设施，不影响现网功能 | - 在 `gameserver` 内新增 `dungeonactor` 空目录与启动入口<br>- 抽象 `ActorRouter`（封装 `actor.Engine` 注册）。<br>- 扩展 `DungeonServerGateway`：增加 `mode`/`RunInProcess` 配置与预留本地派发入口。 |
| Phase 1：共享依赖下沉 | 迁移公共代码 | - 将 `server/service/dungeonserver/internel` 中通用包（`scene`, `entity`, `entitysystem/*`, `attrcalc` 等）移动或复制到 `gameserver/internel/app/dungeonactor/*`，更新 go module 引用。<br>- 保持 `server/service/dungeonserver` 继续引用这些包，确保过渡期可双运行。 |
| Phase 2：入口切换 | GameServer 内部调用 | - 在 GameServer 启动时启动 `dungeonactor` 并注册消息路由。<br>- `DungeonServerGateway` 在配置为“本地模式”时改为 `actorEngine.Send(dungeonActorId, msg)`。<br>- Controller/UseCase 原有 RPC 调用无需改动。<br>- 验证副本进入/结算、技能同步链路。 |
| Phase 3：清理与下线 | 删除独立进程 | - 停用 `cmd/dungeonserver` 启动脚本，文档标注“已下线”。<br>- 删除 `server/service/dungeonserver` 内的入口，仅保留必要工具（如需要则迁移到 `tools/`）。<br>- 更新 `docs/服务端开发进度文档*.md` 中的服务拓扑图。<br>- 最终物理删除 `server/service/dungeonserver` 目录。 |

### 4.3 迁移细节

1. **Actor 启动**：在 `gameserver/internel/app/app_init.go` 中注册 `RegisterDungeonActor(engine)`，确保在 PlayerActor 前启动，避免玩家请求找不到副本。
2. **消息协议**：保留 `proto/dungeonserver/*.proto`，但实现层改为本地调用。若后续需要再拆分，保持协议不变即可重新部署远程进程。
3. **配置与资源**：
   - 地图/怪物配置 (`scene_config`, `monster_config`) 统一移动到 `server/output/config`，GameServer 与 DungeonActor 共享。
   - DungeonActor 运行参数沿用 `gameserver` 现有 `jsonconf`（如 `GameServerConfig.Actor`）与编译期常量：固定 `actor.ModeSingle` + `RunOne` 循环间隔，暂不在 `jsonconf` 中增加新分支（线程数、Tick 等），避免新增配置负担。无论玩家层开关如何，`dungeonactor` 都会在进程内常驻。
4. **日志与监控**：使用 `pkg/log.WithRequester("dungeonactor")`，在单体进程内也能区分模块日志。
5. **数据库访问**：保持 `DungeonActor` 不直接访问玩家数据库，仍通过 `DungeonServerGateway` 回调 GameServer UseCase。若需要临时读配置，可通过共享 `ConfigGateway`。

### 4.4 删除与收尾

1. **代码清单**：以下路径最终将被删除（Phase 3 完成后）：
   - `server/service/dungeonserver/cmd/...`
   - `server/service/dungeonserver/internel/...`
   - `server/service/dungeonserver/main.go`
   - 相关构建脚本与文档引用
2. **文档更新**：
   - `docs/服务端开发进度文档*.md`：更新服务拓扑，注明 DungeonActor 已并入 GameServer。
   - `docs/dungeonserver_CleanArchitecture重构文档.md`：在文首添加“迁移完成”声明，并指向本方案。
3. **测试与验收**：
   - 单体模式下完成副本进入/结算、技能、移动全链路回归。
   - 执行压测场景，确保 `dungeonactor` 与 `playeractor` 在同一进程时不会互相阻塞（Actor 模式单线程，需关注耗时操作）。

### 4.5 最小可运行闭环（MVP）

1. **保留远程 DungeonServer**：在第一版中保持 `server/service/dungeonserver` 编译输出，作为回滚备用。
2. **GameServer 内启动 DungeonActor（常驻）**：`cmd/gameserver/main.go` 启动时读取 `dungeon_mode`（默认 `remote`），在配置为 `inprocess` 时拉起 `RegisterDungeonActor` 并注册 `actor.ModeSingle`。无论玩家层的系统开关如何，DungeonActor 都会运行，用作统一战斗/场景引擎。
3. **Gateway 双写**：`DungeonServerGateway` 新增本地派发路径，当 `mode==inprocess` 时直接投递；否则沿用原 `dungeonserverlink.AsyncCall`。两条链路复用同一 `Request/Response` 结构，方便灰度。
4. **用例验证**：先选 `C2SEnterDungeon` → 副本结算 全链路，确保 Controller/SystemAdapter/UseCase → Gateway → DungeonActor → 回写 Presenter 全流程闭环。
5. **监控与日志**：`pkg/log.WithRequester("dungeonactor")` 输出独立日志前缀；在 `actor.Engine` 层加入 RunOne 耗时统计，便于观察单体模式性能。通过玩家层开关暂停入口时，DungeonActor 仍会记录基础巡检日志，方便排查。
6. **回收远程进程**：当 `inprocess` 模式稳定后，再删除 `server/service/dungeonserver`、编排文件与构建脚本。

---

## 5. 可落地改造路线（满足架构约束）

### 5.1 分层职责落地

1. **Controller（协议入口）**
   - 典型示例：`adapter/controller/fuben_controller.go`、`skill_controller.go`、`item_use_controller.go`。
   - 运行时行为：解析 `C2S`/`D2G` 请求 → 从 `SystemAdapter` Helper（如 `GetFubenSys(ctx)`）取得系统实例 → 判断 `sys != nil && sys.IsOpened()` → 构造 `usecase` 入参。
   - 约束：只做协议层与框架层校验，所有业务分支（副本进入条件、技能升级规则、掉落结算等）全部留在 UseCase。
2. **SystemAdapter（Actor 生命周期胶水层）**
   - 示例：`adapter/system/fuben/adapter.go`、`adapter/system/skill/adapter.go`。
   - 运行时行为：在 `OnInit/OnLogin/OnReconnect/RunOne/OnNewDay` 中决定是否调用 UseCase；负责把 PlayerActor/DungeonActor 生命周期事件转换成 UseCase 调用；维护系统开关状态，必要时在 RunOne 直接 return。
   - 约束：不得写业务循环，仅允许“何时调用”决策；针对 DungeonActor 相关系统（如 FubenSys、SkillSyncSys）需要在 `RunOne` 中通过 Gateway 向 `dungeonactor` 投递消息，但决策逻辑（例如“哪些副本需要同步”）必须来自 UseCase。
3. **UseCase（纯业务逻辑）**
   - 示例：`usecase/fuben/enter_dungeon.go`、`usecase/skill/learn_skill.go`。
   - 运行时行为：依赖 `DungeonServerGateway`、`PlayerRepository`、`Presenter` 等接口；根据输入执行业务并通过 Presenter/Gateway 输出。
   - 约束：不注入任何“系统开关”或 Actor 上下文对象；不直接操作网络/日志，全部依赖接口。

### 5.2 系统可用性校验流程（基于 SysMgr）

1. **获取系统实例**：所有 Controller/SystemAdapter 通过 `system.GetXXXSYS(ctx)` 获取实例；该 Helper 内部依赖 `playerRole.GetSystem` 与 `SysMgr`，确保只有在系统成功初始化、`IsOpened=true` 时才返回对象。
2. **Controller 统一返回**：若 Helper 返回 `nil`，直接返回 `ErrorCode_System_NotEnabled`（或更具体的错误码），并在日志中带上 `systemId` / `sessionId`，方便排查。
3. **SystemAdapter 运行前检查**：在 `OnInit/OnLogin/RunOne/OnNewDay` 等生命周期回调里，第一步检查自身 `IsOpened()`；若为 `false`，直接 return，避免 UseCase 在未准备好的状态下执行。
4. **自动恢复**：当检测到系统处于关闭状态（例如数据缺失或初始化失败），可调用 `SysMgr.CheckAllSysOpen(ctx)` 或重新触发 `system.SetOpened(true)`，确保生命周期与玩家状态保持一致。
5. **错误度量**：建议在“系统未开启”的返回逻辑中统计指标（如 `metrics.SystemDisabledCounter(systemId)`），以便快速发现哪些系统经常被拒绝。

### 5.3 DungeonActor 集成步骤

1. **Gateway 双模实现**
   - 在 `adapter/gateway/dungeon_server_gateway.go` 中增加 `mode` 字段（`Remote`/`InProcess`），并在初始化时根据配置决定是走旧的 `dungeonserverlink.AsyncCall` 还是直接投递到 `dungeonactor`。
   - InProcess 模式下，通过 `actorEngine.Send(dungeonActorId, msg)` 将原 RPC 参数传入 `dungeonactor`。
2. **dungeonactor 启动**
   - 在 `internel/app/app_init.go` 引导 `RegisterDungeonActor(engine)`，并在 `cmd/gameserver/main.go` 中读取配置决定是否启动。
   - `dungeonactor` 内部保持 `ModeSingle` 循环，复用原 DungeonServer 的 `engine`、`entitysystem`、`scene` 子包。
3. **UseCase 无感迁移**
   - 所有 UseCase 继续依赖 `DungeonServerGateway` 接口，无需修改；只需在 Gateway 中使用相同的 `Request/Response` 结构（`proto/rpc.proto`）。
   - 在迁移期间保留远程模式开关，便于灰度和回滚。
4. **系统可用性对接（玩家侧）**
   - 针对副本、战斗录像等玩家入口，Controller 与 SystemAdapter 统一依赖 `SysMgr` 与系统适配器自身的 `IsOpened()` 状态来判断是否执行 UseCase；DungeonActor 自身不参与此判断，始终保持运行。

### 5.4 Controller/SystemAdapter 自检清单

| 检查项 | Controller | SystemAdapter |
| --- | --- | --- |
| 系统可用性 | 入口需通过 `GetXXXSYS(ctx)` 判空/判 `IsOpened`，失败时直接返回 `ErrorCode_System_NotEnabled` | 每个生命周期回调先看 `IsOpened()`，必要时调用 `SysMgr.CheckAllSysOpen(ctx)`，在日志中记录关闭原因 |
| 依赖方向 | 只依赖 UseCase 接口与 Presenter，不引用 `entitysystem`、`playeractor` 具体类型 | 只引用 `usecase` 接口与 `context` Helper，不向 UseCase 暴露 Actor 细节 |
| 框架职责 | 解析协议、做 Session/幂等/权限校验 | 负责生命周期调度、订阅 Actor 事件、管理与 DungeonActor 的消息通道 |
| 业务逻辑 | 所有 if/for/计算应转移到 UseCase；Controller 只做参数转换 | 不写业务循环，仅转发事件（定时检查、刷新）到 UseCase |
| 错误处理 | 统一使用 Presenter 返回错误码；记录 `IRequester` 日志 | 捕获 UseCase 错误并转为 Actor 日志，必要时触发报警/熔断 |

> 每次新增协议/系统时按此表自检，确保 Clean Architecture 依赖方向不被破坏。

### 5.5 C2S 协议与 PlayerActor ↔ DungeonActor 交互方案（InProcess 版）

> 目标：完全取消“客户端直连 DungeonActor / G2D/D2G RPC 枚举驱动”的模式，统一由 PlayerActor Controller 作为所有 C2S 协议入口，PlayerActor 与 DungeonActor 之间通过内部 Actor 消息协作，行为上类似 PublicActor。

1. **C2S 协议统一入口**  
   - 所有来自客户端的 C2S 协议（包括移动、技能、掉落拾取、副本进入等）一律在 GameServer 的 Controller 层注册与处理，禁止在 `dungeonactor/entitysystem` 或 `dungeonactor/clientprotocol` 中注册 C2S。  
   - Controller 负责：解析请求 → 注入 Session/Role → 做系统开启检查/防刷 → 调用对应 UseCase；DungeonActor 只接受来自 GameServer 的内部消息，不直接感知客户端协议。

2. **PlayerActor ↔ DungeonActor 通过 ActorFacade 协作**  
   - 在 `internel/core/gshare/actor_facade.go` 现有 Actor 门面基础上，新增 `IDungeonActorFacade` 接口与 `SetDungeonActorFacade/GetDungeonActorFacade/SendDungeonMessageAsync` 等便捷方法。  
   - `DungeonActor.NewDungeonActor` 在启动时构造并注入 `dungeonActorFacade` 实现（包装内部 `actorMgr`），通过 `gshare.SetDungeonActorFacade` 完成注册。  
   - PlayerActor 侧的 UseCase / Gateway（如 `DungeonServerGateway` 的内部实现）通过 `SendDungeonMessageAsync(sessionKey, msg)` 将“进入副本/移动/释放技能”等意图转换为 Actor 消息投递给 DungeonActor。

3. **去除本地 Dungeon RPC 表与 clientprotocol**  
   - `server/service/gameserver/internel/app/dungeonactor/drpcprotocol`：删除整个包及其在 `fuben/actor_msg.go` 中的注册逻辑，不再使用 `G2D*`/`D2G*` 枚举驱动 Dungeon 内部分发。  
   - `server/service/gameserver/internel/app/dungeonactor/clientprotocol`：删除整个包以及 `move_sys.go` / `fight_sys.go` / `drop_sys.go` / `entity/rolest.go` 中的 `clientprotocol.Register(...)` 调用；这些 C2S 协议改由 GameServer Controller 直接注册。  
   - DungeonActor 内部 Handler 不再以“协议号 + map”的形式暴露给外部，而是作为 Actor 消息处理的一部分，通过 `DungeonActorHandler`/Facade 实现分发。

4. **UseCase 层对 Dungeon 的依赖方式**  
   - UseCase 继续通过 `interfaces.DungeonServerGateway`（或后续精简后的内部接口）依赖战斗能力，而不是直连 DungeonActor 或 G2D/D2G 枚举。  
   - Gateway 内部实现改为：构造内部 DTO / Actor 消息（如 `EnterDungeonMsg/MoveMsg/UseSkillMsg`），通过 `IDungeonActorFacade` 发送到 DungeonActor，保持 UseCase 对通信细节无感知。

5. **与 RPC/Proto 的关系**  
   - Game ↔ Dungeon 之间不再新增任何基于 `proto/csproto/rpc.proto` 的 G2D/D2G RPC；现有 G2D/D2G 枚举和 message 将在 8.7 阶段逐步删除。  
   - DungeonActor 向 GameServer 的回调继续复用现有 D2G* Proto 直到内部接口重构完成，但对外表现为“进程内回调 + Controller Handler”，部署上不再存在独立 DungeonServer。

---

## 6. 验证 Clean Architecture 依赖方向

1. Controller 接收到玩家协议 → 通过 `GetXXXSYS(ctx)` 确认可用 → 调用 UseCase → 使用 Presenter/Gateway 输出。
2. UseCase 依赖 Repository/Gateway 接口（如 `DungeonServerGateway`, `PlayerRepository`），只处理业务逻辑；**禁止**在 UseCase 中：
   - 读取或修改 SystemAdapter 的开启/关闭状态；
   - 直接访问 Actor/Session/网络连接等框架对象；
   - 根据“配置开关”短路整个用例（例如：`if !cfg.FeatureEnabled { return }` 这类逻辑应上移到 Controller 或 SystemAdapter）。
3. SystemAdapter 在 Actor 生命周期中调度 UseCase，不直接做业务逻辑；当系统关闭时，直接 return 或跳过；UseCase 无需感知。
4. `dungeonactor` 作为 GameServer 内的一个 Actor，被视作“系统依赖”的一部分（等同 `PublicActor`），其接口通过 Gateway 暴露给 UseCase，保持依赖倒置。

---

## 7. 风险与后续

- **性能隔离**：同进程运行意味着副本与玩家 Actor 共享 CPU；需监控 `dungeonactor` RunOne 耗时，必要时引入 `actor.Engine` 多线程池或再次拆分服务。
- **资源清理**：迁移时注意引用路径变更，避免遗留对 `server/service/dungeonserver` 的 import。
- **回滚策略**：保留 `DungeonServerGateway` 的远程模式配置，若单体模式不稳定，可以快速切回独立进程（前提是未删除旧代码）。
- **后续拆分**：若未来需要重新拆出 DungeonServer，只需：
  - 把 `dungeonactor` 目录复制为新服务的 `internal`；
  - 将 `DungeonServerGateway` 配置回 `Remote` 模式；
  - 恢复 `cmd/dungeonserver` + 启动脚本。

---

## 8. 开发任务清单（按步骤执行）

> 本清单只列“与 DungeonActor 整合直接相关”的工作，其他 Clean Architecture 与系统开关相关事项仍以 `docs/服务端开发进度文档_full.md` 为权威。

### 8.1 基础架构与目录搭建（仅 InProcess 模式）

- [✅] **创建 `dungeonactor` 目录与入口**
  - 在 `server/service/gameserver/internel/app` 下创建了 `dungeonactor` 目录，并新增 `dungeonactor/actor.go`，实现 GameServer 进程内的 `DungeonActor` 单 Actor 骨架：
    - 使用 `actor.ModeSingle` 与 `actor.NewActorManager` 创建全局唯一 DungeonActor，提供 `NewDungeonActor` 与 `GetDungeonActor` 便于 Gateway 等适配层在 InProcess 模式下访问。
    - 定义 `dungeonActorHandler`，当前 `Loop` 仅基于 `servertime.Now()` 输出调试日志作为 Tick 占位，后续会在此挂载实体/副本的 `RunOne`。
    - 提供 `Start/Stop/AsyncCall/RegisterRPCHandler` 四个方法：目前 `AsyncCall/RegisterRPCHandler` 为占位实现，仅记录日志与预留 handler 映射，保证 UseCase 接口稳定且不破坏 Clean Architecture 依赖方向。
  - 在 `server/service/gameserver/main.go` 中直接创建并启动 DungeonActor（与 PlayerActor/PublicActor 一同随 GameServer 生命周期管理），当前通过 `NewDungeonActor(actor.ModeSingle)` 接入，后续如需要可进一步抽象为 `app_init` 统一初始化流程。

- [✅] **实现仅支持 InProcess 的 DungeonServerGateway**
  - 在 `adapter/gateway/dungeon_server_gateway.go` 中将 `DungeonServerGateway` 改造为 InProcess 占位实现：删除对 `dungeonserverlink` 的直接依赖，通过 `dungeonactor.GetDungeonActor()` 将 `AsyncCall/RegisterRPCHandler` 路由到本地 DungeonActor。
  - `AsyncCall` 当前仅调用 DungeonActor 的占位实现并记录日志；`RegisterRPCHandler` 预留 Handler 注册入口；`IsDungeonProtocol/GetSrvTypeForProtocol/RegisterProtocols/UnregisterProtocols` 统一降级为本地模式下的轻量 stub（返回固定值并输出日志），确保现有 UseCase 与协议注册流程可以继续编译与运行，为后续接入真实消息派发与协议枚举迁移做好准备。

### 8.2 代码迁移（以 DungeonActor 为唯一战斗实现）

- [✅] **8.2-1：复制 DungeonServer 核心代码到 DungeonActor**
  - [✅] 已将 `server/service/dungeonserver/internel/{clientprotocol,devent,drpcprotocol,dshare,engine,entity,entitymgr,entitysystem,fbmgr,fuben,gameserverlink,iface,scene,scenemgr,skill}` 全量复制到 `internel/app/dungeonactor` 并保留原目录结构，覆盖副本/战斗/AI/属性/移动等全部系统。

- [✅] **8.2-2：调整依赖与包路径**
  - [✅] 批量将 import 中的 `postapocgame/server/service/dungeonserver/internel/...` 替换为 `postapocgame/server/service/gameserver/internel/app/dungeonactor/...`，并确认新包未反向依赖 `gameserver` 上层目录。
  - [✅] 迁移过程中统一使用 UTF-8 读写，避免中文注释或首行 `package` 被环境编码破坏。

- [✅] **8.2-3：确认 GameServer + DungeonActor 可以编译**
  - [✅] 在 `server` 目录执行 `go build ./service/gameserver/...`，确保所有新包均可独立编译。
  - [✅] 构建通过后，保留独立 `dungeonserver` 作为 Legacy 参考仓，但后续开发以 GameServer 内的 `dungeonactor` 为唯一代码基线。

- [✅] **8.2-4：在 GameServer 内以 ModeSingle 启动 DungeonActor**
  - [✅] 已在 `main.go` 中创建 `dungeonactor.NewDungeonActor(actor.ModeSingle)`，并在 GameServer 启动/停止流程中调用 `Start/Stop`，确保 DungeonActor 与 PlayerActor/PublicActor 同生命周期。
  - [✅] 启动 GameServer 时可见 `[dungeonactor] Start DungeonActor`、`RegisterRPCHandler stub: msgId=*`、`tick at ... (stub Loop, no game logic yet)` 等日志，说明本地 Actor 已进入 Loop 并可接收后续 Handler 接线。

### 8.3 Gateway 接线与最小闭环

- [✅] **8.3-1：实现 InProcess 派发逻辑**
  - [✅] 在 `adapter/gateway/dungeon_server_gateway.go` 中去掉对 `dungeonserverlink` 的直接依赖，实现 InProcess 版本的 `DungeonServerGatewayImpl`：`AsyncCall` 统一调用 `dungeonactor.GetDungeonActor().AsyncCall`，`RegisterRPCHandler` 将 D2G 回调注册到 DungeonActor 内部的 `rpcHandlers` 映射。
  - [✅] 在 `dungeonactor.DungeonActor` 中增加 `rpcHandlers` 映射与 `DungeonActorMessage` 类型：默认将 G2D 消息封装为 Actor 消息投递到单线程 Loop 中，由 `dungeonActorHandler.HandleMessage` 在 Actor 线程内分发，保持战斗逻辑的单线程语义。

- [✅] **8.3-2：跑通单条链路：进入副本 → 结算**
  - [✅] 在 `DungeonActor.AsyncCall` 中对 `G2DRpcProtocol_G2DEnterDungeon` 做特殊处理：在当前 goroutine 内反序列化 `G2DEnterDungeonReq`，直接构造 `D2GEnterDungeonSuccessReq` 与 `D2GSettleDungeonReq` 两个本地回调消息（最小实现：立即成功结算，奖励列表为空）。
  - [✅] 通过 `invokeRegisteredHandler` 调用在 `fuben_controller_init.go` 中通过 `DungeonServerGateway.RegisterRPCHandler` 注册的 D2G Handler（`HandleEnterDungeonSuccess` 与 `HandleSettleDungeon`），从而触发现有的副本记录/奖励回写逻辑。
  - [✅] 整体链路保持 Controller/UseCase 对实现细节无感知：`HandleEnterDungeon` 仍然只依赖 `DungeonServerGateway` 接口发起 G2D 调用，所有 InProcess 细节都封装在 Gateway 与 DungeonActor 内部，实现 C2SEnterDungeon → G2DEnterDungeon → D2GEnterDungeonSuccess/D2GSettleDungeon 的最小闭环（当前阶段不跑真实战斗流程，只验证进入/结算回调链路）。

### 8.4 扩展到移动 / 技能 / 掉落

- [✅] **8.4-1：移动链路迁移**
  - [✅] `MoveSys` 相关 Handler（`C2SStartMove/C2SUpdateMove/C2SEndMove` 处理逻辑）已随 8.2 阶段一并复制到 `internel/app/dungeonactor/entitysystem/move_sys.go`，并通过 `devent.Subscribe(OnSrvStart)` + `clientprotocol.Register` 在 DungeonActor 进程内注册，逻辑与原 DungeonServer 保持一致（像素坐标 → 格子坐标转换、位置校验、速度容差等）。
  - [✅] 通过增强 `dungeonactor.DungeonActor.Loop`，在单线程 Loop 中驱动 `entitymgr.RunOne(now)`，保证移动系统的 `MovingTime/flushAOIChanges` 等逻辑在 GameServer 进程内按帧执行，符合“单 Actor 驱动整服战斗”的约束。
  - [⏳] 后续在客户端直连 DungeonActor 的链路打通后（C2SStartMove/C2SUpdateMove/C2SEndMove 直接落到 DungeonActor），需在开发环境下用调试客户端做多次移动压力测试，记录 `DungeonActor` RunOne 耗时分布，当前阶段仅完成服务端侧的 RunOne 驱动与内部逻辑迁移。

- [✅] **8.4-2：技能与掉落迁移**
  - [✅] `fight_sys`、`buff_sys` 以及与技能释放/伤害计算相关的逻辑已完全复制到 `internel/app/dungeonactor/entitysystem/fight_sys.go` 等文件，并继续通过 `devent.OnSrvStart` 在 DungeonActor 内注册 `C2SUseSkill` 协议，RunOne 由增强后的 `DungeonActor.Loop` 驱动，技能释放/伤害结算与 Buff 叠加逻辑与原 DungeonServer 一致。
  - [✅] 掉落与拾取相关逻辑（`DropSys` 与 `handlePickupItem` 等）已迁移至 `internel/app/dungeonactor/entitysystem/drop_sys.go`，通过 `gameserverlink.CallGameServer` 触发 `D2GAddItem` RPC；在 InProcess 模式下，`gameserverlink.CallGameServer` 已被桥接为调用 DungeonActor 内部注册的 D2G Handler，从而直接驱动 GameServer 现有的 Bag/Fuben UseCase 完成发奖与背包写入。
  - [✅] 保持 GameServer 侧奖励与背包处理仍通过现有 UseCase 完成：本次仅在 DungeonActor 内通过 `gameserverlink.InProcessCallGameServer = d.callGameServerRPC` 打通 D2G 回调链路，并未修改任何 UseCase 接口或引入新的依赖，确保 Clean Architecture 依赖方向不变。

### 8.5 清理与收尾（可在后期执行）

- [ ] **8.5-1：停止构建旧 dungeonserver（软下线）**
  - [ ] 在你的构建脚本或项目配置中，移除 `server/service/dungeonserver` 可执行文件的构建目标。
  - [ ] 确认日常开发和联调流程中，只需要启动 GameServer + Gateway 即可完成战斗/副本相关功能。

- [✅] **8.5-2：物理删除旧 dungeonserver 代码（硬下线，可选）**
  - [✅] 使用全局搜索确认代码中已不再 import 或引用 `server/service/dungeonserver/...`（仅在文档与历史说明中保留提及，不再出现在 Go 代码 import 路径中）。
  - [✅] 物理删除 `server/service/dungeonserver` 目录下全部 Go 源码文件（入口、engine、entity、entitysystem、scene、skill、fuben 等），仓库中不再保留可编译的独立 DungeonServer 进程代码，仅通过 `internel/app/dungeonactor/*` 维护战斗/副本实现。
  - [✅] 在 `docs/服务端开发进度文档.md` 与 `docs/服务端开发进度文档_full.md` 中补充“DungeonServer 已下线，DungeonActor 为唯一战斗实现”的说明，并更新服务划分/构建章节；如需查看旧实现，可通过版本控制回滚至物理删除前的提交。

### 8.6 开发阶段推荐：只用 DungeonActor，不维护 dungeonserver（简化版路线）

> 适用场景：当前仅在本机做玩法验证/调试，不需要线上灰度与远程回滚能力，希望“尽快把战斗塞进 GameServer 里跑起来”，尽量不再动独立 `dungeonserver`。

- [✅] **8.6-1：仅实现 InProcess 版本的 DungeonServerGateway**
  - [✅] 保证 `DungeonServerGateway` 是 UseCase 唯一依赖的战斗接口实现，所有副本/战斗相关用例仍只依赖 `interfaces.DungeonServerGateway` 接口，不直接访问 DungeonActor 或底层 Actor 框架。
  - [✅] 内部只支持 InProcess：当前实现通过全局 `dungeonactor.GetDungeonActor()` 获取单例 DungeonActor，并直接调用其 `AsyncCall/RegisterRPCHandler`，不再持有远程 `dungeonserverlink`，也不再尝试区分远程/本地模式；如未来需要引入 Remote 模式，可在 Gateway 内重新增加 `mode` 和 `actorEngine/dungeonActorId` 注入，但不影响现有 UseCase 接口。

- [✅] **8.6-2：将现有 dungeonserver 当作“代码来源”**
  - [✅] 在 8.2 阶段已完成从 `server/service/dungeonserver/internel/{entity,entitysystem,scene,...}` 到 `internel/app/dungeonactor/{entity,entitysystem,scene,...}` 的完整拷贝与 import 前缀调整，DungeonActor 内部具备原 DungeonServer 的全部战斗/场景/AI/属性等能力。
  - [✅] 拷贝并接线完成后，GameServer 端业务代码仅依赖 `internel/app/dungeonactor/*`，不再 import `server/service/dungeonserver/...`；随后在 8.5-2 阶段物理删除了旧 `dungeonserver` 目录下的 Go 源码，简化为“只维护 DungeonActor，不维护独立 dungeonserver 可运行性”的开发路径。

- [✅] **8.6-3：GameServer 启动时始终拉起 DungeonActor**
  - [✅] 在 `server/service/gameserver/main.go` 中无条件创建 `dungeonactor.NewDungeonActor(actor.ModeSingle)` 并调用 `Start/Stop`，与 PlayerActor/PublicActor 一起作为基础设施 Actor 随 GameServer 生命周期常驻运行（当前未抽到 `app_init.go`，但行为等价）。
  - [✅] 将 DungeonActor 视为和 PublicActor 一样的基础设施 Actor，即便没有玩家进入副本也允许其空跑或轻量 Tick，战斗/副本逻辑由单 Actor Loop 驱动。
  - [✅] 玩家是否能进入副本/战斗，仍由玩家侧 Controller/SystemAdapter 的系统开启检查控制（通过 `GetFubenSys(ctx)` 等 Helper 判定），DungeonActor 本身不做“玩法开关”判断，仅负责战斗/场景执行。

- [✅] **8.6-4：保留未来重新拆出战斗进程的路径（文档级别即可）**
  - [✅] 在本文件与 `docs/服务端开发进度文档_full.md` 的服务划分/6.3 小节中，补充说明：如需重新拆出独立 DungeonServer，可从 `internel/app/dungeonactor/*` 复制出新服务的 `internal` 目录，并在 `DungeonServerGateway` 内重新增加 Remote 实现（例如通过 TCP 客户端或新的 RPC 层），同时保持现有 UseCase 接口不变。

- [✅] **8.6-5：整理你自己的最小实现顺序**
  - [✅] 按“8.1 → 8.2 → 8.3 → 8.4 → 8.5 → 8.6”顺序在本文件与 `docs/服务端开发进度文档_full.md` 中沉淀阶段性实施记录，作为后续维护与潜在回滚/再拆分的操作手册。
  - [✅] 实际执行顺序为：“先打通 C2SEnterDungeon → 简单结算的最小闭环（8.3），再扩展移动/技能/掉落等高频链路（8.4），确认 DungeonActor 能独立承载战斗负载后完成 dungeonserver 软/硬下线（8.5），最终在 8.6 阶段将日常开发路线收敛为仅维护 DungeonActor 的简化模式”。

此简化路线的核心理念是：**把 DungeonActor 当成 GameServer 内部的“战斗黑盒 Actor”，保持 UseCase/Gateway 架构不变，但短期完全不再维护独立 dungeonserver 的可运行性。**

### 8.7 DungeonServer & Proto 物理移除（最终收尾）

> 前置条件：8.1–8.4 已完成，所有战斗/副本逻辑只跑在 GameServer 内部的 DungeonActor 上；联调流程中不再需要启动独立 `dungeonserver` 进程。

- [ ] **8.7-1：确认运行时完全不依赖 dungeonserver**
  - [ ] 启动 Gateway + GameServer（含 DungeonActor），不启动 `dungeonserver`，用调试客户端完整跑一遍“进入副本 → 战斗/移动 → 结算”链路。
  - [ ] 全局搜索 `dungeonserverlink` 包，确认所有对该包的调用都已被移除或标记为待删除的死代码。

- [ ] **8.7-2：清理 Go 代码中对 dungeonserver 的依赖**
  - [ ] 全局搜索字符串 `server/service/dungeonserver`，确认没有任何 `import` 或路径引用该目录。
  - [ ] 确认 `server/service/gameserver/internel/infrastructure/dungeonserverlink` 不再被 `DungeonServerGateway` 或其他代码实际使用（可删除或在后续步骤一并移除）。
  - [ ] 执行一次 `go build ./server/...` 确认不再有 dungeonserver 相关的编译错误。

- [ ] **8.7-3：重构“Game ↔ Dungeon” RPC 为内部接口（去除 G2D/D2G 枚举依赖）**
  - [ ] 在 DungeonActor 或对应 usecase 接口中，定义内部方法（示例）：`EnterDungeon/SettleDungeon/SyncAttrs/UpdateSkill/PickupDrop`，入参使用 Go 结构体或内部 DTO，而不是直接依赖 `G2D*Req`/`D2G*Req`。
  - [ ] 在 `DungeonServerGateway` 中，增加对这些内部接口的封装：例如 `EnterDungeon(ctx, sessionID, roleID, params)` 直接调用 DungeonActor，而不是构造 `G2DEnterDungeonReq` 并使用 `G2DRpcProtocol_G2DEnterDungeon`。
  - [ ] 依次修改以下文件，使其只依赖内部接口，而不再依赖 G2D/D2G 枚举和 RPC message：
    - [ ] `internel/adapter/controller/fuben_controller.go`：改为调用 `gateway.EnterDungeon/SettleDungeon`，不再手动构造/反序列化 `G2DEnterDungeonReq`、`D2GSettleDungeonReq`、`D2GEnterDungeonSuccessReq`。
    - [ ] `internel/app/playeractor/entity/player_network.go`：移除 `D2GRegisterProtocolsReq`、`D2GUnregisterProtocolsReq`、`D2GSyncPositionReq`、`D2GSyncAttrsReq` 相关的 RPC Handler 注册，改为由 DungeonActor 内部管理协议注册与坐标/属性同步。
    - [ ] `internel/app/playeractor/entity/attr_calculator.go`：改为调用 `gateway.SyncAttrs(...)`，内部调用 DungeonActor，不再构造 `G2DSyncAttrsReq` 或使用 `G2DRpcProtocol_G2DSyncAttrs`。
    - [ ] `internel/adapter/system/skill_system_adapter.go`：改为调用 `gateway.UpdateSkill(...)`，不再构造 `G2DUpdateSkillReq` 或使用 `G2DRpcProtocol_G2DUpdateSkill`。
    - [ ] `internel/adapter/controller/bag_controller.go`：改为调用 `gateway.PickupDrop(...)` 返回结果，而非直接反序列化/返回 `D2GAddItemReq/Resp`。

- [ ] **8.7-4：确认代码中不再使用 G2D/D2G 相关 proto 类型**
  - [ ] 全局搜索 `G2DRpcProtocol_`，确认没有引用（或仅剩你刻意保留的极个别内部工具位置）。
  - [ ] 全局搜索 `D2GRpcProtocol_`，确认没有引用。
  - [ ] 全局搜索以下 message 名称，确认都不再使用：`G2DEnterDungeonReq`、`G2DSyncAttrsReq`、`G2DSyncGameDataReq`、`G2DUpdateHpMpReq`、`G2DUpdateSkillReq`、`D2GRegisterProtocolsReq`、`D2GUnregisterProtocolsReq`、`D2GSettleDungeonReq/Resp`、`D2GEnterDungeonSuccessReq`、`D2GAddItemReq/Resp`、`D2GSyncPositionReq/Resp`、`D2GSyncAttrsReq`。

- [ ] **8.7-5：从 proto/csproto 中移除 DungeonServer 专用 RPC 定义**
  - [ ] 打开 `proto/csproto/rpc.proto`，执行：
    - [ ] 删除整个 `enum G2DRpcProtocol` 定义。
    - [ ] 删除所有 `G2D*` 开头的请求 message：`G2DEnterDungeonReq`、`G2DSyncAttrsReq`、`G2DSyncGameDataReq`、`G2DUpdateHpMpReq`、`G2DUpdateSkillReq`。
    - [ ] 删除整个 `enum D2GRpcProtocol` 定义。
    - [ ] 删除所有 `D2G*` 开头的请求/响应 message：`D2GRegisterProtocolsReq`、`D2GUnregisterProtocolsReq`、`D2GSettleDungeonReq/Resp`、`D2GEnterDungeonSuccessReq`、`D2GAddItemReq/Resp`、`D2GSyncPositionReq/Resp`、`D2GSyncAttrsReq`。
  - [ ] 打开 `proto/csproto/srv_def.proto`，删除 `SrvTypeDungeonServer = 3;` 以及相关注释（保留 Gateway/GameServer 的枚举项）。
  - [ ] 运行 `proto/genproto.sh`（或当前生成脚本），然后 `gofmt`，重新生成 `server/internal/protocol/rpc.pb.go` 等文件。
  - [ ] 执行 `go build ./...`，修复因删除枚举/message 导致的残余引用错误（理论上 8.7-4 已经确保不会有）。

- [ ] **8.7-6：最终文档与架构描述更新**
  - [ ] 在 `docs/服务端开发进度文档.md` 与 `_full.md` 中：
    - [ ] 更新“服务划分”小节，将 `DungeonServer` 描述调整为 “GameServer 进程内的 `DungeonActor`（单 Actor 战斗引擎）”。
    - [ ] 在“DungeonServer → DungeonActor 整合”或 6.3 小节中，增加说明：“`server/service/dungeonserver` 已物理删除，Game ↔ Dungeon 的 RPC 协议已内聚为 GameServer 内部接口，不再通过 csproto 暴露。”
  - [ ] 在本文件 `docs/dungeonactor整合方案.md` 的 8.5/8.7 对应条目打勾，并在版本记录中（若有）增加一条“DungeonServer & Proto 物理移除完成”的记录。

---

> 本文为在 `docs` 目录下关于 DungeonActor 整合的唯一方案描述；
> 若在实施过程中对职责边界或迁移路径有任何调整，请同步更新本文件以及 `docs/服务端开发进度文档_full.md` 中的相关章节。
