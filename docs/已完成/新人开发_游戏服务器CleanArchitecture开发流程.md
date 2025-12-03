## 1. 文档目的与适用人群

- **文档目的**：为新加入项目的开发者提供一条清晰的开发路径，指导如何在现有 GameServer 代码基础上，按照 **Clean Architecture** 规范实现/修改业务功能，并同步更新相关文档。
- **适用范围**：
  - 新人接手 GameServer 相关需求（如：青龙神装、VIP、日常活跃度等）；
  - 老人需要回顾“标准做法”时的快速参考；
  - Code Review 时作为 checklist 使用。

> **前置要求**：能看懂基础 Go 代码；对 Actor 模型、Proto、简单数据库读写有基本概念即可。

---

## 2. 开发前必读清单（必须完成）

在正式写代码之前，**每次新需求都要完成以下阅读**，避免违背现有约束：

- [✅] **阅读服务器总览与进度**
  - `docs/服务端开发进度文档.md`
    - 第 0 章 **开发必读**
    - 第 2 章 **服务器架构**
    - 第 4 章 **已完成功能**（了解现有系统能力）
    - 第 6 章 **待实现 / 待完善功能**（确认你的需求归属哪一块）

- [✅] **阅读 GameServer Clean Architecture 说明**
  - `docs/gameserver_CleanArchitecture重构文档.md`
    - 第 0 章 **导航与任务清单**
    - 第 3 章 **分层设计**
    - 第 4 章 **重构方案（分阶段说明）**

- [✅] **阅读通用架构决策与约束**
  - `docs/服务端开发进度文档.md`
    - 第 7 章 **开发注意事项与架构决策**
      - 7.1 Actor / 并发约束
      - 7.2 数据与存储
      - 7.3 协议与 Proto 规范
      - 7.4 网络与安全
      - 7.10 Clean Architecture 架构决策

> **建议习惯**：新需求先在 `docs/服务端开发进度文档.md` 的对应小节中补一行 TODO，再开始具体设计和编码。

---

## 3. GameServer Clean Architecture 快速脑图

在 GameServer 中，你可以简单记住如下分层：

- **Entities / Domain（业务实体层）**
  - 目录：`server/service/gameserver/internel/domain/...`
  - 职责：**纯业务数据结构 + 纯业务逻辑**，不直接访问数据库、网络、Actor 等框架。

- **Use Cases（用例层）**
  - 目录：`server/service/gameserver/internel/usecase/...`
  - 职责：围绕“一个业务动作”组织逻辑，例如：`AddExp`、`UseItem`、`BuyItem`、`EnterDungeon` 等。
  - 只能依赖：
    - Domain 实体；
    - `usecase/interfaces` 中定义的接口（Repository、Gateway、ConfigManager、EventPublisher 等）。

- **Interface Adapters（接口适配层）**
  - 目录：`server/service/gameserver/internel/adapter/...`
  - 子层：
    - `adapter/controller`：协议入口（C2S / D2G / G2D / PublicActor），负责解析协议、构造上下文、调用 UseCase；
    - `adapter/presenter`：响应构建，负责把 Domain/UseCase 的结果转换成下行 Proto 并通过 NetworkGateway 发送；
    - `adapter/system/<sys>/`：SystemAdapter，承接原来的 EntitySystem 生命周期（OnInit / RunOne / OnNewDay / OnNewWeek 等），内部通过接口调用 UseCase；
    - `adapter/gateway`：各种 Gateway 实现，如 PlayerGateway / DungeonServerGateway / PublicActorGateway / ConfigGateway 等。

- **Frameworks & Drivers（框架层）**
  - 目录：`server/internal/...` 以及部分 `server/service/gameserver/internel/*link` 等
  - 职责：Actor、网络、数据库、配置加载、日志等。
  - **UseCase 和 Domain 不得依赖这一层**，只能通过 Adapter/接口间接访问。

> **记忆口诀**：**“协议到 Controller，逻辑在 UseCase，数据走 Repository/Gateway，系统生命周期在 SystemAdapter”**。

---

## 4. 示例需求：实现“青龙神装”功能的标准步骤

下面以一个虚构需求 **“青龙神装（套装装备 + 属性加成）”** 为例，给出从 0 到 1 的完整流程。你可以把它类比到任何新功能（如新的副本、活动、系统等）。

---

### 4.1 需求归类与范围确认

- [✅] **确认归属系统**
  - 青龙神装 = 新增一组装备 + 套装属性加成 → **归属 EquipSys / AttrSys / Config**；
  - 若涉及获取途径（副本掉落 / 商城售卖 / 活动兑换），则还会关联 FubenSys / ShopSys / RecycleSys / QuestSys 等。

- [✅] **在文档中登记需求**
  - 在 `docs/服务端开发进度文档.md` 中：
    - 若属于 Clean Architecture 重构范畴（例如扩展 EquipSys/AttrSys 的能力），在 **6.1 Clean Architecture 重构** 下新增子项或在已有子项中补充“青龙神装支持”；
    - 若是新玩法/内容，在 **6.2 Phase 2 核心玩法** 或后续 Phase 小节中增加一条复选框 TODO。

---

### 4.2 配置与 Proto 设计

- [✅] **检查 / 设计配置表**
  - 找到装备与套装相关配置：
    - `server/output/config/item_config.json`
    - `server/output/config/equip_set_config.json`（或类似命名，具体以现有代码为准）
  - 需要支持的字段：
    - 套装 ID、套装名称；
    - 套装组成的装备 ID 列表；
    - 套装效果（触发条件：穿戴 N 件；加成属性：攻击/防御/血量等）。
  - 若需新增字段：
    - 修改 `jsonconf` 对应配置解析代码；
    - 在 `usecase/interfaces/config.go` 中补充访问接口；
    - 在 `adapter/gateway/config_gateway.go` 中实现对应方法。

- [✅] **检查 Proto 是否需要改动**
  - 装备数据是否需要增加标记（如“青龙神装套装 ID”）；
  - 属性下发是否需要新增字段（如新的系统 ID / 新的战力来源）。
  - 若需要修改 Proto：
    - 编辑 `proto/csproto/*.proto` 中对应文件；
    - 运行 `proto/genproto.sh`；
    - `gofmt` 生成的 `.pb.go`。

> **注意**：配置更改一定要同时更新 `docs/服务端开发进度文档.md` 中相关小节的“关键配置”描述，方便以后排查问题。

---

### 4.3 Domain & UseCase 设计

> 青龙神装相关逻辑应尽量抽象为 **“装备套装”** 通用能力，而不是在代码中硬编码“青龙”二字。

- [✅] **Domain 层：建模核心概念**
  - 位置：`internel/domain/equip/...` 或与现有 EquipSys/AttrSys 所在目录保持一致。
  - 可能的实体：
    - `EquipSet`：表示一个套装（ID、名称、组成部件、加成效果列表）；
    - `EquipSetEffect`：表示套装加成效果（触发件数、属性加成列表）。
  - 示例（伪代码）：

```go
type EquipSet struct {
    ID        uint32
    Pieces    []uint32      // 装备 ID 列表
    Effects   []EquipSetEffect
}

type EquipSetEffect struct {
    NeedCount int32
    Attrs     []AttrModifier // 可以复用已有的属性增减结构
}
```

- [✅] **UseCase 层：定义套装相关用例**
  - 位置：`internel/usecase/equip/...` 或 `internel/usecase/attr/...` 中与属性计算相关的部分。
  - 常见用例：
    - `RecalcEquipSetBonus`：根据当前已穿戴装备，重新计算套装加成；
    - `OnEquipChanged`：在玩家穿戴/卸下装备时触发，调用上述计算逻辑，并通知 AttrSys 重新计算属性；
    - 若有特殊玩法（如套装激活时发系统邮件、解锁称号），再拆分为独立 UseCase。

- [✅] **通过接口抽象外部依赖**
  - 例如，UseCase 可能需要：
    - 读取玩家当前装备 → 依赖 `PlayerRepository` 或 `EquipRepository`；
    - 访问配置 → 依赖 `ConfigManager`；
    - 触发属性重算 → 依赖 `AttrUseCase` 或类似接口。
  - 这些接口都应定义在 `usecase/interfaces/...` 中，由 Adapter 来实现。

---

### 4.4 SystemAdapter 与属性系统对接

青龙神装本质是额外的属性来源，因此关键是与 **AttrSys** 的集成。

- [✅] **确认 AttrSys 的扩展点**
  - 查阅：`adapter/system/attr/...` 以及 `attrcalc` 相关文档：
    - `docs/属性系统重构文档.md`
    - `server/internal/attrcalc/*`
  - 通常是通过“注册属性计算器”的方式插入新的属性来源：
    - 如已有的 Level、Equip 加成那样，为套装新增一个系统 ID 或在 Equip 计算中追加一段逻辑。

- [✅] **在 SystemAdapter 中集成 UseCase**
  - 位置：`adapter/system/equip/...` 或 `adapter/system/attr/...`。
  - 典型流程：
    - 在玩家登录或加载装备数据时，从配置中构造当前启用的套装效果；
    - 在 OnEquipChanged / OnNewDay / OnNewWeek 等生命周期中，根据需要调用 UseCase 进行重算；
    - 最终通过 AttrSys 暴露的接口（或事件）让属性系统重新计算，并推送到客户端与 DungeonServer。

> **原则**：SystemAdapter 只负责“挂钩生命周期 + 调用 UseCase”，具体业务规则全部放在 UseCase 中，方便测试。

---

### 4.5 Controller / Presenter / RPC 部分（如需要）

如果青龙神装还需要：
- 客户端主动查看“套装激活情况”；
- 运营后台通过 GM 指令发放完整套装；
则需要在控制器层做补充。

- [✅] **Controller：新增协议入口（如有）**
  - 位置：`adapter/controller/equip_controller.go` 或新建相关控制器；
  - 步骤：
    - 在 `*_controller_init.go` 中注册新的 C2S 协议；
    - 在 Handler 中解析协议 → 构造上下文（含 roleId/sessionId）→ 调用对应 UseCase。

- [✅] **Presenter：补充响应格式（如有）**
  - 位置：`adapter/presenter/equip_presenter.go` 或新增文件；
  - 负责：
    - 将 UseCase 结果转换为 Proto 结构；
    - 通过 `NetworkGateway` 下发到客户端。

- [✅] **RPC / PublicActor / DungeonServer 交互（如有）**
  - 若青龙神装效果需要在 DungeonServer 侧生效（通常会通过属性同步已经覆盖），只要确保 AttrSys → DungeonServer 的链路已经支持新的属性即可；
  - 如需跨玩家数据（如“同公会成员套装激活数加成”这类复杂玩法），则需要通过 PublicActorGateway 抽象交互，具体设计参考 Friend/Guild/Auction 的实现模式。

---

## 5. 编码规范与检查清单

### 5.1 Clean Architecture 相关

- [ ] **依赖方向正确**
  - Domain 不依赖任何 Adapter / Framework 包；
  - UseCase 只依赖 Domain 和 `usecase/interfaces` 中的接口；
  - Adapter 实现接口，依赖 `server/internal/*` 和外部库；
  - 避免直接从 UseCase 调用 `database`、`gatewaylink`、`dungeonserverlink` 等。

- [ ] **无循环依赖**
  - Controller ↔ SystemAdapter 不要互相 import；
  - 通过 `usecase/interfaces` 或 `adapter/controller/..._use_case_adapter.go` 这类 *Adapter* 间接依赖；
  - 构建失败有循环依赖时，优先抽接口而不是在原包硬拆文件。

### 5.2 时间 / 日志 / 数据访问规范

- [ ] 时间获取统一使用 `server/internal/servertime`，禁止直接 `time.Now()` / `time.Since()`。
- [ ] 日志统一使用 `server/pkg/log`，需要带上下文时通过 `IRequester` 注入 Session/Role 信息。
- [ ] 数据访问统一通过 Repository / Gateway，避免在 UseCase 中直接操作 `database.*`。
- [ ] Player 数据修改必须围绕 `PlayerRoleBinaryData`，避免散落临时全局状态。

### 5.3 Lint / Build / 测试

- [ ] 本地运行 `go test ./...`（至少覆盖你修改的包）。
- [ ] 运行项目使用的 lint（若有 `golangci-lint` 配置，则执行对应命令）。
- [ ] 确认 `go build ./server/service/gameserver` 无编译错误。

---

## 6. 提交流程与文档同步

每次合入代码前，请务必同步以下信息：

- [ ] **更新开发进度文档**
  - 文件：`docs/服务端开发进度文档.md`
  - 动作：
    - 在第 4 章 **已完成功能** 中，按模块为青龙神装补充一行说明（例如归入 EquipSys / AttrSys）；
    - 在第 6 章中，将对应 TODO 勾选为 `[✅]`，必要时拆分为更细的子项。

- [ ] **更新 Clean Architecture 文档（如涉及架构层面变更）**
  - 文件：`docs/gameserver_CleanArchitecture重构文档.md`
  - 动作：
    - 若引入了新的接口类型 / Gateway / 分层规则，在第 3/4 章补充简要说明；
    - 若是对已有模式的复用（如新增一个 UseCase / SystemAdapter，但模式不变），可以只在“任务清单”中补一行记录。

- [ ] **新增/更新专题文档（如有）**
  - 如青龙神装逻辑较复杂（多阶段解锁、多角色联动效果等），建议新增：
    - `docs/青龙神装功能设计与实现.md`
  - 内容包括：
    - 玩法说明；
    - 配置表字段解释；
    - 关键流程（登录、穿戴、卸下、属性重算、结果下发）；
    - 关键代码位置（Domain / UseCase / SystemAdapter / Controller / Presenter）。

---

## 7. 给新人的几个小建议

- **从仿写开始**：优先找一个和你要做的功能最像的现有系统（例如做装备 → 看 EquipSys / AttrSys；做副本 → 看 FubenSys；做任务 → 看 QuestSys），照着它的 Domain / UseCase / Adapter 结构“抄一遍”，比从 0 开始设计可靠很多。
- **任何时候不要直接改 Legacy EntitySystem**：
  - 若发现 `entitysystem/*.go` 里还有旧实现，优先在 `adapter/system` 中找对应的 Clean Architecture 版本；
  - 若确实还没迁移，请在文档中标记 TODO，并优先按 Clean Architecture 的方式实现新逻辑，然后再考虑迁移旧逻辑。
- **分阶段提交**：
  - 配置与 Proto 改动尽量单独一条 commit；
  - Domain + UseCase + Adapter 的主逻辑作为一条 commit；
  - 修复与调试可以再分一条 commit，方便回滚和代码审查。

---

## 8. 附：青龙神装开发最小 Checklist（可复制到 Issue 或 PR 中）

- [ ] 在 `docs/服务端开发进度文档.md` 中登记“青龙神装”需求（所属章节 + TODO 条目）
- [ ] 完成配置表设计与解析（含 `ConfigGateway` 接口与实现）
- [ ] 完成 Domain 实体（EquipSet 等）与 UseCase 实现（重算套装加成）
- [ ] 在 SystemAdapter 中挂载生命周期（登录 / 装备变更 / 属性重算）
- [ ]（如需）新增查看/操作青龙神装的协议入口（Controller + Presenter）
- [ ] 属性系统成功计算并下发青龙神装带来的属性变化（含 DungeonServer 同步）
- [ ] 完成本地测试 / lint / build
- [ ] 更新 `docs/服务端开发进度文档.md` 与必要的专题文档


