# Godot客户端开发进度文档（单一权威版本）

更新时间：2025-11-20  
责任人：个人独立开发  

> ⚠️ 开发前务必阅读第 0 章与第 7 章；完成新功能后同步“已完成功能 / 待实现 / 注意事项 / 关键代码位置”。

---

## 0. 客户端开发必读

- **先读后写**：遵循本文件与 `docs/服务端开发进度文档.md` 的章节结构，保持术语、协议与流程一致。  
- **统一场景切换**：所有场景切换必须通过 `SceneManager.SwitchSceneAsync`，确保 Loading 流程一致。  
- **协议流程**：新增/修改协议需先更新 `proto/csproto`，运行 `proto/genproto_csharp.bat`，再于 `ProtocolHandler.InitializeProtocolMaps()` 注册。  
- **状态机约束**：`Player_state_machine` 下的状态节点禁止直接访问 UI/网络；统一通过 `Player` 提供的接口以减少耦合。  
- **时间访问**：客户端逻辑若需要服务器时间，统一从 `NetworkManager.ServerTime` 读取；禁止直接依赖本地 `Time.GetTicksMsec()` 做权威判定。  
- **文档同步**：上线能力立即补充“已完成功能”与“关键代码位置”；拆分大需求时在第 5/6 章登记子项。  

---

## 1. 项目概述

- **项目名称**：postapocgame（后启示录横版动作）  
- **客户端引擎**：Godot 4.5.1（Mono）  
- **语言/工具链**：C# + .NET 8.0、Google.Protobuf 28.x  
- **目标平台**：Windows（首发）/ Android / iOS  
- **服务端架构**：Gateway + GameServer + DungeonServer（详见服务端文档）  
- **仓库位置**：`client/` 目录，配套 `proto/` 协议定义  
- **运行模式**：单人开发，AutoLoad 中常驻 `NetworkManager`、`MessageReceiver`、`SceneManager`、`LoadingScreen` 等核心节点  

---

## 2. 客户端架构

| 模块 | 职责 | 关键目录 |
| ---- | ---- | -------- |
| Network | TCP 连接、心跳、断线重连、协议编解码、消息队列 | `client/Scripts/Network` |
| Protocol | Proto 生成结果、协议 ID 枚举 | `client/Scripts/Protocol` |
| Scene | 场景切换、实体管理、AOI 同步、渲染入口 | `client/Scripts/Scene`、`client/Scenes` |
| GameLogic | 玩家控制、移动、战斗、技能、AI 等业务逻辑 | `client/Scripts/GameLogic`, `client/Player/Scripts` |
| UI | 登录、角色、HUD、系统界面 | `client/Scripts/UI`, `client/Scenes/UI` |
| Data | 本地存储、配置缓存、玩家数据缓存 | `client/Scripts/Data` |
| Utils | 日志、资源管理、通用工具 | `client/Scripts/Utils` |

---

## 3. 构建与运行现状

- **Godot 编辑器**：使用 Godot 4.5.1 Mono 版本打开 `client/project.godot`。  
- **命令行构建**：`godot4 --headless --build-solutions`（生成 .csproj）；`dotnet build client/client.csproj`.  
- **测试入口**：默认场景 `AutoLoad.tscn`；登录 → 角色 → （未来）主场景。  
- **协议生成**：`cd proto && .\genproto_csharp.bat`，输出至 `client/Scripts/Protocol/`。  
- **运行依赖**：需配套服务端已启动（Gateway -> GameServer -> DungeonServer），且 `server/output/config/*.json` 齐全。  
- **日志**：`client/Scripts/Utils/Logger.cs` 输出到 Godot 控制台，可按需扩展至文件。  

---

## 4. 已完成功能

### 4.1 引擎与协议基础

- **项目初始化**：完成 Godot 项目创建、目录划分、`client.csproj` 配置与 NuGet 依赖。  
- **Proto 生成链路**：`proto/genproto_csharp.{bat,sh}`、`client/Scripts/Protocol/README.md`、协议注册示例齐备。  
- **网络层**：`NetworkManager`（TCP/心跳/重连/时间差）、`ProtocolHandler`（C2S/S2C 映射）、`MessageSender`、`MessageReceiver`、`Logger` 均可用。  

### 4.2 登录与角色流程

- **LoginUI**：注册/登录、记住账号、错误提示、自动连接、成功后切换场景。  
- **RoleSelectUI**：角色列表、创建角色、进入游戏、详情展示；删除角色弹窗仍待实现。  
- **本地存储**：`LocalStorage` 支持账号信息加密保存。  

### 4.3 场景切换基线

- **Loading 层**：`LoadingScreen.tscn` + `LoadingScreen.cs` + `SceneManager`，支持异步加载、百分比展示、文案透传。  
- **AutoLoad 注册**：`AutoLoad.tscn` 常驻单例，登录 → 角色切换已接入 Loading 流程。  

### 4.4 玩家状态机基线

- `Player`、`Player_state_machine`、`StateIdle`、`StateWalk`、`StateAttack`、`State` 基类已实现；  
- 攻击状态附带特效、音效、减速、动画信号回调；`Player.AnimDirection()`、`Player.IsFacingLeft` 提供统一朝向能力。  

---

## 5. 阶段性方案（参考服务端 Phase 划分）

### 5.1 Phase 3：场景/实体系统（进行中）

1. **SceneManager 扩展**：资源卸载策略、切换动画、场景配置读取。  
2. **主场景 Main.tscn**：承载世界节点 + HUD。  
3. **EntityManager 体系**：基础实体、玩家/怪物/NPC/掉落实现、AOI 同步、位置校准。  
4. **场景渲染管线**：地图、精灵、血条、伤害数字、技能特效。  
5. **调试工具**：可视化碰撞、实体列表、AOI 变更日志。

### 5.2 Phase 4：游戏逻辑（待启动）

- 玩家控制器 / MovementSystem（含服务器时间驱动）；  
- CombatSystem（前摇-判定-后摇 / Buff 显示 / S2C 结果同步）；  
- SkillSystem（技能表、CD、图标）；  
- AutoCombatAI（可选）确保所有动作经服务端校验。  

### 5.3 Phase 5：UI & 数据

- HUD / 主界面、Bag / Equip / Quest / Skill / Mail / Shop / Dungeon / Settings UI；  
- ConfigManager（JSON 载入、缓存、可选热更）；  
- PlayerDataCache（角色、背包、任务等本地快照）；  
- ResourceManager / UtilityFunctions。  

### 5.4 Phase 6：时间与同步

- NetworkManager Ping/Pong 时间差计算、Debug HUD；  
- Movement/Combat 依赖服务器时间戳驱动动画；  
- 动画同步协议（技能/受击/移动/状态）。  

### 5.5 Phase 8+（长期 backlog）

- 战斗统计、快捷栏、背包整理、任务导航、战斗录像等增值能力。  

---

## 6. 待实现 / 待完善功能

1. 场景资源回收与过渡特效。  
2. Main 场景 + 主界面 HUD。  
3. EntityManager / Entity / PlayerEntity / MonsterEntity / NPCEntity / DropItemEntity。  
4. AOI 进入/离开、`S2CEntityAppear/Disappear/Move/StopMove` 处理。  
5. MovementSystem：输入、插值、服务器同步、动画驱动。  
6. CombatSystem：`C2SUseSkill`、`S2CSkillCastResult`、受击表现、伤害数字。  
7. SkillSystem：技能列表、CD/图标/详情、`C2SLearnSkill`/`C2SUpgradeSkill` 接口。  
8. PlayerController：统一按键映射、拾取/对话交互。  
9. AutoCombatAI（可选）流程。  
10. 时间同步模块 + 动画帧线。  
11. 全量 UI（Bag / Equip / Quest / Skill / Mail / Shop / Dungeon / Settings）。  
12. 数据层（ConfigManager、PlayerDataCache）与工具层（ResourceManager、UtilityFunctions）。  
13. Debug & QA：可视化碰撞、日志 HUD、场景内测试菜单。  

---

## 7. 开发注意事项与架构决策

### 7.1 架构原则

- **模块化 & 单例**：Network/Scene/Loading 等通过 AutoLoad 单例统一管理。  
- **事件/信号驱动**：跨模块通信优先 Godot 信号或轻量事件，避免直接引用。  
- **资源管理**：大图/特效进入主场景前预加载，离开场景主动释放，防内存膨胀。  

### 7.2 网络通信

- 所有网络 I/O 走后台线程接收 + 主线程消费队列；  
- 心跳类型 `0x06`，默认 5s，断线自动重连 5 次（3s 间隔）；  
- SessionId 在登录前可为空，Gateway 自动补齐；  
- 时间同步后统一调用 `NetworkManager.ServerTime`。  

### 7.3 玩家状态机约束

1. `Player` 必须包含 `Player_state_machine`，或通过 `StateMachinePath` 显式指定。  
2. 新状态继承 `State`，在 `Enter/Exit/Process/Physics/HandleInput` 内通过返回 State 切换。  
3. 动画前缀由状态决定，调用 `Player.UpdateAnimation("idle"|"walk"|"attack"|...)`，`Player` 自动拼接方向。  
4. `StateAttack` 仅依赖 `AnimationFinished` 信号判定结束，Godot 动画必须 `Loop None`。  
5. `_attackEffectSprite` 镜像逻辑统一交由 `Sprite2D.Scale.X` 控制，不再额外设置 `FlipH`。  
6. 任何状态中需要访问 `AnimationPlayer` 等节点时，通过 `GetParent().GetParent()` 自 `Player` 获取，避免硬编码路径。  
7. 修改状态机脚本前必须自测“八方向移动 + 左右攻击特效”，防止退化。  

### 7.4 UI / 场景

- UI 场景在 `_Ready()` 中注册协议处理器，`_ExitTree()` 必须反注册；  
- Loading UI 文案可自定义，保持 `SceneManager.SwitchSceneAsync(path, hint)` 接口一致；  
- 所有 UI 节点必须支持 16:9 / 18:9 自适应，Godot 控件开启 `Layout -> Full Rect`。  

### 7.5 网络协议实践

- Proto 变更后立即重新生成 C# 代码并 `gofmt` Go 端；  
- `ProtocolHandler` 注册时使用 `C2SProtocol` / `S2CProtocol` 枚举，确保 ID 对齐；  
- `MessageReceiver` 默认在主线程调用 handler，handler 内避免耗时阻塞；  
- 错误统一走 `S2CError`，UI 根据 `ErrorCode` 显示提示。  

### 7.6 代码风格（C#）

- 早返回减少嵌套；  
- 简单分支优先三元表达式或局部变量；  
- 复杂条件拆分布尔变量；  
- 所有日志通过 `Logger.Info/Warn/Error`；  
- 禁止在无 `else` 的 `if` 后直接写逻辑，改为 `if (!cond) return;` 模式。  

### 7.7 客户端脚本维护原则（2025-11-19）

1. 不得移除方向采样、`_sprite.Scale` 镜像等表现逻辑；  
2. `StateAttack.DecelerateSpeed` 非零以保持攻击惯性，调整优先走 Inspector；  
3. 特效节点与主体朝向保持一致，禁止手动 `FlipH`；  
4. 注释说明“为什么修改”方便Golang服务端同学理解；  
5. 改动前后至少冒烟测试“移动 + 攻击 + 切场景”流程。  

---

## 8. 关键代码位置

- `client/Scripts/Network/NetworkManager.cs`：TCP、心跳、时间同步。  
- `client/Scripts/Network/ProtocolHandler.cs`：协议映射注册。  
- `client/Scripts/Network/{MessageSender,MessageReceiver}.cs`：消息发送/接收。  
- `client/Scripts/Protocol/*.cs`：生成的 Proto 枚举与结构。  
- `client/Scripts/UI/{LoginUI,RoleSelectUI}.cs`：登录、角色流程。  
- `client/Scripts/UI/LoadingScreen.cs` + `client/Scenes/LoadingScreen.tscn`：统一 Loading 展示。  
- `client/Scripts/Scene/SceneManager.cs`：异步场景切换。  
- `client/Player/Scripts/{Player,Player_state_machine,State*.cs}`：玩家状态机框架。  
- `client/Scripts/Data/LocalStorage.cs`：账号缓存。  
- `client/Scripts/Utils/Logger.cs`：日志工具。  
- （待补）`client/Scripts/Scene/EntityManager.cs`、`client/Scripts/GameLogic/*`：完成后需登记。  

---

## 9. 核心运行流程（客户端侧）

1. **Proto 生成**：`proto/genproto_csharp.bat` → `client/Scripts/Protocol`；在 `ProtocolHandler.InitializeProtocolMaps()` 注册。  
2. **启动**：Godot 载入 `AutoLoad.tscn`，初始化 Network/Scene/Loading 单例。  
3. **登录链路**：`LoginUI` 注册协议 → 用户输入 → `MessageSender.Send(C2SLogin)` → `MessageReceiver` 分发 `S2CLoginResult` → `SceneManager.SwitchSceneAsync("RoleSelect.tscn")`。  
4. **角色进入**：`RoleSelectUI` 处理列表、创建、进入协议，未来成功后切换到 `Main.tscn`。  
5. **状态机驱动**：`Player_state_machine` 在 `_PhysicsProcess/_UnhandledInput` 调度状态；状态返回 `null` 继续当前状态。  
6. **网络消息分发**：后台线程接收 → 主线程队列 → handler 更新 UI / 场景；需要注销时在 `_ExitTree()` 解除。  

---

## 10. 版本记录

| 日期 | 内容 |
| ---- | ---- |
| 2025-11-20 | 对齐服务端文档结构，重写客户端进度文档，新增开发必读/架构/流程章节。 |
| 2025-11-19 | 完成玩家状态机、攻击状态、Loading 场景切换、登录/角色流程。 |
| 2025-11-12 | 建立 Godot 项目、网络层、协议生成脚本与本地存储工具。 |

---

## 11. 开发流程建议

1. **需求拆解** → 在第 5/6 章登记子项与完成标准；  
2. **协议 & 数据** → 更新 `.proto`、生成代码、注册协议；  
3. **功能开发** → UI/逻辑/网络同步推进，遵循状态机与 SceneManager 约束；  
4. **自测** → 登录 → 角色 → Loading → 对应新能力；  
5. **文档回写** → 更新第 4/6/7/8 章，必要时补关键节点路径；  
6. **与服务端联调** → 确认协议与日志一致，必要时更新服务端文档。  

---

## 12. 相关文档

- `docs/服务端开发进度文档.md`：服务端唯一权威信息源。  
- `docs/godot客户端规划.md`：功能规划与长线目标。  
- `docs/服务器规划.md`、`docs/游戏流程优化建议.md`：整体架构与体验指南。  

---  

**提示**：继续开发前仅需查阅此文档即可掌握客户端当前状态；新增能力后务必回写本文件，保持与服务端文档同样的“唯一权威”原则。  

