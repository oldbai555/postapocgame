# SystemAdapter 验证清单

本文档用于验证 SystemAdapter 的生命周期与事件处理是否符合 Clean Architecture 原则。

## 验证原则

1. **生命周期方法**：只负责"何时调用哪个 UseCase"，不包含业务逻辑
2. **事件订阅**：只负责"订阅哪个事件"，业务逻辑由 UseCase 层处理
3. **状态管理**：只管理与 Actor 运行模型强相关的运行时状态（如 dirty 标记、定时任务）

---

## 系统清单

### BagSystemAdapter（背包系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitBagDataUseCase`（如果存在）
  - 职责：初始化 BagData 结构，重建辅助索引 itemIndex
  - 业务逻辑位置：UseCase 层（`usecase/bag/init_bag_data.go`，如果存在）

#### 事件订阅验证

- [✅] **无事件订阅**
  - 说明：背包系统的事件（OnItemAdd/OnItemRemove/OnBagExpand）由 UseCase 层发布
  - 其他系统如需响应这些事件，应在各自的 UseCase 中订阅

#### 对外接口验证

- [✅] **AddItem/RemoveItem/HasItem**
  - 调用 UseCase：`AddItemUseCase`、`RemoveItemUseCase`、`HasItemUseCase`
  - 职责：路由到对应的 UseCase

---

### MoneySystemAdapter（货币系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitMoneyDataUseCase`
  - 职责：初始化货币数据结构和默认金币
  - 业务逻辑位置：`usecase/money/init_money_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **AddMoney/ConsumeMoney/UpdateBalanceTx**
  - 调用 UseCase：`AddMoneyUseCase`、`ConsumeMoneyUseCase`、`UpdateBalanceTxUseCase`
  - 职责：路由到对应的 UseCase

---

### EquipSystemAdapter（装备系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitEquipDataUseCase`
  - 职责：初始化装备数据（装备列表结构）
  - 业务逻辑位置：`usecase/equip/init_equip_data.go`

#### 事件订阅验证

- [✅] **OnEquipChange**
  - 订阅事件：`gevent2.OnEquipChange`
  - 职责：标记属性系统需要重算（框架状态管理，非业务逻辑）
  - 调用：`AttrCalculator.MarkDirty(SaEquip)`
  - 说明：这是框架层面的状态管理，符合 Clean Architecture 原则

- [✅] **OnEquipUpgrade**
  - 订阅事件：`gevent2.OnEquipUpgrade`
  - 职责：标记属性系统需要重算（框架状态管理，非业务逻辑）
  - 调用：`AttrCalculator.MarkDirty(SaEquip)`

#### 对外接口验证

- [✅] **EquipItem/UnEquipItem**
  - 调用 UseCase：`EquipItemUseCase`、`UnEquipItemUseCase`
  - 职责：路由到对应的 UseCase

---

### LevelSystemAdapter（等级系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitLevelDataUseCase`
  - 职责：初始化等级数据（默认值修正、经验同步）
  - 业务逻辑位置：`usecase/level/init_level_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**
  - 说明：等级系统的事件（OnPlayerLevelUp/OnPlayerExpChange）由 UseCase 层发布
  - 其他系统如需响应这些事件，应在各自的 UseCase 中订阅

#### 对外接口验证

- [✅] **AddExp/LevelUp**
  - 调用 UseCase：`AddExpUseCase`、`LevelUpUseCase`
  - 职责：路由到对应的 UseCase

---

### ItemUseSystemAdapter（物品使用系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitItemUseDataUseCase`
  - 职责：初始化物品使用数据（冷却映射结构）
  - 业务逻辑位置：`usecase/item_use/init_item_use_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **UseItem/CheckCooldown**
  - 调用 UseCase：`UseItemUseCase`
  - 职责：路由到对应的 UseCase

---

### SkillSystemAdapter（技能系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitSkillDataUseCase`
  - 职责：初始化技能数据（按职业配置初始化基础技能）
  - 业务逻辑位置：`usecase/skill/init_skill_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **LearnSkill/UpgradeSkill**
  - 调用 UseCase：`LearnSkillUseCase`、`UpgradeSkillUseCase`
  - 职责：路由到对应的 UseCase

---

### QuestSystemAdapter（任务系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitQuestDataUseCase`、`RefreshQuestTypeUseCase.EnsureRepeatableQuests`
  - 职责：初始化任务数据结构，确保可重复任务存在
  - 业务逻辑位置：`usecase/quest/init_quest_data.go`、`usecase/quest/refresh_quest_type.go`

- [✅] **OnNewDay**
  - 调用 UseCase：`RefreshQuestTypeUseCase.Execute(questCategoryDaily)`
  - 职责：刷新每日任务
  - 业务逻辑位置：`usecase/quest/refresh_quest_type.go`

- [✅] **OnNewWeek**
  - 调用 UseCase：`RefreshQuestTypeUseCase.Execute(questCategoryWeekly)`
  - 职责：刷新每周任务
  - 业务逻辑位置：`usecase/quest/refresh_quest_type.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **AcceptQuest/UpdateQuestProgressByType/SubmitQuest**
  - 调用 UseCase：`AcceptQuestUseCase`、`UpdateQuestProgressUseCase`、`SubmitQuestUseCase`
  - 职责：路由到对应的 UseCase

---

### FubenSystemAdapter（副本系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitDungeonDataUseCase`
  - 职责：初始化副本数据（副本记录容器结构）
  - 业务逻辑位置：`usecase/fuben/init_dungeon_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **EnterDungeon/GetDungeonRecord**
  - 调用 UseCase：`EnterDungeonUseCase`、`GetDungeonRecordUseCase`
  - 职责：路由到对应的 UseCase

---

### ShopSystemAdapter（商城系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：无（商城数据无需初始化）
  - 职责：暂未使用

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **BuyItem**
  - 调用 UseCase：`BuyItemUseCase`
  - 职责：路由到对应的 UseCase

---

### RecycleSystemAdapter（回收系统）

#### 生命周期验证

- [✅] **无生命周期方法**
  - 说明：回收系统以单例形式暴露，不依赖 Actor 生命周期

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **RecycleItem**
  - 调用 UseCase：`RecycleItemUseCase`
  - 职责：路由到对应的 UseCase

---

### ChatSystemAdapter（聊天系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：无（暂无需特殊处理）
  - 职责：初始化聊天系统

#### 事件订阅验证

- [✅] **无事件订阅**

#### 状态管理验证

- [✅] **lastChatTime**
  - 说明：限流逻辑属于框架状态管理，保留在适配层符合 Clean Architecture 原则

#### 对外接口验证

- [✅] **SendWorldChat/SendPrivateChat**
  - 调用 UseCase：`ChatWorldUseCase`、`ChatPrivateUseCase`
  - 职责：路由到对应的 UseCase

---

### MailSystemAdapter（邮件系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：`InitMailDataUseCase`
  - 职责：初始化邮件数据（邮件列表结构）
  - 业务逻辑位置：`usecase/mail/init_mail_data.go`

#### 事件订阅验证

- [✅] **无事件订阅**

#### 对外接口验证

- [✅] **SendMail/ClaimAttachments/ReadAndDelete**
  - 调用 UseCase：`SendCustomMailUseCase`、`SendTemplateMailUseCase`、`ClaimAttachmentsUseCase`、`ReadAndDeleteUseCase`
  - 职责：路由到对应的 UseCase

---

### MessageSystemAdapter（玩家消息系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：无（直接调用 DAO）
  - 职责：加载离线消息并触发回放
  - 说明：主要逻辑为加载离线消息，属于框架层面的消息处理，保留在适配层符合 Clean Architecture 原则

- [✅] **OnRoleLogin**
  - 调用 UseCase：无（直接调用 DAO）
  - 职责：登录时加载离线消息

- [✅] **OnRoleReconnect**
  - 调用 UseCase：无（直接调用 DAO）
  - 职责：重连时加载离线消息

#### 事件订阅验证

- [✅] **无事件订阅**

---

### GMSystemAdapter（GM 系统）

#### 生命周期验证

- [✅] **OnInit**
  - 调用 UseCase：无（初始化 GMManager）
  - 职责：初始化 GMManager
  - 说明：主要逻辑为执行 GM 命令，属于框架层面的命令处理，保留在适配层符合 Clean Architecture 原则

#### 事件订阅验证

- [✅] **OnSrvStart**
  - 订阅事件：`gevent2.OnSrvStart`
  - 职责：注册 GM 协议入口
  - 说明：协议注册属于框架层面的初始化，保留在适配层符合 Clean Architecture 原则

#### 对外接口验证

- [✅] **ExecuteGMCommand**
  - 调用 UseCase：`GMManager.Execute`
  - 职责：路由到 GMManager

---

## 错误/异常情况下的降级行为

### 通用降级策略

1. **UseCase 调用失败**
   - 行为：记录错误日志，返回错误给调用者
   - 示例：`QuestSystemAdapter.OnNewDay` 中如果 `RefreshQuestTypeUseCase.Execute` 失败，记录错误日志但不中断流程

2. **依赖未初始化**
   - 行为：检查依赖是否为 nil，如果为 nil 则返回错误或使用向后兼容的旧方式
   - 示例：`EquipItemUseCase` 中如果 `bagUseCase` 为 nil，调用 `equipItemLegacy`（向后兼容）

3. **数据不存在**
   - 行为：初始化默认数据结构
   - 示例：`BagSystemAdapter.OnInit` 中如果 `BagData` 不存在，创建新的 `BagData`

### 系统特定降级策略

#### QuestSystemAdapter

- **OnNewDay/OnNewWeek 失败**
  - 行为：记录错误日志，不中断流程
  - 影响：任务不会刷新，但不会影响其他系统

#### EquipSystemAdapter

- **OnEquipChange/OnEquipUpgrade 事件处理失败**
  - 行为：记录警告日志，不中断流程
  - 影响：属性系统可能不会立即重算，但会在下次 RunOne 时重算

#### MessageSystemAdapter

- **加载离线消息失败**
  - 行为：记录错误日志，不中断流程
  - 影响：离线消息不会回放，但不会影响玩家登录

---

## 验证方法

1. **代码审查**
   - 检查每个 SystemAdapter 的头部注释是否说明生命周期职责
   - 检查生命周期方法是否只调用 UseCase，不包含业务逻辑
   - 检查事件订阅是否有注释说明

2. **运行时验证**
   - 在测试环境中触发各个生命周期方法
   - 验证是否正确调用了对应的 UseCase
   - 验证错误情况下的降级行为

3. **单元测试**
   - 为 SystemAdapter 的生命周期方法编写轻量集成测试
   - 验证 UseCase 调用是否正确
   - 验证错误处理是否正确

---

## 更新记录

- 2025-01-XX：初始版本，完成所有 SystemAdapter 的验证清单

