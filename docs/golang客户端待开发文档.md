# Golang 文字冒险客户端（server/example）待开发文档

> 请在**每次修改 `server/example`** 前阅读本文件与 `docs/服务端开发进度文档.md` 第 0 章与第 7 章，所有调试客户端的设计/实现约束以此文档为准。

---

## 0. 文档目的与阅读须知

- 对齐 **调试客户端** 的目标、边界与架构约束，避免快迭代时出现协议漂移或临时脚本泛滥。
- 记录“文字冒险式面板”的命令映射与进展，让服务端逻辑能被快速验证。
- 功能完成后需同步 `docs/服务端开发进度文档.md` 的“已完成功能/待实现/注意事项/关键代码位置”四章。
- 如果需要分阶段落地，请在本文件和主进度文档的“待实现”章节拆分清楚。

---

## 1. 背景与目标

| 目标 | 说明 |
| ---- | ---- |
| 文本冒险式调试体验 | 通过“文字 MUD / 单机 RPG”风格的指令面板，复刻客户端关键流程。 |
| 协议一一映射 | 面板命令必须直接映射到现有 `server/service` 协议，方便录制/回放与自动化。 |
| 全链路联调 | 能够覆盖账号、角色、养成、战斗等关键路径，快速发现服务端回归问题。 |
| 可扩展 | 以 `GameClient` Actor + 命令路由为核心，后续可无缝扩展到宏脚本、剧情脚本或机器人。 |

---

## 2. 架构概览

### 2.1 核心组件

| 组件 | 作用 | 关键文件 |
| ---- | ---- | -------- |
| `AdventurePanel` | 文本面板、命令解析、状态机（登录/进场景/战斗）。 | `server/example/internal/panel` |
| `ClientManager` | 复用 ActorManager，为每个客户端建 Actor。 | `server/example/internal/client/manager.go` |
| `GameClient Core` | 网络编解码、协议流、AOI/技能回包监听。 | `server/example/internal/client/core.go` |
| `Systems` | Account / Scene / Move / Combat 系统接口化，供面板/脚本调用。 | `server/example/internal/systems` |
| `ClientHandler` | Actor 消息处理 → `Core` 回调。 | `server/example/internal/client/handler.go` |

### 2.2 消息流

```
AdventurePanel 命令
    ↓
GameClient.SendProto
    ↓ TCP
Gateway → GameServer
    ↓
ClientHandler (Actor)
    ↓
GameClient.OnXXX → 面板输出
```

---

## 3. 命令与协议映射

| 面板命令 | 协议 | 说明 |
| -------- | ---- | ---- |
| `register <acc> <pwd>` | `C2SRegister` | 账号注册，失败原因即时输出。 |
| `login <acc> <pwd>` | `C2SLogin` | 登录成功后保留账号态。 |
| `roles` | `C2SQueryRoles` | 输出角色列表（ID/名字/等级/职业）。 |
| `create-role <name> [job] [sex]` | `C2SCreateRole` | 支持快速创建任意职业/性别。 |
| `enter <roleID>` | `C2SEnterGame` | 阻塞至 `S2CEnterScene`，保证 AOI/属性已同步。 |
| `status` | `S2CEnterScene` + AOI | 展示角色位置/HP/MP/状态。 |
| `move <dx> <dy>` | `C2SStartMove/C2SEndMove` | MoveRunner 逐格上报，带速度容错。 |
| `move-to <x> <y>` | `C2SStartMove/C2SUpdateMove/C2SEndMove` | 自动寻路到指定格子。 |
| `move-resume` | Same as MoveRunner | 继续上一次 `move-to` 目标。 |
| `look` | `S2CEntityMove/Stop + SkillResult` | 基于 `GameClient` 观察缓存输出周围实体。 |
| `attack <handle>` | `C2SUseSkill` | 当前实现普通攻击（skillId=0），等待命中包。 |
| `bag` | `C2SOpenBag`/`S2CBagData` | 拉取背包并本地展示。 |
| `use-item <id> [count]` | `C2SUseItem`/`S2CUseItemResult` | 使用道具，展示成功/失败。 |
| `pickup <handle>` | `C2SPickupItem`/`S2CPickupItemResult` | 拾取掉落物。 |
| `gm <name> [args...]` | `C2SGMCommand`/`S2CGMCommandResult` | 发送 GM 指令，支持参数。 |
| `enter-dungeon <id> [difficulty]` | `C2SEnterDungeon`/`S2CEnterDungeonResult` | 快速验证副本链路。 |
| `script-record <file>` | 面板自实现 | 将后续命令录制到文件。 |
| `script-run <file> [delayMs]` | 面板自实现 | 顺序执行脚本内的命令。 |
| `script-demo` | `ScriptSystem.RunDemo` | 内置巡逻 + 普攻示例。 |

> 设计原则：命令行每条指令都可映射到服务端协议，必要时支持串联（脚本化）而无需改服务端。

---

## 4. 当前进展（2025-11-26）

- ✅ `AdventurePanel` 替换原集成脚本，默认自动连接 Gateway，提供 MUD 风格提示。
- ✅ `GameClient` 抽象账号/角色/场景状态，支持 AOI 快照、技能结果、状态查询。
- ✅ `ClientManager` 支持运行期销毁面板实例，避免连接残留。
- ✅ `server/example` 全量使用 `servertime` 约束，与服务端时间规范一致。
- ✅ 文档化命令映射及使用方式（本文件 + 主进度文档）。
- ✅ 面板采用“标题区 + 日志区 + 命令区”三段式 UI，并提供数字化快捷菜单。
- 🆕 `MoveRunner` 支持直线优先 + A* 寻路、容错重试与 `move-to/move-resume` 命令。
- 🆕 `Inventory/Dungeon/GM` 系统封装背包、GM、副本协议，面板新增 `bag/use-item/pickup/gm/enter-dungeon`。
- 🆕 `script-record/script-run` 支持命令录制与回放，`ScriptSystem` 保留 Demo 巡逻脚本。

---

## 5. 待实现 / 待完善功能

| 编号 | 内容 | 说明 |
| ---- | ---- | ---- |
| GCLI-1 | ✅ 背包/物品指令 | `bag`, `use-item`, `pickup` 与 `ItemSys` 协议映射。 |
| GCLI-2 | 养成面板 | 等级/属性/任务/活跃度等查询与操作面板。 |
| GCLI-3 | 战斗脚本化 | 预设战斗脚本（巡逻/技能循环）、统计命中/伤害。 |
| GCLI-4 | ✅ 副本/匹配流程 | `enter-dungeon <id> [difficulty]` 命令直连副本。 |
| GCLI-5 | ✅ GM/调试命令桥接 | 面板命令映射至 GM 协议，支持批量调试。 |
| GCLI-6 | ✅ 协议回放/录制 | `script-record/script-run` 支持录制与回放。 |
| GCLI-7 | 多 Session 场景 | 面板内创建多客户端（机器人），支持互操作。 |

> 开发上述子项时，请同步更新本表与主进度文档第 6 章，标注依赖与完成标准。

---

## 6. 测试与联调建议

1. 本地同时启动 `dungeonserver → gameserver → gateway`，确认 `server/output/config` 与数据库就绪。
2. 通过 `connect` / `register` / `login` / `create-role` / `enter` 完成基础链路自检。
3. 使用 `move`、`look`、`attack` 验证 AOI 与战斗事件能在日志中正确回显。
4. 若需脚本化，请复用 `AdventurePanel` 的命令解析，禁止直接在 `GameClient` 中写固定流程。
5. 每次扩展命令后补充 README/文档示例，并在主进度文档同步“关键代码位置”。

---

## 7. 关键参考

- 主文档：`docs/服务端开发进度文档.md`
- 协议：`proto/csproto/*.proto`
- 时间源：`server/internal/servertime`
- 公共工具：`server/pkg/log`, `server/pkg/tool`

