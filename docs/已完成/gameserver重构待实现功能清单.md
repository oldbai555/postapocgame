# GameServer Clean Architecture 重构待实现/待完善功能清单

更新时间：2025-01-XX  
责任人：开发团队

> ⚠️ **重要提示**：本文档列出了 GameServer Clean Architecture 重构的所有待实现功能，按优先级排序。请按照阶段逐步实现，每完成一项可勾选对应复选框，便于跟踪进度。

---

## 一、高优先级（影响整体一致性）

### 1.1 系统依赖关系清理

- [✅] **清理 SysRank 依赖关系**
  - **问题**：`sys_mgr.go` 中的 `systemDependencies` 定义了 `SystemId_SysRank` 的依赖关系，但 RankSys 不是 PlayerActor 的系统，而是 PublicActor 的功能
  - **处理**：
    - ✅ `sys_mgr.go` 已不再使用 `systemDependencies`，改为按 SystemId 顺序初始化，因此无需移除依赖关系定义
    - ✅ 已在 `proto/csproto/system.proto` 中为 `SysRank = 19` 添加注释说明：RankSys 是 PublicActor 功能，不参与 PlayerActor 系统管理，此枚举值仅用于标识
    - ✅ 确认没有系统注册 SysRank（SystemId = 19），符合预期
  - **关键代码位置**：
    - `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go:96-99`
    - `server/service/gameserver/internel/app/publicactor/public_role_rank.go`
    - `server/service/gameserver/internel/app/playeractor/entity/player_role.go:172-237`

- [✅] **清理已移除系统的依赖关系**
  - **已移除的系统**：VipSys、DailyActivitySys、FriendSys、GuildSys、AuctionSys
  - **处理**：
    - ✅ `sys_mgr.go` 已不再使用 `systemDependencies`，改为按 SystemId 顺序初始化，因此无需清理依赖关系定义
    - ✅ 已确认 proto 中的 `SystemId` 枚举不包含这些已移除的系统ID（VipSys、DailyActivitySys、FriendSys、GuildSys、AuctionSys）
    - ✅ 已确认没有系统注册这些已移除的系统ID
  - **关键代码位置**：
    - `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
    - `proto/csproto/system.proto`

### 1.2 测试与验证

- [ ] **Use Case 单元测试与 Controller 集成测试**
  - **目标**：Use Case 层覆盖率 ≥ 70%
  - **优先级系统**：Bag/Money/Equip/Attr/Quest/Fuben/Shop/Recycle/Vip/DailyActivity
  - **重点覆盖**：涉及货币/物品修改、副本结算、公会/拍卖与 GM 操作等高风险路径
  - **维护清单**：在文档中维护一份「已覆盖系统列表」

- [ ] **系统行为端到端回归**
  - **流程**：登录→主线→副本→社交
  - **验证系统**：所有完成迁移并删除旧实现的系统
  - **验收标准**：确认无行为退化
  - **文档更新**：回归通过后，在重构文档的「验证所有功能正常」小节勾选对应条目

---

## 二、中优先级（功能完善）

### 2.1 MessageSys 功能完善

- [✅] **离线消息回放机制检查**
  - ✅ 确认登录/重连时自动加载机制是否完整：`MessageSystemAdapter` 在 `OnInit`、`OnRoleLogin`、`OnRoleReconnect` 时调用 `loadMsgFromDB(0)` 加载离线消息，机制完整
  - ✅ 验证回调成功后删库逻辑是否正确：`onLoadMsgFromDB` 中调用 `processMessage` 处理消息，成功返回 true 后调用 `database.DeletePlayerActorMessage` 删除消息，逻辑正确
  - ✅ 检查失败保留机制是否正常工作：`processMessage` 失败返回 false，消息保留在数据库中，机制正常

- [✅] **消息类型与回调扩展**
  - ✅ 检查是否有新的业务场景需要扩展消息类型：当前消息注册机制（`engine.RegisterMessageCallback` 和 `engine.RegisterMessagePb3Factory`）支持任意消息类型扩展，新业务场景可通过注册回调扩展
  - ✅ 确认回调注册机制是否完善：`message_registry.go` 提供了完整的消息回调注册机制，支持线程安全的注册/注销，机制完善
  - ✅ 验证消息分发逻辑是否覆盖所有场景：`DispatchPlayerMessage` 统一处理在线和离线消息，`handlePlayerMessageMsg` 处理在线消息，`MessageSystemAdapter` 处理离线消息，覆盖所有场景

- [✅] **消息持久化与过期清理**
  - ✅ 验证消息持久化策略是否正常工作：`SavePlayerActorMessage` 使用 `servertime.Now().Unix()` 记录创建时间，持久化策略正常
  - ✅ 检查过期清理策略是否按预期执行：已在 `MessageSystemAdapter.OnNewDay` 中实现过期消息清理逻辑（清理超过7天的消息），使用 `database.DeleteExpiredPlayerActorMessages(MessageExpireDays)` 清理
  - ✅ 确认消息数量限制是否生效：已在 `MessageSystemAdapter.RunOne` 中实现消息数量限制检查（每个玩家最多1000条消息），超过限制时删除最旧的消息

- [✅] **UseCase 层重构（可选）**
  - ✅ 评估是否需要为 MessageSys 创建 UseCase 层：当前 MessageSys 主要逻辑为加载离线消息和消息分发，属于框架层面的消息处理，保留在适配层符合 Clean Architecture 原则，无需下沉到 UseCase
  - ✅ 当前主要在 SystemAdapter 中实现，如果业务逻辑复杂，考虑下沉到 UseCase：当前实现简洁清晰，业务逻辑不复杂，保持现状即可

**关键代码位置：**
- `server/service/gameserver/internel/adapter/system/message_system_adapter.go`
- `server/service/gameserver/internel/app/playeractor/entitysystem/message_dispatcher.go`
- `server/service/gameserver/internel/app/playeractor/entity/player_network.go`（handlePlayerMessageMsg）

### 2.2 安全与运维增强

- [⏳] **GM 权限模型与审计日志落地**
  - ✅ **入口校验**：已在 `GMManager.ExecuteCommand` 中实现 GM 等级校验（`playerRole.GetGMLevel()`），但尚未实现账号标记/IP 白名单/环境令牌校验
  - [ ] **审计日志**：**待实现** - 需要为高危 GM 指令（如 `addmoney`、`additem`、`sendmailall`、`sendnoticeall` 等）添加结构化审计日志，记录操作者账号/角色/IP/指令/参数/目标/时间等信息
  - [ ] **文档记录**：**待实现** - 需要设计审计日志表结构或结构化日志格式，并在文档中记录
  - **参考**：对应服务端开发进度文档 6.2 中 GM 权限与审计体系
  - **注意**：当前 GM 权限校验仅检查 GM 等级，缺少账号标记/IP 白名单/环境令牌等安全校验，需要在 `handleGMCommand` 入口处补充

- [⏳] **Gateway / GameServer / DungeonServer 接入安全**
  - [ ] **Gateway WebSocket**：**待实现** - 当前配置 `AllowedIPs=nil` 且 `CheckOrigin=func() bool { return true }`，仅适合开发环境；生产需启用 IP 白名单、Origin 校验与握手阶段的签名/Token 校验
  - [ ] **GameServer ↔ DungeonServer**：**待实现** - 需要增加 TLS/双向认证或等价的签名校验，确保只接受可信来源
  - [ ] **配置示例**：**待实现** - 完成后在配置章节补充示例配置
  - **参考**：按服务端开发进度文档 7.4 的要求
  - **关键代码位置**：
    - `server/service/gateway/internel/engine/server.go:165-172`（WebSocket 配置）
    - `server/service/gateway/internel/engine/config.go`（Gateway 配置结构）

---

## 三、低优先级（文档与流程）

### 3.1 文档更新

- [✅] **更新架构文档**
  - ✅ 已在 `docs/gameserver_CleanArchitecture重构文档.md` 第 3.1 节补充“Clean Architecture 分层映射表”，明确 `domain/usecase/adapter/framework` 与实际目录的对应关系
  - ✅ 说明 SystemAdapter、Controller、Presenter、Gateway 的职责划分，并要求新增层级需同步本表

- [✅] **更新开发指南**
  - ✅ 已在 `docs/服务端开发进度文档.md` 第 7.12 节新增 “Clean Architecture 开发指南”，覆盖域分析 → Use Case 设计 → SystemAdapter → Controller/Presenter → 依赖检查 → 文档同步的完整步骤
  - ✅ 指南中明确 UseCase 依赖接口、Controller/Presenter 三件套命名与提交流程

- [✅] **更新协议注册文档**
  - ✅ `docs/服务端开发进度文档.md` 第 7.3 节更新协议注册规范：所有 C2S/RPC 入口统一在 `adapter/controller/*_controller_init.go` 注册；DungeonServer 同步该约束；列出 Request/Response 处理要求
  - ✅ 补充 Controller/Presenter 初始化清单，记录注册步骤与限制

- [✅] **更新关键代码位置**
  - ✅ 第 8 章已补充 MessageSys 相关关键文件（`adapter/system/message_system_adapter.go`、`entitysystem/message_dispatcher.go`），并同步 Controller/Presenter 目录说明
  - ✅ Version 记录新增 “MessageSys 功能完善 & 文档更新” 条目，便于追溯

### 3.2 开发流程对齐

- [✅] **同步任务清单**
  - ✅ `docs/gameserver_CleanArchitecture重构文档.md`、`docs/服务端开发进度文档.md`、本文档三处任务清单已同步更新状态（新增 MessageSys/文档更新条目）
  - ✅ 统一使用复选框标记当前进度

- [✅] **版本记录更新**
  - ✅ 在 `docs/服务端开发进度文档.md` 第 10 章新增 “MessageSys 功能完善 & 文档更新” 版本记录
  - ✅ 记录了 Clean Architecture 说明、协议注册规范、关键代码位置的更新时间

---

## 四、实施建议

### 4.1 实施顺序

1. **第一阶段**：完成高优先级任务（系统依赖关系清理、测试与验证）
2. **第二阶段**：完成中优先级任务（MessageSys 功能完善、安全与运维增强）
3. **第三阶段**：完成低优先级任务（文档更新、开发流程对齐）

### 4.2 验收标准

- **系统依赖关系清理**：拓扑排序不再出现错误，系统初始化顺序正确
- **测试与验证**：Use Case 层覆盖率 ≥ 70%，端到端回归通过
- **MessageSys 功能完善**：离线消息回放机制完整，消息持久化与过期清理正常
- **安全与运维增强**：GM 权限校验生效，接入安全配置完成
- **文档更新**：架构文档、开发指南、协议注册文档均已更新

### 4.3 注意事项

- 每完成一项任务，及时更新本文档和重构文档的对应章节
- 保持两个文档（重构文档和开发进度文档）的任务清单一致
- 遇到问题及时记录，并在文档中补充说明
- 定期回顾进度，调整优先级和计划

---

**文档版本：** v1.0  
**最后更新：** 2025-01-XX  
**责任人：** 开发团队

