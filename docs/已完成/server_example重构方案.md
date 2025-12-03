# server/example 重构方案（依照 server/service 架构）

> 版本：2025-11-26  
> 适用范围：`server/example` 文本冒险调试客户端  
> 责任人：server/example 维护者

---

## 0. 文档目的与阅读须知

- 本文是 `server/example` 的唯一重构方案，请在开工前与 `docs/服务端开发进度文档.md` 第 0/7 章同步阅读。  
- 所有调试客户端代码仍视为服务端工程，必须复用 `server/service` 既有基础设施（Actor、网络、时间源、配置）。  
- 本方案拆分为三阶段（架构对齐 → 移动链路增强 → 业务扩展），每阶段完成后需同步主进度文档的“已完成功能/待实现/注意事项/关键代码位置”四章。

---

## 1. 参考源： 客户端移动模拟梳理

| 步骤 | 关键函数 | 要点 |
| ---- | -------- | ---- |
| 1. 入口 | `Move.MoveTo` | 记录目标场景/坐标 → 如需跨场景先触发 `PostActorChangeScene`。 |
| 2. 直线段拆解 | `findTitleLineEndPoint` + `astar.FindRoads` | 以 A* 找到路径拐点，再调用 `titleLinePixel` 检测分段直线是否可行走，得到下一段终点。 |
| 3. 移动参数缓存 | `MoveSt.SetMoveLen/SetNpXY` | 保存当前分段的像素终点和距离，供时间驱动计算。 |
| 4. 启动/广播 | `StartMove` (`output.PostStartMove`) | 立即广播 `C2SStartMove`，服务端与旁观者同步开始状态。 |
| 5. 时间驱动更新 | `calcRunMoveAfterPoint` | 基于属性 `speed` 与 `time.Now().UnixMilli()`（后续应替换为 `servertime`）计算本段移动后的像素坐标。 |
| 6. 位置回写 | `ActorLocationUpdate` (`PutActorLocationUpdate`) | 定期上报 `C2SUpdateMove`，同时 `Move.MoveTo` 递归触发下一段移动，形成“Start → Update* → Stop”的循环。 |
| 7. 停止回调 | `Move.StopMove` + `MoveToCallBack` | 抵达目标后发送 `C2SEndMove` 并执行回调，可串联更多行为（AI、寻路）。 |

**关键特性总结**
1. **分段直线 + 路径兜底**：优先直线穿越，受阻后回退到 A* 路径中的下一个合法节点，保证客户端移动既“像玩家”又能绕障碍。  
2. **速度/距离时间驱动**：每次更新基于速度与经过的毫秒数推算坐标，避免纯粹依赖客户端输入造成抖动。  
3. **Start/Update/End 三段式协议完整复刻**：任意一段失败都会重新发起 `MoveTo`，确保服务端实体状态与客户端保持一致。  
4. **移动回调链**：移动结束后能立即触发下一步（继续寻路、释放技能等），方便脚本化测试其他业务。

---

## 2. 现状问题（server/example）

1. **目录未对齐 server/service**：所有源码直接堆在 `server/example` 根目录，缺少 `cmd/internal/pkg` 分层，后续难以拆模块测试。  
2. **移动实现过于理想化**：`NudgeMove` 仅按轴向匀速推进，`sendMoveChunk` 也只做格子级直线移动，缺少 `MoveTo` 式分段、容错和重试。  
3. **缺失移动脚本回调**：无法在移动完成后自动执行下一个动作，业务联调（拾取、怪物交互）只能人工输入。  
4. **业务覆盖窄**：除 `register/login/roles/create/enter/move/attack` 外，没有对接背包、任务、副本、GM 等系统，无法在单仓复现更多链路。  
5. **状态监听散落**：AOI、技能、时间同步在 `GameClient` 内部混杂，缺少“系统化”封装，不利于扩展与复用。

---

## 3. 重构目标与目标架构

### 3.1 目标
1. **与 server/service 同构**：采用与 `gameserver/dungeonserver/gateway` 相同的 `cmd + internal + pkg` 布局，统一 Actor、日志、时间源使用方式。  
2. **移动链路对齐 **：支持分段路径、Start/Update/End 全流程重放、速度/容差校验与自动重试。  
3. **业务能力分层**：将账号/场景/战斗/背包等封装为独立“系统（System）”，面板命令仅调用系统接口。  
4. **脚本化友好**：提供移动和指令的回调/队列能力，可串联自动化流程。  
5. **文档 / 配置 / 测试闭环**：所有新增命令、关键流程和配置写入文档，并附最小验证步骤。

### 3.2 目标目录

```
server/example/
├── cmd/example/main.go               # 入口，初始化日志/配置
├── internal/
│   ├── panel/                        # AdventurePanel & 菜单/命令路由
│   ├── client/
│   │   ├── manager.go                # ActorManager 封装
│   │   ├── game_client.go            # 账号/场景/状态（拆成子系统）
│   │   ├── move_runner.go            # 移动模拟
│   │   ├── flow.go                   # 协程-safe 的协议流通道
│   ├── systems/
│   │   ├── account.go                # register/login/role
│   │   ├── scene.go                  # enter/look/AOI
│   │   ├── move.go                   # Start/Update/End 封装
│   │   ├── combat.go                 # attack/skill
│   │   ├── inventory.go              # (待扩展) 背包/物品
│   │   └── dungeon.go                # (待扩展) 副本/脚本
│   └── scripting/
│       └── sequence.go               # 指令队列/回调（移动完成触发下一步）
├── pkg/
│   └── clamp/..., shared utils       # 仅当公共逻辑与 server/internal 不重复时使用
└── docs/                             # 与本方案配套的使用说明（引用主文档）
```

---

## 4. 分阶段实施计划

### Phase A：架构对齐（Actor/目录/系统化）
> ✅ 已完成（2025-11-26）：`server/example` 现已采用 `cmd/example` + `internal/{client,panel,systems}` 分层；`GameClient` 拆分为 `client.Core`（协议/状态/Flow Waiter）+ `systems` 包（Account/Scene/Move/Combat）；`AdventurePanel` 命令迁移至 `panel/actions.go` 并统一由系统接口驱动。

1. ✅ **目录迁移**：新建 `cmd/example/main.go` 与 `internal/...` 结构，原文件平移并按模块拆分。  
2. ✅ **系统接口化**：将 `GameClient` 拆分为 `client.Core`（连接、收发、状态） + `systems` 包；面板通过接口调用而非直接访问字段。  
3. ✅ **面板命令路由**：把 `AdventurePanel` 的命令实现迁移到 `panel/actions.go`，便于后续脚本重用。  
4. ✅ **流式响应管理**：为注册/登录/进入场景/AOI/技能等回包定义统一 `flow.Waiter`，避免散落通道。

### Phase B：移动链路增强（对齐 ）
> ✅ 已完成（2025-11-26）：`internal/client/move_runner.go` 新增 MoveRunner，支持 A* 寻路+直线裁剪、服务端同步检测、自动重试与回调；`MoveSystem` 与 `AdventurePanel` 已全面切换到新的异步 API。

1. ✅ **MoveRunner**：参考 `/system/move.go`，实现 `MoveTo`（带路径/直线校验）与 `MoveBy/Resume`，完整驱动 `C2SStart/Update/EndMove`。  
2. ✅ **路径裁剪与容错**：接入 `jsonconf.GameMap`，优先直线可达，否则自动 A* 生成路径。  
3. ✅ **速度/时间驱动**：基于 `servertime.Now()` 计算时间片，发送逐格 Update，容忍 1.5× 速度与 200ms 延迟。  
4. ✅ **回调链**：`MoveCallbacks` 可在移动完成/中止后串接后续脚本。  
5. ✅ **自动重试**：检测服务器坐标漂移后自动重算路径，最多 2 次重试。  

### Phase C：业务扩展（“其他业务”对齐服务端系统）
> ✅ 已完成（2025-11-26）：新增 Bag/GM/Dungeon 系统封装，面板增加 `bag/use-item/pickup/gm/enter-dungeon` 命令；录制/回放脚本通过 `script-record/script-run` 实现；`ScriptSystem` 保留 Demo 巡逻示例。

1. ✅ **背包/物品指令**：`InventorySystem` + `bag/use-item/pickup` 命令直接映射 `C2SOpenBag/C2SUseItem/C2SPickupItem`。  
2. ✅ **副本/战斗链路**：`DungeonSystem.Enter` 对接 `C2SEnterDungeon`，面板命令 `enter-dungeon` 快速验证。  
3. ✅ **GM/调试桥接**：`GMSystem.Exec` 与 `gm <name> [args...]` 命令复用 GameServer GM 协议。  
4. ✅ **脚本回放**：新增 `script-record`/`script-run`，可将命令写入文件并按延迟回放；`ScriptSystem.RunDemo` 继续提供示例巡逻。  
5. ✅ **录制控制**：AdventurePanel 支持录制状态切换，默认输出到指定文件，执行脚本时自动暂停录制。  

---

## 5. 移动系统对齐细节

| 能力 |  做法 | server/example 目标实现 |
| ---- | -------------- | ----------------------- |
| 坐标体系 | 客户端内部使用像素，判定时转格子；调用 `base.Pixel2Grid`。 | 继续沿用 `argsdef.TileCoordToPixel/PixelCoordToTile`，移动模块输出格子坐标供面板展示。 |
| 路径寻路 | `astar.FindRoads` + 直线段校验。 | 复用 `server/service/dungeonserver/internel/entitysystem/pathfinding.go` 的逻辑，或实现轻量直线检测程式。 |
| 时间驱动 | `speed*(now-last)/1000` 推算像素位移。 | 使用 `servertime.Now()`，每次 `MoveRunner.Tick` 更新 `posX,posY` 并发送 `C2SUpdateMove`。 |
| 回调 | `MoveToCallBack` 串联行为。 | MoveRunner 暴露 `OnComplete/OnAbort`，供 `scripting.Sequence` 挂接。 |
| 容错 | 移动过快会 `ContinueMove`。 | 当服务端返回的格子与客户端预期不同或 `S2CEntityStopMove.StateFlags` 标记移动异常时自动重试。 |

---

## 6. 关键交付物与测试

1. **文档同步**：每个阶段完成后更新 `docs/服务端开发进度文档.md`（章节 4/6/7/8）及 `docs/golang客户端待开发文档.md`。  
2. **演示脚本**：提供不少于 2 条示例脚本：  
   - `移动 → 拾取 → 攻击` 连锁；  
   - `多段寻路 → 副本匹配 → 技能循环`。  
3. **回归测试**：在本地启动三服后，跑通 `register/login/enter/move/attack` 与新增命令，确保日志面板正常回显。  
4. **配置模板**：若移动模块新增配置（速度容差、路径策略），需在 `server/output/config/example_client.json` 或命令行旗标提供示例。

---

## 7. 验收标准

- 目录结构、依赖与 `server/service` 三大服务保持一致，禁止私造并行框架。  
- 移动链路具备 `Start/Update/End` 全流程、容错和回调能力，可通过日志观察到行为。  
- 面板命令与系统接口解耦，可在不改面板的前提下复用脚本 API。  
- 文档与主进度表保持同步，未完成的子项已登记在“待实现 / 待完善功能”内并写清验收标准。

---

## 8. 后续跟踪

- 负责人在每日开发前需重新检查本方案与主文档，确认是否有新增约束。  
- 若重构过程中需要新增公共工具，请优先考虑 `server/internal` 或 `server/pkg` 是否已有对应能力，避免重复造轮子。  
- 未列出的业务（例如社交、公会）如需接入，请沿用“系统化 + 文档同步”的流程更新本方案。

---

## 9. 关键代码与运行流程

### 9.1 核心代码清单（按职责划分）

1. `server/example/cmd/example/main.go`  
   - CLI 入口；初始化日志、加载 `server/output/config`、创建 `client.Manager` 并启动 `AdventurePanel`。  
2. `server/example/internal/client/core.go`  
   - 调试客户端内核：连接 Gateway、序列化 Proto、维护角色/场景状态、转发 AOI/战斗/背包事件、暴露 `MoveRunner()`。  
3. `server/example/internal/client/flow.go`  
   - 协议等待器注册表；所有命令通过 `flow.Waiter` 等待 `S2C*` 响应，避免阻塞主协程。  
4. `server/example/internal/client/move_runner.go`  
   - 自动寻路与移动重放：直线优先 + A* 寻路、分段发送 `C2SStart/Update/EndMove`、同步校验与自动重试；支持回调链。  
5. `server/example/internal/panel/adventure_panel.go` + `internal/panel/actions.go`  
   - 文字冒险 UI、菜单逻辑与命令解析；数字快捷键（如“4 自动寻路到坐标”）和命令行 `move-to` 共享同一实现。  
6. `server/example/internal/systems/*`  
   - 业务系统集合：`account.go`（账号/角色）、`scene.go`（状态与 AOI）、`move.go`（调用 `MoveRunner`）、`combat.go`、`inventory.go`、`gm.go`、`dungeon.go`、`script.go`。  
7. `server/example/internal/systems/set.go`  
   - 系统注册入口：在 `AdventurePanel.connect` 时一次性创建所有系统，并提供 `p.systems.<Domain>` 访问。  

> 新增命令/系统时务必把关键代码入口补充到本清单，保持与主进度文档第 8 章一致。

### 9.2 “自动寻路到坐标”（菜单选项 4 / 命令 `move-to`）执行流程

1. **面板输入**  
   - 玩家在进入场景后，选择菜单数字 `4`（`actionMoveToPrompt`）或输入命令 `move-to <x> <y>`（`actions.go`）。  
2. **校验与参数获取**  
   - `AdventurePanel` 调用 `requireScene()` 确认账号已登录且角色在场景内，然后从输入读取目标格子坐标。  
3. **调用移动系统**  
   - 面板把坐标传给 `systems.Move.MoveTo(ctx, tileX, tileY, callbacks)`，该系统只是轻量封装，直接转发到 `client.MoveRunner`.  
4. **MoveRunner 准备**  
   - `MoveRunner.MoveTo` 取消正在进行的移动、记录最新目标，并调 `run()`：  
     - 通过 `core.CurrentSceneMap()` 取 `jsonconf.GameMap`，校验目标可行走。  
     - 读取当前格子，若可以直线到达则生成两点路径，否则使用 A*（`makePath → findPath`）求最短路径。  
5. **分段发送移动协议**  
   - `executePath` 遍历路径节点，对每一段调用 `core.sendMoveChunk`：  
     - 发送 `C2SStartMove`（像素坐标 + 速度）→ 周期性发送 `C2SUpdateMove`（逐格像素点）→ 结束时发 `C2SEndMove`。  
     - 本地位置信息通过 `core.updateLocalPosition` 及时刷新，便于后续路径计算。  
6. **等待服务端确认**  
   - `waitForServerPos` 在超时时间内监听 `S2CEntityMove/S2CEntityStopMove`（`core.OnEntityMove` 更新位置），直到角色位置与目标格子进入容差区间；若超时则认定“与服务器不同步”。  
7. **容错与重试**  
   - 遇到 `ErrMoveOutOfSync`、`ErrMoveBlocked` 等错误时，`MoveRunner` 最多重建路径并重试 2 次；若仍失败则触发 `MoveCallbacks.OnAbort` 并在面板日志输出原因。  
8. **完成与回调**  
   - 当到达目标格子，`MoveRunner` 触发 `MoveCallbacks.OnArrived`（若有）并向面板写入“🚶 自动寻路至 (x,y)”日志；脚本系统可在回调中串接下一条行为。  

该流程与 `server/service/dungeonserver` 的移动协议保持一致，可稳定复现客户端移动链路；若用户中断（`Ctrl+C`/输入其他命令）或 `move-resume`，会通过 `cancelFn`/`lastTarget` 控制器继续或停止本次寻路。

### 9.3 server/example 代码概览（系统划分）

- **入口与生命周期**：`cmd/example/main.go` 初始化日志/配置 → 创建 `client.Manager` → 启动 `AdventurePanel`。  
- **客户端核心 (`internal/client`)**：`core.go` 管理连接、协议发送、角色状态、AOI 缓存与 Flow Waiter；`handler.go`/`manager.go` 负责多客户端托管；`move_runner.go` 复刻 Start/Update/End 移动链路。  
- **面板 (`internal/panel`)**：`adventure_panel.go` 绘制 UI、菜单与状态栏，`actions.go` 解析命令（register/login/move/move-to/...），所有命令最终调用下层 `systems`。  
- **系统集合 (`internal/systems`)**：  
  - `account.go`：账号注册/登录、角色列表、进入游戏（依赖 Flow Waiter）。  
  - `scene.go`：查询角色状态、AOI 观察。  
  - `move.go`：轻量封装 `MoveRunner` 的 `MoveBy/MoveTo/Resume`。  
  - `combat.go`：普通攻击与等待 `S2CSkillDamageResult`。  
  - `inventory.go`：`bag/use-item/pickup` 协议。  
  - `dungeon.go`：`enter-dungeon` 流程。  
  - `gm.go`：`gm <name> [args...]` 命令与结果展示。  
  - `script.go`：录制/回放、Demo 巡逻脚本；`set.go` 统一注册系统实例。  
- **脚本与回调**：`systems/script.go`、`client/move_runner.go` 的 `MoveCallbacks` 支持链式动作，可在自动移动完成后继续执行拾取/攻击/副本等步骤。

### 9.4 新增命令开发流程

> 适用于在 AdventurePanel 中添加一条新命令或菜单项；若命令涉及多阶段交付，请先把拆分写入主文档第 6 章“待实现 / 待完善功能”，完成后移至第 4 章并标记 ✅。

1. **梳理需求并登记文档**  
   - 在 `docs/服务端开发进度文档.md` 的“待实现 / 待完善功能”记录命令目标、依赖系统、验收标准；若是调试客户端专属能力，同步 `docs/golang客户端待开发文档.md`。  
2. **确认所属系统**  
   - 检查 `internal/systems` 是否已有对应业务模块；如果没有，按 9.5 的流程先新增系统。  
3. **实现系统方法**  
   - 在对应 `systems/<domain>.go` 中新增公共方法，内部调用 `client.Core` 的协议封装（`SendClientProto`、`flow.Waiter`、`MoveRunner` 等），并返回明确的错误/结果。  
4. **注册面板命令**  
   - 在 `internal/panel/actions.go` 中增加命令解析与帮助文案；若希望通过菜单快捷键调用，更新 `AdventurePanel.currentMenuOptions()` 并添加 `actionXXX`。  
5. **脚本/录制兼容**  
   - 若命令需要脚本化或自动回放，更新 `systems/script.go` 的命令映射或 Demo，确保录制文件可复现。  
6. **联调与文档同步**  
   - 启动三服 + example，验证命令；完成后在主文档“已完成功能/关键代码位置/注意事项”中补充条目并以 ✅ 标记，说明命令入口与依赖系统。

### 9.5 新增系统开发流程

1. **需求拆分与文档记录**  
   - 在 `docs/服务端开发进度文档.md` 第 6 章登记系统目标、阶段划分、前置依赖与验收方式；若系统属于调试客户端，也在本方案与 `docs/golang客户端待开发文档.md` 中同步。  
2. **创建系统骨架**  
   - 在 `internal/systems` 下新建 `<name>.go`，按照现有模板定义 `type <Name>System struct { core *client.Core ... }` 及 `New<Name>System`；只允许通过 Core 访问网络/时间/Actor。  
3. **接入 Systems Set**  
   - 修改 `internal/systems/set.go`：将新系统实例化并挂载到 `Set` 结构，供面板统一访问。  
4. **实现协议与状态**  
   - 在系统内部使用 `core.SendClientProto`、`core.Flow()` 等落地协议调用与响应处理；需要移动/脚本支持的复用 `MoveRunner` 与 `MoveCallbacks`。  
5. **暴露命令与脚本 API**  
   - 在 `panel/actions.go` 注册对应命令/菜单；如需脚本化，扩展 `systems/script.go` 的调度与录制。  
6. **测试与文档同步**  
   - 自测全链路后，将系统入口路径加入本文件 9.1 的关键代码清单，并在主文档第 4/6/8 章同步状态（完成项前写“✅”）。

### 9.6 FlowRegistry（`server/example/internal/client/flow.go`）使用说明

调试客户端大量命令需要等待特定 `S2C` 回包（注册、登录、进入游戏、背包更新、GM 结果等）。为避免在业务代码中手动管理 channel，本项目在 `flow.go` 定义了 `flowRegistry`，提供“发送请求 → 等待响应”的统一模式：

1. **注册 Waiter**  
   - `flowRegistry` 包含多个 `waiter[T]` 字段（泛型结构，内部是带缓冲的 channel）。  
   - 在 `client.NewCore` 中通过 `newFlowRegistry()` 初始化，确保每个 `Core` 独立持有一套 Waiter。
2. **发送请求前获取 Waiter**  
   - 业务系统（如 `systems.Account.Register`）在发送 `C2SRegister` 之前，不需要额外创建 channel，只需调用 `c.flow.register.Wait(timeout)` 等方法等待结果。  
   - 若命令需要流式读取（例如 AOI 实体更新），可以直接使用 `c.flow.aoi.Chan()` 非阻塞读取。
3. **协议回调中 Deliver**  
   - `client/core.go` 的 `OnRegisterResult`、`OnLoginResult`、`OnBagData` 等方法在收到 `S2C` 后调用对应 `waiter.Deliver(resp)`。`Deliver` 使用非阻塞写入：如果 channel 为空则写入成功；如果 channel 已有消息且没有人在等待，新消息会被丢弃（避免阻塞网络线程，但不会覆盖旧消息）。  
4. **等待响应**  
   - `waiter.Wait(timeout)` 使用 `servertime.Now()` + `time.Timer` 实现超时机制，调用方只需处理 `(resp, err)`。  
   - 典型用法：

```70:110:server/example/internal/client/core.go
func (c *Core) RegisterAccount(username, password string) error {
    ...
    if err := c.sendProtoMessage(...); err != nil {
        return err
    }
    resp, err := c.flow.register.Wait(defaultClientTimeout)
    if err != nil {
        return err
    }
    if !resp.Success {
        return fmt.Errorf("register failed: %s", resp.Message)
    }
    return nil
}
```

5. **并行场景**  
   - 每个 Waiter 是独立的 channel，允许多个协议并行等待（例如同时等待 `bagData` 与 `gmResult`）；同一 Waiter 的 channel 容量为 1，如果已有消息且没有人在等待，新消息会被丢弃（适合“请求 → 单次响应”场景，需及时调用 `Wait` 获取结果）。  
   - 对于需要连续事件的场景（如 AOI），使用 `waiter.Chan()` 并结合 select 或非阻塞读获取最新实体快照。

通过 FlowRegistry，将所有协议交互的并发、超时和回调逻辑集中管理，使 `systems/*` 和 `panel/actions.go` 的流程保持简洁，同时复用统一的时间源与错误处理。


