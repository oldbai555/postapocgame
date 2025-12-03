## 1. 文档目的与适用范围

- **文档目的**：给出一份针对 `server/service/gameserver/internel/adapter/system` 的优化演进规划，明确：
  - 这层在 Clean Architecture 下应当承担的职责边界；
  - 各系统从“胖 SystemAdapter”向“瘦胶水层 + 纯 UseCase”的迁移步骤；
  - 可按阶段执行的待办清单（含复选框），便于你逐步实现与检查。
- **适用范围**：仅针对 GameServer 内的 SystemAdapter（玩家 Actor / PublicActor 侧）；不直接约束 DungeonServer。
- **阅读前置**：
  - 《`docs/gameserver_CleanArchitecture重构文档.md`》第 3、4 章；
  - 《`docs/gameserver_目录与调用关系说明.md`》第 4.6、5 章；
  - 《`docs/系统移除方案.md`》：了解已移除的系统（Vip、DailyActivity、Friend、Guild、Auction）和重构的系统（Attr 改为工具类）。

> 约定：本文件内所有任务统一使用复选框跟踪进度——`[ ]` 表示待做，`[⏳]` 表示进行中，`[✅]` 表示已完成但仍保留在清单中用于追溯。

> **重要更新**（2025-01-XX）：
> - ~~SysVip、SysDailyActivity、SysFriend、SysGuild、SysAuction~~ 已按最小粒度游戏需求移除
> - ~~SysAttr~~ 已重构为工具类 `AttrCalculator`，注入到 `PlayerRole` 中，不再是 SystemAdapter
> - 本文档已同步更新，移除了所有已删除系统的相关内容

---

## 2. 当前 `adapter/system` 的角色（现状定位）

在现有实现中，`internel/adapter/system/*` 主要承担三类职责：

1. **生命周期适配**  
   - 把 PlayerActor 的 `OnInit/RunOne/OnNewDay/OnNewWeek/OnLogout` 等钩子，转换为对“本系统”的调用。
2. **业务编排与状态管理（偏多）**  
   - 保存系统级运行时状态（缓存、脏标记、冷却等）；
   - 直接操作 `PlayerRoleBinaryData` 或调用其他系统；
   - 甚至在适配层中混入部分业务规则，实现本应属于 UseCase 的逻辑。
3. **外部交互**  
   - 少量 SystemAdapter 直接调用 Gateway / DungeonServerGateway / PublicActorGateway / Event 等外部设施。

这导致：

- UseCase 层已经涵盖了大量业务逻辑，但 SystemAdapter 仍残留不少“业务 + 状态 + 框架”混合代码；
- 若简单删除 `adapter/system` 并把逻辑挪入 UseCase，会让 UseCase 感知 Actor/帧循环/事件，削弱 Clean Architecture 的优势。

后续演进的目标不是移除 `adapter/system`，而是**让它变成一个更薄的“胶水层 + 调度层”**。

---

## 3. 目标状态与演进原则

### 3.1 目标状态（理想形态）

对于每一个业务系统（Bag/Money/Equip/Skill/Quest/Fuben/ItemUse/Shop/Recycle/Chat/Mail/Message/GM 等），理想形态是：

> **注意**：以下系统已移除或重构：
> - ~~Attr~~：已重构为工具类 `AttrCalculator`，注入到 `PlayerRole` 中，不再是 SystemAdapter
> - ~~Vip/DailyActivity/Friend/Guild/Auction~~：已按最小粒度游戏需求移除

- **UseCase 层**：
  - 包含“单次业务动作”的全部规则（AddItem、EnterDungeon、BuyItem、SubmitQuest 等）；
  - 只依赖接口（Repository、ConfigManager、DungeonServerGateway、PublicActorGateway 等）；
  - 不感知 Actor/帧循环/SessionID，仅通过输入/输出建模业务。
- **SystemAdapter 层**：
  - 只做两类事情：
    1. 生命周期事件 → UseCase 调用的编排（包括 `OnInit/RunOne/OnNewDay/OnNewWeek/OnLogout` 等）；
    2. 管理与 Actor 运行模型强相关的运行时状态（如属性 dirty 标记、定时任务、订阅玩家级事件）。
  - 不直接操作数据库 / 网络；需要外部交互时，优先调用 UseCase 或接口。

### 3.2 演进原则

- **能下沉到 UseCase 的业务规则，尽量下沉**：
  - 校验逻辑、奖励构建、条件判断、状态变更等，都应该在 UseCase 层统一实现。
- **SystemAdapter 只保留“何时调用 / 调用谁”的逻辑**：
  - 比如：“每帧 RunOne 时检查哪些系统需要重算属性”，而具体“如何重算属性”在 UseCase 中。
- **跨系统调用统一通过接口**：
  - 不在 SystemAdapter 中直接 import 其他 SystemAdapter，而是通过 `usecase/interfaces` + `usecaseadapter` 实现依赖；
  - 避免产生新的循环依赖。
- **改一处，补一处文档与测试**：
  - 每个系统完成从旧 SystemAdapter → 目标形态的迁移后，补充一次小范围单测/集成测试，并在对应文档中记录。

---

## 4. 分阶段演进计划（总览清单）

### 4.1 阶段 A：梳理职责边界（查看 + 标记）

- [✅] **A1：列出现有 SystemAdapter 清单与职责归类**  
  - 目标：为每个 `adapter/system/*` 记录“生命周期编排 / 业务逻辑 / 外部交互”三类代码的占比与大致位置。
  - 行动建议：
    - 针对每个系统（Bag/Money/Equip/Skill/Quest/Fuben/ItemUse/Shop/Recycle/Chat/Mail/Message/GM），简单写一段注释或本文件子节，描述当前 SystemAdapter 的职责。
    - **注意**：~~Attr~~ 已重构为工具类，~~Vip/DailyActivity/Friend/Guild/Auction~~ 已移除，不再需要描述。

- [✅] **A2：标记可下沉到 UseCase 的逻辑块**  
  - 目标：在代码中通过 TODO 或注释标出“纯业务逻辑”，准备迁移到 UseCase。
  - 行动建议：
    - 优先关注：奖励计算、条件校验、状态变更、冷却判定、次数限制等逻辑。

#### 4.1.A1 SystemAdapter 职责现状梳理（按系统）

- **BagSystemAdapter（背包）**  
  - **生命周期编排**：在 `OnInit` 中确保 `BagData` 初始化并重建物品索引，用于后续快速查询。  
  - **业务逻辑**：封装增删查物品、快照生成与恢复、事务型增删（`AddItemTx/RemoveItemTx`）以及基础容量校验、堆叠规则等兼容旧逻辑的能力。  
  - **外部交互**：通过 `PlayerGateway` 读取/写入 `PlayerRoleBinaryData.BagData`，通过 `EventPublisher`/`ConfigGateway` 间接参与经济与配置驱动逻辑，对其他系统提供 `AddItem/RemoveItem/HasItem` 等对外入口。

- **MoneySystemAdapter（货币）**  
  - **生命周期编排**：在 `OnInit` 中保证 `MoneyData` 与 `MoneyMap` 初始化，并在空余额时注入默认金币，避免后续访问空 map。  
  - **业务逻辑**：封装加钱/扣钱接口、兼容旧接口（`CostMoney/UpdateBalanceTx/UpdateBalanceOnlyMemory`）以及余额不足检测等规则，并通过 `MoneyUseCaseImpl` 为其他系统提供统一货币用例实现。  
  - **外部交互**：依赖 `PlayerGateway`、`EventPublisher`、`ConfigGateway`，与其他系统通过 UseCase 接口协作。

- **EquipSystemAdapter（装备）**  
  - **生命周期编排**：在 `OnInit` 中保证 `EquipData` 初始化，并记录当前玩家装备数量。  
  - **业务逻辑**：自身只持有 `Equip/UnEquip` 用例引用，几乎不做复杂规则计算；装备位合法性、属性加成等规则主要由 UseCase 与 `AttrCalculator` 工具类完成。  
  - **外部交互**：通过 `PlayerGateway` 访问 `EquipData`，并通过 `EventPublisher/ConfigGateway` 在 UseCase 层完成配置驱动和事件发布；通过 `PlayerRole.GetAttrCalculator()` 访问属性计算器。

- ~~**AttrSystemAdapter（属性）**~~  
  - **已重构为工具类**：属性系统已从 SystemAdapter 重构为 `AttrCalculator` 工具类，注入到 `PlayerRole` 中。  
  - **职责变更**：属性计算、战力计算、属性同步等逻辑现在由 `entity/attr_calculator.go` 中的 `AttrCalculator` 承担，通过 `PlayerRole.GetAttrCalculator()` 访问。  
  - **生命周期**：`AttrCalculator` 在 `PlayerRole` 初始化时创建并调用 `Init(ctx)`，在 `RunOne` 中调用 `RunOne(ctx)`，在重连时调用 `PushFullAttrData(ctx)`。

- **LevelSystemAdapter（等级）**  
  - **生命周期编排**：在玩家初始化与经验变更场景中，负责触发等级相关 UseCase（加经验、升级、属性加成刷新等）。  
  - **业务逻辑**：大部分经验/升级规则（经验曲线、等级上限、属性加成）已下沉到 `usecase/level` 与 `attrcalc`，SystemAdapter 主要作为"何时触发"的调度层。  
  - **外部交互**：与 `AttrCalculator`/`MoneySystemAdapter` 通过 UseCase 接口协作，避免直接依赖其他 SystemAdapter。

- **ItemUseSystemAdapter（物品使用）**  
  - **生命周期编排**：注册并处理与物品使用相关的生命周期事件（如登录后重置可用次数、每日刷新）。  
  - **业务逻辑**：物品使用效果、前置条件、冷却/次数限制等主要在 UseCase 层实现，SystemAdapter 负责在合适的时机调用 UseCase 并与任务/活跃度等系统协作。  
  - **外部交互**：通过 `BagUseCase`/`LevelUseCase` 等接口访问背包和等级信息，避免直接跨系统引用。

- **SkillSystemAdapter（技能）**  
  - **生命周期编排**：在玩家上线/换技能栏等场景中，负责加载技能数据并在必要时触发重算属性。  
  - **业务逻辑**：学习/升级技能的条件校验、消耗与奖励逻辑主要在 UseCase 中实现；SystemAdapter 保留极少量与帧循环或事件订阅相关的逻辑。  
  - **外部交互**：通过 `ConsumeUseCase`、`LevelUseCase` 等接口与货币、等级系统协作，并通过 `DungeonServerGateway` 参与战斗服技能数据同步。

- **QuestSystemAdapter（任务）**  
  - **生命周期编排**：在 `OnNewDay/OnNewWeek` 中刷新任务进度，在关键事件（击杀、通关、副本结算等）发生时分发给任务 UseCase。  
  - **业务逻辑**：任务接取/更新/提交规则、奖励构建等集中在 `usecase/quest`，SystemAdapter 主要负责"事件到任务"的映射与调度。  
  - **外部交互**：通过 `RewardUseCase` 等接口与奖励系统协作，并依赖配置接口获取任务与 NPC 场景配置。

- **FubenSystemAdapter（副本）**  
  - **生命周期编排**：在玩家请求进入/结算副本时，协调调用 UseCase 与 DungeonServer RPC，管理与 Actor 生命周期相关的副本状态。  
  - **业务逻辑**：副本条件校验、记录更新、奖励构建等逻辑主要在 `usecase/fuben` 完成；SystemAdapter 负责选择何时发起进入/结算调用。  
  - **外部交互**：通过 `DungeonServerGateway` 进行 `Enter/Settle` 等 RPC 调用，并通过 `RewardUseCase/LevelUseCase/ConsumeUseCase` 处理收益与成本。

- **ShopSystemAdapter（商城）**  
  - **生命周期编排**：在玩家登录或特定事件中刷新商城状态（若有），并驱动购买流程入口。  
  - **业务逻辑**：购买条件校验、消耗与奖励列表构建在 `usecase/shop` 内实现；SystemAdapter 仅负责将请求路由到正确的 UseCase。  
  - **外部交互**：通过 `ConsumeUseCase/RewardUseCase` 与货币/奖励系统协作，通过 `ConfigGateway` 获取商城/消耗/奖励配置。

- **RecycleSystemAdapter（回收）**  
  - **生命周期编排**：注册回收系统到 Actor 生命周期，保证在玩家初始化时可用。  
  - **业务逻辑**：回收物品的校验与奖励计算主要由 `usecase/recycle` 承担；SystemAdapter 作为入口与辅助索引/推送封装层。  
  - **外部交互**：统一依赖 `BagUseCase`、`RewardUseCase` 与配置接口完成物品扣除与奖励发放。

- **ChatSystemAdapter（聊天系统）**  
  - **生命周期编排**：在玩家登录/登出时与 PublicActor 同步聊天等全局状态；订阅与社交事件相关的玩家级事件。  
  - **业务逻辑**：聊天限频等规则均集中在 `usecase/chat`；SystemAdapter 只保留将生命周期与玩家事件映射到正确 UseCase 的逻辑。  
  - **外部交互**：统一通过 `PublicActorGateway` 与 PublicActor 通信，遵循 Phase 3 的社交经济架构决策。

- **MailSystemAdapter（邮件）**  
  - **生命周期编排**：在玩家登录/新邮件到达等节点触发邮件加载与推送。  
  - **业务逻辑**：发件、收件、附件领取等规则在 `usecase/mail` 中实现；SystemAdapter 负责挂载到 Actor 生命周期并向 UseCase 传递上下文。  
  - **外部交互**：与 `GMSys` 及运营工具通过 UseCase 协同发送系统邮件，不直接操作数据库。

- **MessageSystemAdapter（玩家消息）**  
  - **生命周期编排**：在登录/重连时加载离线玩家消息并触发回放，在 Actor 生命周期结束或条件满足时清理消息。  
  - **业务逻辑**：消息类型注册、回放时机、失败重试策略主要在 UseCase 与消息注册中心中实现；SystemAdapter 负责在生命周期钩子中统一调用。  
  - **外部交互**：通过 `PlayerActorMessage` DAO 与数据库交互，通过 `NetworkGateway` 推送消息结果。

- **GMSystemAdapter（GM）**  
  - **生命周期编排**：在系统初始化时注册 GM 协议入口与 GM 工具函数，挂载到玩家 Actor 生命周期。  
  - **业务逻辑**：GM 指令解析、权限校验与实际执行逻辑逐步下沉到 GM 用例与工具函数中；SystemAdapter 本身主要协调协议入口与系统邮件/奖励等子系统。  
  - **外部交互**：通过 `NetworkGateway` 回传执行结果，并复用 Mail/Bag/Money 等 UseCase 完成 GM 操作产生的副作用。

### 4.2 阶段 B：下沉业务逻辑到 UseCase（按系统逐步推进）

> 建议优先处理“高复用 + 高复杂”的核心玩法系统，再处理社交/辅助系统。

- [⏳] **B1：核心玩法系统（Level / Bag / Money / Equip / ItemUse / Skill / Quest / Fuben / Shop / Recycle）**  
  - [⏳] B1.1 为每个系统补足/统一 UseCase 接口（如 `FubenUseCase` 等），确保：
    - 所有"单次业务动作"都有对应 UseCase 方法；
    - 方法签名只依赖 DTO / 接口，不引入 Actor/Session。
    - ✅ Money：已新增 `UpdateBalanceTxUseCase`（事务型余额更新）、`InitMoneyDataUseCase`（货币数据初始化）
    - ✅ Bag：已新增 `AddItemTxUseCase`、`RemoveItemTxUseCase`（事务型物品增删）
    - ✅ Level：已新增 `InitLevelDataUseCase`（等级数据初始化）
    - ✅ ItemUse：已新增 `InitItemUseDataUseCase`（物品使用数据初始化）
    - ✅ Equip：已新增 `InitEquipDataUseCase`（装备数据初始化）
    - ✅ Skill：已新增 `InitSkillDataUseCase`（技能数据初始化，按职业配置初始化基础技能）
    - ✅ Quest：已新增 `InitQuestDataUseCase`（任务数据初始化，任务桶结构与基础任务类型集合）
    - ✅ Attr（工具类）：已新增 `CalculateSysPowerUseCase`（系统战力计算）、`CompareAttrVecUseCase`（属性向量比较），由 `AttrCalculator` 工具类调用
    - ✅ Fuben：已新增 `InitDungeonDataUseCase`（副本数据初始化）、`GetDungeonRecordUseCase`（副本记录查找）
    - ✅ Shop：主要业务逻辑已在 `BuyItemUseCase` 中，适配层已精简
    - ✅ Recycle：主要业务逻辑已在 `RecycleItemUseCase` 中，适配层已精简
    - ✅ Chat：限流逻辑属于框架状态管理，保留在适配层
    - ✅ Mail：已新增 `InitMailDataUseCase`（邮件数据初始化）
    - ✅ Message：主要逻辑为加载离线消息，属于框架层面，保留在适配层
    - ✅ GM：主要逻辑为执行 GM 命令，属于框架层面，保留在适配层
  - [⏳] B1.2 将 SystemAdapter 中的业务逻辑块（非生命周期/非调度）迁移到 UseCase：
    - 迁移后，SystemAdapter 中只保留：
      - 生命周期事件分发；
      - 对"是否需要调用 UseCase"的简单判断。
    - ✅ Money：`UpdateBalanceTx` 已改为调用 UseCase，余额计算与不足校验规则已下沉；`OnInit` 已改为调用 UseCase，MoneyData 初始化与默认金币注入逻辑已下沉
    - ✅ Bag：`AddItemTx`、`RemoveItemTx` 已改为调用 UseCase，物品配置校验、堆叠规则、容量检查等业务逻辑已下沉
    - ✅ Level：`OnInit` 已改为调用 UseCase，等级默认值修正、经验同步等初始化逻辑已下沉
    - ✅ ItemUse：`OnInit` 已改为调用 UseCase，冷却映射结构初始化逻辑已下沉
    - ✅ Equip：`OnInit` 已改为调用 UseCase，装备列表结构初始化逻辑已下沉
    - ✅ Skill：`OnInit` 已改为调用 UseCase，按职业配置初始化基础技能列表逻辑已下沉
    - ✅ Quest：`OnInit` 已改为调用 UseCase，任务桶结构与基础任务类型集合初始化逻辑已下沉
    - ✅ Attr（工具类）：属性计算逻辑已迁移到 `AttrCalculator` 工具类，通过 `CalculateSysPowerUseCase` 和 `CompareAttrVecUseCase` 实现
    - ✅ Fuben：`OnInit` 已改为调用 UseCase，副本数据初始化逻辑已下沉；`GetDungeonRecord` 已改为调用 UseCase，副本记录查找逻辑已下沉
    - ✅ Shop：主要业务逻辑已在 UseCase 中，适配层已精简为薄胶水层
    - ✅ Recycle：主要业务逻辑已在 UseCase 中，适配层已精简为薄胶水层
    - ✅ Chat：限流逻辑属于框架状态管理，保留在适配层，符合 Clean Architecture 原则
    - ✅ Mail：`OnInit` 已改为调用 UseCase，邮件数据初始化逻辑已下沉
    - ✅ Message：主要逻辑为加载离线消息，属于框架层面，保留在适配层
    - ✅ GM：主要逻辑为执行 GM 命令，属于框架层面，保留在适配层
  - [⏳] B1.3 对迁移后的 UseCase 编写/补充单元测试。
    - ✅ 已为以下 UseCase 编写单元测试：
      - `InitMoneyDataUseCase`：覆盖初始化、默认金币注入、已存在数据不覆盖等场景
      - `InitLevelDataUseCase`：覆盖初始化、等级/经验修正、经验同步到货币系统等场景
      - `AddItemTxUseCase`：覆盖添加新物品、堆叠、背包容量检查、配置校验等场景
      - `RemoveItemTxUseCase`：覆盖移除物品、跨物品移除、物品不足检查等场景
      - `CompareAttrVecUseCase`：覆盖属性向量比较的各种场景（nil、相等、不等、顺序无关等）
      - `CalculateSysPowerUseCase`：覆盖系统战力计算的各种场景（空数据、单个系统、多个系统、加成率等）
    - [ ] 继续为其他 UseCase 编写测试（如 `UpdateBalanceTxUseCase`、`InitSkillDataUseCase` 等）

- [✅] **B2：社交系统（Chat）**  
  - [✅] B2.1 核对当前 UseCase 是否已完整承载"聊天消息发送、限频、敏感词过滤"等逻辑（已完成）
  - [✅] B2.2 将 SystemAdapter 中残余的业务逻辑迁移至相应 UseCase（限流逻辑属于框架状态管理，保留在适配层）
  - [✅] B2.3 确保与 PublicActor 交互统一由 UseCase 通过 `PublicActorGateway` 完成（已完成）

> **注意**：~~Friend / Guild / Auction~~ 系统已移除，不再需要处理。

- [✅] **B3：辅助系统（Mail / Message / GM）**  
  - [✅] B3.1 明确每个辅助系统的用例集合（发邮件、追加玩家消息、执行 GM 指令等）（已完成）
  - [✅] B3.2 将其在 SystemAdapter 中的业务逻辑迁移到 UseCase（已完成）
  - [✅] B3.3 根据需要补充 `usecase/interfaces` 与适配器，避免直接在 SystemAdapter 中跨系统调用（已完成）

> **注意**：~~Vip / DailyActivity~~ 系统已移除，不再需要处理。

### 4.3 阶段 C：精简生命周期与事件处理（保留"胶水"逻辑）

- [✅] **C1：统一 SystemAdapter 生命周期签名与职责说明**  
  - 目标：确保所有 SystemAdapter 的生命周期方法风格一致（如 `OnInit/OnEnterGame/OnLogout/RunOne/OnNewDay/OnNewWeek` 等），并在头部注释中说明：
    - 本系统在每个生命周期阶段"仅做哪些调度行为"；
    - 具体业务由哪些 UseCase 承担。
  - ✅ 已完成：为所有 SystemAdapter 添加了统一的头部注释，说明生命周期职责和业务逻辑位置
  - ✅ 已完成：在 `BaseSystemAdapter` 中添加了详细的职责说明注释

- [✅] **C2：将复杂定时/调度逻辑下沉或抽象**  
  - 目标：SystemAdapter 不直接操作复杂的时间轮或手写 tick 逻辑，而是：
    - 在合适的时机调用"带有时间窗口参数"的 UseCase；
    - 或引入一个更通用的调度工具（如统一的 Scheduler Adapter），由它来管理时间与重试策略。
  - ✅ 已完成：创建了 `RefreshQuestTypeUseCase`，将任务刷新逻辑（`refreshQuestType`、`shouldRefresh`、`ensureRepeatableQuests`）下沉到 UseCase 层
  - ✅ 已完成：`QuestSystemAdapter.OnNewDay/OnNewWeek` 现在只调用 UseCase，不再包含业务逻辑

- [✅] **C3：事件订阅精简**  
  - 目标：将事件处理逻辑拆分为：
    - SystemAdapter 层只负责"订阅哪个事件 / 在事件到来时调用哪个 UseCase"；
    - UseCase 层处理事件业务本身。
  - ✅ 已完成：删除了空的事件订阅（Bag 系统的 OnItemAdd/OnItemRemove/OnBagExpand，Level 系统的 OnPlayerLevelUp/OnPlayerExpChange）
  - ✅ 已完成：为保留的事件订阅（Equip 系统的 OnEquipChange/OnEquipUpgrade）添加了注释，说明这是框架层面的状态管理

### 4.4 阶段 D：测试与文档补全

- [⏳] **D1：为关键 UseCase 编写单元测试**  
  - 覆盖：背包增删、货币增减、装备穿脱、属性重算、副本进入与结算、任务接受与提交、商城购买与回收等。
  - ✅ 已完成测试：
    - `InitMoneyDataUseCase`：覆盖初始化、默认金币注入、已存在数据不覆盖等场景
    - `InitLevelDataUseCase`：覆盖初始化、等级/经验修正、经验同步到货币系统等场景
    - `AddItemTxUseCase`：覆盖添加新物品、堆叠、背包容量检查、配置校验等场景
    - `RemoveItemTxUseCase`：覆盖移除物品、跨物品移除、物品不足检查等场景
    - `CompareAttrVecUseCase`：覆盖属性向量比较的各种场景（nil、相等、不等、顺序无关等）
    - `CalculateSysPowerUseCase`：覆盖系统战力计算的各种场景（空数据、单个系统、多个系统、加成率等）
  - [ ] 待补充测试：
    - Equip：`EquipItemUseCase`、`UnEquipItemUseCase`
    - Fuben：`EnterDungeonUseCase`、`SettleDungeonUseCase`
    - Quest：`AcceptQuestUseCase`、`SubmitQuestUseCase`、`RefreshQuestTypeUseCase`
    - Shop：`BuyItemUseCase`
    - Recycle：`RecycleItemUseCase`
    - Skill：`LearnSkillUseCase`、`UpgradeSkillUseCase`
    - ItemUse：`UseItemUseCase`

- [✅] **D2：为 SystemAdapter 编写轻量集成测试或脚本验证清单**  
  - 针对生命周期与事件，列出：
    - 在什么时机会调用哪个 UseCase；
    - 错误/异常情况下的降级行为。
  - ✅ 已完成：创建了 `docs/SystemAdapter验证清单.md`，包含所有 SystemAdapter 的生命周期验证、事件订阅验证、对外接口验证和错误/异常情况下的降级行为

- [✅] **D3：更新文档**  
  - [✅] D3.1 在《`docs/gameserver_CleanArchitecture重构文档.md`》中新增一小节，记录 SystemAdapter 的目标职责与当前状态（已完成：在第 15.3 节添加了 SystemAdapter 职责与演进章节）
  - [✅] D3.2 在《`docs/服务端开发进度文档.md`》第 6 章中，为本演进计划增加一条待办描述，并在完成阶段时标记为 ✅（已完成：已在第 6.1 节中记录）

### 4.5 阶段 E：清理 Legacy 代码与防退化机制

- [✅] **E1：清理不再需要的 legacy SystemAdapter 逻辑或字段**  
  - 在确认所有业务逻辑已迁移到 UseCase 且测试通过后：
    - ✅ 删除或精简 SystemAdapter 中历史遗留的字段/方法（已完成：修复了 item_use_system_adapter.go 中的 TODO，使用 servertime）
    - ✅ 移除重复/不再使用的工具函数（已完成：检查并确认所有方法都有明确的用途）

- [✅] **E2：增加防退化检查**  
  - 目标：避免后续开发者再次把业务逻辑写回 SystemAdapter。
  - 手段示例：
    - ✅ 在 SystemAdapter 头部注释中明确标注"禁止编写业务规则逻辑，只允许调用 UseCase 与管理生命周期"（已完成：为所有 SystemAdapter 添加了防退化说明）
    - ✅ 在 Code Review 清单中新增"SystemAdapter 是否含有可下沉到 UseCase 的逻辑"检查项（已完成：创建了 `docs/SystemAdapter_CodeReview清单.md`）
    - ✅ 在 BaseSystemAdapter 中添加了详细的防退化机制说明

---

## 5. 按系统的粗略优先级（可选参考）

若需要规划一条“从易到难 / 从基础到玩法”的推进顺序，可以参考：

- [ ] **P1（基础 + 高复用，高优先级）**  
  - LevelSys、BagSys、MoneySys、EquipSys、ItemUseSys
  - ~~AttrSys~~（已重构为工具类 `AttrCalculator`，注入到 `PlayerRole` 中）
- [ ] **P2（核心玩法）**  
  - SkillSys、QuestSys、FubenSys、ShopSys、RecycleSys
- [ ] **P3（社交）**  
  - ChatSys
  - ~~FriendSys、GuildSys、AuctionSys~~（已移除，按最小粒度游戏需求）
- [ ] **P4（辅助/运营）**  
  - MailSys、MessageSys、GMSys
  - ~~VipSys、DailyActivitySys~~（已移除，按最小粒度游戏需求）

> 建议：每完成一个 P1/P2 系统的迁移，就同步勾选对应阶段 B/C/D 中的细分复选框，避免“全部做完再一次性勾选”导致信息不透明。

---

## 6. 使用建议

- 开发前：先在本文件中选择本次要推进的阶段（A~E）与系统（P1~P4），圈定小范围任务；
- 开发中：在代码中用 TODO 或注释标记“已迁移/待迁移”的逻辑块，对照本文件更新复选框；
- 开发后：补充/更新用例测试与文档，并在 `docs/服务端开发进度文档.md` 的“待实现 / 待完善功能”下登记阶段性成果。


