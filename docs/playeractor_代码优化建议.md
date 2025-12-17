## 1. 文档目的

- **目标**：给 `server/service/gameserver/internel/app/playeractor` 的后续优化提供一份可执行 checklist，避免“口头建议”无法落地。
- **范围**：仅覆盖当前已存在代码的改造建议，不包含新玩法/新系统设计。
- **使用方式**：建议按“优先级从高到低 + 模块聚合”的顺序推进，每完成一条就在本地勾选对应复选框。

---

## 2. 优先级 A（建议优先处理）

### 2.1 Actor 适配层并发与错误处理

- [✅] **修正 `PlayerRoleActor` 中异步包装的错误处理与并发问题**  
  - 位置：`internel/app/playeractor/adapter.go`  
  - 问题点：
    - `SendMessageAsync` / `RemoveActor` / `Init` / `Start` / `Stop` 等方法使用 `routine.Run` 启动 goroutine，并在闭包内写入外层 `err` 变量，调用方几乎总是拿到默认值 `nil`，且存在 data race。
  - 建议改造：
    - 方案一（推荐，语义简单）：这些方法**改为同步调用** `actorMgr`，去掉 `routine.Run` 包裹，由调用方自行决定是否在更外层异步化。
    - 方案二（如确实需要异步）：方法签名改为“显式异步、不返回 error”，内部负责打日志，不再把 `error` 往外冒，避免“看起来同步、其实不可靠”的 API。

- [✅] **理顺 `NewPlayerRoleActor` / `Init` 的生命周期调用**  
  - 位置：`internel/app/playeractor/adapter.go`  
  - 问题点：
    - `NewPlayerRoleActor` 中显式调用了一次 `playerHandler.OnInit()`，`Init()` 成功后又调用了一次 `playerHandler.OnInit()`，语义含混，后续维护容易踩坑。
  - 建议改造：
    - 将 `OnInit` 的真正逻辑统一收敛到 `Init()`，`NewPlayerRoleActor` 只做“字段初始化、不带副作用”；  
    - 如确实需要“模板初始化 + 运行期初始化”两类行为，则拆分为两个明确命名的方法（例如 `InitTemplate()` / `InitRuntime()`），并在注释中解释清楚。

### 2.2 Runtime 依赖收敛与上下文注入

- [✅] **统一通过 `Runtime` 访问依赖，不再在 PlayerRole 中直接使用 `deps.NewXXX()`**  
  - 位置：`internel/app/playeractor/entity/player_role.go`  
  - 当前用法示例：
    - `NewPlayerRole` 中已经构造了 `runtime.NewRuntime(...)` 并挂在 `PlayerRole.runtime` 上；
    - 但 `sendPublicActorMessage` / `CallDungeonServer` 仍然直接调用 `deps.NewPublicActorGateway()` / `deps.NewDungeonServerGateway()`。
  - 建议改造：
    - 为 `runtime.Runtime` 增加必要的访问器（若尚未有）：
      - `PublicGateway() iface.PublicActorGateway`  
      - `DungeonGateway() iface.DungeonServerGateway`
    - 将 `sendPublicActorMessage` / `CallDungeonServer` 内的依赖访问改为通过 `pr.runtime` 获取，避免在 PlayerRole 中重新 new gateway。

- [✅] **增强 `PlayerRole.WithContext`，让上下文同时携带 Role / Runtime / SessionId**  
  - 位置：`internel/app/playeractor/entity/player_role.go`  
  - 当前行为：
    - 仅通过 `gshare.ContextKeyRole` 注入了 `IPlayerRole`，`Runtime` 和 SessionId 没有进入 `context.Context`。
  - 建议改造：
    - 在 `WithContext` 中追加：
      - `ctx = pr.runtime.WithContext(ctx)`，使 `runtime.FromContext(ctx)` 生效；  
      - `ctx = context.WithValue(ctx, gshare.ContextKeySession, pr.SessionId)`，方便日志和下游链路从 ctx 还原 Session。
    - 统一要求：Controller / SystemAdapter / UseCase 在需要 Context 时优先通过 `PlayerRole.WithContext` 构造，而不是直接 `context.Background()`。

### 2.3 SystemAdapter 依赖装配（以 Level 为切入点）

- [✅] **为 Level 系统引入本地 `Deps`，替代直接使用 `deps.NewXXX()`**  
  - 位置：`internel/app/playeractor/level/system.go`  
  - 当前问题：
    - `NewLevelSystemAdapter` 中多处直接调用 `deps.NewPlayerGateway()` / `deps.NewEventPublisher()` / `deps.NewConfigManager()` / `deps.NewNetworkGateway()`；
    - 业务方法（`GetLevel` / `GetExp` / `GetLevelData` / `CalculateAttrs` 等）也直接通过 `deps.NewPlayerGateway()` 访问仓储与配置。
  - 建议改造步骤：
    1. 在 `level` 包内定义 `type Deps struct { PlayerRepo iface.PlayerRepository; EventPublisher iface.EventPublisher; ConfigManager iface.ConfigManager; /* 如有需要可加入 NetworkGateway 等 */ }`。
    2. 将 `LevelSystemAdapter` 结构体增加一个 `deps Deps` 字段，并在构造函数中接受 `Deps` 参数：
       - `func NewLevelSystemAdapter(d Deps) *LevelSystemAdapter { ... }`
    3. `NewLevelSystemAdapter` 内部不再直接调用 `deps.NewXXX()`，全部改为使用传入的 `Deps`。
    4. 在系统注册处（当前在 `init()` 中注册工厂）调整为从 `Runtime` 构造 `Deps`：
       - 若短期无法直接拿到 Runtime，可先保留一个过渡层，例如 `level.NewDepsFromGlobal()`，但中期目标是从 `PlayerRole.GetRuntime()` 或 `runtime` 注入。

- [✅] **改造 Level System 的配置访问，统一走 `ConfigManager` 接口**  
  - 位置：`internel/app/playeractor/level/system.go`  
  - 当前问题：
    - `CalculateAttrs` / `CalculateAddRate` 直接调用 `jsonconf.GetConfigManager()`，属于全局单例访问，不利于测试与替换实现。
  - 建议改造：
    - 将配置访问封装到 `Deps.ConfigManager` 中，例如：
      - `cfg := a.deps.ConfigManager.GetLevelAttrsConfig(...)`  
      - `cfg := a.deps.ConfigManager.GetAttrAddRateConfig()`
    - 对应在 `iface.ConfigManager` 中补齐必要接口（如尚未定义），保证 UseCase / SystemAdapter 只依赖端口接口。

### 2.4 Controller 上下文与健壮性

- [✅] **统一 Controller 中对 SessionId 的获取方式，避免直接类型断言导致 panic**  
  - 位置：`internel/app/playeractor/controller/player_network_controller.go`  
  - 问题点：
    - `HandleEnterGame` / `HandleRunOneMsg` 等使用 `ctx.Value(gshare.ContextKeySession).(string)`，一旦上游缺失该键，会直接 panic。
  - 建议改造：
    - 抽一个安全 helper，例如：
      - `func getSessionID(ctx context.Context) (string, error)`，内部做 `.(string)` + 判空，返回统一错误码；
    - Controller 内全部改为使用该 helper，出现异常时：
      - 日志带上合理的错误信息；
      - 返回 `ErrorCode_Internal_Error` 或适当错误给客户端。

- [✅] **调用 DungeonActor 时传递带 Session / Role 的 Context，而不是 `context.Background()`**  
  - 位置：
    - `internel/app/playeractor/controller/player_network_controller.go`（`enterGame` 中调用 DungeonActor）；  
    - 以及将来新增的“PlayerActor → DungeonActor” Controller/UseCase 接口。
  - 建议改造：
    - 使用 `playerRole.WithContext` 或统一的 `newSessionContext` 构造 ctx，再传递给 `AsyncCall`；
    - 确保 DungeonActor 侧日志也能通过 ctx 还原出 Session / RoleId。

---

## 3. 优先级 B（中期可以顺手推进）

### 3.1 系统开启状态与 SysMgr 行为

- [ ] **为系统开启状态预留更精细的控制入口，替代“登录时全开”的实现**  
  - 位置：`internel/app/playeractor/entitysystem/sys_mgr.go`  
  - 当前实现：
    - `OnRoleLogin` 中调用 `CheckAllSysOpen`，遍历所有已挂载系统，对每个系统做：
      - `iPlayerRole.SetSysStatus(system.GetId(), true)`  
      - `system.SetOpened(true)`
    - 实际效果是“所有挂载系统一律开启”，和“按等级 / 任务 / 配置控制系统开放”的目标有差距。
  - 建议改造思路：
    - 保留 `CheckAllSysOpen` 作为迁移期兜底，但新增更细粒度的接口，例如：
      - `EnsureSysOpen(ctx context.Context, sysId uint32) error`：根据配置 / 等级 / 任务状态决定是否真正开启系统，并更新 `SysOpenStatus`；
    - Controller 层在处理协议前，统一通过 `GetXXXSys(ctx)` + `EnsureSysOpen` 检查系统是否开启，未开启时用统一错误码返回。

- [ ] **为 `GetDefaultSystemIds` 引入配置化或最少文档化约束**  
  - 位置：`internel/app/playeractor/entitysystem/system_registry.go`  
  - 现状：
    - 默认系统列表硬编码为 Level/Bag/Equip/Skill/Money/FuBen/Message。
  - 建议：
    - 短期：在代码注释中明确“新增系统时需要更新这里，否则不会被挂载”；  
    - 中期：考虑从配置文件或环境变量读取系统列表，以支持“按服裁剪玩法”。

### 3.2 `deps` 工厂与 `runtime` 的职责边界

- [✅] **限制 `deps.NewXXX()` 的使用范围，防止新的隐式全局依赖回潮**  
  - 位置：`internel/app/playeractor/deps/deps.go` + 各调用点  
  - 现状：
    - `deps` 已经从“全局单例容器”转为“工厂函数集合”，这是正确方向；
    - 但 `playeractor` 下仍有大量业务代码直接调用 `deps.NewXXX()`，没有经过 `Runtime` / `Deps` 注入。
  - 建议：
    - 约定：`deps.NewXXX()` 只允许在以下场景使用：
      - `Runtime.NewRuntime` 内部；
      - 极少量 bootstrapping 代码（例如为 `register.RegisterAll` / 系统工厂构建初始依赖）。
    - 对于 SystemAdapter / Controller / Service 层：
      - 一律通过 `runtime.Runtime` 或各自定义的 `Deps` 结构注入依赖；
      - 在 Code Review checklist 中增加一条：“新代码禁止直接使用 `deps.NewXXX()`，需说明理由”。

---

## 4. 优先级 C（长期改进方向）

### 4.1 Controller → Repository 抽象的补全

- [✅] **将 PlayerActor Controller 直接访问数据库的逻辑收敛到 Repository 接口**  
  - 位置：`internel/app/playeractor/controller/player_network_controller.go`  
  - 已完成：
    - ✅ 在 `runtime.Runtime` 中添加了 `RoleRepository` 字段和 `RoleRepo()` 访问器
    - ✅ 将 `HandleEnterGame` 中的 `database.GetPlayerByID(...)` 替换为通过 `RoleRepository.GetRoleByID(...)` 接口访问
    - ✅ Controller 优先从 Runtime 获取 RoleRepository，如果 Runtime 不可用则使用 `deps.NewRoleRepository()` 作为回退（符合 3.2 的约定）
  - 待完善（长期）：
    - `SavePlayerActorMessage` 仍直接调用 `database.SavePlayerActorMessage`，后续可考虑创建 `MessageRepository` 接口统一消息存储访问

### 4.2 属性与战力计算的职责再梳理

- [✅] **梳理 `LevelSystemAdapter` 中与属性计算相关的逻辑，评估是否下沉到专门的属性服务**  
  - 位置：`internel/app/playeractor/level/system.go` + `internel/app/playeractor/attrcalc/*`  
  - 已完成：
    - ✅ 在 `CalculateAttrs` 和 `CalculateAddRate` 方法中添加了职责说明注释，明确当前实现保留在 Level 系统的原因（需要访问等级数据和配置）
    - ✅ 注释中说明了后续改进方向：可考虑创建独立的属性计算服务，或将属性计算逻辑下沉到 UseCase 层
  - 当前状态：
    - Level 系统中的 `CalculateAttrs` / `CalculateAddRate` 通过 `attrcalc.Register` 注册到属性计算总线，符合当前架构设计
    - 这些方法需要访问 Level 系统的数据（等级、配置），暂时保留在 Level 系统中是合理的
    - 后续如需进一步解耦，可创建专门的属性计算服务，Level 只提供数据访问接口

---

## 5. 建议实施顺序（参考）

- **阶段 1（短期，1~2 次迭代内完成）**
  - 完成 2.1/2.2（Actor 并发与 Runtime/Context 收敛），确保框架层没有明显 bug，日志和依赖访问路径统一。
- **阶段 2（中期，按系统逐步推进）**
  - 以 Level 系统为模板完成 2.3/3.2，之后复用同一模式推广到其他系统（Money/Bag/Skill/Equip/Fuben/Recycle）。
- **阶段 3（长期，视需求排期）**
  - 根据业务需要推进 3.1/4.1/4.2，为系统开关/Repository 抽象/属性系统进一步演进留出空间。

> 后续每完成一批改造，建议同步更新：  
> - `docs/服务端开发进度文档.md` 中的「已完成功能」和「开发注意事项与架构决策」；  
> - 如涉及 SystemAdapter/Runtime 原则变更，也请同步更新：`docs/已完成/gameserver_adapter_system演进规划.md`。


