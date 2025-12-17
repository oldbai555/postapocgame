# 游戏服务器开发进度文档（单一权威版本）

更新时间：2025-12-17  
责任人：个人独立开发  

> ⚠️ 自本次更新起，原 `docs/Phase3社交经济架构设计方案.md` 已完全整合到本文，未来所有开发、评审与交接均以本文件为唯一权威信息源。请在每次开发前完整阅读第 0 章与第 7 章，完成新功能后同步"已完成功能 / 待实现 / 注意事项 / 关键代码位置"四个章节。

---

## 0. 开发必读

- **先读后写**：任何新任务开始前必须先阅读"服务器架构""开发注意事项与架构决策"以及"Phase 3 社交经济一体化方案"，确保遵循 Actor/无锁约束与既定数据规范。  
- **文档同步规则**：新功能上线后，将条目上移至"已完成功能"，补充实现细节，并在"关键代码位置"登记入口；若功能仍需分阶段推进，需在"待实现 / 待完善功能"拆分子项。  
- **架构决策记录**：任何重要结构或跨模块约束，均记录在第 7 章；若涉及流程或数据流转变更，请同步绘制在第 5 章相关子节。  
- **关键代码位点**：发现新的核心入口、易踩坑逻辑、隐含约束时立即更新第 8 章，避免知识散落。  
- **阶段性需求拆分**：遇到大需求需分多迭代交付时，将剩余工作写入第 6 章对应小节，并标注前置依赖/完成标准。
- **时间访问规范**：除 proto/第三方库外，所有服务端业务代码必须通过 `server/internal/servertime` 获取时间；禁止直接调用 `time.Now()`、`time.Since()` 等标准库接口，以保证多服务统一时间与可配置偏移。

---

## 1. 项目概述

- **项目名称**：postapocgame（后启示录横版动作）  
- **客户端**：Godot（计划中），运行于 Windows / Android / iOS  
- **服务器语言**：Golang 1.24.10（`go 1.24` toolchain）  
- **服务集合**：`gateway`、`gameserver`（内置 `DungeonActor` 单 Actor 战斗引擎） 单仓管理  
- **数据层**：SQLite（`server/output/postapocgame.db`）+ GORM，所有玩家数据落于 `PlayerRoleBinaryData`  
- **配置驱动**：`server/output/config/*.json` 26+ 表；需存在于输出目录方可启动  
- **目标部署**：开发使用 Windows；线上兼容 Windows/Linux

---

## 2. 服务器架构

### 2.1 服务划分

| 服务            | 职责                                                                 | 关键目录 |
| --------------- | -------------------------------------------------------------------- | -------- |
| Gateway         | TCP/WS 接入、Session 生命周期、消息压缩/分发、限流、日志             | `server/service/gateway` |
| GameServer      | 玩家长连接逻辑（任务、成长、经济、社交、公会、GM），一玩家一 Actor；内置 `DungeonActor` 作为战斗/副本引擎 | `server/service/gameserver` |
| Shared Packages | Actor 框架、事件、网络编解码、Proto、配置、日志、错误码              | `server/internal`, `server/pkg`, `proto/` |
| DungeonActor    | 在 GameServer 进程内以单 Actor 形式承载战斗/副本/场景逻辑，通过 `DungeonServerGateway` 暴露统一接口；通过 `gshare.IDungeonActorFacade` 与 PlayerActor 协作，使用 `DungeonActorMsgId` / `PlayerActorMsgId` 枚举统一管理内部消息 | `server/service/gameserver/internel/app/dungeonactor` |

### 2.2 拓扑与通信

```
Client (TCP / WebSocket)
      |
Gateway (SessionManager + ClientHandler)
      | ForwardMessage / SessionEvent
GameServer (per-player Actor + PublicActor + DungeonActor)
```

- **会话分发**：Gateway 维护 Session，所有 C2S 消息封装为 `ForwardMessage` 异步转发到 GameServer；Session 事件通过 `gameserverlink` 广播。  
- **Actor 调度**：GameServer 使用 `actor.ModePerKey`（key=SessionId）；`PlayerHandler.Loop()` 通过 `actorCtx.ExecuteAsync` 注入 `gshare.DoRunOneMsg`，确保所有系统 RunOne 在 Actor 主线程执行；DungeonActor 使用 `actor.ModeSingle` 作为进程内战斗/副本引擎骨架。  
- **内部 Actor 通信**：PlayerActor ↔ DungeonActor 通过 `gshare.IDungeonActorFacade` 发送内部 Actor 消息（`DungeonActorMsgId` / `PlayerActorMsgId`），UseCase 层统一通过 `DungeonServerGateway` 接口访问 DungeonActor 能力，避免直接感知 Actor 实现细节。  
- **数据持久化**：玩家系统数据序列化为 `protocol.PlayerRoleBinaryData`，落库字段 `Player.BinaryData`；全局数据（公会/拍卖行等）将由 PublicActor 驱动持久化。  
- **公共 Actor**：GameServer 内包含 `PublicActor`（单 Actor）用于社交经济全局数据、在线映射、排行榜等逻辑。

### 2.3 GameServer 文件作用速查（`server/service/gameserver`）

- **顶层入口**
  - `main.go`：GameServer 进程入口，加载配置与依赖后启动 Engine。
  - `requires.go`：编译期依赖收敛，保证工具链/模块要求。
- **进程级基础（`internel/engine`）**
  - `config.go`：读取/校验 GameServer 配置（端口、数据库、日志等）。
  - `server.go`：构建 Actor 引擎、注册 PlayerActor/PublicActor/DungeonActor，启动循环。
  - `message_registry.go`：注册玩家消息回放用的回调映射。
- **全局管理（`internel/manager`）**
  - `export.go`：对外导出管理能力（DI 暴露）。
  - `role_mgr.go`：玩家角色管理（session→role 映射、关服刷盘）。
- **公共设施（`internel/gatewaylink`）**
  - `export.go`：GatewayLink 对外导出。
  - `handler.go`：处理 Gateway → GameServer 的 Session/ForwardMessage 事件。
  - `sender.go`：封装向客户端发送消息的网关转发。
- **事件与接口（`internel/gevent`、`internel/iface`）**
  - `gevent/enum.go`：游戏内事件枚举。
  - `gevent/event.go`：事件发布/订阅实现。
  - `iface/iserver.go`：GameServer 抽象接口。
  - `iface/irole.go`：PlayerRole 抽象接口。
  - `iface/isystem.go`：SystemAdapter 基础接口。
- **横切工具（`internel/gshare`）**
  - `actor_facade.go`：DungeonActor/PlayerActor 公用的 Actor 消息收发门面接口。
  - `context_helper.go`：从上下文提取 Session/Role/SystemAdapter 入口。
  - `log_helper.go`：注入 IRequester，统一日志前缀。
  - `message_sender.go`：玩家消息发送，在线直投 + 离线落库回放。
  - `protocol.go`：内部协议/消息 ID 定义。
  - `srv.go`：区服信息、开服日等全局配置访问。
- **应用组合（`internel/app`）**
  - `playeractor/adapter.go`：将 PlayerActor 装配为 Actor，可被 Engine 注册。
  - `playeractor/handler.go`：PlayerActor 主循环与消息分发。
  - `playeractor/deps/deps.go`：依赖装配入口（Gateway/Repository/EventBus 等）。
  - `playeractor/domain/chat/chat.go`：聊天领域模型与限频值对象。
  - `playeractor/domain/model/account.go`：账号领域实体。
  - `playeractor/domain/model/role.go`：角色领域实体。
  - `playeractor/domain/repository/account_repository.go`：账号仓储接口定义。
  - `playeractor/domain/repository/player_repository.go`：玩家数据仓储接口。
  - `playeractor/domain/repository/role_repository.go`：角色仓储接口。
  - `playeractor/entity/player_data.go`：玩家运行时数据结构（Session、缓存等）。
  - `playeractor/entity/player_role.go`：PlayerRole 核心实现，封装公共操作。
  - `playeractor/entity/reconnect_key.go`：重连 Key 管理。
  - `playeractor/entitysystem/sys_mgr.go`：系统管理器，生命周期驱动。
  - `playeractor/entitysystem/system_registry.go`：系统工厂注册表。
  - `playeractor/entitysystem/message_dispatcher.go`：离线消息回放调度。
  - `playeractor/controller/player_network_controller.go`：入口控制器，处理 ForwardMessage/PlayerActorMsg。
  - `playeractor/controller/bag_controller.go`：背包协议解析与业务 Service 调用。
  - `playeractor/controller/chat_controller.go`：聊天协议入口与限频检查。
  - `playeractor/controller/dungeon_item_controller.go`：副本内拾取等物品协议转发。
  - `playeractor/controller/equip_controller.go`：装备穿脱协议入口。
  - `playeractor/controller/fuben_controller.go`：副本进入/结算协议入口。
  - `playeractor/controller/money_controller.go`：货币查看/变更协议入口。
  - `playeractor/controller/move_controller.go`：移动协议路由到 DungeonActor。
  - `playeractor/controller/player_account_controller.go`：注册/登录协议入口。
  - `playeractor/controller/player_role_controller.go`：角色创建/列表/进入游戏协议。
  - `playeractor/controller/recycle_controller.go`：回收协议入口。
  - `playeractor/controller/revive_controller.go`：复活协议入口。
  - `playeractor/controller/skill_controller.go`：技能学习/升级协议入口。
  - `playeractor/event/event_adapter.go`：事件发布适配器。
  - `playeractor/gateway/account_gateway.go`：账号仓储实现（DB/GORM）。
  - `playeractor/gateway/client_gateway.go`：客户端 Session 访问/发包封装。
  - `playeractor/gateway/config_gateway.go`：配置读取实现。
  - `playeractor/gateway/dungeon_server_gateway.go`：DungeonActor 调用适配（内部 Actor 消息）。
  - `playeractor/gateway/interfaces.go`：Gateway 抽象接口集合。
  - `playeractor/gateway/player_gateway.go`：玩家数据仓储实现。
  - `playeractor/gateway/public_actor_gateway.go`：与 PublicActor 交互的发送封装。
  - `playeractor/gateway/role_gateway.go`：角色数据访问实现。
  - `playeractor/gateway/token_generator.go`：登录 Token 生成工具。
  - `playeractor/system/attrcalc/bus.go`：属性计算器 Provider 总线（具体计算器接口统一在 `server/internal/attrcalc`）。
  - `playeractor/presenter/push_helpers.go`：下行推送公共辅助。
  - `playeractor/presenter/player_auth_presenter.go`：注册/登录结果回包。
  - `playeractor/presenter/player_role_presenter.go`：角色列表/进入游戏回包。
  - `playeractor/presenter/bag_presenter.go`：背包相关 S2C 构建。
  - `playeractor/presenter/chat_presenter.go`：聊天回包构建。
  - `playeractor/presenter/equip_presenter.go`：装备回包构建。
  - `playeractor/presenter/fuben_presenter.go`：副本回包构建。
  - `playeractor/presenter/item_use_presenter.go`：物品使用回包。
  - `playeractor/presenter/money_presenter.go`：货币回包。
  - `playeractor/presenter/quest_presenter.go`：任务回包。
  - `playeractor/presenter/recycle_presenter.go`：回收回包。
  - `playeractor/presenter/shop_presenter.go`：商城回包。
  - `playeractor/presenter/skill_presenter.go`：技能回包。
  - `playeractor/router/protocol_registry.go`：协议注册表。
  - `playeractor/router/protocol_router_controller.go`：协议路由控制器，按 MsgId 分发。
  - `playeractor/system/base_sys.go`：SystemAdapter 基类，声明生命周期职责。
  - `playeractor/system/attr_use_case_adapter.go`：属性用例适配，供跨系统调用。
  - `playeractor/system/bag_sys.go`：背包生命周期调度与缓存索引。
  - `playeractor/system/bag_use_case_adapter.go`：背包 UseCase 适配器（解循环依赖）。
  - `playeractor/system/chat_sys.go`：聊天系统生命周期与事件订阅。
  - `playeractor/system/equip_sys.go`：装备系统生命周期调度。
  - `playeractor/system/fuben_sys.go`：副本系统生命周期调度。
  - `playeractor/system/item_use_sys.go`：物品使用生命周期与冷却管理。
  - `playeractor/system/level_sys.go`：等级系统生命周期调度。
  - `playeractor/system/level_use_case_adapter.go`：等级 UseCase 适配器。
  - `playeractor/system/message_sys.go`：玩家离线消息加载/回放调度。
  - `playeractor/system/money_sys.go`：货币系统生命周期调度与余额初始化。
  - `playeractor/system/quest_sys.go`：任务系统生命周期与事件映射。
  - `playeractor/system/recycle_sys.go`：回收系统生命周期调度。
  - `playeractor/system/shop_sys.go`：商城系统生命周期调度。
  - `playeractor/system/skill_sys.go`：技能系统生命周期调度。
  - `playeractor/system/attrcalc/add_rate_bus.go`：属性加成注册总线（加成项汇总）。
  - `playeractor/system/attrcalc/bus.go`：属性计算器注册/调度总线。
  - `playeractor/service/interfaces/attr.go`：属性相关用例接口。
  - `playeractor/service/interfaces/bag.go`：背包用例接口。
  - `playeractor/service/interfaces/blacklist.go`：黑名单接口。
  - `playeractor/service/interfaces/chat_rate_limiter.go`：聊天限频接口。
  - `playeractor/service/interfaces/config.go`：配置读取接口。
  - `playeractor/service/interfaces/consume.go`：通用扣耗接口。
  - `playeractor/service/interfaces/event.go`：事件发布接口。
  - `playeractor/service/interfaces/event_subscriber.go`：事件订阅接口。
  - `playeractor/service/interfaces/level.go`：等级接口。
  - `playeractor/service/interfaces/money.go`：货币接口。
  - `playeractor/service/interfaces/player_role_manager.go`：角色管理接口。
  - `playeractor/service/interfaces/public_actor.go`：PublicActor 网关接口。
  - `playeractor/service/interfaces/reward.go`：通用发奖接口。
  - `playeractor/service/interfaces/rpc.go`：DungeonActor RPC 接口。
  - `playeractor/service/interfaces/runone.go`：RunOne 调度接口。
  - `playeractor/service/interfaces/time_callback.go`：定时回调接口。
  - `playeractor/service/interfaces/token_generator.go`：Token 生成接口。
  - `playeractor/service/attr/calculate_sys_power.go`：计算系统战力用例。
  - `playeractor/service/attr/compare_attr_vec.go`：属性向量比较用例。
  - `playeractor/service/bag/add_item.go`：非事务加物品用例。
  - `playeractor/service/bag/add_item_tx.go`：事务加物品用例。
  - `playeractor/service/bag/remove_item.go`：非事务移除物品用例。
  - `playeractor/service/bag/remove_item_tx.go`：事务移除物品用例。
  - `playeractor/service/bag/has_item.go`：检查物品是否存在。
  - `playeractor/service/chat/chat_private.go`：私聊用例。
  - `playeractor/service/chat/chat_world.go`：世界聊天用例。
  - `playeractor/service/consume/consume_use_case.go`：通用扣耗实现（复用 Bag/Money）。
  - `playeractor/service/equip/equip_item.go`：装备穿戴用例。
  - `playeractor/service/equip/unequip_item.go`：装备卸下用例。
  - `playeractor/service/fuben/enter_dungeon.go`：进入副本用例。
  - `playeractor/service/fuben/get_dungeon_record.go`：查询副本记录用例。
  - `playeractor/service/fuben/settle_dungeon.go`：副本结算用例。
  - `playeractor/service/item_use/use_item.go`：物品使用用例。
  - `playeractor/service/level/add_exp.go`：加经验用例。
  - `playeractor/service/level/init_level_data.go`：等级数据初始化用例。
  - `playeractor/service/level/level_up.go`：升级用例。
  - `playeractor/service/mail/claim_attachments.go`：领取附件用例。
  - `playeractor/service/mail/read_and_delete.go`：邮件阅读/删除用例。
  - `playeractor/service/mail/send_custom_mail.go`：发送自定义邮件用例。
  - `playeractor/service/mail/send_template_mail.go`：发送模板邮件用例。
  - `playeractor/service/money/add_money.go`：加货币用例。
  - `playeractor/service/money/consume_money.go`：扣货币用例。
  - `playeractor/service/money/init_money_data.go`：货币数据初始化。
  - `playeractor/service/money/money_use_case_impl.go`：MoneyUseCase 组合实现。
  - `playeractor/service/money/update_balance_tx.go`：事务余额更新。
  - `playeractor/service/playerauth/login.go`：登录用例。
  - `playeractor/service/playerauth/register.go`：注册用例。
  - `playeractor/service/playerrole/create_role.go`：创角用例。
  - `playeractor/service/playerrole/query_roles.go`：角色列表查询。
  - `playeractor/service/quest/accept_quest.go`：接取任务用例。
  - `playeractor/service/quest/init_quest_data.go`：任务数据初始化。
  - `playeractor/service/quest/refresh_quest_type.go`：日/周任务刷新用例。
  - `playeractor/service/quest/submit_quest.go`：提交任务用例。
  - `playeractor/service/quest/update_progress.go`：任务进度更新。
  - `playeractor/service/recycle/recycle_item.go`：回收物品用例。
  - `playeractor/service/reward/reward_use_case.go`：通用发奖实现。
  - `playeractor/service/shop/buy_item.go`：商城购买用例。
  - `playeractor/service/skill/init_skill_data.go`：技能数据初始化。
  - `playeractor/service/skill/learn_skill.go`：学习技能用例。
  - `playeractor/service/skill/upgrade_skill.go`：升级技能用例。
  - `publicactor/adapter.go`：PublicActor 适配器。
  - `publicactor/handler.go`：PublicActor 主循环。
  - `publicactor/register.go`：消息注册。
  - `publicactor/message_handler.go`：处理来自玩家的公共消息。
  - `publicactor/public_role.go`：全局角色视图管理（在线、缓存）。
  - `publicactor/public_role_chat.go`：聊天相关公共数据。
  - `publicactor/public_role_online.go`：在线状态管理。
  - `publicactor/public_role_rank.go`：排行榜快照与更新。
  - `publicactor/public_role_offline_data.go`：离线快照加载/写库。
  - `publicactor/public_role_offline_message.go`：离线消息桥接。
  - `publicactor/public_handler_util.go`：公共处理辅助函数。
  - `publicactor/offlinedata/manager.go`：OfflineDataManager 实现。
  - `publicactor/send_client.go`：下发客户端广播/通知。
  - `dungeonactor/adapter.go`：DungeonActor 适配器，实现 IDungeonActorFacade。
  - `dungeonactor/handler.go`：DungeonActor 主循环与消息处理。
  - `dungeonactor/register.go`：注册 DungeonActor 消息处理器。
  - `dungeonactor/message.go`：DungeonActor 消息定义/派发。
  - `dungeonactor/entity/entity.go`：实体基础结构。
  - `dungeonactor/entity/drop_item.go`：掉落物实体与拾取逻辑。
  - `dungeonactor/entity/monster.go`：怪物实体定义与属性计算。
  - `dungeonactor/entity/revive_provider.go`：复活能力提供者。
  - `dungeonactor/entity/rolest.go`：玩家角色在副本内镜像。
  - `dungeonactor/entitymgr/mgr.go`：实体管理器，生命周期维护。
  - `dungeonactor/entitymgr/handle.go`：实体管理事件处理。
  - `dungeonactor/entitysystem/ai_state.go`：AI 状态枚举与切换。
  - `dungeonactor/entitysystem/ai_sys.go`：AI 系统，驱动自动移动/攻击。
  - `dungeonactor/entitysystem/aoi_sys.go`：AOI 系统。
  - `dungeonactor/entitysystem/attr_sys.go`：副本内属性系统（重算/广播）。
  - `dungeonactor/entitysystem/attrcalc/bus.go`：副本侧属性计算器注册。
  - `dungeonactor/entitysystem/broadcast.go`：场景广播工具。
  - `dungeonactor/entitysystem/buff_sys.go`：Buff 系统。
  - `dungeonactor/entitysystem/debuff_helper.go`：减益效果工具。
  - `dungeonactor/entitysystem/drop_sys.go`：掉落分配/拾取系统。
  - `dungeonactor/entitysystem/fight_sys.go`：战斗系统（技能/伤害结算）。
  - `dungeonactor/entitysystem/move_sys.go`：移动系统（Start/Update/EndMove 校验）。
  - `dungeonactor/entitysystem/pathfinding.go`：寻路算法（A* 等）。
  - `dungeonactor/entitysystem/skill_util.go`：技能计算辅助。
  - `dungeonactor/entitysystem/state_machine.go`：实体状态机。
  - `dungeonactor/fbmgr/mgr.go`：副本管理器（房间/实例调度）。
  - `dungeonactor/fuben/default.go`：默认副本实现。
  - `dungeonactor/fuben/fubenst.go`：副本状态定义。
  - `dungeonactor/fuben/actor_msg.go`：与 PlayerActor 的进入成功消息。
  - `dungeonactor/fuben/settlement.go`：副本结算、发消息回 PlayerActor。
  - `dungeonactor/fuben/timed_provider.go`：限时副本时间提供者。
  - `dungeonactor/iface/iaoi.go`：AOI 接口。
  - `dungeonactor/iface/iattr.go`：属性接口。
  - `dungeonactor/iface/ibuff.go`：Buff 接口。
  - `dungeonactor/iface/idrop.go`：掉落接口。
  - `dungeonactor/iface/ientity.go`：实体接口。
  - `dungeonactor/iface/ifight.go`：战斗接口。
  - `dungeonactor/iface/ifuben.go`：副本接口。
  - `dungeonactor/iface/imonster.go`：怪物接口。
  - `dungeonactor/iface/imove.go`：移动接口。
  - `dungeonactor/iface/irole.go`：副本内角色接口。
  - `dungeonactor/iface/iscene.go`：场景接口。
  - `dungeonactor/scene/scenest.go`：场景结构、地图绑定与出生点校验。
  - `dungeonactor/scene/scene_aoi.go`：场景 AOI 具体实现。
  - `dungeonactor/scenemgr/mgr.go`：场景管理器。
  - `dungeonactor/skill/skill.go`：技能主流程与目标筛选。
  - `dungeonactor/skill/skill_cast_result.go`：技能施放结果结构。
  - `dungeonactor/skill/skill_damage_calculator.go`：技能伤害计算。
  - `dungeonactor/skill/skill_hit_result.go`：技能命中结果结构。
  - `dungeonactor/skill/skill_result.go`：技能结果聚合。

---

## 3. 构建与运行现状

- **Go 版本**：`go 1.24.0`，toolchain `1.24.10`。  
- **最新可执行文件**：  
  - `go build -o server/output/gameserver.exe ./server/service/gameserver`  
  - `gateway.exe` 已存在历史构建产物  
- **配置**：`server/output/{gateway,gamesrv}.json` + `server/output/config/*.json` 必须齐备。  
- **数据库**：`server/output/postapocgame.db` 随 GameServer 自动迁移，表定义位于 `server/internal/database/*.go`。  
- **日志**：各服务默认输出到 `server/output/log/<service>.log`，同时打印至控制台。  
- **启动顺序建议**：GameServer → Gateway → 客户端。

---

## 4. 已完成功能

> 若新增子系统或完成阶段性能力，请在对应子节补充实现细节，并描述关键入口/注意事项。

### 4.1 Gateway（接入层）

- 双协议接入：TCP + WebSocket
- Session 生命周期管理、消息转发、限流与资源保护
- 关键代码：`server/service/gateway/internel/clientnet`、`server/service/gateway/internel/engine`

### 4.2 GameServer（玩家主逻辑）

**账号 / 角色 / Session**
- 注册、登录、Token 认证（bcrypt）
- 角色创建/删除/查询/进入游戏流程，账号最多 3 角色
- Session 扩展（AccountID/Token）

**玩家 Actor 与系统框架**
- PlayerRoleActor（ModePerKey）、EntitySystem 动态注册
- BinaryData 加载/保存、事件总线、RunOne 定期存盘

**统一时间与服务器广播**
- `server/internal/servertime` 统一 UTC 时间源
- `timesync.Broadcaster` 每秒广播服务器时间

**成长 / 经济 / 玩法系统（当前保留）**
- 背包、货币、装备、属性、等级、离线收益、GM、反作弊（任务 / 商城 / 物品使用系统已移除，仅保留历史记录）

**副本与限时玩法支撑**
- ✅ 副本进入/结算链路：`FubenSys` 负责 `DungeonData` 读写，支持副本记录按天重置、次数限制与进入冷却（`GetDungeonRecord/GetOrCreateDungeonRecord/CheckDungeonCD/EnterDungeon`）
- ✅ 限时副本校验：`handleEnterDungeon` 按 `DungeonConfig` 校验副本存在性、类型（限时）、难度、每日进入次数与消耗物品，失败场景均返回明确错误码与文案
- ✅ 副本结算回写：`handleSettleDungeon` 在副本成功后更新进入记录，并根据奖励类型拆分为经验/货币/物品，由 `LevelSys/GrantRewards` 分别落到角色数据与背包

**防作弊与 GM 能力**
- ✅ `AntiCheatSys`：基于 `SiAntiCheatData` 维护操作计数与每日重置，支持 10 秒窗口 100 次的频率限制、可疑计数与 1 小时临时封禁/永久封禁（`CheckOperationFrequency/RecordSuspiciousBehavior/BanPlayer`）
- ✅ `GMSys`：玩家级 GM 系统，支持通过 `C2SGMCommand` 协议下发 GM 指令，由 `GMManager` 执行业务逻辑；GM 执行结果通过 `S2CGMCommandResult` 回传客户端
- ✅ 系统广播与系统邮件：`SendSystemNotification* / SendSystemMail* / SendSystemMailByTemplate* / GrantRewardsByMail` 支持单人/全服广播与系统邮件发放，可复用为运营活动工具
- 🆕 2025-12-10：修复 GM 邮件用临时仓储实现，接口签名改为携带 `context.Context`，与 `PlayerRepository` 规范一致，避免编译报错并确保通过 Context 注入 PlayerRole。
- 🆕 2025-12-10：nilaway 全量通过，配置读取与实体访问统一判空并返回明确错误码，GM 工具仓储在玩家缺失时返回空结构，gatewaylink/deps/playerRoleManager 全局依赖安全初始化，寻路优先队列拒绝 nil 节点。

**副本协作**
- `FubenSys`、`SkillSys` 负责副本进入、技能同步、掉落拾取
- ✅ 所有客户端 C2S 协议统一在 PlayerActor Controller 层处理，通过 `gshare.SendDungeonMessageAsync` 将请求转发到 DungeonActor，使用 `DungeonActorMsgId` 枚举统一管理内部消息

**Phase 3 社交经济系统（全部完成）**
- ✅ **PublicActor 框架**：单 Actor 框架、在线状态管理、消息路由
- ✅ **聊天系统**：世界/私聊、频率限制、敏感词过滤、离线消息（持久化、数量限制、过期清理）
- ✅ **好友系统**：申请/同意/拒绝、列表查询、在线状态联动
- ✅ **排行榜系统**：查询、快照注册、数值更新、自动刷新（上线/等级变化）
- ✅ **公会系统**：创建/解散、申请加入、审批流程、权限管理（会长/副会长/组长/成员）、数据持久化
- ✅ **拍卖行系统**：上架/购买/浏览、过期处理、货币结算、物品交付、数据持久化
- ✅ **社交安全系统**：配置化敏感词库、交易审计、黑名单机制
- ✅ **离线数据管理器（RankSnapshot 首版）**：PublicActor 引入 `OfflineDataManager`（`publicactor/offlinedata`），落地 `OfflineData` 数据表、`UpdateOfflineDataMsg` 协议与 `PublicActorMsgIdUpdateOfflineData`，完成玩家上线/定时/下线的离线快照写库、DB 启动加载与 60s 定时 Flush，`FriendSys`/排行榜查询可直接读取离线快照；详见《`docs/离线数据管理器开发文档.md`》
- ✅ **PublicActor 交互统一**：`PlayerRole` 登录/登出/排行榜快照/离线数据/QueryRank 全部通过 `PublicActorGateway` 发送消息，新建 `sendPublicActorMessage` 辅助函数消除业务侧直接调用 `gshare.SendPublicMessageAsync`
- 🆕 **玩家消息系统（阶段一：数据库层）**：新增 `PlayerActorMessage` 表（消息类型+序列化数据+时间戳），提供 `Save/Load/Delete/Count` 等 DAO 封装，支持按 `msgId` 增量加载；详见《`docs/玩家消息系统开发文档.md`》
- 🆕 **玩家消息系统（阶段二：回放框架）**：实现 `engine/message_registry.go` 消息注册中心 + `entitysystem/message_sys.go` 离线消息回放（OnInit/登录/重连自动加载、回调成功后删库，失败保留），并在 `proto/csproto/system.proto` 增加 `SysMessage`
- 🆕 **玩家消息系统（阶段三：发送入口）**：新增 `gshare.SendPlayerActorMessage`（在线直接向 Actor 投递，失败/离线回落入库）、`player_network.handlePlayerMessageMsg`（Actor 内调度消息回调）以及 `rpc.proto/AddActorMessageMsg`，完成在线/离线统一链路

**属性系统阶段一（基础结构）**
- `entitysystem/attr_sys.go` 支持 `sysAttr/sysAddRateAttr/sysPowerMap` 缓存、差异化重算与 `ResetSysAttr` 对外接口；`SyncAttrData` 追加 `AddRateAttr` 字段，下行仅同步变更系统。

**属性系统阶段二（加成与推送）**
- `attrcalc/add_rate_bus.go` 提供加成计算注册；示例 `level_sys` 基于角色等级注入 HP/MP 回复加成，`AttrSys.calcTotalSysAddRate` 自动汇总并写入 `sysAddRateAttr`。
- `proto/sc.proto` 的 `S2CAttrDataReq` 携带 `SyncAttrData+sys_power_map`；GameServer 在属性变更、首次登录、重连时通过 `AttrSys.pushAttrDataToClient` 推送属性快照。

**Clean Architecture 重构（进行中）**
- 🆕 **系统依赖关系清理**：已完成 SysRank 和已移除系统的依赖关系清理
  - ✅ 已在 `proto/csproto/system.proto` 中为 `SysRank = 19` 添加注释说明：RankSys 是 PublicActor 功能，不参与 PlayerActor 系统管理
  - ✅ 已确认 `sys_mgr.go` 不再使用 `systemDependencies`，改为按 SystemId 顺序初始化
  - ✅ 已确认 proto 中不包含已移除的系统ID（VipSys、DailyActivitySys、FriendSys、GuildSys、AuctionSys）
  - ✅ 已确认没有系统注册 SysRank 和已移除的系统ID，符合预期
- 🆕 **MessageSys 功能完善**：已完成离线消息回放机制检查、消息类型与回调扩展检查、消息持久化与过期清理实现
  - ✅ 离线消息回放机制：`MessageSys` 在 `OnInit`、`OnRoleLogin`、`OnRoleReconnect` 时自动加载离线消息，回调成功后删库，失败保留
  - ✅ 消息类型与回调扩展：消息注册机制完善，支持任意消息类型扩展，消息分发逻辑覆盖所有场景
  - ✅ 消息持久化与过期清理：已在 `OnNewDay` 中实现过期消息清理（超过7天的消息），在 `RunOne` 中实现消息数量限制（每个玩家最多1000条消息）
  - ✅ UseCase 层评估：当前实现简洁清晰，业务逻辑不复杂，保持现状即可
- 🆕 **阶段一：基础结构搭建**：已完成目录结构创建、基础接口定义、基础设施适配层实现、系统生命周期适配器、依赖注入容器框架
  - ✅ 创建了 `domain/repository/`、`service/interfaces/`、`controller/`、`system/`、`gateway/`、`presenter/`、`router/`、`event/` 等目录结构
  - ✅ 定义了所有基础接口（Repository、EventPublisher、PublicActorGateway、DungeonServerGateway、ConfigManager 等）
  - ✅ 实现了所有 Gateway 和 Adapter（NetworkGateway、PublicActorGateway、DungeonServerGateway、EventAdapter、ConfigGateway、PlayerGateway）
  - ✅ 实现了 BaseSystemAdapter 和 Context Helper
  - ✅ 实现了依赖注入容器基础框架
- 🆕 **2025-12-10 瘦身动作（P0）**：删除 `di/` 与旧 context 适配层，统一依赖装配到 `playeractor/deps`，Context 取值收敛到 `core/gshare/context_helper.go`（`cd server && go test ./...` 通过）
- 🆕 **2025-12-10 UseCase 精简**：删除 Equip/ItemUse/Fuben 的 Init*UseCase 空实现，Equip 用例去掉 BagSys 旧兼容分支，SystemAdapter OnInit 只保留必要调度，避免无意义空调用。
- 🆕 **试点系统重构（LevelSys）**：已完成 LevelSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/level/add_exp.go` 和 `usecase/level/level_up.go`（提取业务逻辑）
  - ✅ 创建了 `adapter/system/level_system_adapter.go`（系统生命周期适配器）
  - ✅ 实现了 `GetLevelSys(ctx)` 函数和系统注册
  - ✅ 实现了属性计算器支持（CalculateAttrs 和 levelAddRateCalculator）
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **核心系统重构（BagSys）**：已完成 BagSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/bag/add_item.go`、`remove_item.go`、`has_item.go`（提取业务逻辑）
  - ✅ 创建了 `adapter/controller/bag_controller.go`（协议处理：C2SOpenBag、D2GAddItem）
  - ✅ 创建了 `adapter/presenter/bag_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/bag_system_adapter.go`（系统生命周期适配器）
  - ✅ 索引/快照逻辑收敛到 `usecase/bag/accessor.go`，SystemAdapter 只路由 UseCase，不再维护本地索引或快照
  - ✅ 注册了系统适配器工厂和协议处理器
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **核心系统重构（MoneySys）**：已完成 MoneySys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/money/add_money.go`、`consume_money.go`（提取业务逻辑）
  - ✅ 创建了 `usecase/money/money_use_case_impl.go`（实现 MoneyUseCase 接口，供 LevelSys 使用）
  - ✅ 创建了 `adapter/controller/money_controller.go`（协议处理：C2SOpenMoney）
  - ✅ 创建了 `adapter/presenter/money_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/money/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 实现了 `GetMoneySys(ctx)` 函数和系统注册
  - ✅ 实现了 MoneyUseCase 接口，支持 LevelSys 依赖注入
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **核心系统重构（EquipSys）**：已完成 EquipSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/equip/equip_item.go`、`unequip_item.go`（提取业务逻辑）
  - ✅ 创建了 `adapter/controller/equip_controller.go`（协议处理：C2SEquipItem）
  - ✅ 创建了 `adapter/presenter/equip_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/equip/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 创建了 `adapter/controller/bag_use_case_adapter.go`（实现 BagUseCase 接口，解决循环依赖）
  - ✅ 实现了 `GetEquipSys(ctx)` 函数和系统注册
  - ✅ 通过接口依赖 BagSys，避免循环依赖
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **核心系统重构（AttrSys）**：已完成 AttrSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/attr/mark_dirty.go`、`calc_attr.go`、`run_one.go`（接口定义）
  - ✅ 创建了 `adapter/system/attr/attr_system_adapter.go`（系统适配器，实现核心逻辑）
  - ✅ 创建了 `adapter/system/attr/attr_system_adapter_helper.go`（GetAttrSys 函数）
  - ✅ 创建了 `adapter/system/attr/attr_system_adapter_init.go`（系统注册）
  - ✅ 实现了 `RunOne` 方法（计算变动的系统属性并同步到DungeonActor）
  - ✅ 实现了 `MarkDirty` 方法（标记需要重算的系统）
  - ✅ 实现了 `CalculateAllAttrs` 方法（计算所有系统的属性）
  - ✅ 通过 `attrcalc` 包注册的计算器获取各系统属性（LevelSys 和 EquipSys 已注册）
  - ✅ 属性同步到 DungeonActor（通过 DungeonServerGateway）
  - ✅ 属性推送到客户端（通过 NetworkGateway）
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **统一数据访问和网络发送**：已完成阶段二系统的统一验证
  - ✅ 所有 Use Case 层通过 `PlayerRepository` 接口访问数据
  - ✅ 所有 System Adapter 层通过 `PlayerGateway`（实现 `PlayerRepository`）访问数据
  - ✅ `PlayerGateway` 正确实现接口，保持 BinaryData 共享引用模式
  - ✅ 所有 Presenter 通过 `NetworkGateway` 发送消息
  - ✅ 所有 Controller 通过 Presenter 构建和发送响应
  - ✅ AttrSys 通过 `NetworkGateway` 发送消息
  - ✅ 创建了验证文档 `docs/统一数据访问和网络发送验证.md`
- 🆕 **阶段三：玩法系统重构（ItemUseSys）**：已完成 ItemUseSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/item_use/use_item.go`（使用物品用例）
  - ✅ 创建了 `usecase/interfaces/level.go`（LevelUseCase 接口定义）
  - ✅ 创建了 `adapter/controller/item_use_controller.go`（协议处理：C2SUseItem）
  - ✅ 创建了 `adapter/controller/level_use_case_adapter.go`（LevelUseCase 适配器）
  - ✅ 创建了 `adapter/presenter/item_use_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/item_use/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 实现了 `GetItemUseSys(ctx)` 函数和系统注册
  - ✅ 通过接口依赖 BagSys 和 LevelSys，避免循环依赖
  - ✅ 完善了 ConfigManager 接口（添加 GetItemUseEffectConfig、GetJobConfig、GetEquipSetConfig）
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
  - ⏳ TODO: 完善 HP/MP 同步到 DungeonActor 的逻辑（通过事件或接口）
- 🆕 **阶段三：玩法系统重构（SkillSys）**：已完成 SkillSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/skill/learn_skill.go`、`upgrade_skill.go`（提取业务逻辑）
  - ✅ 创建了 `usecase/interfaces/consume.go`（ConsumeUseCase 接口定义）
  - ✅ 创建了 `adapter/controller/skill_controller.go`（协议处理：C2SLearnSkill、C2SUpgradeSkill）
  - ✅ 创建了 `adapter/controller/consume_use_case_adapter.go`（ConsumeUseCase 适配器）
  - ✅ 创建了 `adapter/presenter/skill_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/skill/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 完善了 `usecase/interfaces/level.go`（添加 GetLevel 方法）
  - ✅ 实现了 `GetSkillSys(ctx)` 函数和系统注册
  - ✅ 通过接口依赖 LevelSys 和 ConsumeUseCase，避免循环依赖
  - ✅ 实现了技能同步到 DungeonActor 的逻辑（通过 DungeonServerGateway）
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **阶段三：玩法系统重构（FubenSys）**：已完成 FubenSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/fuben/enter_dungeon.go`、`settle_dungeon.go`（提取业务逻辑）
  - ✅ 创建了 `adapter/controller/fuben_controller.go`（协议处理：C2SEnterDungeon、D2GSettleDungeon、D2GEnterDungeonSuccess）
  - ✅ 创建了 `adapter/controller/reward_use_case_adapter.go`（RewardUseCase 适配器）
  - ✅ 创建了 `adapter/presenter/fuben_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/fuben/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 实现了 RPC 处理器注册（通过 DungeonServerGateway）
  - ✅ 实现了 `GetFubenSys(ctx)` 함수和系统注册
  - ✅ 通过接口依赖 ConsumeUseCase、LevelUseCase、RewardUseCase，避免循环依赖
  - ✅ 实现了进入副本和副本结算的完整流程
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
- 🆕 **阶段三：玩法系统重构（QuestSys）**：已完成 QuestSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/quest/accept_quest.go`、`update_progress.go`、`submit_quest.go`（提取业务逻辑）
  - ✅ 创建了 `usecase/interfaces/daily_activity.go`（DailyActivityUseCase 接口定义）
  - ✅ 创建了 `adapter/controller/quest_controller.go`（协议处理：C2STalkToNPC）
  - ✅ 创建了 `adapter/controller/daily_activity_use_case_adapter.go`（DailyActivityUseCase 适配器）
  - ✅ 创建了 `adapter/presenter/quest_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/quest/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 完善了 `usecase/interfaces/config.go`（添加 GetQuestConfigsByType、GetNPCSceneConfig）
  - ✅ 实现了 `OnNewDay` 和 `OnNewWeek` 方法（每日/每周刷新任务）
  - ✅ 实现了 `GetQuestSys(ctx)` 函数和系统注册
  - ✅ 订阅玩家事件（OnNewDay、OnNewWeek）用于刷新日常/周常任务
  - ✅ 通过接口依赖 LevelUseCase、RewardUseCase、DailyActivityUseCase，避免循环依赖
  - ✅ 更新了 SkillSys，在学习技能时触发任务进度更新
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
  - ⏳ TODO: DailyActivitySys 重构后完善 DailyActivityUseCase 适配器
- 🆕 **阶段三：玩法系统重构（ShopSys）**：已完成 ShopSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/shop/buy_item.go`（提取业务逻辑：购买商品、构建消耗/奖励列表）
  - ✅ 创建了 `adapter/controller/shop_controller.go`（协议处理：C2SShopBuy）
  - ✅ 创建了 `adapter/presenter/shop_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/shop/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - ✅ 完善了 `usecase/interfaces/config.go`（添加 GetShopConfig、GetConsumeConfig、GetRewardConfig）
  - ✅ 完善了 `adapter/gateway/config_gateway.go`（实现新的配置接口方法）
  - ✅ 实现了 `GetShopSys(ctx)` 函数和系统注册
  - ✅ 通过接口依赖 ConsumeUseCase、RewardUseCase，避免循环依赖
  - ✅ 购买成功后推送背包和货币数据更新（通过 Presenter）
  - ✅ 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
  - ⏳ TODO: purchaseCounters 当前为内存数据，如需持久化可后续完善
- 🆕 **阶段三：玩法系统重构（RecycleSys）**：已完成 RecycleSys 的 Clean Architecture 重构
  - ✅ 创建了 `usecase/recycle/recycle_item.go`（回收物品用例，负责校验配置、扣除物品、发放奖励）
  - ✅ 创建了 `adapter/controller/recycle_controller.go`（协议处理：C2SRecycleItem）
  - ✅ 创建了 `adapter/presenter/recycle_presenter.go`（响应构建）
  - ✅ 创建了 `adapter/system/recycle/` 包（适配器、Helper、Init）
  - ✅ 实现了 `GetRecycleSys(ctx)` 函数，并以单例形式暴露回收能力
  - ✅ 完善了 `usecase/interfaces/config.go` 与 `adapter/gateway/config_gateway.go`（新增 `GetItemRecycleConfig`）
  - ✅ 新增 `adapter/controller/push_helpers.go`，统一回收/商城后的背包与货币推送逻辑
  - 🆕 Legacy RecycleSys 清理：删除 `playeractor/entitysystem/recycle_sys.go`，旧 `clientprotocol` 注册与 Gateway 推送逻辑统一转移至 Clean Architecture 控制器
  - ✅ 通过接口依赖 BagUseCase、RewardUseCase，避免直接依赖 EntitySystem
  - ✅ 保持了向后兼容性（接口化依赖，可与旧系统并行）
- 🆕 **阶段四：社交系统重构（Guild/Chat/Auction）**：已完成 GuildSys / ChatSys / AuctionSys 的 Clean Architecture 重构
  - ✅ GuildSys：`domain/guild/guild.go`、`usecase/guild/*`、`adapter/controller|presenter|system/guild/`，通过 PublicActorGateway 异步创建/加入/退出公会，删除旧 `entitysystem/guild_sys.go`
  - ✅ ChatSys：`domain/chat/chat.go`、`usecase/chat/chat_world.go|chat_private.go`、`usecase/interfaces/chat_rate_limiter.go`、`adapter/system/chat/`（限频）以及新的 Controller/Presenter；敏感词从 ConfigGateway 拉取，全部消息经过 PublicActor
  - ✅ AuctionSys：`domain/auction/auction.go`、`usecase/auction/put_on.go|buy.go|query.go`、`adapter/controller|presenter|system/auction/`；所有上架/购买/查询请求经 PublicActorGateway 转发，删除旧 `entitysystem/auction_sys.go`
- 🆕 **阶段四：社交系统重构（FriendSys）**：已完成 FriendSys 的 Clean Architecture 重构
  - ✅ 创建了 `domain/friend/friend.go`（封装好友数据初始化、列表增删工具）
  - ✅ 创建了 `usecase/friend/`（发送/响应好友申请、删除好友、查询列表、黑名单操作）
  - ✅ 创建了 `usecase/interfaces/blacklist.go` 与 `adapter/gateway/blacklist_repository.go`（黑名单仓储接口及实现）
  - ✅ 创建了 `adapter/controller/friend_controller.go` + `adapter/presenter/friend_presenter.go`（统一协议入口与响应构建）
  - ✅ 创建了 `adapter/system/friend/`（系统适配器、Helper、Init，负责工厂注册与协议订阅）
  - ✅ 通过 `PublicActorGateway` 转发 `AddFriendReq/Resp/FriendListQuery` 异步消息，保持与 PublicActor 交互解耦
  - ✅ 删除旧版 `entitysystem/friend_sys.go`，所有逻辑迁移到 Clean Architecture 层次
- 🆕 **阶段三：玩法系统重构（RPC 调用链路）**：已完成 GameServer ↔ DungeonServer RPC 管理重构
- 🆕 **阶段六：Legacy EntitySystem 清理**：删除 `server/service/gameserver/internel/app/playeractor/entitysystem` 下 Bag/Money/Level/Skill/Quest/Fuben/ItemUse/Shop/Attr/Equip 等旧实现，仅保留 `sys_mgr.go` 与 `message_dispatcher.go`；所有调用统一改用 `adapter/system` + UseCase/Controller，避免 Legacy 入口与循环依赖
  - ✅ 新增 `adapter/controller/protocol_router_controller.go`，集中处理 `C2S` 协议解析、上下文注入与 DungeonActor 转发
  - ✅ `player_network_controller.go` 承接 EnterGame/QueryRank/PlayerActorMsg 处理，注册统一收敛到 `adapter/controller/register_all.go`
  - ✅ `DungeonServerGateway` 扩展 `RegisterProtocols/UnregisterProtocols`、`GetSrvTypeForProtocol` 返回自定义枚举，统一 RPC 入口
  - ✅ `adapter/system/bag_system_adapter_init.go` 等路径均改为通过 `DungeonServerGateway.RegisterRPCHandler` 注册回调
  - ✅ `docs/服务端开发进度文档.md` 第 7.8 节与 `docs/gameserver_CleanArchitecture重构文档.md` 19.3.2 节记录最新架构决策
  - ✅ 新增《`docs/gameserver_目录与调用关系说明.md`》，系统性梳理 `server/service/gameserver` 目录结构与调用关系，便于新人按 Clean Architecture 视角理解整体架构

**服务器开服信息**
- `server/internal/database/server_info.go` 存储开服时间，`gshare.GetOpenSrvDay()` 获取开服天数

### 4.3 DungeonActor（副本与战斗）

- 单 Actor 主循环、实体层（角色/怪物/掉落/AOI）
- 状态机、Buff、AI、技能、战斗系统
- ✅ **移动系统重构**：重新实现移动系统
  - ✅ **服务端时间驱动移动**：实现 `MovingTime` 方法，根据时间计算当前位置，支持服务端驱动移动
  - ✅ **客户端位置校验**：实现 `LocationUpdate` 方法，校验客户端上报的位置是否合理，防止移动过快或瞬移
  - ✅ **坐标系统统一**：内部使用像素坐标进行精确计算，与场景系统交互时转换为格子坐标
  - ✅ **移动状态管理**：维护移动开始时间、起始位置、目标位置、移动距离等状态信息
  - ✅ **速度容差处理**：给与客户端速度1.5倍与200ping的容忍度，兼容网络延迟和服务端卡顿
  - ✅ **移动协议流程优化**：
    - ✅ **C2SStartMove 处理**：收到 C2SStartMove 执行 handleStart，记录 move_data（包含目的地），广播 S2CStartMove（携带实体hdl和move_data）
    - ✅ **C2SUpdateMove 处理**：收到 C2SUpdateMove 执行 handleUpdate，判定客户端要更新的坐标和服务端计算的坐标误差是否会很大，支持有1s的误差，如果差距很大就要结束移动执行 handleEnd，广播S2CEndMove，通知客户端结束移动，最终坐标为上一个点；否则就更新客户端给过来最新的坐标
    - ✅ **C2SEndMove 处理**：收到 C2SEndMove 执行 handleEnd，广播S2CEndMove，通知客户端结束移动
- ✅ **移动系统重构**：`MoveSys` 专注于移动功能，移除所有AI相关业务逻辑；AI系统通过组合调用 `HandleStartMove` → `HandleUpdateMove` → `HandleEndMove` 模拟客户端移动协议，保持移动系统代码简洁干净
- 场景地图加载：`scene_config.json` 通过 `mapId` 关联 `map_config.json`，启动时转换为 `GameMap`（宽高、阻挡、随机出生点、移动校验一致）
- 出生点判定：`scene_config.BornArea` 作为玩家/实体出生范围，进入场景时随机点位需落在 `GameMap` 可行走区域，否则回退到全局随机可行走点
- **寻路系统**：支持 A* 算法（绕障碍、贴墙走）和直线寻路（最短直线、遇障碍自动绕过），`MonsterAIConfig` 可配置巡逻/追击时使用的寻路算法，`MoveSys.MoveToWithPathfinding` 提供寻路移动接口
- 技能释放校验（`Effects` 为空时返回错误）
- 技能广播（释放成功/伤害结果，客户端驱动动画）
- **消息注册**：DungeonActor 不再直接注册或处理任何 C2S 协议，所有客户端协议统一在 PlayerActor Controller 层处理；DungeonActor 通过 `gshare.IDungeonActorFacade.RegisterHandler` 注册内部消息处理器（`DungeonActorMsgId`），在 `register.go` 中按业务模块拆分注册
- **坐标系统优化**：
  - ✅ 移动系统：客户端发送的像素坐标自动转换为格子坐标进行校验，距离计算和速度校验基于格子大小
  - ✅ 寻路算法：明确输入输出为格子坐标，距离计算使用格子距离（曼哈顿距离或欧几里得距离）
  - ✅ 场景系统：统一使用格子坐标，所有位置相关函数明确标注坐标类型
  - ✅ 客户端（server/example）：统一规范为发送像素坐标，服务端自动转换为格子坐标
  - ✅ 协议定义：在移动、实体、技能相关协议中添加坐标类型注释
  - ✅ 距离和范围计算：技能范围、攻击范围统一使用格子距离，添加详细注释

**属性系统阶段一（聚合同步）**
- `entitysystem/attr_sys.go` 提供 `ApplySyncData/ensureAggregated`，可直接聚合 GameServer 下发的系统属性与 `AddRateAttr`；`entity/rolest.go` 的 `UpdateAttrs/initAttrSys` 仅透传 `SyncAttrData`，移除旧的逐属性增减逻辑。

**属性系统阶段四（DungeonActor 本地属性计算与 RunOne 机制）**
- ✅ **属性计算器注册管理器**：`entitysystem/attrcalc/bus.go` 提供 `RegIncAttrCalcFn/RegDecAttrCalcFn` 和 `GetIncAttrCalcFn/GetDecAttrCalcFn`，支持注册怪物基础属性计算器（`MonsterBaseProperty`）、Buff 属性计算器（`SaBuff`）等
- ✅ **ResetSysAttr 方法**：`AttrSys.ResetSysAttr` 通过注册管理器触发属性重算，支持增量和减量属性计算
- ✅ **RunOne 机制**：`AttrSys.RunOne` 每帧调用 `ResetProperty` 和 `CheckAndSyncProp`，在 `BaseEntity.RunOne` 中统一调用
- ✅ **非战斗属性变化跟踪**：使用 `extraUpdateMask`（map）跟踪非战斗属性变化，`CheckAndSyncProp` 检查并同步变化
- ✅ **初始化完成标志**：`bInitFinish` 标志控制是否广播属性（初始化完成前不广播），`SetInitFinish` 方法标记初始化完成
- ✅ **实体属性重置**：`MonsterEntity.ResetProperty` 先调用 `ResetSysAttr` 计算基础属性，再调用 `AttrSys.ResetProperty` 触发完整流程
- ✅ **属性同步到 PlayerActor**：属性变化时通过 `gshare.SendMessageAsync` 发送 `PlayerActorMsgIdSyncAttrs` 消息给 PlayerActor

### 4.4 共享基础能力

- Actor 框架（ModeSingle / ModePerKey）
- 网络编解码、压缩、连接池
- 统一时间源（servertime）
- 事件总线（gevent）
- Proto 体系
- 日志、错误码、工具函数
- **日志上下文请求器**：`server/pkg/log` 新增 `IRequester` 接口及 `NewRequester/WithRequester` 系列全局函数，可在调用侧声明前缀与栈深，`gshare`/`player_network` 的上下文日志已全部改为通过请求器注入 `session/role` 信息，堆栈定位指向真实业务调用者
- **坐标系统定义与转换**：在 `server/internal/argsdef/position.go` 中定义了格子大小常量（TileSize=128、TileCenterOffset=64）和坐标转换函数（格子坐标↔像素坐标），统一了坐标系统的定义和使用规范

### 4.5 调试客户端（Golang 文字冒险面板）

- 新增 `server/example` 文本冒险式调试客户端，默认自动连接 Gateway，提供 `register/login/roles/create-role/enter/status/look/move/attack` 等命令，完整覆盖账号→角色→进场景→战斗链路
- 基于 `AdventurePanel` + `GameClient` + Actor 管理器实现，所有命令均直接映射到现有协议，便于录制/回放
- `docs/golang客户端待开发文档.md` 记录后续扩展计划与命令映射，是调试客户端的唯一规划文档
- 交互层采用“标题区 + 日志区 + 命令区”三段式布局，支持数字/快捷键菜单操作，贴合古早文字冒险体验
- ✅ `move` 命令前会加载 `map_config` → `GameMap`，自动裁剪并校验坐标，避免向服务器发送越界/不可行走的移动请求
- ✅ `move` 命令发送链路已对齐正式客户端：按 `C2SStartMove → C2SUpdateMove（逐格）→ C2SEndMove` 顺序逐步上报像素坐标，服务端移动系统可稳定复现客户端行为
- 🆕 新增《`docs/server_example重构方案.md`》，定义与 `server/service` 同构的客户端目录、脚本化回调与业务扩展阶段目标
- 🆕 **Phase A 架构对齐完成**：`server/example` 采用 `cmd/example + internal/{client,panel,systems}` 结构；`GameClient` 拆分为 `client.Core` + `systems` 四子系统（Account/Scene/Move/Combat）；`AdventurePanel` 菜单与命令实现迁移至 `panel/actions.go`；协议等待通道统一由 `internal/client/flow.go` 管理
- 🆕 **Phase B MoveRunner**：`internal/client/move_runner.go` + `systems.Move` 支持直线优先+A* 寻路、速度容错、自动重试与回调；面板提供 `move-to/move-resume`、脚本可复用同一链路
- 🆕 **Phase C 背包/副本/GM/脚本**：`systems.Inventory/Dungeon/GM/Script` 封装背包、GM、副本与 Demo 脚本；CLI 新增 `bag/use-item/pickup/gm/enter-dungeon/script-record/script-run` 指令，支持录制/回放

---

## 5. Phase 3 社交经济架构设计

### 5.1 架构原则

**PublicActor 职责**
1. 全局数据管理：公会、拍卖行、排行榜等权威数据
2. 跨玩家消息路由：聊天、好友请求、公会审批等
3. 在线状态管理：`roleId → SessionId` 映射
4. 快照缓存：角色展示、排行榜展示数据
5. 数据持久化：公会与拍卖行必须落库（独立表）

**PlayerActor 职责（已完成 Clean Architecture 重构）**
1. 管理自身数据：好友、公会归属、拍卖上架记录等社交数据均存于 `PlayerRoleBinaryData`，通过 Repository 接口读写。
2. Clean Architecture 承载业务：`FriendSys`、`GuildSys`、`AuctionSys` 等均按领域（`domain`）、用例（`usecase`）、适配层（`adapter/{controller,system,presenter}`）拆分实现，不再放在 `entitysystem` 目录。
3. 数据更新 → 主动通知 PublicActor（排行榜数值、快照、在线状态），通过 `PublicActorGateway` 与 `PlayerRole.sendPublicActorMessage` 统一发送。
4. 客户端协议入口：由 Controller 层接入请求、做合法性校验并调用 UseCase，必要时通过 PublicActor 回调。

**协作模式**
```
客户端
  ↓
PlayerActor（校验 + 防刷）
  ↓
PublicActor（全局逻辑 / 数据持久化）
  ↓
PlayerActor（通知/回写）
  ↓
客户端
```

### 5.2 Proto 与数据定义

> 修改 Proto 后务必运行 `proto/genproto.sh && gofmt`.

- `proto/csproto/social_def.proto`：ChatType、ChatMessage、RankType、RankData、PlayerRankSnapshot、GuildPosition、GuildMember、GuildData、AuctionItem
- `proto/csproto/system.proto`：SiFriendData、SiGuildData、SiAuctionData
- `proto/csproto/player.proto`：PlayerRoleBinaryData 新增社交字段
- `proto/csproto/rpc.proto`：PublicActor ↔ PlayerActor 内部消息定义

详细定义见代码，此处不再重复。

### 5.3 社交系统实现模式（Clean Architecture 版）

- **数据归属**：好友、公会归属、拍卖上架记录等仍存放在 `PlayerRoleBinaryData` 中的社交相关字段，由对应 Repository 接口（如 `FriendRepository`、`GuildRepository` 等）读写。  
- **分层结构**：社交系统按 Clean Architecture 落地，分别位于 `domain/{friend,guild,chat,auction}`、`usecase/{friend,guild,chat,auction}` 与 `adapter/{controller,presenter,system}/{friend,guild,chat,auction}`，不再依赖旧的 `entitysystem/*_sys.go`。  
- **PublicActor 协作**：所有跨玩家、全局状态与广播能力统一经 `PublicActorGateway` 访问 PublicActor；PlayerActor 侧禁止直接调用 `gshare.SendPublicMessageAsync`，必须通过 `PlayerRole.sendPublicActorMessage` 完成。  
- **系统注册**：社交相关的 SystemAdapter 通过 `adapter/system/{friend,guild,chat,auction}` 在 `sys_mgr.go` 中注册，不再占用 `protocol.SystemId` 中已移除的 `SysFriend/SysGuild/SysAuction` 等系统枚举。  

### 5.4 PublicActor 持久化策略

- **公会**：`database.Guild` 表，变更时立即持久化
- **拍卖行**：`database.AuctionItem` 表，定期过期检测
- **数据恢复**：GameServer 启动时加载所有公会/拍卖数据到 PublicActor
- **日志对齐**：重要事件需写 `pkg/log` 并可选落地到审计表

---

## 6. 待实现 / 待完善功能

> 完成后请及时上移至第 4 章，并描述实现细节。  
> **约定**：本章统一使用复选框形式跟踪进度——`[ ]` 表示待做，`[⏳]` 表示进行中，`[✅]` 表示已完成但仍保留在清单中用于追溯。

### 6.1 Clean Architecture 重构（进行中）

- [✅] **执行指南同步**：2025-12-03 新增《`docs/gameserver_CleanArchitecture重构实施手册.md`》，沉淀分层映射、阶段路线、子系统任务清单与验收标准，后续迭代直接在该文档维护阶段性待办。
- [⏳] **兼容代码清理规划**：2025-12-03 新增《`docs/gameserver_兼容代码清理规划.md`》，梳理 AntiCheat/MessageSys/GM Tools 等遗留兼容层的迁移与删除路线；后续删除 legacy 代码时需按此文档执行并在本章同步状态。
  - [✅] 阶段一：物理删除空壳领域目录 `internel/domain/vip` 与 `internel/domain/dailyactivity`，仅影响目录结构，不涉及业务代码，详见《`docs/gameserver_兼容代码清理规划.md`》2.1 节。
- [✅] **阶段一：基础结构搭建**
  - ✅ 目录结构创建
  - ✅ 基础接口定义
  - ✅ 基础设施适配层实现
  - ✅ 系统生命周期适配器
  - ✅ 依赖注入容器框架
  - ✅ 试点系统重构（LevelSys）
  - [⏳] 编写单元测试和集成测试
  - [⏳] 验证功能正常

- [✅] **阶段二：核心系统重构**
  - ✅ BagSys（背包系统）
  - ✅ MoneySys（货币系统）
  - ✅ EquipSys（装备系统）
  - ✅ AttrSys（属性系统）——曾重构为工具类 `AttrCalculator`，当前版本已移除属性计算工具类，仅保留基础数据结构
  - [⏳] 统一数据访问和网络发送验证

- [✅] **阶段三：玩法系统重构**
  - ✅ SkillSys、FubenSys、RecycleSys（QuestSys / ItemUseSys / ShopSys 已在后续阶段整体移除）
  - ✅ RPC 调用链路（ProtocolRouterController + DungeonServerGateway）
  - [⏳] 阶段性联调与旧 EntitySystem 清理
- [ ] **PlayerActor Phase2A：特性分片**
  - [✅] Bag 系统收拢至 `internel/app/playeractor/bag`（Controller/System/Presenter/UseCase + 适配器），保留 init 注册与外部接口，SystemAdapter 使用本地基类避免循环依赖。
  - [✅] Money 系统收拢至 `internel/app/playeractor/money`，用 `money.Deps` 聚合依赖，小 service 化用例，SystemAdapter 使用本地基类避免循环依赖。
  - [ ] 规划 Money / Equip / Skill 等系统的分片路径与依赖拆解，输出迁移清单与循环依赖风险评估。
- [ ] **PlayerActor Phase2B：Bag 过程化改造**
  - Bag 用例由 struct+方法改为“函数/小 service + 显式依赖装配”，过渡期按文档的 `depsFromGlobalOrApp(ctx)` 方案落地。

- [✅] **阶段四：社交系统重构**
  - ✅ FriendSys（好友系统）——完成 Use Case / Controller / Presenter / System Adapter、黑名单仓储接口，统一通过 PublicActorGateway 交互
  - ✅ GuildSys（公会系统）——完成领域/用例/Controller/Presenter/System Adapter，所有创建/加入/退出统一走 PublicActorGateway
  - ✅ ChatSys（聊天系统）——完成用例拆分（世界/私聊）、限频接口与 System Adapter、敏感词配置接入
  - ✅ AuctionSys（拍卖行系统）——完成上架/购买/查询用例、Controller/Presenter/System Adapter，全部拍卖请求通过 PublicActorGateway
  - ✅ PublicActor 交互重构——玩家登录/登出/排行榜/查询排行榜统一走 `PublicActorGateway`，新增 `PlayerRole.sendPublicActorMessage` 封装发送逻辑

- [⏳] **阶段五：辅助系统重构**
  - [✅] MailSys（邮件系统）——完成领域/用例/SystemAdapter 迁移，GMSys 系统邮件工具统一经 `usecase/mail` + `gmPlayerRepository` 写入 `PlayerRoleBinaryData.MailData`
  - [✅] VipSys（VIP 系统）——**已移除**（按最小粒度游戏需求，移除 VIP 系统）
  - [✅] DailyActivitySys（日常活跃度系统）——**已移除**（按最小粒度游戏需求，移除日常活跃系统）
  - [ ] MessageSys（玩家消息系统）——已将旧 `MessageSys` 生命周期迁移到 `adapter/system/message`，保留 `DispatchPlayerMessage` 作为统一回放入口；`message_dispatcher.go` 现仅包含“数据库消息 → Proto 工厂 → 回调”单一路径，不再依赖旧 EntitySystem；后续仅需在有新业务场景时在 `engine/message_registry.go` 注册消息类型与回调，并补齐控制台/监控能力（见 6.1“玩家消息系统阶段四”）
  - [✅] GMSys（GM 系统，GM 命令与工具函数 Clean Architecture 化）——新增 `adapter/system/gm`（GMSystemAdapter + GMManager + GM 工具函数），在系统适配器中注册 `C2SGMCommand` 协议处理；旧 `entitysystem/gm_sys.go` 与 `gm_manager.go` 已删除，GM 工具函数集中于 `adapter/system/gm/gm_tools.go`，仅作为系统通知/系统邮件的 Helper，并通过 `usecase/mail` 完成发货逻辑，不再承担底层业务规则

- [⏳] **阶段六：清理与优化**
  - Legacy EntitySystem 移除进度：
    - [✅] RecycleSys：删除 `server/service/gameserver/internel/app/playeractor/entitysystem/recycle_sys.go`（目录现已不存在），客户端协议与 DungeonServer 回写链路全部由 `adapter/system/recycle` + `RecycleController` 托管
    - [✅] BagSys / MoneySys / LevelSys / SkillSys / QuestSys / FubenSys / ItemUseSys / ShopSys / AttrSys / EquipSys：逐一确认所有调用切换到 SystemAdapter 与 UseCase 后已删除 legacy `*_sys.go`，`entitysystem` 目录目前仅保留系统管理器与离线消息分发器
  - [⏳] 完善测试
  - [⏳] 文档更新

- [⏳] **阶段七：SystemAdapter 职责精简与 UseCase 下沉（按 `docs/gameserver_adapter_system演进规划.md` 执行）**
  - [✅] 阶段 A：梳理职责边界（A1 已在文档中完成系统清单与职责归类；A2 在关键逻辑处添加 TODO 标记，后续迭代可继续补充）
  - [⏳] 阶段 B：业务逻辑下沉到 UseCase（进行中）
  - [✅] 阶段 C：精简生命周期与事件处理（保留"胶水"逻辑）
    - ✅ C1：统一 SystemAdapter 生命周期签名与职责说明（所有 SystemAdapter 已添加统一的头部注释，BaseSystemAdapter 已添加详细的职责说明）
    - ✅ C2：将复杂定时/调度逻辑下沉或抽象（QuestSystemAdapter 的 OnNewDay/OnNewWeek 已只调用 UseCase，任务刷新逻辑已下沉到 RefreshQuestTypeUseCase）
    - ✅ C3：事件订阅精简（已删除空的事件订阅，保留的事件订阅已添加注释说明）
    - ✅ Money 系统：
      - `UpdateBalanceTxUseCase` 已创建，`MoneySystemAdapter.UpdateBalanceTx` 改为调用 UseCase，余额计算与不足校验规则已下沉
      - `InitMoneyDataUseCase` 已创建，`MoneySystemAdapter.OnInit` 改为调用 UseCase，MoneyData 初始化与默认金币注入逻辑已下沉
    - ✅ Bag 系统：`AddItemTxUseCase` 与 `RemoveItemTxUseCase` 已创建，`BagSystemAdapter` 的 Tx 方法改为调用 UseCase，物品配置校验、堆叠规则、容量检查等业务逻辑已下沉
    - ✅ Level 系统：`InitLevelDataUseCase` 已创建，`LevelSystemAdapter.OnInit` 改为调用 UseCase，等级默认值修正、经验同步等初始化逻辑已下沉
    - ✅ ItemUse 系统：`InitItemUseDataUseCase` 已创建，`ItemUseSystemAdapter.OnInit` 改为调用 UseCase，冷却映射结构初始化逻辑已下沉
    - ✅ Equip 系统：`InitEquipDataUseCase` 已创建，`EquipSystemAdapter.OnInit` 改为调用 UseCase，装备列表结构初始化逻辑已下沉
    - ✅ Skill 系统：`InitSkillDataUseCase` 已创建，`SkillSystemAdapter.OnInit` 改为调用 UseCase，按职业配置初始化基础技能列表逻辑已下沉
    - ✅ Quest 系统：`InitQuestDataUseCase` 已创建，`QuestSystemAdapter.OnInit` 改为调用 UseCase，任务桶结构与基础任务类型集合初始化逻辑已下沉
    - ✅ Attr 系统：
      - `CalculateSysPowerUseCase` 已创建，`AttrSystemAdapter.calcSysPowerMap` 改为调用 UseCase，系统战力计算逻辑已下沉
      - `CompareAttrVecUseCase` 已创建，`AttrSystemAdapter.calculateSystemAttr` 中的属性向量比较逻辑已下沉
    - ✅ Fuben 系统：
      - `InitDungeonDataUseCase` 已创建，`FubenSystemAdapter.OnInit` 改为调用 UseCase，副本记录容器结构初始化逻辑已下沉
      - `GetDungeonRecordUseCase` 已创建，`FubenSystemAdapter.GetDungeonRecord` 改为调用 UseCase，副本记录查找逻辑已下沉
    - ✅ Shop 系统：主要业务逻辑已在 `BuyItemUseCase` 中，适配层已精简为薄胶水层
    - ✅ Recycle 系统：主要业务逻辑已在 `RecycleItemUseCase` 中，适配层已精简为薄胶水层
    - ✅ Friend 系统：`InitFriendDataUseCase` 已创建，`FriendSystemAdapter.OnInit` 改为调用 UseCase，好友列表与好友申请列表结构初始化逻辑已下沉
    - ✅ Guild 系统：`InitGuildDataUseCase` 已创建，`GuildSystemAdapter.OnInit` 改为调用 UseCase，公会数据初始化逻辑已下沉
    - ✅ Chat 系统：限流逻辑（`CanSend`、`MarkSent`）属于框架状态管理，保留在适配层，符合 Clean Architecture 原则
    - ✅ Auction 系统：`InitAuctionDataUseCase` 已创建，`AuctionSystemAdapter.OnInit` 改为调用 UseCase，拍卖ID列表结构初始化逻辑已下沉
    - ✅ Mail 系统：`InitMailDataUseCase` 已创建，`MailSystemAdapter.OnInit` 改为调用 UseCase，邮件列表结构初始化逻辑已下沉
    - ✅ Vip 系统：**已移除**（按最小粒度游戏需求，移除 VIP 系统）
    - ✅ DailyActivity 系统：**已移除**（按最小粒度游戏需求，移除日常活跃系统）
    - ✅ Message 系统：主要逻辑为加载离线消息，属于框架层面的消息处理，保留在适配层
    - ✅ GM 系统：主要逻辑为执行 GM 命令，属于框架层面的命令处理，保留在适配层
  - [⏳] 阶段 B：下沉业务逻辑到 UseCase（按玩法/社交/辅助系统分批推进，并补充用例单测）
    - ✅ B1.3 单元测试：已为关键 UseCase 编写单元测试
      - Money：`InitMoneyDataUseCase` 测试完成
      - Level：`InitLevelDataUseCase` 测试完成
      - Bag：`AddItemTxUseCase`、`RemoveItemTxUseCase` 测试完成
      - Attr：`CompareAttrVecUseCase`、`CalculateSysPowerUseCase` 测试完成
  - [✅] 阶段 C：精简生命周期与事件处理（保留"胶水"逻辑，引入统一调度工具）
    - ✅ C1：为所有 SystemAdapter 添加统一的头部注释，说明生命周期方法的职责
    - ✅ C2：创建 `RefreshQuestTypeUseCase`，将任务刷新逻辑下沉到 UseCase 层
    - ✅ C3：精简事件订阅，删除空订阅，为保留的订阅添加注释说明
  - [⏳] 阶段 D：测试与文档补全（在 CleanArchitecture 文档与本文件中补充结果）
    - ✅ D2：为 SystemAdapter 编写轻量集成测试或脚本验证清单（已完成：创建了 `docs/SystemAdapter验证清单.md`）
    - ✅ D3：更新文档（已完成：在 Clean Architecture 文档中新增 SystemAdapter 章节，更新了服务端开发进度文档）
    - [⏳] D1：为关键 UseCase 编写单元测试（进行中：已有部分测试，待补充 Equip、Fuben、Quest、Shop、Recycle、Skill、ItemUse 等系统的测试）
  - [✅] 阶段 E：清理 Legacy 代码与防退化机制（删除不再需要的字段/方法，增加防退化检查项）
    - ✅ E1：清理不再需要的 legacy SystemAdapter 逻辑或字段
      - ✅ 修复了 item_use_system_adapter.go 中的 TODO（使用 servertime 获取当前时间）
      - ✅ 检查并确认所有方法都有明确的用途，无遗留的未使用方法
    - ✅ E2：增加防退化检查
      - ✅ 在 BaseSystemAdapter 中添加了详细的防退化机制说明
      - ✅ 为所有 SystemAdapter 头部注释添加了防退化说明（"禁止编写业务规则逻辑，只允许调用 UseCase 与管理生命周期"）
      - ✅ 创建了 Code Review 清单文档（`docs/SystemAdapter_CodeReview清单.md`），包含职责边界检查、代码质量检查、注释和文档检查、防退化机制检查等检查项
    - **关键代码位置**：
      - BaseSystemAdapter：`server/service/gameserver/internel/adapter/system/base_system_adapter.go`
      - Code Review 清单：`docs/SystemAdapter_CodeReview清单.md`
      - 所有 SystemAdapter：`server/service/gameserver/internel/adapter/system/*_system_adapter.go`
  - [✅] **阶段八：Controller 层系统开启检查优化（按 `docs/SystemAdapter系统开启检查优化方案.md` 执行）**
    - [✅] 为所有 Controller 添加系统开启状态检查（在调用 UseCase 之前检查系统是否开启）
      - ✅ BagController：`HandleOpenBag`、`HandleAddItem`
      - ✅ MoneyController：`HandleOpenMoney`
      - ✅ EquipController：`HandleEquipItem`
      - ✅ SkillController：`HandleLearnSkill`、`HandleUpgradeSkill`
      - ✅ QuestController：`HandleTalkToNPC`
      - ✅ FubenController：`HandleEnterDungeon`、`HandleSettleDungeon`
      - ✅ ItemUseController：`HandleUseItem`
      - ✅ ShopController：`HandleShopBuy`
      - ✅ RecycleController：`HandleRecycleItem`（已更新错误码）
      - ✅ ChatController：`HandleWorldChat`、`HandlePrivateChat`（已更新错误码）
      - ✅ GM 系统：`HandleGMCommand`（已更新错误码）
    - [✅] 统一错误处理（为系统未开启的情况定义统一的错误码和错误消息）
      - ✅ 在 `proto/csproto/error_code.proto` 中添加了 `System_NotFound` 和 `System_NotEnabled` 错误码
      - ✅ 在 `server/internal/protocol/error_code_init.go` 中注册了新的错误码
      - ✅ 所有 Controller 统一使用 `customerr.NewErrorByCode(int32(protocol.ErrorCode_System_NotEnabled), "系统未开启")` 返回错误
    - [⏳] 补充测试（为每个 Controller 添加系统未开启场景的测试用例）
    - **关键代码位置**：
      - 优化方案文档：`docs/SystemAdapter系统开启检查优化方案.md`
      - Controller 层：`server/service/gameserver/internel/adapter/controller/*`
      - SystemAdapter Helper：`server/service/gameserver/internel/adapter/system/*_system_adapter_helper.go`
    - **问题背景**：当前 Controller -> UseCase 的流程没有经过 System 的开启/关闭检查，如果系统未开启，UseCase 仍然会被执行，不符合预期。期望流程为：Controller -> [检查 System 是否开启] -> UseCase

- 🆕 **Clean Architecture 重构实施手册**：新增《`docs/gameserver_CleanArchitecture重构实施手册.md`》，汇总分层映射、阶段路线、子系统任务与验收清单，作为 GameServer 重构执行指南；后续变更请同步本手册与本文第 4/6/7/8 章。

### 6.2 系统精简与移除（进行中）

- [✅] **阶段一：移除 SysVip (VIP系统)**
  - ✅ 删除 proto 文件中的 `SysVip` 枚举和 `SiVipData` message 定义
  - ✅ 删除所有 VIP 系统相关代码文件（Adapter、UseCase、Domain）
  - ✅ 从系统管理器移除 SysVip 注册和依赖关系
  - ✅ 清理所有 VIP 相关引用（包括 `money/add_money.go`、`money_controller.go`、`money_system_adapter.go`）
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`、`proto/csproto/sc.proto`、`proto/csproto/player.proto`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - 依赖清理：`server/service/gameserver/internel/usecase/money/add_money.go`、`server/service/gameserver/internel/adapter/controller/money_controller.go`、`server/service/gameserver/internel/adapter/system/money_system_adapter.go`

- [✅] **阶段二：移除 SysDailyActivity (日常活跃系统)**
  - ✅ 删除 proto 文件中的 `SysDailyActivity` 枚举和 `SiDailyActivityData` message 定义
  - ✅ 删除所有 DailyActivity 系统相关代码文件（Adapter、UseCase、Domain、Interfaces）
  - ✅ 从系统管理器移除 SysDailyActivity 注册和依赖关系（包括从 SysQuest 的依赖列表中移除）
  - ✅ 清理所有 DailyActivity 相关引用：
    - `money/add_money.go`、`money/consume_money.go`：移除活跃点处理逻辑
    - `quest/submit_quest.go`：移除活跃点添加逻辑
    - `money_controller.go`、`quest_controller.go`：移除 DailyActivity 用例注入
    - `money_system_adapter.go`、`quest_system_adapter.go`：移除 DailyActivity 用例注入
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`、`proto/csproto/player.proto`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - 依赖清理：`server/service/gameserver/internel/usecase/money/`、`server/service/gameserver/internel/usecase/quest/submit_quest.go`、相关 Controller 和 SystemAdapter
- [✅] **阶段三：移除 SysFriend (好友系统)**
  - ✅ 删除 proto 文件中的 `SysFriend` 枚举和 `SiFriendData` message 定义
  - ✅ 删除所有 Friend 系统相关代码文件（Adapter、UseCase、Domain、Controller、Presenter）
  - ✅ 删除 PublicActor 中的 `public_role_friend.go` 文件
  - ✅ 从系统管理器移除 SysFriend 注册和依赖关系（包括从 SysGuild 和 SysChat 的依赖列表中移除）
  - ✅ 从 `publicactor/register.go` 中移除 `RegisterFriendHandlers` 调用
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`、`proto/csproto/player.proto`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - PublicActor：`server/service/gameserver/internel/app/publicactor/public_role_friend.go`、`register.go`
- [✅] **阶段四：移除 SysGuild (公会系统)**
  - ✅ 删除 proto 文件中的 `SysGuild` 枚举和 `SiGuildData` message 定义
  - ✅ 删除所有 Guild 系统相关代码文件（Adapter、UseCase、Domain、Controller、Presenter）
  - ✅ 删除 PublicActor 中的 Guild 相关文件（`public_role_guild.go`、`public_role_guild_application.go`、`public_role_guild_persist.go`、`guild_permission.go`）
  - ✅ 从系统管理器移除 SysGuild 注册和依赖关系（包括从 SysChat 和 SysRank 的依赖列表中移除）
  - ✅ 从 `publicactor/register.go` 中移除 `RegisterGuildHandlers` 调用
  - ✅ 从 `publicactor/public_role.go` 中移除 Guild 相关字段（guildMap、guildApplicationMap、nextGuildId、guildIdMu）
  - ✅ 从 `publicactor/handler.go` 中移除加载公会数据的逻辑
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`、`proto/csproto/player.proto`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - PublicActor：`server/service/gameserver/internel/app/publicactor/` 中的多个 Guild 相关文件、`register.go`、`public_role.go`、`handler.go`
- [✅] **阶段五：移除 SysAuction (拍卖行系统)**
  - ✅ 删除 proto 文件中的 `SysAuction` 枚举和 `SiAuctionData` message 定义
  - ✅ 删除所有 Auction 系统相关代码文件（Adapter、UseCase、Domain、Controller、Presenter）
  - ✅ 删除 PublicActor 中的 `public_role_auction.go` 文件
  - ✅ 从系统管理器移除 SysAuction 注册和依赖关系
  - ✅ 从 `publicactor/register.go` 中移除 `RegisterAuctionHandlers` 调用
  - ✅ 从 `publicactor/public_role.go` 中移除 Auction 相关字段（auctionMap、nextAuctionId、auctionIdMu）
  - ✅ 从 `publicactor/handler.go` 中移除加载拍卖行数据的逻辑
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`、`proto/csproto/player.proto`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - PublicActor：`server/service/gameserver/internel/app/publicactor/public_role_auction.go`、`register.go`、`public_role.go`、`handler.go`
- [✅] **阶段六：重构 SysAttr 为工具类（属性系统）**
  - ✅ 从 proto 文件中删除 `SysAttr` 枚举项（保留 AttrVec、AttrSt 等数据结构）
  - ✅ 创建 `AttrCalculator` 工具类（`attr_calculator.go`），包含原 AttrSystemAdapter 的核心计算逻辑
  - ✅ 将 `AttrCalculator` 注入到 `PlayerRole` 中，并在初始化时调用 `Init(ctx)`
  - ✅ 添加 `GetAttrCalculator()` 方法供其他包访问
  - ✅ 创建 `attr_calculator_helper.go` 提供 `GetAttrCalculator(ctx)` 辅助函数
  - ✅ 删除所有 AttrSystemAdapter 相关文件（`attr_system_adapter.go`、`attr_system_adapter_init.go`、`attr_system_adapter_helper.go`）
  - ✅ 从系统管理器移除 SysAttr 注册和依赖关系（包括从 Quest、FuBen、ItemUse、Rank 的依赖列表中移除）
  - ✅ 修改所有使用 `GetAttrSys` 的地方，改为使用 `GetAttrCalculator` 或直接访问 `PlayerRole.attrCalculator`
  - ✅ 修改 `player_role.go`、`adapter/controller/player_network_controller.go`、`fuben_controller.go`、`equip_system_adapter_init.go`、`attr_use_case_adapter.go` 等文件
  - ✅ 验证编译和 lint 检查通过
  - **关键代码位置**：
    - Proto 修改：`proto/csproto/system.proto`
    - 工具类：`server/service/gameserver/internel/app/playeractor/entity/attr_calculator.go`
    - PlayerRole：`server/service/gameserver/internel/app/playeractor/entity/player_role.go`
    - 辅助函数：`server/service/gameserver/internel/adapter/system/attr_calculator_helper.go`
    - 系统注册：`server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`

> 详细移除方案见 `docs/系统移除方案.md`

### 6.3 DungeonActor 架构优化（已完成基础，持续完善）

- [✅] **架构与目录规划**
  - [✅] 在 `server/service/gameserver/internel/app` 下新增 `dungeonactor` 目录，定义 `DungeonActor` 单 Actor 骨架（ModeSingle），拆分为 `adapter.go`、`handler.go`、`register.go`、`message.go`。
  - [✅] 在 `main.go` 中接入 DungeonActor 的创建与启动/停止流程，将其视为与 PublicActor 类似的基础设施 Actor，随 GameServer 一同常驻运行。
- [✅] **Gateway 实现**
  - [✅] 在 `adapter/gateway/dungeon_server_gateway.go` 中实现 InProcess 版 `DungeonServerGateway`：对 UseCase 暴露统一的异步接口，内部直接路由到 GameServer 内部的 DungeonActor，仅支持 InProcess 模式。
- [✅] **DungeonActor 启动与运行**
  - [✅] 已将原 `server/service/dungeonserver/internel` 下的核心代码（`entity`、`entitysystem`、`scene`、`fuben`、`fbmgr`、`skill` 等）完整复制到 `internel/app/dungeonactor`，统一 import 前缀并保持 UTF-8 编码。
  - [✅] GameServer 启动时自动创建并启动 ModeSingle 的 DungeonActor，在 `handler.go` 的 `NewDungeonActorHandler` 中初始化默认副本。
  - [✅] DungeonActor 的 `Loop` 方法驱动 `entitymgr.RunOne` 和 `fbmgr.RunOne`，实现实体和副本的每帧更新。
- [✅] **交互与协议收敛**
  - [✅] 删除 `dungeonactor/{drpcprotocol,clientprotocol}` 包，移除所有 C2S/G2D 本地注册逻辑。
  - [✅] 在 `gshare.actor_facade.go` 中引入 `IDungeonActorFacade` 接口，由 `dungeonactor.NewDungeonActor` 在启动时注入实现。
  - [✅] 在 `proto/csproto/rpc.proto` 中新增 `DungeonActorMsgId` 和 `PlayerActorMsgId` 枚举，统一管理内部消息ID。
  - [✅] 所有客户端 C2S 协议统一在 PlayerActor Controller 层处理，通过 `gshare.SendDungeonMessageAsync` 转发到 DungeonActor。
  - [✅] DungeonActor → PlayerActor 的消息（添加物品、副本结算、坐标同步、属性同步等）通过 `gshare.SendMessageAsync` 发送，使用 `PlayerActorMsgId` 枚举。
- [✅] **清理与收尾**
  - [✅] 物理删除 `server/service/dungeonserver` 代码目录及相关构建脚本，所有战斗相关改动统一在 `internel/app/dungeonactor` 维护。
- [⏳] **性能评估与优化**
  - [⏳] 完成副本进入/结算、移动/技能、掉落/拾取的完整联调与性能评估。

#### 阶段 8.7 DungeonActor 交互与协议收敛实现概要（已完成）

1. 明确 DungeonActor 只作为 GameServer 内部的战斗/副本 Actor，不再直接暴露任何网络协议或 G2D/D2G RPC 枚举；所有客户端 C2S 协议统一由 PlayerActor 侧的 Controller 处理（包括移动、技能、掉落拾取、副本进入/退出等）。  
2. 为保持 Clean Architecture 依赖方向，UseCase 仍然只依赖 `interfaces.DungeonServerGateway` 接口，Gateway 负责将“进入副本/同步属性/更新技能/拾取掉落”等领域意图转换为 DungeonActor 可理解的内部消息，通过 `gshare.SendDungeonMessageAsync` 发送。  
3. PlayerActor ↔ DungeonActor 的通信按照 PublicActor 的模式落地：在 `internel/core/gshare/actor_facade.go` 上扩展出 `IDungeonActorFacade` 接口与 `SetDungeonActorFacade/GetDungeonActorFacade/SendDungeonMessageAsync` 等方法，由 `dungeonactor.NewDungeonActor` 在启动时注入具体实现。  
4. GameServer 侧的 Controller/UseCase 需要驱动 DungeonActor 时，通过 `gshare.SendDungeonMessageAsync` 发送内部消息，使用 `DungeonActorMsgId` 枚举统一内部消息ID；已在 `adapter/controller` 中补齐移动/技能/拾取/复活四条链路，通过 `router.RegisterProtocolHandler` + `SendDungeonMessageAsync("global", NewBaseMessage(ctxWithSession, DungeonActorMsgId_*, data))` 将 C2S 请求转发给 DungeonActor。  
5. DungeonActor 不再注册/处理任何 C2S 协议：已物理删除旧的 `dungeonactor/{clientprotocol,drpcprotocol}` 包，并在 `entitysystem/{move_sys,fight_sys,drop_sys}.go`、`entity/rolest.go` 中移除所有 C2S 注册逻辑，协议注册统一迁移到 GameServer Controller 层。  
6. DungeonActor 通过 `gshare.IDungeonActorFacade.RegisterHandler` 注册内部消息处理器，在 `register.go` 中按业务模块拆分注册（`RegisterMoveHandlers`、`RegisterFightHandlers`、`RegisterFuBenHandlers`）。  
7. DungeonActor → PlayerActor 的消息（添加物品、副本结算、坐标同步、属性同步等）统一通过 `gshare.SendMessageAsync` 发送，使用 `PlayerActorMsgId` 枚举，在 `drop_sys.go`、`move_sys.go`、`settlement.go`、`actor_msg.go`、`attr_sys.go` 中实现对应的发送函数。  
8. 已完成架构收敛：物理删除 `server/service/dungeonserver` 代码目录，所有战斗相关改动统一在 `internel/app/dungeonactor` 维护；通过“Controller → UseCase → Gateway → ActorFacade → DungeonActor”链路，支撑所有战斗/副本相关交互。

#### 阶段 8.2 代码迁移实现概要（已完成）

1. 以批量脚本方式，将原 `server/service/dungeonserver/internel` 下的核心目录（`entity`、`entitymgr`、`entitysystem`、`fbmgr`、`fuben`、`gameserverlink`、`iface`、`scene`、`scenemgr`、`skill` 等）全部复制到 `server/service/gameserver/internel/app/dungeonactor`，保持原有目录结构。  
2. 统一将所有 `postapocgame/server/service/dungeonserver/internel/...` import 前缀替换为 `postapocgame/server/service/gameserver/internel/app/dungeonactor/...`，使新包完全脱离旧路径。  
3. `entitysystem` 下的系统（移动、战斗、AI、AOI、Buff、属性、掉落等）以及 `scene/scenemgr`、`entitymgr`、`fbmgr`、`fuben`、`gameserverlink` 等辅助模块现已与 GameServer 同仓维护。  
4. 执行 `go build ./service/gameserver/...` 验证所有新包均可独立编译，避免隐藏的循环依赖或编码异常。  
5. 物理删除 `server/service/dungeonserver` 代码目录，所有战斗相关改动统一在 `internel/app/dungeonactor` 维护。

#### 阶段 8.3 Gateway 接线与消息交互实现概要（已完成）

1. 在 GameServer 侧保持 Clean Architecture：Controller（如 `FubenController`）仍然只依赖 `DungeonServerGateway` 接口，不直接感知 DungeonActor 实现。  
2. `DungeonServerGatewayImpl.AsyncCall` 直接调用本地 `dungeonactor.GetDungeonActor().AsyncCall`，将消息投递到 DungeonActor 的单线程 Loop。  
3. `DungeonActor` 内使用 `DungeonActorMessage` 封装消息，通过单线程 Actor Loop + `HandleMessage` 分发，保证战斗逻辑仍在单 Actor 线程上执行，符合并发约束。  
4. DungeonActor 通过 `gshare.IDungeonActorFacade.RegisterHandler` 注册内部消息处理器，在 `register.go` 中按业务模块拆分注册。  
5. 整个链路为：客户端 `C2SEnterDungeon` → `FubenController` 做系统开关与参数校验 → `EnterDungeonUseCase` → `DungeonServerGateway.AsyncCall` → `gshare.SendDungeonMessageAsync` → DungeonActor 处理 `DungeonActorMsgIdEnterDungeon` → 通过 `gshare.SendMessageAsync` 发送 `PlayerActorMsgIdEnterDungeonSuccess` / `PlayerActorMsgIdSettleDungeon` → PlayerActor 处理并回写客户端。  
6. 该实现保证 UseCase 与 Controller 完全不感知 Actor 实现细节，只依赖接口；同时保持 Actor 单线程语义，避免跨 goroutine 并发访问战斗状态。  
7. DungeonActor → PlayerActor 的消息（添加物品、副本结算、坐标同步、属性同步等）统一通过 `gshare.SendMessageAsync` 发送，使用 `PlayerActorMsgId` 枚举。

#### 阶段 8.4 DungeonActor → PlayerActor 消息发送实现概要（已完成）

1. 对 `dungeonactor.DungeonActorHandler.Loop` 进行增强：在单线程 Loop 中驱动 `entitymgr.RunOne(now)` 与 `fbmgr.GetFuBenMgr().RunOne(now)`，复用原战斗逻辑的实体与副本主循环，使移动、技能、掉落等系统的 RunOne 能在 GameServer 进程内按帧执行。  
2. 实现 DungeonActor → PlayerActor 的消息发送：
   - `drop_sys.go`：拾取物品后发送 `PlayerActorMsgIdAddItem` 消息
   - `move_sys.go`：移动结束后发送 `PlayerActorMsgIdSyncPosition` 消息
   - `settlement.go`：副本结算时发送 `PlayerActorMsgIdSettleDungeon` 消息
   - `actor_msg.go`：进入副本成功后发送 `PlayerActorMsgIdEnterDungeonSuccess` 消息
   - `attr_sys.go`：属性变化时发送 `PlayerActorMsgIdSyncAttrs` 消息
3. 所有消息统一通过 `gshare.SendMessageAsync` 发送，使用 `PlayerActorMsgId` 枚举，消息体使用对应的 `D2G*Req` Proto 定义（如 `D2GAddItemReq`、`D2GSyncPositionReq` 等）。  
4. PlayerActor 侧通过 `gshare.RegisterHandler` 注册这些消息的处理器，现已集中在 `adapter/controller/register_all.go`（涵盖 Bag/Fuben/PlayerNetwork 等处理器）。  

### 6.3 Phase 2 核心玩法（进行中）

- [ ] 战斗录像：录制 / 回放 / 存储链路
- [ ] 多人副本匹配 / 排队：匹配算法 + 排队策略
- [ ] Boss 特殊机制：阶段技能、掉落表
- [ ] 防作弊链路完善：协议级频率检测、移动测速、伤害 / CD 校验
- [ ] 数据统计与分析：关键事件打点、性能监控、行为分析
- [ ] 技能配置巡检：对 `skillCfg.Effects` 为空的技能补齐配置，避免运行期被判定为释放失败
- [ ] 地图生产工具：`map_config` 可视化编辑、行列与 `scene_config` 校验、阻挡可视化导出
- [✅] 属性系统重构（进行中，高优先级）：参考 `server/server` 实现，分阶段完成 `server/service` 中 GameServer 与 DungeonActor 的属性系统
  - ✅ 阶段一（基础结构）：GameServer `entitysystem/attr_sys.go` 支持系统级缓存、差异化同步与 `AddRateAttr` 扩展；DungeonActor `entitysystem/attr_sys.go` / `entity/rolest.go` 可直接聚合 `SyncAttrData`
  - ✅ 阶段二（加成与推送）：`attrcalc/add_rate_bus.go` 提供加成计算注册，`level_sys` 实现示例加成；`S2CAttrDataReq` 携带 `SyncAttrData + sys_power_map`，GameServer 在变更/登录/重连时推送属性
  - ✅ 阶段三（战力与广播）：DungeonActor `AttrSys` 使用共享 `AttrSet` + `ResetProperty`，在应用 `attr_config.json.formula`（转换/百分比）后向客户端与 GameServer 双向广播；GameServer 复用 `attrpower` 战力计算、`attrpush` 推送配置与统一的属性/加成计算器接口
  - ✅ 阶段四（DungeonActor 本地属性计算）：为 DungeonActor 创建属性计算器注册管理器（`server/service/gameserver/internel/app/dungeonactor/entitysystem/attrcalc/bus.go`），实现 `ResetSysAttr` 方法，支持怪物基础属性计算器（`MonsterBaseProperty`）、Buff 系统属性计算器（`SaBuff`）；属性汇总逻辑：`GameServer属性 + Buff属性 + 其他战斗服系统属性`；在 `proto/csproto/attr_def.proto` 中添加 `SaBuff` 和 `MonsterBaseProperty` 系统ID；详见《`docs/属性系统重构文档.md`》第 8 章
- [⏳] 坐标系统优化（基本完成）：
  - ✅ 坐标系统定义与转换：已完成格子大小常量定义和坐标转换函数（`TileCoordToPixel`、`PixelCoordToTile`、`IsSameTile`），详见 `server/internal/argsdef/position.go`
  - ✅ 移动系统优化：已完成客户端像素坐标到格子坐标的转换，距离计算和速度校验基于格子大小，容差调整为格子距离（详见 `server/service/gameserver/internel/app/dungeonactor/entitysystem/move_sys.go`）
  - ✅ 寻路算法优化：已确保寻路算法输入输出明确为格子坐标，距离计算使用格子距离，添加了详细注释（详见 `server/service/gameserver/internel/app/dungeonactor/entitysystem/pathfinding.go`）
  - ✅ 场景系统优化：已确认场景系统统一使用格子坐标，所有位置相关函数明确标注坐标类型（详见 `server/service/gameserver/internel/app/dungeonactor/scene/scenest.go`）
  - ✅ 客户端优化：调试客户端已明确坐标类型，添加详细注释说明坐标系统（详见 `server/example/game_client.go`）
  - ✅ 协议定义检查：已在移动、实体、技能相关协议中添加坐标类型注释（详见 `proto/csproto/cs.proto`、`proto/csproto/sc.proto`、`proto/csproto/base.proto`）
  - ✅ 距离和范围计算：技能范围、攻击范围已统一使用格子距离，添加了详细注释（详见 `server/service/gameserver/internel/app/dungeonactor/skill/skill.go`）
  - [⏳] 配置和常量优化：当前格子大小硬编码为128，未来可考虑配置化（低优先级，详见 `docs/坐标系统优化建议.md` 第8点）
- [⏳] 日志上下文请求器推广（新）：`server/pkg/log` 与 `gshare` 已支持 `IRequester` 注入 Session/Role 前缀，后续需在 Gateway/DungeonActor/工具脚本等路径替换旧的 `GetSkipCall+字符串拼接` 做法，确保跨模块日志能定位真实调用者

- [⏳] GM 权限与审计体系（新，高优先级）  
  - 目前 `GMSys` 通过 `C2SGMCommand` 协议直接对接 `GMManager`，尚未在协议入口统一做 GM 账号权限校验（账号标记/IP 白名单/签名令牌）与调用频率限制  
  - 需要为 GM 指令接入统一的权限模型（账号等级/角色标签）、审计日志（包含操作者账号/角色/IP/指令/参数/目标）与可配置的 IP/环境白名单，防止线上误操作或被恶意利用  
  - 建议将 GM 相关操作纳入独立审计通道（库表或结构化日志），便于后续问题追踪

- [⏳] 网关 / 副本接入安全加固（新，高优先级）  
  - Gateway WebSocket 当前配置 `AllowedIPs=nil` 且 `CheckOrigin=func() bool { return true }`，仅适合开发环境；生产需启用 IP 白名单、Origin 校验与握手阶段的签名/Token 校验  
  - 以上接入防护应与 7.4 中的 TLS/鉴权要求一并落地，并在配置中提供灰度/开关能力

- [⏳] 玩家消息系统（进行中）
  - ✅ 阶段一：数据库表 `PlayerActorMessage` 与 DAO 接口  
  - ✅ 阶段二：消息回调注册中心 + MessageSys 系统加载/回放  
  - ✅ 阶段三：消息发送接口、PlayerActor 入口与 Proto 扩展  
  - [⏳] 阶段四：控制台/监控与过期清理策略

### 6.2 Phase 4 PvP（待规划）

- [ ] 竞技场系统：匹配、战斗、积分与奖励
- [ ] 实时配对系统：ELO/MMR 算法、断线处理
- [ ] 评分系统：积分计算、排名调整、赛季重置

### 6.3 Golang 文字冒险面板（待完善）

- [✅] GCLI-1 背包/物品指令：`bag/use-item/pickup` 命令已接入 `C2SOpenBag/C2SUseItem/C2SPickupItem`
- [ ] GCLI-2 养成/任务面板：等级/属性/任务/活跃度等信息展示与常用交互
- [ ] GCLI-3 战斗脚本化：录制/回放巡逻与技能循环，输出命中/伤害统计
- [✅] GCLI-4 副本/匹配流程：提供 `enter-dungeon <id> [difficulty]` 命令快速联调副本
- [✅] GCLI-5 GM/调试命令桥接：`gm <name> [args...]` 命令接入 GM 协议，可用于批量脚本
- [✅] GCLI-6 协议录制回放：新增 `script-record/script-run`，支持命令录制与回放
- [ ] GCLI-7 多 Session 支持：单面板管理多客户端（机器人），验证社交/战斗
- [✅] GCLI-9 移动链路增强：`MoveRunner` + `move-to/move-resume` 指令已上线，支持直线优先+BFS 寻路、自动重试与回调
- [✅] GCLI-10 业务扩展脚本：新增 `bag/use-item/pickup/gm/enter-dungeon/script-record/script-run` 命令，完成背包/副本/GM/脚本能力首个版本

### 6.4 离线数据管理器（二期规划）

- [ ] 数据类型扩展：为公会展示、拍卖行陈列、形象外观等新增 `OfflineDataType`，补充 PlayerActor 采集与 PublicActor 解码逻辑。
- [ ] 性能与懒加载：按 `data_type` 分批加载、提供冷启动限流与单玩家按需拉取能力，避免全量加载对大区造成压力。
- [ ] 监控与观测：接入离线数据 Flush 统计（dirtyCount、耗时、失败次数），并提供命令行/GM 查询接口与过期清理策略。
- [ ] 容错：实现磁盘/数据库异常时的降级策略（缓存保留 + 重试队列），同时在 `docs/离线数据管理器开发文档.md` 记录扩展方案。

---

## 7. 开发注意事项与架构决策

### 7.1 Actor / 并发约束

- 每位玩家仅一个 Actor；DungeonActor 单 Actor。禁止在业务逻辑中额外创建 goroutine 访问玩家状态
- 所有玩家系统禁止使用 `sync.Mutex`；数据只能源自 `PlayerRoleBinaryData`
- PublicActor 负责所有跨玩家数据，保持无锁；如需并行计算，必须在 Actor 外封装异步任务并通过消息返回
- `MoveSys` 只专注于移动功能，不包含AI业务逻辑；AI系统通过组合调用移动协议方法（`HandleStartMove` → `HandleUpdateMove` → `HandleEndMove`）实现移动，模拟客户端行为

### 7.2 数据与存储

- 数据库仅存 `PlayerRoleBinaryData`（玩家侧）+ 公会/拍卖等全局数据（PublicActor 持久化）
- 属性由 `AttrSys` 实时计算，禁止持久化最终属性值
- 定期存盘：默认每 5 分钟，可按需调整 `_5minChecker`
- 所有错误日志必须包含关键上下文（RoleID、ItemID、QuestID 等）
- `server/internal/database` 的元数据时间列统一使用秒级 Unix 时间戳（int64），禁止直接写入 `time.Time`
- `scene_config` 的 `width/height` 会在加载 `map_config` 后被 `GameMap` 覆盖，`map_config.tileData` 的 `row/col` 必须与场景尺寸一致；缺失地图时仅退化为随机地图（调试用），上线前必须提供权威地图
- 游戏实体进入场景时必须调用 `SceneSt.GetSpawnPos()`，确保出生点命中 `BornArea` 且可行走；`BornArea` 不合法时需回退随机可行走点并记录告警
- `GameMap` 必须支持坐标 ↔ index 双向转换（`CoordToIndex`/`IndexToCoord`），供移动校验、寻路和客户端验证直接复用，避免自行计算造成越界
- **寻路算法配置**：`MonsterAIConfig.PatrolPathfinding`（巡逻）和 `ChasePathfinding`（追击）分别控制对应状态下的寻路算法，1=直线寻路（最短直线、遇障碍自动绕过），2=A*寻路（绕障碍、贴墙走）；默认巡逻用直线、追击用A*；`AISys` 在 `handleIdle`/`handleChase`/`handleReturn` 中根据配置调用 `moveTowardsWithPathfinding`，自动管理路径缓存和分段移动；直线寻路在遇到障碍时会优先保持直线方向，尝试左右两侧可行走点绕过障碍，确保能够到达目标点
- **坐标系统规范**：
  - **服务端统一使用格子坐标**：所有服务端业务逻辑（移动校验、寻路、距离计算、技能范围、AOI等）统一使用格子坐标进行处理
  - **客户端坐标处理**：客户端根据格子坐标自行转换为像素坐标用于显示和特效；客户端发送给服务端的坐标统一为像素坐标
  - **服务端坐标转换**：客户端发送的坐标统一为像素坐标，服务端接收到后自动转换为格子坐标进行业务处理
  - **坐标定义**：客户端格子大小为 128×128 像素，服务端坐标 (x, y) 代表格子坐标，玩家在格子中心（像素坐标 x*128+64, y*128+64）
  - **坐标转换函数**：已定义坐标转换函数（`argsdef.TileCoordToPixel`、`argsdef.PixelCoordToTile`、`argsdef.IsSameTile`），详见 `server/internal/argsdef/position.go`
  - **已优化模块**：移动系统、寻路算法、场景系统、技能系统、距离计算已统一使用格子坐标；协议定义已明确客户端发送像素坐标；服务端统一将像素坐标转换为格子坐标
- **属性同步规范**：GameServer 仅下发差异化的 `SyncAttrData.AttrData` + 汇总 `AddRateAttr`；DungeonActor 必须通过 `entitysystem.AttrSys.ApplySyncData` 聚合，禁止在 `RoleEntity` 等处自行累加或直接写入属性值。
- **属性推送规范**：GameServer 在属性变更、首登与重连时通过 `S2CAttrData` 推送最新 `SyncAttrData + sys_power_map`，客户端禁止本地推导。
- **玩家消息持久化规范**：离线消息（非聊天）统一写入 `PlayerActorMessage` 表，`MsgData` 存储完整 Proto 字节，时间字段采用 `servertime.Now().Unix()`（秒）；所有 DAO 调用需位于 `server/internal/database/player_actor_message.go`，禁止直接拼 SQL。

### 7.3 协议与 Proto 规范

- 共享枚举/结构统一位于 `proto/csproto/*.proto`；修改后执行 `proto/genproto.sh` 并 `gofmt`
- 枚举放入 `*_def.proto`；系统数据放入 `system.proto`；玩家数据放入 `player.proto`；协议消息在 `cs.proto/sc.proto`；内部消息在 `rpc.proto`
- 若数据无法使用 Proto，放入 `server/internal/argsdef/`
- **协议注册规范（2025-01-XX 更新）**：
  1. **GameServer**：所有 C2S 协议、RPC、事件入口统一在 `adapter/controller/register_all.go` 注册（包含 EnterGame/QueryRank/PlayerActor 消息），`gevent.OnSrvStart` 触发一次注册；Controller 内部通过 UseCase → Presenter 完成执行业务与回包。禁止在 SystemAdapter、EntitySystem 或 `player_network_controller.go` 外新增注册逻辑。  
     - （2025-12-10 更新）账号/角色协议已拆分为 `PlayerAccountController`（注册/登录/验证）与 `PlayerRoleController`（角色列表/创角），EnterGame/QueryRank/PlayerActorMsg 入口迁移至 `player_network_controller.go` 并在 `register_all.go` 统一注册。
  2. **DungeonActor**：不再直接注册或处理任何 C2S 协议，所有客户端协议统一在 PlayerActor Controller 层处理；DungeonActor 通过 `gshare.IDungeonActorFacade.RegisterHandler` 注册内部消息处理器（`DungeonActorMsgId`），在 `register.go` 中按业务模块拆分注册。
  3. **公共约束**：协议处理函数禁止直接访问数据库/网关/RPC，必须经 UseCase/Adapter；注册时需同时声明 Request/Response Proto，并在文档记录关键链路。
- **Proto 生成文件禁止手工修改（2025-12-04 更新）**：
  - 任何对 `server/internal/protocol/*.pb.go` 的改动必须来自重新执行 `proto/genproto.sh`；禁止在 PR 中直接编辑生成文件。
  - CI 增加 “执行 protoc + gofmt + git diff 必须为空” 的检查。发现手工修改视为阻断项。
  - 推荐在本地配置 pre-commit 钩子，自动运行 `proto/genproto.sh` 并拒绝包含 `.pb.go` 修改的提交。
- **Controller/Presenter 初始化清单**：
  - Controller 负责：协议注册、Request 解析、上下文注入（SessionId、RoleId）、调用 UseCase。
  - Presenter 负责：将 UseCase 输出转换为 `S2C`/`Rpc`，统一封装错误码/提示语，所有跨服务回包必须走 Presenter。
  - Controller/Presenter 文件命名统一为 `{system}_controller.go` / `{system}_presenter.go`，注册集中在 `register_all.go`，避免在各处散落 init。

### 7.4 网络与安全

- Gateway Session `Stop` 必须幂等；GameServer 需正确处理 `SessionEventClose`
- `pkg/log` 为唯一日志入口，禁止 `fmt.Println`
- `pkg/log` 支持 `IRequester`，带上下文日志必须组合 `NewRequester/WithRequester` 传入 session/role 信息，同时设置 `GetLogCallStackSkip()` 以确保堆栈定位到真实业务函数
- WebSocket/TCP 尚未加 TLS/鉴权，线上前需补齐
- 防作弊：频率检测、移动测速、伤害验证、CD 校验需逐步接入协议链路
- GameServer 在 Controller 层处理所有 C2S 协议，需要转发到 DungeonActor 的通过 `gshare.SendDungeonMessageAsync` 发送内部 Actor 消息，使用 `DungeonActorMsgId` 枚举
- **CI 审查（2025-12-04 更新）**：
  - ✅ 已实现静态分析检查（2025-12-04）：
    - 创建 `scripts/ci_check.sh` 和 `scripts/ci_check.ps1` 脚本，集成 `go vet` 和 `staticcheck` 检查
    - 支持 Windows (PowerShell) 和 Linux/Mac (Bash) 环境
    - 对 `server/service/gameserver` 目录执行 `go vet ./...` 和 `staticcheck ./...`
  - ✅ 已实现 gatewaylink 导入检查（2025-12-04）：
    - 创建 `scripts/check_gatewaylink_imports.sh` 和 `scripts/check_gatewaylink_imports.ps1` 脚本
    - 白名单包括：PlayerActor 相关文件（`adapter/controller/player_network_controller.go`, `player_role.go`）、Gateway 适配器、gatewaylink 自身、engine/server.go、dungeonserverlink
    - 已集成到 CI 检查脚本中，违规导入将导致检查失败
  - ⏳ 待实现：Proto 防护检查（`scripts/check_proto_clean.sh`）
  - ⏳ 待实现：PR 模板新增检查项：是否运行过 proto 生成脚本、go fmt、go test、静态分析、gatewaylink 使用检查；未通过禁止合入。
- **S2C 发送链路统一（2025-12-04 更新）**：
  - `DungeonActor`、`PublicActor` 以及其它系统代码禁止直接引用 `gatewaylink`；所有发往客户端的协议必须交由 PlayerActor 统一透传。
  - 发送 S2C 时构造 `PlayerActorMsgIdSendToClient`（`PlayerActorMsgIdSendToClientReq{msg_id,data}`）消息，通过 `gshare.SendMessageAsync` 投递给 PlayerActor，对应 Handler `handleSendToClient` 再调用 `gatewaylink`。
  - 该准则保证 Session 校验、频控、日志注入全部集中在 PlayerActor，维持 Actor 单线程语义并避免跨 Actor 直接操作网络。
  - **自动审计检查（2025-12-04 新增）**：
    - ✅ 已提供 `scripts/check_gatewaylink_imports.sh`（Linux/Mac）和 `scripts/check_gatewaylink_imports.ps1`（Windows）脚本，自动检测违规的 `gatewaylink` 直接引用。
    - ✅ 白名单路径：`app/playeractor/controller/player_network_controller.go`、`app/playeractor/entity/player_role.go`、`adapter/gateway/*`、`infrastructure/gatewaylink/*`、`app/engine/server.go`、`infrastructure/dungeonserverlink/dungeon_cli.go`。
    - ✅ 已集成到 `scripts/ci_check.sh` 和 `scripts/ci_check.ps1` 中，CI 流程中应集成此检查，发现违规直接失败；本地开发建议在 pre-commit hook 中运行。
    - 若确实需要直接使用 `gatewaylink`（如基础设施层），需在文档中登记并更新白名单，避免后续重复讨论。

**接入与权限补充约束**
- Gateway WebSocket 在生产环境必须开启 IP 白名单与 Origin 校验：`WSServerConfig.AllowedIPs` 需配置为可信网段，`CheckOrigin` 禁止返回常量 true，应校验域名/协议与预期前端一致  
- 所有 GM 协议（`C2SGMCommand`）必须在协议入口做权限校验：仅允许具备 GM 标记的账号/角色调用，必要时叠加来源 IP、环境变量或临时令牌校验  
- 针对高危 GM 指令（发放货币/道具、踢封玩家等）必须输出结构化审计日志或写入审计表，日志字段需至少包含：操作者账号/角色、目标账号/角色、指令名、参数、时间与来源 IP

### 7.8 RPC 与上下文使用

- 所有 PlayerActor ↔ DungeonActor 调用一律通过 `DungeonServerGateway` 适配层完成，UseCase 层统一通过该接口访问 DungeonActor 能力；PlayerActor → DungeonActor 通过 `gshare.SendDungeonMessageAsync` 发送内部 Actor 消息，DungeonActor → PlayerActor 通过 `gshare.SendMessageAsync` 发送内部 Actor 消息  
- 客户端协议转发由 `ProtocolRouterController` 托管，PlayerActor 消息与 EnterGame/QueryRank 入口集中在 `adapter/controller/register_all.go` + `player_network_controller.go`；任何新协议转发逻辑必须集中在 Controller 层，通过 `gshare.SendDungeonMessageAsync` 转发到 DungeonActor  
- 在 Actor 消息发送场景中禁止直接使用 `context.Background()` 发起长链路调用，上线前需统一改为携带超时的 `context.WithTimeout` 或服务级别的请求上下文  
- 对于 fire-and-forget 类通知可以继续使用带超时的短期上下文，但必须保证底层 TCP 客户端在失败时不会无限重试或阻塞 Actor 主线程  
- 新增 RPC 时需在本节登记调用方/被调方、上下文策略（带/不带超时）、失败重试与降级策略

### 7.5 Phase 3 特殊决策（结合 Clean Architecture 后的现状）

- **PublicActor + PlayerActor 协作** 仍然是所有社交经济功能的唯一方案，差异仅在于 Player 侧从 EntitySystem 迁移为 Clean Architecture 分层实现。  
- 排行榜仅存 key/value；展示数据来自离线快照，缺失时通过 `OfflineDataManager` 异步补全。  
- **社交系统实现位置**：好友/公会/聊天/拍卖行等社交系统现统一按 Clean Architecture 落地在 `domain` + `usecase` + `adapter/{controller,system,presenter}`，不再在 `playeractor/entitysystem` 下新增 `*_sys.go`；原 `SysFriend/SysGuild/SysAuction` 等系统枚举已从 `protocol.SystemId` 移除。  
- 公会/拍卖行仍必须持久化并在 GameServer 启动时加载至 PublicActor，对应逻辑位于 `server/internal/database/{guild,auction_item}.go` 与 `publicactor` 目录。  
- **PublicActor 消息处理**：所有消息处理器函数签名继续遵循 `actor.HandlerMessageFunc`，新增消息必须在 `publicactor/register.go` 中集中注册。  
- **在线状态管理**：玩家上线时通过 `RegisterOnlineMsg` 注册到 PublicActor，下线时通过 `UnregisterOnlineMsg` 注销，`PlayerRole` 统一通过 `sendPublicActorMessage` 封装发送逻辑。  
- **PublicActor 内部消息 ID 规范**：所有 PublicActor 内部消息 ID 必须统一使用 Proto 中的 `PublicActorMsgId` 枚举定义，新类型需同步更新文档与枚举注释。  

### 7.6 时间同步与客户端动画职责

- 所有服务一律通过 `server/internal/servertime` 读取权威 UTC 时间，禁止直接调用 `time.Now()`
- 服务器时间广播由 GameServer 的 `timesync.Broadcaster` 统一承担
- 技能/动作动画完全由客户端驱动：服务端只计算技能释放成功与伤害批次并广播，客户端根据协议结果自行播放

### 7.7 技能配置校验

- 所有技能必须至少包含一个 `skillCfg.Effects` 条目；若为空，`skill.Skill.Use` 会直接返回 `ErrSkillCannotCast`

### 7.8 调试客户端约束

- `server/example` 视为服务端代码，必须使用 `servertime` 读取时间，禁止 `time.Now()` 直接调用
- 文字冒险面板命令需与现有协议一一映射，禁止在 `GameClient` 中硬编码临时逻辑
- 所有扩展设计、命令列表、待办事项统一记录在 `docs/golang客户端待开发文档.md`
- 扩展命令前优先复用 `AdventurePanel` 解析与 `GameClient` 能力，禁止创建平行脚本
- server/example 任意重构与新增命令必须遵循《`docs/server_example重构方案.md`》，并保持移动模拟
- server/example 现已使用 `cmd/example` 入口与 `internal/{client,panel,systems}` 分层，新增能力需按模块拆分；协议等待/回调统一复用 `internal/client/flow.go`
- 自动寻路、脚本与 AI 必须通过 `systems.Move`/`MoveRunner` 实现，禁止绕过统一容错逻辑；背包/GM/副本等能力需复用 `systems.Inventory/GM/Dungeon` 读写协议

### 7.9 系统管理器约束（2025-12-04 新增）

- **系统注册**：所有系统工厂通过 `entitysystem.RegisterSystemFactory` 在模块级注册（`*_system_adapter_init.go` 的 `init()` 函数中），不再使用全局 `globalFactories`。
- **系统初始化**：`SysMgr` 使用显式配置的系统列表初始化，不再遍历枚举范围；`NewSysMgr()` 默认使用 `GetDefaultSystemIds()` 返回的系统列表，可通过 `NewSysMgrWithSystems(systemIds)` 自定义。
- **调试支持**：使用 `SysMgr.ListMountedSystems()` 获取已挂载系统的信息（SysId、Opened、HasImpl），便于 UseCase/Controller 定位问题。
- **新增系统**：新增系统时只需在对应 `*_system_adapter_init.go` 中注册工厂，并在 `GetDefaultSystemIds()` 中添加系统ID（或通过配置/环境变量控制），无需修改枚举范围。

### 7.10 离线数据管理器约束（新）

- PublicActor 是唯一的离线快照写入口，`OfflineDataManager` 内部状态禁止在 PlayerActor 中直接访问。
- 所有离线数据时间戳、定时器均需使用 `servertime`；禁止 `time.Now()`。
- 新增离线数据类型必须通过注册表声明 `data_type`、序列化函数与落库策略，并在 `docs/离线数据管理器开发文档.md` 记录。
- GameServer 启动时必须加载离线数据（或按需懒加载），并在 `rank` 查询等路径上允许离线玩家返回快照。
- 持久化失败需要保留 `dirty` 状态并监控日志，避免吞掉落库失败。
- PlayerActor 必须通过 `PublicActorMsgIdUpdateOfflineData`（`UpdateOfflineDataMsg`）上报最新快照；若消息发送失败需在 Actor 端重试。

### 7.10 Clean Architecture 架构决策（新）

**SystemAdapter 职责边界（阶段 C 完成）**：
- SystemAdapter 只负责生命周期适配、事件订阅和框架状态管理
- 所有业务逻辑（包括定时/调度逻辑）应在 UseCase 层实现
- 事件订阅：SystemAdapter 层只负责"订阅哪个事件"，业务逻辑由 UseCase 层处理
- 生命周期方法：应在头部注释中明确说明每个生命周期阶段"仅做哪些调度行为"和"具体业务由哪些 UseCase 承担"
- 参考文档：`docs/gameserver_adapter_system演进规划.md` 第 4.3 节

**Controller 层系统开启检查（阶段八）**：
- Controller 层负责框架层面的检查，包括系统开启/关闭状态检查
- 在调用 UseCase 之前，Controller 应先通过 SystemAdapter helper 函数（如 `GetBagSys(ctx)`）检查系统是否开启
- 如果系统未开启，Controller 应直接返回错误，不执行 UseCase
- UseCase 层保持纯业务逻辑，不应该感知系统开启/关闭状态（这是框架层面的职责）
- 参考文档：`docs/SystemAdapter系统开启检查优化方案.md`

**SystemAdapter 防退化机制（阶段 E）**：
- BaseSystemAdapter 和所有 SystemAdapter 头部注释中明确标注"禁止编写业务规则逻辑，只允许调用 UseCase 与管理生命周期"
- Code Review 时必须检查 SystemAdapter 是否含有可下沉到 UseCase 的逻辑
- 所有业务逻辑必须在 UseCase 层实现，SystemAdapter 只负责"何时调用哪个 UseCase"的调度
- 参考文档：`docs/SystemAdapter_CodeReview清单.md`

### 7.11 Clean Architecture 架构决策（新）

- **架构原则**：GameServer（包含 DungeonActor）将按照 Clean Architecture（清洁架构）原则进行重构，实现业务逻辑与框架解耦。
- **分层结构**：
  - **Entities 层**：纯业务实体和值对象，不依赖任何框架
  - **Use Cases 层**：业务用例和业务规则，依赖 Entities 和 Repository 接口
  - **Interface Adapters 层**：协议处理（Controllers）、响应构建（Presenters）、数据访问（Gateways）、RPC 适配（RPC Adapters）
  - **Frameworks & Drivers 层**：Actor 框架、网络、数据库、配置管理器等
- **依赖规则**：
  - 内层定义接口，外层实现接口（依赖倒置原则）
  - 业务逻辑（Use Cases）禁止直接依赖框架层
  - 所有框架调用必须通过 Adapter 层封装
- **文件组织原则**（新增）：
  - Clean Architecture 主要关注依赖方向（内层不依赖外层），而不是文件组织结构
  - 建议将 `adapter/system` 下的文件按系统分包（如 `adapter/system/level/`、`adapter/system/bag/`）
  - 只要保持依赖方向正确，文件组织可以灵活调整
  - 分包方式更清晰，便于维护，不会影响 Clean Architecture 原则
- **重构策略**：
  - 采用渐进式重构，新旧代码可以并存
  - 优先重构核心系统（如 MoveSys、FightSys、AttrSys）
  - 每个系统迁移后必须编写单元测试
- **测试要求**：
  - Use Case 层必须可独立测试（通过 Mock Repository 和接口）
  - Controller 层编写集成测试
  - 单元测试覆盖率目标 > 70%
- **参考文档**：
  - Gateway 重构：详见《`docs/gateway_CleanArchitecture重构文档.md`》
  - GameServer 重构：详见《`docs/gameserver_CleanArchitecture重构文档.md`》
- **注意事项**：
  - 重构过程中保持向后兼容，不破坏现有功能
  - 保持 Actor 框架的单线程特性
  - 避免过度抽象导致性能下降
  - DungeonActor 作为实时战斗引擎，需特别注意性能影响
- FriendSys 迁移后，`C2SAddFriend/Respond/Query` 必须通过 Use Case → `PublicActorGateway` 异步转发，禁止在 Controller 中直接调用 `gshare.SendPublicMessageAsync`
- Friend/Blacklist 数据统一存储在 `PlayerRoleBinaryData.FriendData` 与 `BlacklistRepository`，任何模块需要访问必须依赖对应 UseCase 或接口，不得直接操作数据库
- PlayerActor 层统一使用 `PlayerRole.sendPublicActorMessage`（内部依赖 `PublicActorGateway`）发送所有 PublicActor 消息，禁止直接调用 `gshare.SendPublicMessageAsync`
- RecycleSys 仅通过 `adapter/controller/recycle_controller.go` → `adapter/system/recycle` 处理协议与业务逻辑，禁止重新引入 `entitysystem/recycle_sys.go` 等 legacy 入口

### 7.11 FriendSys 迁移注意事项（新）
- Friend 列表查询是异步链路：Controller 只负责触发 Use Case，实际 `S2CFriendList` 由 PublicActor 汇总快照后下发
- 黑名单增删查询统一通过 `usecase/interfaces/blacklist.go` + `adapter/gateway/blacklist_repository.go`，保持 Clean Architecture 依赖方向
- 旧 `entitysystem/friend_sys.go` 已删除，如需 legacy 行为请改为调用 `adapter/system/friend` 暴露的 `GetFriendSys(ctx)`

### 7.12 Clean Architecture 开发指南（2025-01-XX 更新）
1. **域分析**：在 `domain/` 中定义/复用实体，确认状态存放在 `PlayerRoleBinaryData` 或 PublicActor。若需新字段，先更新 proto + 数据迁移。
2. **Use Case 设计**：在 `usecase/{system}` 下创建用例，并通过 `usecase/interfaces` 声明依赖（Gateway、Repository、Presenter Adapter 等）。所有数据库/网关访问必须通过接口。
3. **SystemAdapter 角色**：`adapter/system/{system}` 仅负责 PlayerActor 生命周期（Init/Login/RunOne）与 UseCase 编排；禁止在其中写协议解析或直接操作网络。
4. **Controller/Presenter**：Controller 解析协议 → 调用 UseCase；Presenter 统一封装 `S2C` 响应与错误文案。每个系统必须提供 `*_controller.go`、`*_presenter.go`、`*_controller_init.go` 三件套。
5. **依赖检查**：提交前确认 `go list ./...` 无循环依赖；新增 Gateway/Adapter 文件需在文档“关键代码位置”章节登记，确保他人可以快速定位。
6. **文档同步**：完成功能后必须同步两个文档：`docs/服务端开发进度文档.md`（已完成功能/注意事项/关键代码）与 `docs/gameserver重构待实现功能清单.md`（待办列表）。

---

## 8. 关键代码位置

### 8.1 Gateway
- `server/service/gateway/internel/clientnet`：Session 与消息适配
- `server/service/gateway/internel/engine`：Gateway 配置解析及 TCP/WS 启停
- `docs/gateway_CleanArchitecture重构文档.md`：Gateway Clean Architecture 重构方案文档，包含详细的重构步骤和检查清单

### 8.2 GameServer - 玩家 Actor
- `server/service/gameserver/internel/app/playeractor`：玩家 Actor Handler / 协议注册入口 / `PlayerRole` 实体
- `server/service/gameserver/internel/app/manager/role_mgr.go`：玩家角色管理器，使用 `sessionIndex map[string]uint64` 实现 O(1) 的 SessionId 查找，支持分批保存（`FlushAndSave` 支持 `batchSize` 参数），已注册到 DI 容器（2025-12-04 优化）
- `server/service/gameserver/internel/app/playeractor/deps`：PlayerActor 依赖装配入口（取代 `di/`），暴露 Gateway/Repository/EventPublisher 等依赖以及兼容性的 Container 视图；Network/Session Gateway 合并为 `adapter/gateway/client_gateway.go`，统一由 `deps.ClientGateway()` 提供
- `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`：系统管理器，使用显式配置的系统列表初始化（不再遍历枚举范围），支持 `NewSysMgrWithSystems(systemIds)` 自定义系统列表，提供 `ListMountedSystems()` 调试函数
- `server/service/gameserver/internel/app/playeractor/entitysystem/system_registry.go`：系统注册表（替代全局 `globalFactories`），使用读写锁保护，提供模块级注册接口
- `server/service/gameserver/internel/app/playeractor/entitysystem/message_dispatcher.go`：离线消息分发入口，配合 `adapter/system/message_system_adapter.go` 与 `engine/message_registry.go` 完成玩家消息回放
- `server/service/gameserver/internel/app/playeractor/controller/player_network_controller.go`：客户端协议与 PlayerActor 消息入口；所有注册集中于 `adapter/controller/register_all.go`
- `server/service/gameserver/internel/app/playeractor/entity/player_role.go`：PlayerRole 主体逻辑与 `sendPublicActorMessage` 封装，统一通过 `PublicActorGateway` 发送消息
- `server/service/gameserver/internel/app/playeractor/bag/*`：Bag 特性分片包（Controller/SystemAdapter/Presenter/UseCase/适配器），提供 `GetBagSys`、`NewBagUseCaseAdapter` 供跨系统调用。
- `server/service/gameserver/internel/app/playeractor/money/*`：Money 特性分片包（Controller/SystemAdapter/Presenter/UseCase/适配器），使用 `money.Deps` 聚合依赖，提供 `GetMoneySys` 供跨系统调用。
- `server/service/gameserver/internel/adapter/controller/register_all.go`：所有 GameServer 控制器的协议/RPC 注册入口，SystemAdapter 不再直接依赖 controller 包
- `server/service/gameserver/internel/adapter/controller/protocol_router_controller.go`：协议路由控制器，负责解析 C2S 消息、注入上下文并通过 `gshare.SendDungeonMessageAsync` 转发到 DungeonActor
- `server/service/gameserver/internel/adapter/controller/friend_controller.go`：好友系统协议入口，负责发送/响应好友申请、查询好友/黑名单
- `server/service/gameserver/internel/adapter/controller/guild_controller.go`：公会系统协议入口（创建/加入/退出/查询）
- `server/service/gameserver/internel/adapter/controller/chat_controller.go`：聊天系统协议入口（世界/私聊，带冷却与敏感词过滤）
- `server/service/gameserver/internel/adapter/controller/auction_controller.go`：拍卖行协议入口（上架/购买/查询）
- `server/service/gameserver/internel/adapter/controller/recycle_controller.go`：回收协议入口，负责解析 `C2SRecycleItem` 并驱动回收用例
- `server/service/gameserver/internel/adapter/system/recycle/`：回收系统适配层，封装回收用例、配置访问与 RPC 注册
- `server/service/gameserver/internel/adapter/system/message_system_adapter.go`：MessageSys 适配层，负责登录/重连/初始化时加载离线消息、RunOne 时进行数量限制、OnNewDay 清理过期消息
- `server/service/gameserver/internel/app/playeractor/entitysystem/message_dispatcher.go`：统一的离线消息分发入口，结合 `engine/message_registry.go` 完成消息回调
- `server/service/gameserver/internel/gshare/log_helper.go`：日志上下文辅助（自动输出 Session/Role 信息）
- `server/service/gameserver/internel/core/gshare/context_helper.go`：Context 取值工具（Session/Role），取代 `adapter/context` 包

### 8.3 GameServer - PublicActor
- `server/service/gameserver/internel/publicactor/adapter.go`：PublicActor 适配器，单 Actor 模式
- `server/service/gameserver/internel/publicactor/handler.go`：PublicActor 消息处理器
- `server/service/gameserver/internel/publicactor/public_role.go`：公共角色数据管理（在线状态、排行榜、公会、拍卖行）
- `server/service/gameserver/internel/publicactor/message_handler.go`：PublicActor 消息处理函数
- `server/service/gameserver/internel/publicactor/public_role_offline_data.go`：离线数据加载、消息入口与周期 Flush
- `server/service/gameserver/internel/publicactor/offlinedata/*`：OfflineDataManager（内存缓存、Load/Update/FlushDirty）
- `docs/离线数据管理器开发文档.md`：离线快照/排行榜持久化重构方案与实施步骤

### 8.4 GameServer - 其他
- `server/service/gameserver/internel/adapter/gateway/dungeon_server_gateway.go`：DungeonServerGateway 实现，封装 `AsyncCall/ RegisterRPCHandler / RegisterProtocols`
- `server/service/gameserver/internel/adapter/gateway/blacklist_repository.go`：黑名单仓储适配器
- `server/service/gameserver/internel/adapter/presenter/{friend,guild,chat,auction}_presenter.go`：社交系统响应构建器
- `server/service/gameserver/internel/usecase/{friend,guild,chat,auction}/`：社交系统业务用例
- `server/service/gameserver/internel/adapter/system/{friend,guild,chat,auction}/`：Clean Architecture System Adapter（生命周期 + Helper + init）
- `server/service/gameserver/internel/dungeonserverlink`：DungeonServer RPC 客户端
  - `dungeon_cli.go`：`AsyncCall` 对 `msgId=0` 的调用封装 `MsgTypeClient` 透传，保持 sessionId
- `server/service/gameserver/internel/gatewaylink`：与 Gateway 的 Session 映射与消息转发
- `server/service/gameserver/internel/timesync`：服务器时间广播器
- `server/service/gameserver/internel/gshare/srv.go`：平台/区服常量、开服时间、`GetOpenSrvDay`
- `server/service/gameserver/internel/manager/role_mgr.go`：玩家角色管理、关服刷盘 `FlushAndSave`
- `server/service/gameserver/internel/engine/message_registry.go`：玩家消息回调注册中心
- `server/service/gameserver/internel/gshare/message_sender.go`：玩家消息发送入口（在线 Actor 投递 + 离线回退）
- `proto/csproto/rpc.proto`：`AddActorMessageMsg` 内部消息定义
 - `docs/gameserver_目录与调用关系说明.md`：GameServer 目录结构与调用关系说明文档，从分层、依赖方向和调用链路三个角度解释其如何满足 Clean Architecture 规范
- `docs/gameserver_CleanArchitecture重构实施手册.md`：GameServer Clean Architecture 重构执行手册，提供阶段路线、子系统任务清单、跨模块约束与测试/验收标准

### 8.5 DungeonActor
- `server/service/gameserver/internel/app/dungeonactor/adapter.go`：DungeonActor 适配器，实现 `IDungeonActorFacade` 接口
- `server/service/gameserver/internel/app/dungeonactor/handler.go`：DungeonActor 消息处理器，实现 `Loop` 方法驱动实体和副本管理
- `server/service/gameserver/internel/app/dungeonactor/register.go`：消息处理器注册，按业务模块拆分（`RegisterMoveHandlers`、`RegisterFightHandlers`、`RegisterFuBenHandlers`）
- `server/service/gameserver/internel/app/dungeonactor/entitysystem/*`：AI、Buff、Attr、Move、Fight、StateMachine、AOI
  - `entitysystem/pathfinding.go`：寻路算法实现（A*、直线寻路），`FindPath` 提供统一入口；所有坐标参数和返回值都是格子坐标；直线寻路在遇到障碍时自动绕过，优先保持直线方向并选择最接近目标的方向
  - `entitysystem/move_sys.go`：移动系统，专注于移动功能，不包含AI业务逻辑
    - `HandleStartMove`：处理 C2SStartMove，记录 move_data（包含目的地），广播 S2CStartMove
    - `HandleUpdateMove`：处理 C2SUpdateMove，判定坐标误差（支持1s误差），如果差距大就结束移动，否则更新坐标
    - `HandleEndMove`：处理 C2SEndMove，广播 S2CEndMove，通知客户端结束移动；移动结束后发送 `PlayerActorMsgIdSyncPosition` 消息给 PlayerActor
    - `HandMove`：处理客户端移动请求的核心逻辑
    - `LocationUpdate`：客户端位置更新校验，防止移动过快或瞬移
    - `MovingTime`：服务端时间驱动移动，根据时间计算当前位置
  - `entitysystem/fight_sys.go`：战斗系统，管理技能释放、伤害结算、Buff 应用
  - `entitysystem/ai_sys.go`：AI系统，通过组合调用移动协议方法（`HandleStartMove` → `HandleUpdateMove` → `HandleEndMove`）实现移动，模拟客户端行为；`moveTowardsWithPathfinding` 根据配置选择寻路算法并管理路径缓存；`distanceBetween` 函数计算格子距离
  - `entitysystem/attr_sys.go`：属性系统，使用 `attrcalc.AttrSet`+`ResetProperty`，广播 `S2CAttrData`（Dungeon即时属性）并复用 `attr_config.json.power` 战力公式；支持 `ApplySyncData` 接收 GameServer 属性，`ResetSysAttr` 支持本地系统属性计算（怪物、Buff 等）；`RunOne` 每帧调用 `ResetProperty` 和 `CheckAndSyncProp`；`extraUpdateMask` 跟踪非战斗属性变化，`bInitFinish` 控制是否广播属性；属性变化时发送 `PlayerActorMsgIdSyncAttrs` 消息给 PlayerActor
  - `entitysystem/attrcalc/bus.go`：属性计算器注册管理器，提供 `RegIncAttrCalcFn/RegDecAttrCalcFn` 和 `GetIncAttrCalcFn/GetDecAttrCalcFn`，支持注册怪物基础属性计算器（`MonsterBaseProperty`）、Buff 属性计算器（`SaBuff`）等
  - `entitysystem/drop_sys.go`：掉落系统，拾取物品后发送 `PlayerActorMsgIdAddItem` 消息给 PlayerActor
  - `entity/monster.go`：怪物实体，在 `NewMonsterEntity` 中调用 `ResetProperty()` 触发完整属性计算流程；`MonsterEntity.ResetProperty()` 方法先调用 `ResetSysAttr(MonsterBaseProperty)` 计算基础属性，再调用 `AttrSys.ResetProperty()` 触发属性汇总、转换、百分比加成和广播；`monsterBaseProperty` 函数从怪物配置表读取属性并写入 `AttrSet`
  - `entitysystem/buff_sys.go`：Buff 系统，在 `AddBuff/RemoveBuff/ClearAllBuffs` 时触发 `ResetSysAttr(SaBuff)`；`buffAttrCalc` 函数遍历所有 Buff 汇总属性加成
- `server/service/gameserver/internel/app/dungeonactor/fuben` & `fbmgr`：副本实例与结算逻辑
  - `fuben/settlement.go`：副本结算时发送 `PlayerActorMsgIdSettleDungeon` 消息给 PlayerActor
  - `fuben/actor_msg.go`：进入副本成功后发送 `PlayerActorMsgIdEnterDungeonSuccess` 消息给 PlayerActor
- `server/service/gameserver/internel/app/dungeonactor/scene/scenest.go`：`GameMap` 绑定、出生点随机校验（`BornArea`）、移动校验；所有位置相关函数统一使用格子坐标
- `server/service/gameserver/internel/app/dungeonactor/skill/skill.go`：技能目标筛选、配置校验、结果填充；距离计算和范围判断统一使用格子距离

### 8.6 数据库
- `server/internal/database`：账号/角色/Token 访问层
- `server/internal/database/server_info.go`：平台/区服开服信息表
- `server/internal/database/guild.go`：公会数据持久化
- `server/internal/database/auction_item.go`：拍卖行数据持久化
- `server/internal/database/offline_message.go`：离线消息持久化
- `server/internal/database/offline_data.go`：离线数据表 `OfflineData` 定义及 Upsert/查询
- `server/internal/database/player_actor_message.go`：玩家消息表 DAO（增量加载、单条/批量删除、计数）
- `server/internal/database/transaction_audit.go`：交易审计
- `server/internal/database/blacklist.go`：黑名单管理

### 8.7 CI 检查脚本（2025-12-04 新增）
- `scripts/ci_check.sh` 和 `scripts/ci_check.ps1`：CI 检查主脚本，集成 `go vet`、`staticcheck` 和 gatewaylink 导入检查
  - 支持 Windows (PowerShell) 和 Linux/Mac (Bash) 环境
  - 对 `server/service/gameserver` 目录执行静态分析检查
- `scripts/check_gatewaylink_imports.sh` 和 `scripts/check_gatewaylink_imports.ps1`：检查 gatewaylink 导入规则
  - 禁止除白名单外的包引用 gatewaylink
- 白名单包括：PlayerActor 相关文件（`adapter/controller/player_network_controller.go`, `player_role.go`）、Gateway 适配器、gatewaylink 自身、engine/server.go、dungeonserverlink
  - 已集成到 CI 检查脚本中

### 8.8 共享基础
- `server/internal/actor`：Actor 框架
- `server/internal/network`：消息编解码、压缩
- `server/internal/network/codec.go` & `message.go`：前向消息池化、缓冲复用
- `server/internal/servertime`：统一时间源
- `server/internal/event`：事件总线
- `server/internal/jsonconf`：配置加载与缓存
  - `monster_config.go`：怪物配置，`MonsterAIConfig` 包含 `PatrolPathfinding`/`ChasePathfinding` 寻路算法配置
  - `map_config.go`：`map_config.json` → `GameMap` 转换、场景绑定
    - `CoordToIndex` / `IndexToCoord`：坐标与一维数组索引互转，供移动校验/寻路调用
- `server/internal/argsdef`：参数定义与工具函数
  - `position.go`：坐标系统定义与转换（`TileSize`、`TileCenterOffset`、`TileCoordToPixel`、`PixelCoordToTile`、`IsSameTile`）
  - `gridst.go`：AOI格子系统（`GrIdSt`、`GrIdSize`、`GetGrIdSt`、`GetNineGrIds`），注意：AOI格子大小（100）与游戏坐标格子大小（128）不同
- `server/internal/attrcalc`：属性计算工具包（待完善）
  - `attrcalc.go`：战斗属性计算器（`CombatAttrCalc`）
  - `extraattrcalc.go`：非战斗属性计算器（`ExtraAttrCalc`）
  - 待实现：`attr_set.go`（属性集合管理，参考 `server/server/base/attrcalc/attr_set.go`）
- `server/internal/attrdef`：属性类型定义（`attrdef.go`）
- `server/pkg/log`：统一日志组件（级别控制、结构化字段、文件轮转）
  - `logger.go` / `requester.go`：`IRequester` 接口与 `*_WithRequester` 全局函数，用于注入 Session/Role 前缀与自定义栈深；GameServer `gshare/log_helper.go` 通过该能力补充上下文
- `proto/csproto/*.proto`：协议定义
- `server/internal/protocol/*.pb.go`：协议生成结果

### 8.8 调试客户端（server/example）
- `server/example/cmd/example/main.go`：调试客户端入口，初始化日志/配置并启动面板
- `server/example/internal/client`：`Core`（协议收发、状态、Flow Waiter、MoveRunner）、`Manager`、`ClientHandler`
- `server/example/internal/panel`：`AdventurePanel` UI、命令解析与脚本触发
- `server/example/internal/systems`：`Account/Scene/Move/Combat/Inventory/Dungeon/GM/Script` 系统封装
- `docs/golang客户端待开发文档.md`：调试客户端规划、命令映射、待办
- `docs/server_example重构方案.md`：server/example 重构与移动对齐指南，分阶段任务与验收标准

### 8.9 参考项目（server/server）
- `server/server/gameserver/logicworker/actorsystem/attr_sys.go`：GameServer 属性系统参考实现（系统属性存储、加成属性计算、战力计算、属性推送）
- `server/server/fightsrv/entitysystem/attr_sys.go`：FightServer 属性系统参考实现（AttrSet 结构、属性重置计算、属性同步处理、属性广播）
- `server/server/base/attrcalc/`：属性计算器参考实现（`fight_attr.go`、`extra_attr.go`、`attr_set.go`）
- `docs/属性系统重构文档.md`：属性系统重构方案文档，包含详细的重构步骤和检查清单
- `docs/属性系统阶段三联调记录.md`：阶段三端到端自测流程记录

---

## 9. 核心运行流程（摘要）

1. **登录**：Client → Gateway → GameServer (`C2SRegister/C2SLogin`) → `database.Account` 校验 → Token + Session 扩展
2. **角色管理**：`C2SQueryRoles/C2SCreateRole/C2SEnterGame` → 加载 `PlayerRoleBinaryData` → 初始化系统 → `S2CLoginSuccess`
3. **进入副本**：GameServer Controller 通过 `gshare.SendDungeonMessageAsync` 发送 `DungeonActorMsgIdEnterDungeon` 消息到 DungeonActor，DungeonActor 创建实体并发送 `PlayerActorMsgIdEnterDungeonSuccess` 消息回 PlayerActor
4. **移动与战斗**：客户端发送 `C2SStart/Update/EndMove` 和 `C2SUseSkill`，GameServer Controller 通过 `gshare.SendDungeonMessageAsync` 转发到 DungeonActor；DungeonActor 校验并广播，必要信息通过 `PlayerActorMsgId` 消息回传 PlayerActor
5. **掉落拾取**：DungeonActor 判定归属/距离，拾取成功后发送 `PlayerActorMsgIdAddItem` 消息给 PlayerActor，PlayerActor 处理背包逻辑
6. **副本结算**：DungeonActor 发送 `PlayerActorMsgIdSettleDungeon` 消息 → PlayerActor `FubenSys` 更新记录、发奖励
7. **断线重连**：Gateway 关闭连接 → GameServer 接收 `SessionEventClose`，标记离线并允许在 `ReconnectKey` 有效期内重连

---

## 10. 版本记录

| 日期 | 内容 |
| ---- | ---- |
| 2025-12-17 | **PlayerActor 目录扁平化（更 Go 风格）**：将 `playeractor/adapter/*` 与 `playeractor/usecase/*` 整合为 `controller/`、`system/`、`service/`、`gateway/`、`presenter/`、`router/`、`event/`，减少层级与跨包跳转成本；`cd server && go test ./...` 通过。 |
| 2025-12-10 | **PlayerActor 瘦身 P0**：删除 `di/` 与旧 context 适配层，新增 `playeractor/deps` 聚合依赖，合并 Network/Session Gateway 为 `ClientGateway`，Context 取值收敛到 `core/gshare/context_helper.go`；通用消耗/奖励收敛到业务层（`playeractor/service/consume`、`playeractor/service/reward`），删除旧 `usecaseadapter`。 |
| 2025-12-10 | **Controller 注册收敛 & 事件总线按 Actor 注入**：各 Controller 在 `init()` 中注册 C2S/PlayerActor 消息，`player_network_controller.go` 承接 EnterGame/QueryRank/PlayerActorMsg；玩家事件总线改为 `gevent.NewPlayerEventBus()` 按 Actor 构造；升级奖励改用 Reward 业务用例（去除 PlayerRole 直发奖），删除 PlayerRole 直接扣耗/发奖接口（依赖 Consume/Reward 业务用例）。 |
| 2025-12-17 | **PlayerActor Bag 特性分片（Phase2A）**：Bag 控制器 / SystemAdapter / Presenter / 用例收拢到 `playeractor/bag`，提供 `NewBagUseCaseAdapter` 供跨系统使用，SystemAdapter 使用本地基类避免循环依赖，保持 init 注册与外部接口不变，`go test ./...` 通过。 |
| 2025-12-17 | **PlayerActor Money 特性分片+过程化（Phase2A/2B）**：Money 控制器 / SystemAdapter / Presenter / 用例收拢到 `playeractor/money`，用 `money.Deps` 聚合依赖，小 service 化用例，SystemAdapter 使用本地基类避免循环依赖；Consume/Reward/Level 等依赖已指向新包，`go test ./...` 通过。 |
| 2025-12-03 | **兼容代码清理阶段一**：按《`docs/gameserver_兼容代码清理规划.md`》2.1 小节执行，物理删除 `server/service/gameserver/internel/domain/vip` 与 `server/service/gameserver/internel/domain/dailyactivity` 两个空壳领域目录，仅清理历史残留目录，不改动任何 Go 代码；当前本地环境尚未初始化 Go module，未执行完整 `go build`，但通过文件搜索确认无引用。 |
| 2025-01-XX | **系统依赖关系清理**：完成 SysRank 和已移除系统的依赖关系清理；已在 `proto/csproto/system.proto` 中为 `SysRank = 19` 添加注释说明（RankSys 是 PublicActor 功能，不参与 PlayerActor 系统管理）；确认 `sys_mgr.go` 不再使用 `systemDependencies`，改为按 SystemId 顺序初始化；确认 proto 中不包含已移除的系统ID（VipSys、DailyActivitySys、FriendSys、GuildSys、AuctionSys） |
| 2025-01-XX | **MessageSys 功能完善 & 文档更新**：补充 Clean Architecture 分层说明、Controller/Presenter 开发指南、协议注册规范；记录 MessageSys 关键代码位置与运行机制（OnInit/OnRoleLogin/OnNewDay/RunOne）并在 3.1/7.3/8.2 章节同步；版本记录新增此条目，方便后续追溯 |
| 2025-11-26 | **server/example Phase B & C**：引入 `MoveRunner` + `move-to/move-resume`、`bag/use-item/pickup/gm/enter-dungeon` 命令与 `script-record/script-run` 录制回放；`systems` 包补全 Inventory/Dungeon/GM/Script，面板全量经由系统接口调用 |
| 2025-12-XX | **移动系统代码简化**：移除 `MoveSys` 中的AI相关业务逻辑（`TickAutoMove`、`MoveTo`、`MoveToWithPathfinding`等），保持移动系统代码简洁干净；AI系统通过组合调用移动协议方法（`HandleStartMove` → `HandleUpdateMove` → `HandleEndMove`）实现移动，模拟客户端行为 |
| 2025-11-25 | **DungeonServer 移动系统重构**：重新实现移动系统，添加服务端时间驱动移动（`MovingTime`）和客户端位置校验（`LocationUpdate`），统一坐标系统（内部使用像素坐标，与场景交互时转换为格子坐标），支持速度容差和网络延迟容忍度 |
| 2025-11-25 | **DungeonServer 移动协议流程优化**：优化移动协议处理流程，实现 `HandleStartMove`（记录 move_data 并广播 S2CStartMove）、`HandleUpdateMove`（判定坐标误差，支持1s误差，差距大时结束移动）、`HandleEndMove`（广播 S2CEndMove）；在 proto 中定义 MoveData 和 S2CStartMove、S2CEndMove 协议 |
| 2025-11-25 | **DungeonServer 移动重构**：`MoveSys` 改为仅由 `AISys.TickAutoMove` 驱动（玩家不再调用 RunOne），自动移动失败日志输出格子坐标；`server/example` 的 `move` 命令严格按照 `C2SStart/C2SUpdate/C2SEnd` 顺序逐格上报像素坐标，便于复现客户端移动链路 |
| 2025-11-24 | **DungeonServer 协议注册重构**：参考 GameServer 的协议注册方式，将 DungeonServer 的协议注册归到具体业务系统中；将 `clientprotocol/skill_handler.go` 的 `handleUseSkill` 移至 `entitysystem/fight_sys.go`，使用 `devent.Subscribe(OnSrvStart)` 在服务器启动时注册协议；更新开发注意事项，明确协议注册规范 |
| 2025-11-24 | **坐标系统规范统一**：统一规范为客户端只发送像素坐标，服务端统一转换为格子坐标进行业务处理；更新协议定义、代码注释和开发注意事项，明确坐标系统规范 |
| 2025-11-24 | **坐标系统优化（完整）**：完成坐标系统优化建议第2-7项，移动系统、寻路算法、场景系统、客户端、协议定义、距离和范围计算全部优化完成；服务端统一使用格子坐标进行业务处理；已在开发注意事项中明确坐标系统规范 |
| 2025-11-24 | **坐标系统优化（移动/寻路/场景）**：完成坐标系统优化建议第2、3、4项，移动系统实现客户端像素坐标到格子坐标的自动转换，距离计算和速度校验基于格子大小；寻路算法明确输入输出为格子坐标；场景系统统一使用格子坐标并添加详细注释 |
| 2025-11-24 | **坐标系统定义与转换**：完成坐标系统优化建议第1项，在 `server/internal/argsdef/position.go` 中添加格子大小常量（TileSize=128、TileCenterOffset=64）和坐标转换函数（TileCoordToPixel、PixelCoordToTile、IsSameTile），统一了坐标系统的定义和使用规范 |
| 2025-11-24 | 日志管理优化：`server/pkg/log` 新增结构化字段 & 环境级别控制，完成 11.2.9 |
| 2025-11-24 | 优化资源清理：ForwardMessage 池化、Gateway 转发链路释放消息，完成 11.2.7 资源清理检查 |
| 2025-11-21 | **生产环境优化修复**：修复时间访问规范违反（chat_sys.go、session_mgr.go）、添加数据库连接池配置、在 Actor 消息处理中添加 Panic 恢复机制 |
| 2025-11-20 | 重新梳理文档结构，简化已完成功能描述，清理待实现章节，优化文档可读性 |
| 2025-11-21 | 新增 `server` 表存储开服信息，提供 `gshare.GetOpenSrvDay()`；统一数据库时间字段为秒级 Unix 时间戳 |
| 2025-11-20 | 完成 Phase 3 社交经济系统：聊天、好友、排行榜、公会、拍卖行、社交安全系统全部完成 |
| 2025-11-20 | 完成 Phase 3 社交经济基础框架：PublicActor 框架、Proto 定义、EntitySystem 基础 |
| 2025-11-17 | 依据当前代码补全架构说明、已完成功能、关键代码位置与注意事项 |


