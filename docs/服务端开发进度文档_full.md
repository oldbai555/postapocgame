# 游戏服务器开发进度文档（单一权威版本）

更新时间：2025-12-23  
责任人：个人独立开发  

> 本版以重置后的 GameServer 骨架为唯一基线。旧目录/兼容层已移除，所有新增功能需直接按当前分层与接口重写。

---

## 0. 开发必读

- 先读后写：开始任务前务必阅读「当前架构」「开发注意事项与架构决策」，确保遵守 Actor 单线程、Clean Architecture、统一时间等约束。
- 文档同步：功能完成→移入「已完成功能」并在「关键代码位置」补入口；分期功能→写入「待实现 / 待完善」拆子项。
- 架构决策：跨模块约束或重要选择写入第 6 章；不在此文档粘贴函数级细节。
- 时间访问：除 proto/三方库外统一使用 `server/internal/servertime`，数据库时间字段写 Unix 秒；禁止直接 `time.Now()`。
- 无兼容层：不再保留旧 entitysystem 玩法代码或过渡 wrapper，新功能直接按现有目录/接口实现。

---

## 1. 项目与架构概览

- 项目：postapocgame（后启示录横版动作），后端 Go 1.24.x，单仓包含 `gateway`、`gameserver`。
- 数据：SQLite + GORM，玩家数据存 `PlayerRoleBinaryData`；当前无 PublicActor，全局社交/经济待重建。
- 配置：`server/output/config/*.json` 必须齐备；服务配置 `server/output/{gateway,gamesrv}.json`。
- 拓扑：
  ```
  Client (TCP/WS)
        |
  Gateway (SessionManager)
        | ForwardMessage / SessionEvent
  GameServer (PlayerActor per player + DungeonActor single)
  ```
- 通信：PlayerActor ↔ DungeonActor 通过 `gshare.IDungeonActorFacade` 发送内部消息（`DungeonActorMsgId` / `PlayerActorMsgId`），禁止阻塞调用。
- 分层：Controller 解析/检查 → UseCase/Service 业务 → Presenter 回包；SystemAdapter 仅做生命周期与事件调度。

---

## 2. 构建与运行

- Go：`go 1.24.0`（toolchain 1.24.10）。
- 构建：`go build -o server/output/gameserver.exe ./server/service/gameserver`；Gateway 同理。
- 运行依赖：`server/output/{gateway,gamesrv}.json`、`server/output/config/*.json`、数据库文件自动迁移。
- 日志：`server/output/log/<service>.log` + 控制台。

---

## 3. 当前基线与已完成功能

### 3.1 Gateway

- TCP + WebSocket 接入；Session 生命周期管理；ForwardMessage 转发；基础限流/资源保护。  
  关键目录：`server/service/gateway/internel/{clientnet,engine}`

### 3.2 GameServer（2025-12-23 重置后）

- 账号/角色：注册、登录、Token、角色创建/进入游戏；Session 挂账号/角色信息。
- Actor 框架：每玩家单 Actor；SystemRegistry 仅挂 `Level`、`Skill`；定期落盘、无锁单线程。
- Controller：`player_account_controller`、`player_role_controller`、`move_controller`（转 DungeonActor）、`controller/skill_controller.go`（转 DungeonActor）；注册集中于 `router/protocol_registry.go`。
- 依赖装配：`playeractor/deps` 同时承载 Runtime + 工厂（gateway/repo），Context 取值在 `gshare/context_helper.go`。
- 消息派发：`player_network_controller.go` 处理 `ForwardMessage`/`PlayerActorMsg`，经 `gshare.SendDungeonMessageAsync` 调用 DungeonActor。
- 系统：`level`、`skill` 基于 `sysbase`，运行于 Actor 单线程。
- 瘦身：删除未用的 PublicActor 网关/事件发布器占位与多余玩家事件枚举，精简 PlayerActor 运行时依赖；补齐 `DAMEnterGame` 处理，进入游戏时直接分配默认副本/场景并下发 `EnterScene`/AOI Appear；移除未使用的 message registry/dispatcher，技能控制器统一到 `controller` 目录；新增清理与 proto 无关的怪物/AI/寻路/掉落接口，DungeonActor 仅保留玩家 AOI/移动/技能链路；配置层仅保留 `job/skill/scene/map`，删除 item/level/monster/monsterscene 相关结构体与配置文件。
- 技能定义：SkillCastResult/SkillHitResult 等结构统一在 `proto/csproto/skill_def.proto` 生成，skill 包删除重复结构体并移除 FightSys 未用字段。

### 3.3 DungeonActor（单 Actor 战斗/副本引擎）

- 主循环：ModeSingle Actor；场景/实体管理、RunOne 驱动。
- 系统：AOI、移动、战斗、技能、属性聚合；已移除怪物、AI、寻路、掉落逻辑，副本骨架仍保留。
- 坐标/移动：客户端像素坐标→服务端格子坐标校验；Start/Update/EndMove 流程含容错。
- 消息：通过 `DungeonActorMsgId` / `PlayerActorMsgId` 与 PlayerActor 交互；不直接处理任何 C2S 协议。
- 副本：保留常驻默认副本，限时副本 provider 已移除（如需再开请重建）。

### 3.4 共享基础

- Actor 框架、事件总线、servertime、日志、Proto、gatewaylink 透传、上下文/日志辅助。  
  关键目录：`server/internal/{actor,servertime,jsonconf,argsdef}`、`server/pkg/log`
- 调试客户端：对齐当前 `cs/sc.proto`，仅保留注册/登录/角色/移动/技能命令，移除背包、GM、副本、脚本录制等旧逻辑。  
  关键目录：`server/example/internal/{client,panel,systems}`

---

## 4. 待实现 / 待完善

- [ ] 重建玩法/经济系统：Bag/Money/Equip/Fuben/Recycle/Quest/Shop/GM/AntiCheat 等，直接用当前分层与接口，无旧兼容。
- [ ] 恢复 PublicActor 与社交：好友/公会/排行/拍卖/离线快照，链路为 Gateway → PlayerActor → PublicActor。
- [ ] Controller 层系统开启检查与 UseCase 单测补齐（现仅 Level/Skill）。
- [ ] 玩家消息系统 Phase4：监控与过期/清理策略，避免消息表膨胀。
- [ ] 接入安全：Gateway WS/IP/Origin/签名校验；GM 权限模型与审计日志。
- [ ] 多人副本匹配、战斗录像等扩展（待核心玩法重建后排期）。
- [ ] 调试客户端扩展：多 Session/脚本化战斗回放，对齐新协议。

---

## 5. 开发注意事项与架构决策（精要）

- Actor 约束：玩家状态只能在 PlayerActor 线程修改；DungeonActor 单线程；禁止业务 goroutine 直接改玩家数据。
- 时间统一：业务取时间一律用 `servertime`；数据库时间字段写 Unix 秒。
- 控制器职责：解析 + 权限/系统开关检查 + 调用 UseCase/Service；回包经 Presenter，禁止在 SystemAdapter/EntitySystem 注册协议。
- S2C 发送：任何下行都经 PlayerActor → `gatewaylink`，其他 Actor 禁止直发。
- 接口归档：端口接口集中于 `server/service/gameserver/internel/iface`，新增接口先放这里。
- 坐标规范：服务端全部使用格子坐标做校验/寻路/范围；客户端上送像素坐标，由服务端转换。
- 无兼容层：旧 entitysystem 玩法代码已移除，不再保留 wrapper，新功能直接按现有目录实现。
- 事件注册：使用 gevent 事件总线，控制器与 DungeonActor 在 OnSrvStart 时注册，PlayerRole 登录通过事件驱动系统管理器。
- 技能结果：逻辑层使用 proto 生成的 SkillCastResult/SkillHitResult，不重复定义内部结构。
- 停服流程：收到退出信号发布 `OnSrvStop`，先触发所有在线玩家的 OnDisconnect/Close 并移除 Actor，再走批量落盘与服务停止。
- DungeonActor 仅支持 `ModeSingle`，配置为多 Actor 直接拒绝启动。

---

## 6. 关键代码位置

- Gateway：`server/service/gateway/internel/{clientnet,engine}`。
- GameServer 入口：`server/service/gameserver/main.go`、`internel/engine/{config.go,server.go}`、`internel/gatewaylink/{handler.go,sender.go,export.go}`。
- PlayerActor：`internel/playeractor/{adapter.go,handler.go}`、`controller/{player_account_controller.go,player_role_controller.go,move_controller.go,skill_controller.go}`、`router/protocol_registry.go`、`register/register.go`、`runtime/runtime.go`、`deps/deps.go`、系统注册 `entitysystem/{sys_mgr.go,system_registry.go}`、`level/*`、`skill/*`、`sysbase/base_system.go`。
- DungeonActor：`internel/dungeonactor/{adapter.go,handler.go,register.go}`、`entity/*`、`entitysystem/*`、`scene/*`、`scenemgr/*`、`fbmgr/*`、`fuben/*`、`skill/*`、`iface/*`。
- 基础库：`server/internal/{actor,servertime,jsonconf,argsdef}`、日志 `server/pkg/log`。
- 调试客户端：`server/example/cmd/example`、`server/example/internal/{client,panel,systems}`。

---

## 7. 版本记录

- 2025-12-23：GameServer 重置为最小可运行骨架，仅保留账号/角色登录、Move/UseSkill 转发、Level/Skill 系统与 DungeonActor 基础；移除 Bag/Money/Equip/Fuben/Recycle/Quest/Shop/GM/PublicActor 等旧实现与兼容层。
- 2025-12-23（瘦身补充）：删除未用 PublicActor 网关/事件发布器占位和冗余玩家事件枚举，PlayerActor 运行时依赖进一步收敛。
- 2025-12-23（瘦身补充2）：DungeonActor 补齐 `DAMEnterGame` 处理并去除限时副本 provider；调试客户端裁剪为注册/登录/角色/移动/技能命令。
- 2025-12-23（瘦身补充3）：删除未使用的 message registry / player message dispatcher，技能协议控制器统一放入 `controller/skill_controller.go`。
- 2025-12-23（瘦身补充4）：移除与当前 proto 无关的怪物/AI/寻路/掉落接口，DungeonActor 仅保留玩家 AOI/移动/技能链路。
- 2025-12-23（瘦身补充5）：配置层仅加载 `job/skill/scene/map`，删除 item/level/monster/monsterscene 结构体及对应 json。
- 2025-12-23（瘦身补充6）：技能 Cast/Hit 结果结构移入 `skill_def.proto`，删除本地 Cast/HitResult 结构与 FightSys 未用字段，保持逻辑/协议一致。
 