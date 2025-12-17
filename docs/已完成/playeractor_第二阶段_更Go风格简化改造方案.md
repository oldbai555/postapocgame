# PlayerActor 第二阶段：更面向过程、更 Go 风格的简化改造方案

> 适用范围：`server/service/gameserver/internel/app/playeractor/*`  
> 目标读者：本项目后端开发者（你将按本文进行开发）  
> 前置阶段：已完成第一阶段“目录扁平化”（`controller/ system/ service/ gateway/ presenter/ router/ event/`）  

---

## 阶段任务清单（按步骤勾选）

- [x] 基线确认：第一阶段目录扁平化完成，`cd server && go test ./...` 通过
- [x] 接口归口：PlayerActor 相关端口/用例接口统一迁移到 `server/service/gameserver/internel/iface`，gateway/interfaces 改为 type alias
- [x] Phase 2A（功能切片）：Bag 系统代码迁入 `playeractor/bag` 包（含 controller/system/presenter/service），聚合包使用 blank import
- [x] Phase 2A（功能切片）：其他系统（Money/Equip/Skill/Fuben/Recycle 等）逐一迁入各自包并保持接线（全部完成，Chat 已移除）
- [x] Phase 2B（过程化）：Bag 系统的 UseCase 改为函数/小 service，依赖通过轻量 `BagDeps` 注入
- [x] Phase 2B（过程化）：其他系统的 UseCase 改为函数/小 service（Money/Equip/Skill/Fuben/Recycle 全部完成）
- [x] Phase 2D：引入 `app/runtime` 聚合依赖 + 显式注册替代 init()（已完成 2025-12-17）
  - ✅ 创建 `playeractor/runtime` 包聚合依赖（PlayerRepo/ConfigManager/EventPublisher/NetworkGateway/DungeonGateway/PublicGateway）
  - ✅ PlayerRole 新增 `runtime` 字段和 `GetRuntime()` 方法
  - ✅ 各系统统一使用 `depsFromRuntime(rt)` 函数，**完全移除** `depsFromGlobal()`
  - ✅ 创建 `playeractor/register` 包，提供 `RegisterAll(rt)` 显式注册所有系统
  - ✅ 在 `main.go` 启动时调用 `register.RegisterAll(globalRuntime)`
  - ✅ 所有 Controller/SystemAdapter 构造函数统一接受 `*runtime.Runtime` 参数
  - ✅ 移除所有系统的 `init()` 函数，迁移至统一注册
  - ✅ 编译通过，`go test ./...` **完全通过**
- [ ] 回归验证：`cd server && go test ./...`，以及 example 客户端关键链路（登录/背包/副本拾取/重连）

---

## 0. 背景与问题

当前 PlayerActor 代码虽已做“目录扁平化”，但仍保留了较强的 Clean Architecture / OOP 痕迹：

- Controller 多为 `NewXXXController()` + 持有多个 `*XXXUseCase` 字段，业务调用集中在 `uc.Execute(...)`
- SystemAdapter 也持有 UseCase 指针，形成“System→UseCase→deps 全局单例”的长链路
- 接口定义集中在 `service/interfaces`，容易演化成“大接口集合”
- 依赖装配主要靠 `deps` 全局单例（隐式依赖、测试成本高）

第二阶段目标是在不破坏 Actor 单线程约束、协议接线方式与现有功能的前提下，让代码更贴近 Go 常见工程风格：

- 更少层级、更少对象壳、更短调用链
- 面向过程：**业务行为用函数/小型 service 直达**，依赖显式传入
- 接口集中管理：**所有接口统一收归到 `server/service/gameserver/internel/iface`**，避免散落定义与重复抽象
- 迁移策略统一：**从本阶段起，所有系统的重构与迁移一律直接按新方案落地，不再保留旧代码路径、兼容层或“过渡 wrapper（compat adapter）”**

---

## 1. 必须遵守的硬约束（不要为了“更 Go”破坏这些）

- **Actor 单线程/无锁约束**：任何玩家状态修改必须在 PlayerActor 主线程内完成，禁止在业务里起 goroutine 改玩家数据
- **时间统一**：业务时间统一用 `server/internal/servertime`
- **网络发送链路**：发客户端消息必须经 PlayerActor（不允许其他模块直接调 gatewaylink）
- **DungeonActor 交互**：UseCase/业务层访问 DungeonActor 必须通过 Gateway（如 `DungeonServerGateway`），不要直接 `gshare.SendDungeonMessageAsync`
- **系统开关检查**：Controller 负责系统是否开启检查，业务层（service）不感知开关状态

---

## 2. 当前基线（第一阶段完成后）

现有目录（示例）：

```
server/service/gameserver/internel/app/playeractor/
  controller/  (协议入口，init 注册)
  system/      (生命周期胶水，init 注册 system factory)
  service/     (业务逻辑，仍以 *XXXUseCase.Execute 为主)
  gateway/ presenter/ router/ event/ ...
```

接线方式（重要）：

- `server/service/gameserver/requires.go` 通过空导入触发 init 注册：
  - `_ ".../playeractor/controller"`
  - `_ ".../playeractor/system"`

这条链路建议在第二阶段继续沿用（减少改动面），直到你愿意切到“显式注册”方案。

### 2.1 当前仓库已落地的改造点（与本文保持一致）

为对齐“接口统一收归 `internel/iface`”的约束，当前仓库已将 PlayerActor 相关的端口/用例接口迁移到：

- `server/service/gameserver/internel/iface/`
  - `playeractor_player_repository.go` / `playeractor_account_repository.go` / `playeractor_role_repository.go`
  - `playeractor_config_manager.go` / `playeractor_event_publisher.go`
  - `playeractor_public_actor_gateway.go` / `playeractor_dungeon_server_gateway.go` / `playeractor_token_generator.go`
  - `playeractor_usecase_*.go`（Bag/Money/Level/Consume/Reward 等）
  - `playeractor_client_gateway.go`（Network/Session/ClientGateway）

并将 `playeractor/gateway/interfaces.go` 改为对 `internel/iface` 的 type alias（确保接口定义只在 `iface`）。

---

## 3. 第二阶段总体路线（建议按顺序做）

### 3.1 Phase 2A：按“功能切片”收拢代码（目录/包级别的整合）

目标：以一个功能为单位（Bag/Money/Equip/...），把该功能涉及的 controller + system + service + presenter 放在同一个包里，减少跨包跳转。

**推荐目标形态（以 Bag 为例）：**

```
playeractor/
  bag/
    register.go        # init 注册协议 & actor msg handler（可选拆分）
    system.go          # SysBag 的 SystemAdapter（init 注册 system factory）
    handler.go         # C2S/D2G handler（原 controller）
    presenter.go       # S2C 构建与发送（原 presenter）
    service.go         # 业务函数/小 service（原 service/bag）
    (无 ports.go)       # 接口统一放到 internel/iface（见 Phase 2C）
```

其余系统可仿照：
`playeractor/money`、`playeractor/equip`、`playeractor/skill`、`playeractor/fuben`、`playeractor/recycle`、`playeractor/chat` 等。

#### 3.1.1 其它系统分片排期（可逐项勾选）

- [x] **Money（货币） → `playeractor/money`**
  - controller：`controller/money_controller.go`
  - system：`system/money_sys.go`
  - service：`service/money/*`
  - presenter：`presenter/money_presenter.go`
  - 风险点：与 Bag/Consume/Reward 的交互较多，迁移时需确认 `ConsumeUseCase`、`RewardUseCase` 的依赖是否正确指向新包。
- [x] **Equip（装备） → `playeractor/equip`**
  - controller：`controller/equip_controller.go`
  - system：`system/equip_sys.go`
  - service：`service/equip/*`
  - presenter：`presenter/equip_presenter.go`
  - 风险点：依赖 Bag/Attr/Money，多条链路（装备成功后属性刷新、背包变化推送），迁移时需重点跑 example 的装备相关脚本。
- [x] **Skill（技能） → `playeractor/skill`**
  - controller：`controller/skill_controller.go`
  - system：`system/skill_sys.go`
  - service：`service/skill/*`
  - presenter：`presenter/skill_presenter.go`
  - 风险点：与 Level/Fuben/DungeonActor 的联动较多，需确保 `DungeonServerGateway` 调用链不被破坏。
- [x] **Fuben（副本） → `playeractor/fuben`**
  - controller：`controller/fuben_controller.go`
  - system：`system/fuben_sys.go`
  - service：`service/fuben/*`
  - presenter：`presenter/fuben_presenter.go`
  - 风险点：强依赖 DungeonActor 与 Reward/Bag/Money，建议在迁移后优先回归"进入副本→结算→掉落"的完整流程。
- [x] **Recycle（回收） → `playeractor/recycle`**
  - controller：`controller/recycle_controller.go`
  - system：`system/recycle_sys.go`
  - service：`service/recycle/*`
  - presenter：`presenter/recycle_presenter.go`
  - 风险点：依赖 Bag + Reward，注意 `NewRecycleSystemAdapter` 内的依赖构造是否仍然简单清晰。
- [x] **Chat（聊天）系统已移除**
  - 已删除所有 Chat 相关代码（controller/system/service/presenter/domain）。

### 3.2 Phase 2B：从 “UseCase 对象” 改为 “函数/小 service”（更面向过程）

目标：把 `type AddItemUseCase struct {...}; func (uc *AddItemUseCase) Execute(...)` 逐步改成：

- **纯函数风格：**
  - `func AddItem(ctx context.Context, deps Deps, roleID uint64, itemID, count, bind uint32) error`
- 或 **小 service 风格（仍是 Go 常见写法，但更轻）：**
  - `type Service struct { repo Repo; cfg Cfg; events Events }`
  - `func (s *Service) AddItem(...) error`

两种都比“多个 UseCase struct 分散在不同目录”更 Go。

#### 3.2.1 关键落点：把“依赖装配”从 `NewXXXController()` 里搬出来

以当前 Bag 为例，现状常见是：

- Controller：`NewBagController()` 内部直接 `bag.NewAddItemUseCase(deps.PlayerGateway(), deps.EventPublisher(), ...)`
- System：`NewBagSystemAdapter()` 内部同样 new 多个 UseCase

这会导致：

- 依赖来源隐式（global deps 到处被调用）
- Controller/System 变成“巨型构造器 + 持有多个 UseCase 指针”的 OOP 形态

第二阶段建议把依赖聚合成一个轻量结构（仅数据，不要有复杂逻辑）：

- `type BagDeps struct { Repo iface.IPlayerRepository; Cfg iface.IConfigManager; Events iface.IEventPublisher; Net iface.INetworkGateway }`

然后在注册处（`init` 或 `RegisterAll`）构造 `BagDeps`，并将 handler 注册为闭包/函数，避免 Controller struct。

> 注意：这里的 `iface.*` 只是示意名称，实际以你在 `internel/iface` 中定义/迁移后的接口为准。

#### 3.2.2 `depsFromGlobalOrApp(ctx)` 这类占位符必须落地为可编译代码

文档后续示例会用到 `depsFromGlobalOrApp(ctx)`（表示“从全局 deps 或未来的 app/runtime 中取依赖”）。
为了让你按文档开发时不会卡住，建议在每个 feature 包里放一个非常薄的装配函数（Phase 2D 前先这么做）：

```go
// 仅示意：在 Phase 2D（引入 app/runtime）之前，可先用全局 deps 过渡。
func depsFromGlobal() Deps {
    return Deps{
        Repo:   deps.PlayerGateway(),   // 需满足 iface.IBagRepository / iface.IPlayerRepository 等接口
        Cfg:    deps.ConfigGateway(),   // 需满足 iface.IConfigManager / iface.IBagConfig 等接口
        Events: deps.EventPublisher(),  // 需满足 iface.IEventPublisher / iface.IBagEventPublisher 等接口
    }
}
```

Phase 2D 后再替换为 `depsFromApp(ctx)`（从上下文/PlayerRole 获取 app/runtime 实例）。

### 3.3 Phase 2C：接口收归 `internel/iface`（不要在功能包里到处定义接口）

目标：把 PlayerActor 侧所有“端口接口”（Repo/Gateway/Publisher/Config 等）统一收归：

- `server/service/gameserver/internel/iface`

并按 **“最小接口 + 领域前缀 + 分文件”** 管理，避免变成新的“大接口合集”。

#### 3.3.1 命名与分文件规范（建议强制执行）

- **文件命名**：按领域拆文件，使用明确前缀，避免一个文件越来越大：
  - `server/service/gameserver/internel/iface/playeractor_bag.go`
  - `server/service/gameserver/internel/iface/playeractor_money.go`
  - `server/service/gameserver/internel/iface/playeractor_common.go`（少量通用接口）
- **类型命名（与项目既有风格一致）**：
  - 当前 `internel/iface` 已大量使用 `I*`（如 `IPlayerRole`、`ISystem`），为避免风格混乱，**新增接口建议也使用 `I*` 前缀**，并用领域词避免泛化：
    - 推荐：`IPlayerRepository`、`IConfigManager`、`IEventPublisher`、`IDungeonServerGateway`
    - 如果必须拆得更细：`IBagRepository`、`IBagConfig`、`IBagEventPublisher`
- **接口粒度**：一个接口只表达一个用途（读写/发布/查询），不要把多能力塞进一个接口里。
- **依赖方向**：`internel/iface` 只依赖稳定底层包（`context`、`server/internal/*`、标准库等），**禁止依赖 `internel/app/*`**，否则很容易引入循环依赖。

#### 3.3.1.1 与现有 `iface/irole.go` 的关系（重要）

`internel/iface` 里已经有 `IPlayerRole / ISystem / ISystemMgr` 等“Actor 内聚对象接口”。第二阶段做“更 Go / 更过程化”时，建议遵循：

- **handler/controller 层**：可以依赖 `IPlayerRole`（用于系统开关检查、取 RoleID/Session 等上下文信息）
- **业务层（service）**：尽量不要直接依赖 `IPlayerRole`，而是依赖更小的端口接口（Repository/Gateway/Config/EventPublisher），这样业务逻辑更纯、更可测试，也更符合依赖倒置

如果业务层确实只需要 `GetBagData()` 这种简单读操作，你也可以选择依赖 `IPlayerSiDataRepository`，但要注意它通常是“内存视图（无 error）”，与需要 `ctx+error` 的仓储接口语义不同。

#### 3.3.2 Bag 示例（接口定义在 iface）

例如 Bag 功能需要的最小接口可以在 `internel/iface/playeractor_bag.go` 中定义（示意）：

```go
package iface

import (
    "context"
    "postapocgame/server/internal/event"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/internal/jsonconf"
)

type IBagRepository interface {
    GetBagData(ctx context.Context) (*protocol.SiBagData, error)
    // 其它 Bag 业务需要的最小方法...
}

type IBagConfig interface {
    GetItemConfig(itemID uint32) *jsonconf.ItemConfig
    GetBagConfig(bagType uint32) *jsonconf.BagConfig
}

type IBagEventPublisher interface {
    // 建议直接沿用现有 EventPublisher 的签名（event.Type + 可变参数），减少迁移成本
    PublishPlayerEvent(ctx context.Context, eventType event.Type, args ...interface{})
}
```

由 `playeractor/gateway` 与 `playeractor/event`（或未来的 runtime/app）提供具体实现。

> 备注：上面只是示意签名。实际项目里事件类型与参数类型请以现有 `EventPublisher`/`gevent` 设计为准，核心原则是“最小接口 + 统一收敛到 iface”。

#### 3.3.3 从现状迁移到 `internel/iface`（操作步骤）

现状中已经存在接口包（例如 `playeractor/service/interfaces`、`playeractor/domain/repository`）。按“统一收归 iface”的约束，建议按下面顺序迁移，尽量做到“每一步都可编译、可回归”：

1. **在 `internel/iface` 新增分领域文件并定义最小接口**
   - 示例：
     - `server/service/gameserver/internel/iface/playeractor_bag.go`
     - `server/service/gameserver/internel/iface/playeractor_money.go`
   - 规则：
     - `iface` 包**只能依赖更底层的稳定包**（如 `server/internal/protocol`、`server/internal/jsonconf`、`context`），不要依赖 `internel/app/playeractor/*`，否则容易形成循环依赖。

2. **让实现方“对齐接口”但不强制改实现代码结构**
   - `playeractor/gateway/*`、`playeractor/event/*` 等实现包，通常只需要：
     - 调整方法签名（如果旧接口签名不同）
     - 确保 `var _ iface.XXX = (*Impl)(nil)` 能通过编译

3. **逐步替换业务层/系统层构造参数类型**
   - 将 `playeractor/service/*` 里的构造函数参数，从旧的 `playeractor/service/interfaces` 改为 `internel/iface`：
     - `func NewService(repo iface.IPlayerRepository, cfg iface.IConfigManager, ...)`
     - 或更细粒度：`func NewService(repo iface.IBagRepository, cfg iface.IBagConfig, ...)`
   - 这一步建议按系统逐个做（先 Bag，再 Money/Equip/...），避免一次性全量替换引发大范围冲突。

4. **删除旧接口包（或先保留薄 wrapper 过渡）**
   - 当 `playeractor/service/interfaces`、`playeractor/domain/repository` 不再被引用后：
     - 直接删除（推荐）
     - 或保留一层过渡 wrapper（不推荐长期存在），以降低合并期间的冲突成本

5. **回归检查**
   - `cd server && go test ./...`
   - 用 `server/example` 完成关键链路回归（登录/背包/副本拾取/重连等）

#### 3.3.4 建议的“直接迁移清单”（更贴合现有仓库）

为了让你按文档开发时改动更可控，建议先把现有接口包“迁移/折叠”进 `internel/iface`，而不是立刻发明大量全新接口：

- **迁移接口定义（推荐先做这一步，最省心）**
  - 来源：
    - `server/service/gameserver/internel/app/playeractor/service/interfaces/*.go`
    - `server/service/gameserver/internel/app/playeractor/domain/repository/*.go`
  - 目标（示例）：
    - `server/service/gameserver/internel/iface/playeractor_ports_*.go`
    - `server/service/gameserver/internel/iface/playeractor_repository_*.go`

对应命令示例（在仓库根目录执行，建议“按文件迁移”，最终 package 统一为 `iface`）：

```powershell
git mv server/service/gameserver/internel/app/playeractor/service/interfaces/event.go `
      server/service/gameserver/internel/iface/playeractor_event_publisher.go

git mv server/service/gameserver/internel/app/playeractor/service/interfaces/config.go `
      server/service/gameserver/internel/iface/playeractor_config_manager.go

git mv server/service/gameserver/internel/app/playeractor/service/interfaces/rpc.go `
      server/service/gameserver/internel/iface/playeractor_dungeon_gateway.go

git mv server/service/gameserver/internel/app/playeractor/domain/repository/player_repository.go `
      server/service/gameserver/internel/iface/playeractor_player_repository.go
```

然后逐步：

- 把 `package interfaces` / `package repository` 改为 `package iface`
- 解决 import（注意 `iface` 不要 import `internel/app/playeractor/*`）
- 用 `go test` 驱动逐步修编译

> 说明：上面只展示了几类典型接口的迁移方式，你可以继续按文件迁移其余接口文件；关键是“接口最终归口到 `internel/iface`，并保持依赖方向正确”。  

> 关于 `ConfigManager`：现有 `service/interfaces/config.go` 是一个“全量配置接口”（方法很多）。为了降低迁移成本，建议先原样迁到 `iface.IConfigManager`；后续如果要拆分成 `IBagConfig/IMoneyConfig...` 也可以，但仍然放在 `internel/iface`，并通过分文件控制规模。

> 额外提示：`domain/repository/player_repository.go` 里目前还定义了一组 `ErrXxxNotFound`，这类错误值是否迁移到 `iface` 可自行决定；若迁移，建议单独放 `internel/iface/playeractor_errors.go`，避免接口文件塞进太多业务语义。

### 3.4 Phase 2D：移除全局 `deps` 单例（可选，但最终最“Go”）

目标：依赖显式注入，不再到处 `deps.PlayerGateway()`：

- 引入 `playeractor/app`（或 `playeractor/runtime`）对象，启动时构造一次并挂到 `PlayerRole` 或 Actor 上
- Controller/System/Service 通过上下文拿到 `app`，再取子依赖

你可以先在 Bag 系统试点，不用一次性全量替换。

---

## 4. Bag 系统完整试点：从现状到目标（可照抄执行）

下面给出一个“最小风险”的迁移路径：先做 Phase 2A（目录切片），再做 Phase 2B（函数化）。

### 4.1 Step 0：准备工作

在仓库根目录：

```powershell
git checkout -b refactor/playeractor-phase2
cd server
go test ./...
```

确保基线通过后开始改造。

### 4.2 Step 1：创建新包 `playeractor/bag`

创建目录：

```
server/service/gameserver/internel/app/playeractor/bag/
```

把现有 Bag 相关文件移动过来（建议用 `git mv` 保留历史）：

- 从 `controller/` 移：
  - `controller/bag_controller.go` → `bag/handler.go`
- 从 `system/` 移：
  - `system/bag_sys.go` → `bag/system.go`
  - `system/bag_use_case_adapter.go` → `bag/compat_ports.go`（先保留，后续按 Phase 2C 把接口迁到 `internel/iface` 后再删）
- 从 `presenter/` 移：
  - `presenter/bag_presenter.go` → `bag/presenter.go`
- 从 `service/bag/` 移：
  - `service/bag/*.go` → `bag/*.go`（可合并成 `service.go` / `accessor.go` 等）

示例命令（在仓库根目录执行）：

```powershell
git mv server/service/gameserver/internel/app/playeractor/controller/bag_controller.go `
      server/service/gameserver/internel/app/playeractor/bag/handler.go

git mv server/service/gameserver/internel/app/playeractor/system/bag_sys.go `
      server/service/gameserver/internel/app/playeractor/bag/system.go

git mv server/service/gameserver/internel/app/playeractor/presenter/bag_presenter.go `
      server/service/gameserver/internel/app/playeractor/bag/presenter.go

git mv server/service/gameserver/internel/app/playeractor/service/bag `
      server/service/gameserver/internel/app/playeractor/bag/service
```

> 注：上面最后一行把目录整体移为 `bag/service/*` 也可以；后续你再把代码合并成单文件/少文件。

### 4.3 Step 2：保持 `requires.go` 不变：在聚合包中 blank import

当前 `requires.go` 空导入的是 `playeractor/controller` 与 `playeractor/system`。为了避免你还要改接线，建议：

- `playeractor/controller` 包变为“聚合注册包”，仅 blank import 各功能包（Bag/Money/...）
- `playeractor/system` 同理

在 `playeractor/controller/` 下新增 `imports.go`：

```go
package controller

import (
    _ "postapocgame/server/service/gameserver/internel/app/playeractor/bag"
)
```

在 `playeractor/system/` 下新增 `imports.go`：

```go
package system

import (
    _ "postapocgame/server/service/gameserver/internel/app/playeractor/bag"
)
```

然后把原先 `controller/` 里与 Bag 相关的文件移走后，这两个聚合包仍会触发 Bag 包的 `init()` 注册。

> 关键点：Bag 包不要 import `controller`/`system` 聚合包，否则会形成循环依赖。

### 4.4 Step 3：修包名与 import

把移动后的文件包名统一为 `package bag`，并修正引用：

- 原 `controller/bag_controller.go`：
  - `package controller` → `package bag`
  - 对外调用 `router.RegisterProtocolHandler`、`gshare.RegisterHandler` 逻辑保留
- 原 `presenter/bag_presenter.go`：
  - `package presenter` → `package bag`
- 原 `system/bag_sys.go`：
  - `package system` → `package bag`
  - `GetBagSys(ctx)` 保留或迁移为 `bag.GetSys(ctx)`

做到这一步后，先确保仍能编译：

```powershell
cd server
go test ./...
```

### 4.5 Step 4（可选）：把 UseCase 改成更过程化（Phase 2B）

以“添加物品”为例，现状一般是：

```go
type AddItemUseCase struct { ... }
func (uc *AddItemUseCase) Execute(ctx, roleID, itemID, ...) error
```

推荐改为 Bag 包内的过程化写法：

#### 方案 B1：纯函数 + 显式 deps

```go
type Deps struct {
    Repo   iface.IBagRepository
    Cfg    iface.IBagConfig
    Events iface.IBagEventPublisher
}

func AddItem(ctx context.Context, d Deps, roleID uint64, itemID, count, bind uint32) error {
    // 原 Execute 的逻辑搬过来
    // ...
    return nil
}
```

Controller 调用时：

```go
err := AddItem(ctx, depsFromGlobalOrApp(ctx), req.RoleId, req.ItemId, req.Count, 0)
```

System 调用时：

```go
return AddItem(ctx, depsFromGlobalOrApp(ctx), roleID, itemID, count, bind)
```

#### 方案 B2：小 service（依旧显式依赖，但更便于复用）

```go
type Service struct {
    repo   iface.IBagRepository
    cfg    iface.IBagConfig
    events iface.IBagEventPublisher
}

func NewService(repo iface.IBagRepository, cfg iface.IBagConfig, events iface.IBagEventPublisher) *Service { ... }

func (s *Service) AddItem(ctx context.Context, roleID uint64, itemID, count, bind uint32) error {
    // ...
    return nil
}
```

> 建议：先对 Bag 做到 B1/B2 之一；确认风格满意后再推广到 Money/Equip/...，避免一次性大爆炸。

---

## 5. 显式注册 + 全局 deps 移除 + sysbase 基类（已完成）

**完成时间**：2025-12-17

### 5.1 显式注册

- 创建 `playeractor/register/register.go` 提供 `func RegisterAll(rt *runtime.Runtime)`（注册协议、注册 actor msg handler、注册 system factory）
- 在 `gameserver` 启动时调用 `register.RegisterAll(globalRuntime)`（替代 `requires.go` blank import）
- 移除所有 feature 包中的 `init()` 函数

优点：无隐式 init，调用链更清晰，启动流程显式可控  
已实施：所有系统（Bag/Money/Equip/Skill/Fuben/Recycle）均已迁移

### 5.2 全局 deps 移除

- 移除 `deps.go` 中的全局单例（`deps` 变量、`depsOnce`、`ensure()` 函数等）
- 将所有 `deps.XXXGateway()` 全局访问器改为工厂函数 `deps.NewXXXGateway()`
- 保留 `deps.GetPlayerRoleManager()` 用于访问全局 Manager 单例（Manager 本身是单例设计）
- 所有原 `deps.PlayerGateway()` 等调用改为：
  - 通过 `runtime` 传递（feature 包内部）
  - 或直接 `deps.NewPlayerGateway()`（临时兼容层或非 feature 包）

**影响**：`deps` 包职责从"全局容器"变为"工厂函数集合"，依赖关系更显式，便于单元测试。

### 5.3 sysbase 基类（统一 ISystem 实现）

- 新增 `playeractor/sysbase` 包，提供 `type BaseSystem struct`，完整实现 `iface.ISystem` 接口
- 各业务系统统一改为组合 `*sysbase.BaseSystem`：
  - `bag.BagSystemAdapter`
  - `money.MoneySystemAdapter`
  - `equip.EquipSystemAdapter`
  - `skill.SkillSystemAdapter`
  - `fuben.FuBenSystemAdapter`
  - `recycle.RecycleSystemAdapter`（单例系统，使用 `SystemIdNil` 作为占位 ID）
- `playeractor/system` 包中的 `LevelSystemAdapter`、`MessageSystemAdapter` 也改为使用 `sysbase.BaseSystem`
- 旧的 `system.BaseSystemAdapter` 标记为 Deprecated，仅保留空壳占位，避免误用

**影响**：所有系统都通过同一个基础实现满足 `ISystem` 接口，减少重复样板代码，同时保持 feature 包不依赖聚合 `system` 包，避免循环依赖。

---

## 6. 推广到其他系统的清单（照 Bag 复制）

对每个系统（Money/Equip/Skill/Fuben/Recycle/Chat/Message...）重复执行：

1. 创建 `playeractor/<feature>/` 包
2. `git mv` 把该系统相关 controller/system/service/presenter 合并进去
3. 在聚合包 `playeractor/controller` 与 `playeractor/system` 增加 blank import
4. 编译通过后，逐步把 `*XXXUseCase.Execute` 改成函数/小 service
5. 收敛接口：把功能只用到的最小接口放到 `server/service/gameserver/internel/iface/playeractor_<feature>.go`
6. 删除旧目录与旧接口定义（确保无引用）

---

## 7. 验收与回归检查（每一步都要做）

### 7.1 编译/测试

```powershell
cd server
go test ./...
```

### 7.2 防止循环依赖

- 新 feature 包不要 import `playeractor/controller`（聚合包）与 `playeractor/system`（聚合包）
- 聚合包仅做 blank import，不承载业务代码

### 7.3 运行时行为检查（建议用 example 客户端）

- 登录/进游戏
- 打开背包（C2SOpenBag）
- 副本拾取触发 D2GAddItem / PlayerActorMsgIdAddItem
- 断线重连后背包数据仍正确

---

## 8. 常见坑（按踩坑概率排序）

1. **包循环依赖**：feature 包 import 了聚合包（controller/system）会直接炸
2. **注册遗漏**：移动文件后忘记保留 `init()` 或忘记在聚合包里 blank import
3. **系统开关检查下沉到了业务层**：请保持在 Controller 层（业务层不感知开关）
4. **跨 Actor 发包**：发客户端消息必须经 PlayerActor 的 NetworkGateway（Presenter 负责）
5. **引入 goroutine 改玩家数据**：强制禁止

---

## 9. 本文与其他文档的关系

- 本文是“第二阶段：更 Go 风格简化”的可执行改造指南
- 第一阶段（目录扁平化）的结论与入口位置已记录在：
  - `docs/服务端开发进度文档.md`
  - `docs/服务端开发进度文档_full.md`


