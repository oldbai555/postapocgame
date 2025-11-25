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


