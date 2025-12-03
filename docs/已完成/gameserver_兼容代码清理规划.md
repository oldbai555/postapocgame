# GameServer 向后兼容代码清理规划

## 0. 文档目的与阅读须知

- **文档目的**：梳理 `server/service/gameserver` 中仍然存在的向后兼容/过渡代码，给出分阶段的“新增新架构 → 切流量 → 删除旧实现”路线，避免再次在旧层上叠加逻辑。  
- **适用范围**：只覆盖 GameServer 进程（`server/service/gameserver`），不讨论 Gateway / DungeonServer 的兼容代码。  
- **前置阅读**：建议先阅读：  
  - 《`docs/服务端开发进度文档.md`》第 6.1、7 章  
  - 《`docs/gameserver_CleanArchitecture重构文档.md`》与《`docs/gameserver_CleanArchitecture重构实施手册.md`》  
- **执行原则**：任何删除操作必须满足：**编译通过 + 关键链路自测通过 + 文档同步**（本文件 + 进度文档 4/6/7/8 章）。  

---

## 1. 总览：当前仍存在的兼容/过渡层

> 以下目录/文件是当前确认仍存在**旧实现、兼容桥接或历史空壳**的主要位置；本规划按“可立即删 → 需先重构 → 防退化”三类给出路线。

- **空壳目录（已无代码，仅历史残留）**
  - `internel/domain/vip/`
  - `internel/domain/dailyactivity/`
- **旧 EntitySystem / 过渡胶水**
  - `internel/app/playeractor/entitysystem/sys_mgr.go`（系统 ID/初始化顺序管理，已瘦身，但仍是入口）
  - `internel/app/playeractor/entitysystem/message_dispatcher.go`（离线消息分发器）
- **仍带有“兼容色彩”的工具/适配层**
  - `internel/adapter/system/gm_tools.go`（GM 工具，用于背包满奖励等兼容场景）
  - AntiCheat 旧实现（位于 `entitysystem`，待完全 Clean 化，具体文件名以实际代码为准）
  - MessageSys 相关：
    - `internel/adapter/system/message_system_adapter.go`
    - `internel/app/engine/message_registry.go`
    - `internel/app/playeractor/entitysystem/message_dispatcher.go`

---

## 2. 阶段一：立刻可清理项（无业务风险）

### 2.1 删除空壳领域目录

- **目标**：删除已物理清空但仍保留的领域目录，降低新人阅读干扰。  
- **对象**：
  - `server/service/gameserver/internel/domain/vip/`
  - `server/service/gameserver/internel/domain/dailyactivity/`

> 2025-12-03：已完成。上述两个目录已从代码仓库物理删除，仅为目录结构清理，不涉及任何 `.go` 源码文件变更。

**操作步骤**：
1. 确认目录下无 `.go` 文件（当前已为空壳，仅目录存在）。  
2. 删除整个目录。  
3. 执行：`go list ./...` 或 `go build ./server/service/gameserver`，确认无编译错误。  
4. 在《`docs/服务端开发进度文档.md`》第 6.1 小节备注“空壳目录已物理删除”，并在版本记录中补一行说明。  

**验收标准**：
- GameServer 编译通过。  
- 代码搜索 `internel/domain/vip`、`internel/domain/dailyactivity` 无引用。  

---

## 3. 阶段二：AntiCheat 旧 EntitySystem 清理规划

> 目标：将 AntiCheat 从旧 `entitysystem` 迁移到 Clean Architecture 分层（Domain + UseCase + SystemAdapter），完成后删除旧 Sys 文件。  
> 现状：当前代码仓库中已不存在 `playeractor/entitysystem/anti_cheat_sys.go` 等 AntiCheat 旧实现，仅在 Proto 层保留了 `SiAntiCheatData` 与 `PlayerRoleBinaryData.anti_cheat_data` 字段，尚未落地实际防作弊系统代码。本阶段暂作为**未来新增 AntiCheat 系统时的架构规范指引**，无需再做“旧实现迁移与删除”的清理工作。

### 3.1 目标结构

- **Domain 层**：`internel/domain/anti_cheat/*`
  - 定义 AntiCheat 领域实体（计数器、窗口、封禁记录等）。
  - 提供纯函数/方法用于：频率统计、是否超过阈值、封禁决策。
- **UseCase 层**：`internel/usecase/anti_cheat/*`
  - 提供操作入口：`CheckOperationFrequencyUseCase`、`RecordSuspiciousBehaviorUseCase`、`BanPlayerUseCase` 等。
  - 依赖接口（在 `usecase/interfaces` 中声明）：
    - `AntiCheatRepository`（持久化/加载作弊数据，如需）
    - `TimeProvider`（统一使用 `servertime`）
    - `PlayerRepository` / `GMUseCase`（若需要触发封禁通知/GM 操作）。
- **SystemAdapter 层**：`internel/adapter/system/anti_cheat/*`
  - 实现 AntiCheatSystemAdapter：
    - 生命周期：`OnInit`（加载数据）、`OnRunOne`（定期清理/过期窗口）、必要事件订阅。  
    - 对外提供 helper：`GetAntiCheatSys(ctx)`；内部仅调 UseCase，不写规则。  

### 3.2 迁移步骤

> 注：以下步骤在“已有 AntiCheat 旧 EntitySystem 实现”场景下适用；当前仓库已无相关 Sys 文件，未来若直接按 Clean Architecture 方式新增 AntiCheat，可从第 2 步开始执行。

1. **锁定旧实现位置（仅适用于已有 legacy Sys 的历史版本）**  
   - 在 `internel/app/playeractor/entitysystem` 下查找 AntiCheat 相关文件（例如 `anti_cheat_sys.go`）。  
   - 通读当前逻辑：频率控制点、封禁条件、数据存储方式。  
2. **抽取 Domain（未来新增 AntiCheat 时直接按此结构设计）**  
   - 在 `internel/domain/anti_cheat` 新建领域模型：计数结构、时间窗口、封禁状态枚举等。  
   - 将纯业务计算（如“10 秒 100 次”、“是否封禁”）搬进 Domain。  
3. **编写 UseCase**  
   - 在 `internel/usecase/interfaces` 增加 `anti_cheat.go`，定义需要的接口。  
   - 在 `internel/usecase/anti_cheat` 下实现对应用例，将 Actor/日志等细节通过接口注入。  
4. **实现 SystemAdapter**  
   - 在 `internel/adapter/system/anti_cheat` 新建 adapter：注册到 `sys_mgr`，在生命周期中调用 UseCase。  
5. **切换调用方**  
   - 将原来直接访问 AntiCheat EntitySystem 的调用（例如在其他系统里做频率检测）改为通过 UseCase 接口或 `GetAntiCheatSys(ctx)`。  
6. **删旧文件（仅适用于已有 legacy Sys 的历史版本）**  
   - 确认无任何引用后，删除 `entitysystem/anti_cheat_sys.go` 等旧实现文件。  
   - 更新《`docs/系统移除方案.md`》增加“AntiCheat 移除/迁移”章节。  

### 3.3 验收标准

- `usecase/anti_cheat` 有单元测试，覆盖关键规则。  
- 所有调用 AntiCheat 的地方均通过 UseCase/Adapter，不再 import `entitysystem` 旧文件。  
- GameServer 编译以及常规登录/玩法链路自测通过。  

---

## 4. 阶段三：MessageSys 兼容层收缩规划

> 现状：MessageSys 已通过 `adapter/system/message_system_adapter.go`、`engine/message_registry.go`、`message_dispatcher.go` 完成“玩家消息系统（离线回放）”的 Clean 化；当前实现路径已经是“数据库 → 注册中心 → 回调”的单一路径，不再存在旧 EntitySystem 的兼容分支。

### 4.1 目标

- 保留 MessageSys 作为**框架级消息回放机制**，但：  
  - 明确 SystemAdapter 仅负责生命周期 + 调度；  
  - 对外暴露统一的 `DispatchPlayerMessage` 入口；  
  - 强化监控/清理策略后，删除不再使用的兼容分支或过期代码路径。

### 4.2 需要关注的关键文件

- `internel/adapter/system/message_system_adapter.go`  
- `internel/adapter/system/message_system_adapter_init.go`  
- `internel/app/engine/message_registry.go`  
- `internel/app/playeractor/entitysystem/message_dispatcher.go`  

### 4.3 收缩步骤

1. **梳理现有调用点（已完成，2025-12-03）**  
   - `DispatchPlayerMessage`：由 `playeractor/entity/player_network.go` 中的 `handlePlayerMessageMsg`（在线消息）与 `MessageSystemAdapter.loadMsgFromDB`（离线消息回放）统一调用。  
   - `MessageSystemAdapter`：在 `adapter/system/message_system_adapter.go` 中实现生命周期逻辑（OnInit/OnRoleLogin/OnRoleReconnect/OnNewDay/RunOne），通过 `message_system_adapter_init.go` 注册为 `SysMessage`。  
   - `message_registry`：`engine/message_registry.go` 仅负责注册/查询回调与 Proto 工厂，不含业务逻辑；当前仓库中尚未有具体业务消息类型注册，未来如有需要可按 UseCase 维度注册。  
2. **监控与过期策略落地（部分完成，2025-12-03 前已实现）**  
   - 数量/过期策略：`MessageSystemAdapter` 中已实现：  
     - `OnNewDay` 调用 `database.DeleteExpiredPlayerActorMessages(MessageExpireDays)` 清理超过 7 天的消息。  
     - `RunOne` 周期性调用 `GetPlayerActorMessageCount` 与 `DeleteOldestPlayerActorMessages`，保证每个玩家最多保留 `MaxPlayerActorMessages=1000` 条消息。  
   - 控制台/监控：仍按《服务端开发进度文档》6.1 中“玩家消息系统阶段四”的规划，在后续迭代中补充控制台查看与监控指标（当前仅日志输出，无专门监控项）。  
3. **删除兼容分支/死代码（已完成，2025-12-03）**  
   - 现有 `message_dispatcher.go` 中仅保留：  
     - `engine.GetMessagePb3` → 解码 Proto（若有数据）；  
     - `engine.GetMessageCallback` → 触发注册中心回调，未注册时输出告警并安全返回；  
   - 代码中已不存在旧 EntitySystem 的分支逻辑或特殊兼容入口，无需进一步删除。  
4. **文档同步（已完成，2025-12-03）**  
   - 在本文件与《服务端开发进度文档》7.2/8.2 节中，MessageSys 已按“框架级消息回放机制 + Adapter/Registry/Dispatcher 三段式”记为最终形态，且明确不再依赖旧 EntitySystem。

### 4.4 验收标准

- 所有玩家离线消息场景（重连、OnInit、OnNewDay）验证通过。  
- 无任何代码再 import 已删除的旧 MessageSys 文件。  
- 日志/监控可观察到消息数量/过期清理行为。  

---

## 5. 阶段四：GM Tools 兼容逻辑清理规划

> 现状：GM 系统已按 Clean Architecture 落地，`adapter/system/gm/gm_manager.go` 负责命令注册与执行，`adapter/system/gm/gm_tools.go` 仅保留 GM 命令复用的通知/邮件 Helper，并通过 `usecase/mail` 完成发货逻辑；不再存在旧 EntitySystem 层的 GM 实现或直连数据库/背包的兼容代码。

### 5.1 目标

- 保持 GM 相关“业务规则”位于 UseCase 层（例如系统邮件发放在 `usecase/mail` 中实现），`gm_tools.go` 仅作为：  
  - 与 GM 命令解析紧耦合的薄封装 Helper（如 `SendSystemNotification*` / `SendSystemMail*`）；  
  - 或后续进一步精简时的删除候选（由 UseCase 或 Gateway 直接提供等价能力）。  

### 5.2 迁移步骤（2025-12-03 状态）

1. **审查 `gm_tools.go` 内容**  
   - 当前导出函数主要包括：`SendSystemNotification*`、`SendSystemMail*`、`SendSystemMailByTemplate*`、`HandleGMCommand` 等。  
   - 其中系统邮件相关逻辑已全部委托给 `usecase/mail`（`NewSendCustomMailUseCase` / `NewSendTemplateMailUseCase`），自身不再编写发货规则，仅负责拼装上下文与仓储。  
2. **为每类业务建立 UseCase**  
   - 发放奖励/系统邮件：已在 `usecase/mail` 中通过 UseCase 完成实现，`gm_tools.go` 仅作为 GM 入口的 Helper。  
3. **切换调用点**  
   - GM 命令实现统一在 `gm_manager.go` 中通过命令注册调用 `gm_tools` 提供的 Helper，不再直接操作 EntitySystem 或数据库。  
4. **后续精简方向（可选，尚未执行）**  
   - 若未来希望进一步下沉 GM 入口，可考虑：  
     - 将通知/邮件 Helper 下沉到专门的 Gateway/UseCase；  
     - 或保持当前结构，但在 Code Review 中将 `gm_tools.go` 视为薄辅助层，禁止新增业务规则。  
5. **文档同步**  
   - 在《服务端开发进度文档.md》6.1/7.10 节中，将 GM 系统描述为：命令/工具已 Clean Architecture 化，GM 工具函数只作为系统邮件/通知的入口 Helper，不再承担底层逻辑。  

### 5.3 验收标准

- 所有 GM 指令（尤其是发货、改资源类）链路通过 `gm_manager` + 对应 UseCase（如 `usecase/mail`）完成，测试验证通过。  
- `gm_tools.go` 中不再包含可下沉到 UseCase 的业务规则逻辑，仅保留与 GM 命令解析和上下文组装紧耦合的 Helper；如后续删除 Helper，则由 UseCase/Gateway 提供等价能力。  

---

## 6. 阶段五：防退化与新代码约束

> 本阶段不删除已存在代码，而是防止“旧风格”代码再次出现。

### 6.1 禁止新增 `entitysystem/*_sys.go` 业务逻辑

- **约束**：  
  - GameServer 新增功能不得在 `internel/app/playeractor/entitysystem` 下创建新的 `*_sys.go`。  
  - 所有新系统必须通过：`domain/{system}` + `usecase/{system}` + `adapter/system/{system}` 实现。  
- **执行方式**：  
  - 在 Code Review 清单（`docs/SystemAdapter_CodeReview清单.md`）中加入检查项：  
    - “是否在 entitysystem 目录新增了系统实现文件？”  
  - 对违反约束的变更直接退回。  

### 6.2 SystemAdapter 防退化

- 依照《服务端开发进度文档》7.10 与 `BaseSystemAdapter` 注释：  
  - SystemAdapter 仅允许：生命周期管理、事件订阅、调用 UseCase；  
  - 禁止直接写业务规则、配置解析、网络/数据库操作。  
- 建议在 CI 或本地脚本中增加简单的 grep 检查：  
  - 在 `adapter/system` 下查找 `jsonconf.` / `database.` / `gatewaylink.` 等关键字，作为人工 review 提示。  

---

## 7. 执行顺序建议

1. **先做阶段一**：删除空壳目录（立刻收益，零风险）。  
2. **评估并规划 AntiCheat**：若近期有防作弊需求，优先完成阶段二，顺手把旧 AntiCheat EntitySystem 清掉。  
3. **根据开发节奏推进 MessageSys 与 GM Tools**：  
   - 若近期在做玩家消息系统，就顺带完成阶段三；  
   - 若近期在做运营/GM 相关功能，就顺带完成阶段四。  
4. **长期坚持阶段五的防退化规则**，确保不会再在 `entitysystem` 下长出新的“旧风格系统”。  

---

## 8. 文档与进度同步要求

每完成一个阶段或删除一块兼容代码，需要同步：

1. 《`docs/服务端开发进度文档.md`》：  
   - 第 4 章“已完成功能”：补充简要说明和关键代码位置。  
   - 第 6.1 章“Clean Architecture 重构”：更新复选框状态。  
   - 第 7 章“开发注意事项与架构决策”：必要时新增或更新约束条目。  
   - 第 8 章“关键代码位置”：如有核心入口变化需更新。  
2. 《`docs/gameserver_CleanArchitecture重构实施手册.md`》：  
   - 在相关子章节（AntiCheat、MessageSys、GM）补充“已完成/已删除”的说明。  
3. 本文档：  
   - 在对应阶段小节标注完成日期与简要结果，便于后续追踪。  

---

> 建议每次实际删除代码时，都在提交信息中带上本文件对应的小节编号（例如“cleanup: remove legacy AntiCheat entitysystem (doc 3.x)”），方便未来审计和回溯。


