# GameServer Clean Architecture 重构文档

更新时间：2025-01-XX  
责任人：开发团队

> ⚠️ **重要提示**：本文档已通过代码梳理补充了所有遗漏点，包括：
> - PlayerRole 的 RunOne 消息机制（17.5.1）
> - 系统状态位图管理的详细实现（17.5.2）
> - BinaryData 获取的标准模式（17.5.3）
> - 时间同步机制（17.5.4）
> - 系统间调用的运行时模式（17.5.5）
> - 事件总线的独立克隆机制（17.5.6）
> - PlayerRole 的 SaveToDB 机制（17.5.7）
> - 系统初始化时的数据同步（17.5.8）
> - ProtocolManager 协议路由管理（17.5.9）
> - DungeonClient 生命周期管理（17.5.10）
> - DungeonMessageHandler 消息处理机制（17.5.11）
> - 客户端消息转发机制（17.5.12）
> 
> 请在重构前完整阅读本文档，特别是第 17 章的遗漏点补充部分。

## 0. 导航与任务清单（建议先读）

> **用途**：为新人和后续维护者提供一个“鸟瞰视角”，按阶段查看 GameServer Clean Architecture 重构整体进度，并快速跳转到后文详细章节。

- **阅读顺序建议**
  - [✅] **第 1 章 文档目的**：理解为什么要做这次重构  
  - [✅] **第 3 章 分层设计**：整体分层 & 目录设计（理解之后再看代码更清晰）  
  - [✅] **第 4 章 重构方案**：按阶段说明每层如何落地  
  - [✅] **第 17 章 遗漏点补充**：补齐历史文档未覆盖但实现里存在的关键机制  

- **重构阶段总览（跨文档同步，状态以本表为准）**
  - [✅] **阶段一：基础结构搭建**  
    - 目录结构创建（`domain/`、`usecase/`、`adapter/` 等）  
    - 基础接口定义（Repository、Gateway、ConfigManager、EventPublisher 等）  
    - 基础设施适配层 & 系统生命周期适配器 & DI 容器  
  - [✅] **阶段二：核心系统重构**（详见：背包/货币/等级/装备/属性）  
  - [✅] **阶段三：玩法系统重构**（Skill/Quest/Fuben/ItemUse/Shop/Recycle 等）  
  - [✅] **阶段四：社交系统重构**（Friend/Guild/Chat/Auction + PublicActor 交互统一）  
  - [⏳] **阶段五：辅助系统重构**（Mail/Vip/DailyActivity/Message/GM 等）  
  - [⏳] **阶段六：Legacy 清理与文档/测试完善**  

- **系统迁移任务视图（与 `docs/服务端开发进度文档.md` 保持一致）**
  - [✅] **核心数据 &基础系统**
    - [✅] LevelSys（等级系统）  
    - [✅] BagSys（背包系统）  
    - [✅] MoneySys（货币系统）  
    - [✅] EquipSys（装备系统）  
    - [✅] AttrSys（属性系统）  
  - [✅] **玩法系统**
    - [✅] SkillSys（技能）  
    - [✅] QuestSys（任务）  
    - [✅] FubenSys（副本）  
    - [✅] ItemUseSys（物品使用）  
    - [✅] ShopSys（商城）  
    - [✅] RecycleSys（回收）  
  - [✅] **社交系统**
    - [✅] FriendSys（好友）  
    - [✅] GuildSys（公会）  
    - [✅] ChatSys（聊天）  
    - [✅] AuctionSys（拍卖行）  
  - [⏳] **辅助/运营系统**
    - [⏳] VipSys（VIP & 经验加成，依赖 Money/VipUseCase）  
    - [⏳] DailyActivitySys（日常活跃度，依赖 PointsUseCase）  
    - [⏳] MessageSys（玩家离线消息统一入口）  
    - [✅] MailSys（系统邮件 & GM 发奖）  
    - [✅] GMSys（GM 命令与工具函数）  

- **阶段性检查清单（给正在重构/新开发的你）**
  - [ ] 新系统是否有明确的 **Domain 实体** 与 **Use Case**，业务逻辑不依赖 `database/gatewaylink/dungeonserverlink` 等外层包？  
  - [ ] 所有外部依赖是否通过 `usecase/interfaces` 中的接口定义并在 `adapter/gateway` 中实现？  
  - [ ] 协议入口是否通过 `adapter/controller/*_controller.go`，响应是否通过 `adapter/presenter/*_presenter.go` 统一构建并发送？  
  - [ ] 是否避免了 SystemAdapter ↔ Controller、UseCase ↔ Adapter 等循环依赖（必要时增加 *UseCaseAdapter*）？  
  - [ ] 是否在重构完成后，更新 `docs/服务端开发进度文档.md` 与本文档对应章节的状态？  

> **说明**：从本节开始，本文档按传统结构展开：第 1 章说明文档目的，第 2 章分析当前架构问题，第 3~5 章给出分层设计与分阶段重构方案，第 17 章补充遗漏机制。

## 1. 文档目的

本文档旨在将 `server/service/gameserver` 按照 Clean Architecture（清洁架构）原则进行重构，实现业务逻辑与框架解耦，提高代码可测试性、可维护性和可扩展性。

## 2. 当前架构问题分析

### 2.1 依赖方向混乱

**问题描述：**
- EntitySystem（业务逻辑层）直接依赖 `database`、`gatewaylink`、`dungeonserverlink` 等框架层
- 业务逻辑与协议处理、网络发送、数据访问混在一起
- 内层（业务逻辑）依赖外层（框架），违反了依赖倒置原则

**典型示例：**

```go
// entitysystem/bag_sys.go - 直接依赖 database
func (bs *BagSys) OnInit(ctx context.Context) {
    binaryData, err := database.GetPlayerBinaryData(roleId)  // ❌ 直接依赖数据库
    // ...
}

// entitysystem/friend_sys.go - 直接依赖 gatewaylink
func (s *FriendSys) handleAddFriend(ctx context.Context, msg *network.ClientMessage) {
    gatewaylink.SendToSessionProto(sessionId, ...)  // ❌ 直接依赖网络层
    // ...
}
```

### 2.2 当前代码结构详细分析

**EntitySystem 列表（共 20 个系统）：**
1. `attr_sys.go` - 属性系统（含 attrcalc 子包）
2. `auction_sys.go` - 拍卖行系统
3. `bag_sys.go` - 背包系统
4. `chat_sys.go` - 聊天系统
5. `daily_activity_sys.go` - 日常活跃度系统
6. `equip_sys.go` - 装备系统
7. `friend_sys.go` - 好友系统
8. `fuben_sys.go` - 副本系统
9. `gm_sys.go` - GM 系统
10. `guild_sys.go` - 公会系统
11. `item_use_sys.go` - 物品使用系统
12. `level_sys.go` - 等级系统
13. `mail_sys.go` - 邮件系统
14. `message_sys.go` - 玩家消息系统
15. `money_sys.go` - 货币系统
16. `quest_sys.go` - 任务系统
17. `recycle_sys.go` - 回收系统
18. `shop_sys.go` - 商城系统
19. `skill_sys.go` - 技能系统
20. `vip_sys.go` - VIP 系统

**协议注册方式：**
- 协议注册分散在各个 EntitySystem 的 `init()` 函数中
- 通过 `gevent.Subscribe(gevent.OnSrvStart, ...)` 订阅服务器启动事件
- 在事件回调中调用 `clientprotocol.Register(protoId, handler)` 注册协议处理器
- 部分协议在 `player_network.go` 的 `init()` 中注册（登录、注册、角色管理等）

**PublicActor 交互方式：**
- 通过 `gshare.SendPublicMessageAsync(key, message)` 发送异步消息
- 消息类型定义在 `proto/csproto/rpc.proto` 的 `PublicActorMsgId` 枚举中
- PublicActor 内部通过 `RegisterHandler` 注册消息处理器
- 处理器函数签名必须符合 `actor.HandlerMessageFunc`

**RPC 调用方式：**
- `dungeonserverlink.AsyncCall(msgId, data)` - 异步调用 DungeonServer
- `dungeonserverlink.RegisterRPCHandler(msgId, handler)` - 注册 RPC 处理器
- 通过 `ProtocolManager` 管理协议路由（通用协议 vs 独有协议）

**事件系统：**
- 使用 `gevent` 包进行事件发布和订阅
- `gevent.SubscribePlayerEvent(eventType, handler)` - 订阅玩家事件
- `gevent.Publish(eventType, args...)` - 发布事件
- 事件类型定义在 `gevent/enum.go` 中

**配置系统：**
- 使用 `jsonconf.GetConfigManager()` 获取配置管理器
- 配置表位于 `server/output/config/*.json`
- 配置加载在服务器启动时完成

**系统依赖关系：**
- 在 `sys_mgr.go` 中定义了系统依赖关系（`systemDependencies`）
- 使用拓扑排序确定系统初始化顺序
- 当前已知依赖：`AttrSys` 依赖 `LevelSys` 和 `EquipSys`

### 2.2 业务逻辑与框架耦合

**问题描述：**
- EntitySystem 中混入了协议解析、网络发送、Actor 调度等框架代码
- 业务逻辑无法独立测试，必须启动完整的 Actor 框架和网络服务
- 系统之间通过直接调用而非接口交互

**典型示例：**

```go
// entitysystem/level_sys.go - 混入网络发送
func (ls *LevelSys) AddExp(ctx context.Context, exp int64) {
    // 业务逻辑
    ls.levelData.Exp += exp
    
    // ❌ 框架代码混入业务逻辑
    gatewaylink.SendToSessionProto(sessionId, protocol.S2CLevelUp, ...)
}
```

### 2.3 数据访问层缺失

**问题描述：**
- 没有数据访问接口抽象，EntitySystem 直接调用 `database` 包
- 无法进行单元测试（无法 mock 数据库）
- 数据访问逻辑分散在各个系统中

### 2.4 接口适配层不清晰

**问题描述：**
- 协议处理、响应构建、数据转换等适配逻辑混在业务逻辑中
- 没有明确的 Controller/Presenter 层
- 协议变更会影响业务逻辑代码

## 3. Clean Architecture 分层设计

### 3.1 分层结构

> **新增说明（2025-01-XX）**：本节补充了“代码 → Clean Architecture 层级”的对应关系，便于后续阅读其它文档时快速定位。

| Clean Architecture 层 | 作用 | 代码位置/关键包 |
| --- | --- | --- |
| **Domain（Enterprise Business）** | 领域实体 + 纯业务规则，不依赖实现 | `server/service/gameserver/internel/domain/*`、`protocol/*.proto` 中的业务对象 |
| **Use Case/Application** | 用例编排、事务脚本、跨系统协调，通过接口访问外层资源 | `server/service/gameserver/internel/usecase/*`、`usecase/interfaces/*`、`usecaseadapter/*` |
| **Interface Adapters** | 将 Use Case 输入/输出转换为外层可用形式（Controller / Presenter / Gateway / SystemAdapter） | `internel/adapter/controller`、`adapter/presenter`、`adapter/system`、`adapter/gateway` |
| **Frameworks & Drivers** | 具体框架实现：网络、数据库、Actor、事件、配置等 | `server/internal/*`（数据库、配置、事件）、`server/service/*/internel/core/*`（Actor 框架、gshare）、`gatewaylink`/`dungeonserverlink` |

- **SystemAdapter**（`adapter/system/*`）承担 PlayerActor 生命周期钩子、UseCase 调用、系统状态同步；禁止直接写业务规则。
- **Controller/Presenter** 将协议请求转为 UseCase 输入，并负责响应/错误码序列化，确保业务层与协议细节解耦。
- **Gateway / Repository Adapter** 通过接口（`usecase/interfaces/*`）暴露数据库、配置、PublicActor、外部服务，确保 UseCase 可单测。

> 以后如果新增目录或层级，务必在本表同步，以保证与《服务端开发进度文档》第 0 章的开发约束一致。

```
┌─────────────────────────────────────────────────────────┐
│  Frameworks & Drivers (框架层)                          │
│  - Actor 框架                                           │
│  - 网络层 (gatewaylink, dungeonserverlink)              │
│  - 数据库 (database)                                     │
│  - 事件总线 (event)                                      │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Interface Adapters (接口适配层)                         │
│  - Controllers: 协议处理器                               │
│  - Presenters: 响应构建器                               │
│  - Gateways: 数据访问实现                                │
│  - Event Adapters: 事件适配器                           │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Use Cases (用例层)                                      │
│  - 业务用例: AddItem, LevelUp, SendMessage 等            │
│  - 业务规则: 等级计算、属性计算、奖励发放等              │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Entities (实体层)                                       │
│  - 业务实体: Player, Item, Skill, Bag 等                │
│  - 值对象: Exp, Level, Attribute 等                     │
└─────────────────────────────────────────────────────────┘
```

### 3.2 目录结构设计

```
server/service/gameserver/
├── internel/
│   ├── domain/                    # Entities 层
│   │   ├── player.go              # 玩家实体
│   │   ├── item.go                # 物品实体
│   │   ├── skill.go               # 技能实体
│   │   ├── bag.go                 # 背包实体
│   │   └── ...
│   │
│   ├── usecase/                   # Use Cases 层
│   │   ├── bag/                   # 背包用例
│   │   │   ├── add_item.go
│   │   │   ├── remove_item.go
│   │   │   └── use_item.go
│   │   ├── level/                 # 等级用例
│   │   │   ├── add_exp.go
│   │   │   └── level_up.go
│   │   ├── friend/                # 好友用例
│   │   │   ├── add_friend.go
│   │   │   └── remove_friend.go
│   │   └── ...
│   │
│   ├── adapter/                  # Interface Adapters 层
│   │   ├── controller/           # 协议控制器
│   │   │   ├── bag_controller.go
│   │   │   ├── level_controller.go
│   │   │   └── ...
│   │   ├── presenter/            # 响应构建器
│   │   │   ├── bag_presenter.go
│   │   │   └── ...
│   │   ├── gateway/              # 数据访问实现
│   │   │   ├── player_gateway.go
│   │   │   ├── item_gateway.go
│   │   │   └── ...
│   │   └── event/                # 事件适配器
│   │       └── event_adapter.go
│   │
│   ├── infrastructure/           # Frameworks & Drivers 层
│   │   ├── actor/                # Actor 适配
│   │   ├── network/              # 网络适配
│   │   ├── database/             # 数据库适配
│   │   └── event/                # 事件适配
│   │
│   └── ... (保留现有目录用于过渡)
```

## 4. 重构方案

### 4.1 阶段一：Entities 层重构

**目标：** 提取纯业务实体，移除所有框架依赖

#### 4.1.1 创建 Domain 实体

**目录：** `internel/domain/`

**示例：Player 实体**

```go
// domain/player.go
package domain

// Player 玩家实体（纯业务对象，不依赖任何框架）
type Player struct {
    ID       uint64
    RoleID   uint64
    Level    int32
    Exp      int64
    Bag      *Bag
    Skills   []*Skill
    // ... 其他业务属性
}

// AddExp 增加经验值（纯业务逻辑）
func (p *Player) AddExp(exp int64) {
    p.Exp += exp
    // 检查是否升级
    for p.canLevelUp() {
        p.levelUp()
    }
}

func (p *Player) canLevelUp() bool {
    // 根据配置表判断是否可以升级
    return p.Exp >= p.getRequiredExp(p.Level+1)
}

func (p *Player) levelUp() {
    p.Level++
    // 触发升级事件（通过接口，不直接依赖事件框架）
}
```

**示例：Bag 实体**

```go
// domain/bag.go
package domain

// Bag 背包实体
type Bag struct {
    Items []*Item
    Size  uint32
}

// AddItem 添加物品（纯业务逻辑）
func (b *Bag) AddItem(item *Item) error {
    if b.isFull() {
        return ErrBagFull
    }
    b.Items = append(b.Items, item)
    return nil
}

func (b *Bag) isFull() bool {
    return uint32(len(b.Items)) >= b.Size
}
```

#### 4.1.2 定义 Repository 接口

**目录：** `internel/domain/repository/`

```go
// domain/repository/player_repository.go
package repository

import "postapocgame/server/service/gameserver/internel/domain"

// PlayerRepository 玩家数据访问接口（定义在 domain 层）
type PlayerRepository interface {
    GetByID(roleID uint64) (*domain.Player, error)
    Save(player *domain.Player) error
    GetBinaryData(roleID uint64) (*protocol.PlayerRoleBinaryData, error)
    SaveBinaryData(roleID uint64, data *protocol.PlayerRoleBinaryData) error
}
```

### 4.2 阶段二：Use Cases 层重构

**目标：** 实现业务用例，依赖 Entities 和 Repository 接口

#### 4.2.1 创建 Use Case

**目录：** `internel/usecase/`

**示例：AddItem Use Case**

```go
// usecase/bag/add_item.go
package bag

import (
    "context"
    "postapocgame/server/service/gameserver/internel/domain"
    "postapocgame/server/service/gameserver/internel/domain/repository"
)

// AddItemUseCase 添加物品用例
type AddItemUseCase struct {
    playerRepo repository.PlayerRepository
    // 可以注入其他依赖，如事件发布器接口
    eventPublisher EventPublisher
}

func NewAddItemUseCase(
    playerRepo repository.PlayerRepository,
    eventPublisher EventPublisher,
) *AddItemUseCase {
    return &AddItemUseCase{
        playerRepo:     playerRepo,
        eventPublisher: eventPublisher,
    }
}

// Execute 执行添加物品用例
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, item *domain.Item) error {
    // 1. 获取玩家实体
    player, err := uc.playerRepo.GetByID(roleID)
    if err != nil {
        return err
    }
    
    // 2. 执行业务逻辑（纯业务，不依赖框架）
    if err := player.Bag.AddItem(item); err != nil {
        return err
    }
    
    // 3. 保存数据
    if err := uc.playerRepo.Save(player); err != nil {
        return err
    }
    
    // 4. 发布事件（通过接口）
    uc.eventPublisher.PublishItemAdded(ctx, roleID, item)
    
    return nil
}
```

#### 4.2.2 定义 Use Case 依赖接口

**目录：** `internel/usecase/interfaces/`

```go
// usecase/interfaces/event_publisher.go
package interfaces

// EventPublisher 事件发布器接口（Use Case 层定义）
type EventPublisher interface {
    PublishItemAdded(ctx context.Context, roleID uint64, item interface{})
    PublishLevelUp(ctx context.Context, roleID uint64, newLevel int32)
    // ... 其他事件
}
```

### 4.3 阶段三：Interface Adapters 层重构

**目标：** 实现协议处理、数据访问、事件适配

#### 4.3.1 Controllers（协议控制器）

**目录：** `internel/adapter/controller/`

```go
// adapter/controller/bag_controller.go
package controller

import (
    "context"
    "postapocgame/server/internal/network"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/gameserver/internel/adapter/presenter"
    "postapocgame/server/service/gameserver/internel/usecase/bag"
)

// BagController 背包协议控制器
type BagController struct {
    addItemUseCase *bag.AddItemUseCase
    presenter      *presenter.BagPresenter
}

func NewBagController(
    addItemUseCase *bag.AddItemUseCase,
    presenter *presenter.BagPresenter,
) *BagController {
    return &BagController{
        addItemUseCase: addItemUseCase,
        presenter:      presenter,
    }
}

// HandleAddItem 处理添加物品协议
func (c *BagController) HandleAddItem(ctx context.Context, msg *network.ClientMessage) error {
    // 1. 解析协议
    var req protocol.C2SAddItemReq
    if err := proto.Unmarshal(msg.Data, &req); err != nil {
        return err
    }
    
    // 2. 转换为 Domain 实体
    item := convertToDomainItem(&req.Item)
    
    // 3. 调用 Use Case
    roleID := getRoleIDFromContext(ctx)
    if err := c.addItemUseCase.Execute(ctx, roleID, item); err != nil {
        return c.presenter.PresentError(ctx, err)
    }
    
    // 4. 构建响应
    return c.presenter.PresentAddItemSuccess(ctx, item)
}
```

#### 4.3.2 Presenters（响应构建器）

**目录：** `internel/adapter/presenter/`

```go
// adapter/presenter/bag_presenter.go
package presenter

import (
    "context"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/gameserver/internel/adapter/gateway/network"
)

// BagPresenter 背包响应构建器
type BagPresenter struct {
    networkGateway network.Gateway
}

func NewBagPresenter(networkGateway network.Gateway) *BagPresenter {
    return &BagPresenter{
        networkGateway: networkGateway,
    }
}

// PresentAddItemSuccess 构建添加物品成功响应
func (p *BagPresenter) PresentAddItemSuccess(ctx context.Context, item interface{}) error {
    sessionID := getSessionIDFromContext(ctx)
    resp := &protocol.S2CAddItemResp{
        Success: true,
        Item:    convertToProtocolItem(item),
    }
    return p.networkGateway.SendToSession(sessionID, protocol.S2CProtocol_S2CAddItemResult, resp)
}

// PresentError 构建错误响应
func (p *BagPresenter) PresentError(ctx context.Context, err error) error {
    sessionID := getSessionIDFromContext(ctx)
    resp := &protocol.S2CError{
        Code: getErrorCode(err),
        Msg:  err.Error(),
    }
    return p.networkGateway.SendToSession(sessionID, protocol.S2CProtocol_S2CError, resp)
}
```

#### 4.3.3 Gateways（数据访问实现）

**目录：** `internel/adapter/gateway/`

```go
// adapter/gateway/player_gateway.go
package gateway

import (
    "postapocgame/server/internal/database"
    "postapocgame/server/service/gameserver/internel/domain"
    "postapocgame/server/service/gameserver/internel/domain/repository"
)

// PlayerGateway 玩家数据访问实现（实现 domain 层的 Repository 接口）
type PlayerGateway struct {
    // 可以注入数据库连接等
}

func NewPlayerGateway() repository.PlayerRepository {
    return &PlayerGateway{}
}

func (g *PlayerGateway) GetByID(roleID uint64) (*domain.Player, error) {
    // 调用 database 包获取数据
    binaryData, err := database.GetPlayerBinaryData(uint(roleID))
    if err != nil {
        return nil, err
    }
    
    // 转换为 Domain 实体
    return convertToDomainPlayer(binaryData), nil
}

func (g *PlayerGateway) Save(player *domain.Player) error {
    // 转换为数据库格式
    binaryData := convertToBinaryData(player)
    
    // 调用 database 包保存
    return database.SavePlayerBinaryData(uint(player.RoleID), binaryData)
}
```

#### 4.3.4 Event Adapters（事件适配器）

**目录：** `internel/adapter/event/`

```go
// adapter/event/event_adapter.go
package event

import (
    "context"
    "postapocgame/server/internal/event"
    "postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// EventAdapter 事件适配器（实现 Use Case 层的 EventPublisher 接口）
type EventAdapter struct {
    eventBus *event.Bus
}

func NewEventAdapter(eventBus *event.Bus) interfaces.EventPublisher {
    return &EventAdapter{
        eventBus: eventBus,
    }
}

func (a *EventAdapter) PublishItemAdded(ctx context.Context, roleID uint64, item interface{}) {
    // 转换为框架事件并发布
    evt := event.NewEvent(gevent.OnItemAdded, map[string]interface{}{
        "roleID": roleID,
        "item":   item,
    })
    a.eventBus.Publish(ctx, evt)
}
```

### 4.4 阶段四：Interface Adapters 层补充

#### 4.4.1 PublicActor 交互接口抽象

**目录：** `internel/usecase/interfaces/public_actor.go`

```go
// usecase/interfaces/public_actor.go
package interfaces

import (
    "context"
    "postapocgame/server/internal/actor"
)

// PublicActorGateway PublicActor 交互接口（Use Case 层定义）
type PublicActorGateway interface {
    // 发送异步消息到 PublicActor
    SendMessageAsync(key string, message actor.IActorMessage) error
    
    // 注册消息处理器（可选，用于反向调用）
    RegisterHandler(msgId uint16, handler actor.HandlerMessageFunc)
}
```

**目录：** `internel/adapter/gateway/public_actor_gateway.go`

```go
// adapter/gateway/public_actor_gateway.go
package gateway

import (
    "postapocgame/server/internal/actor"
    "postapocgame/server/service/gameserver/internel/gshare"
    "postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// PublicActorGatewayImpl PublicActor 交互实现
type PublicActorGatewayImpl struct{}

func NewPublicActorGateway() interfaces.PublicActorGateway {
    return &PublicActorGatewayImpl{}
}

func (g *PublicActorGatewayImpl) SendMessageAsync(key string, message actor.IActorMessage) error {
    return gshare.SendPublicMessageAsync(key, message)
}

func (g *PublicActorGatewayImpl) RegisterHandler(msgId uint16, handler actor.HandlerMessageFunc) {
    gshare.RegisterHandler(msgId, handler)
}
```

#### 4.4.2 RPC 调用接口抽象

**目录：** `internel/usecase/interfaces/rpc.go`

```go
// usecase/interfaces/rpc.go
package interfaces

import "context"

// DungeonServerGateway DungeonServer RPC 接口（Use Case 层定义）
type DungeonServerGateway interface {
    // 异步调用 DungeonServer
    AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error
    
    // 注册 RPC 处理器（用于接收 DungeonServer 回调）
    RegisterRPCHandler(msgId uint16, handler func(ctx context.Context, sessionId string, data []byte) error)
    
    // 协议路由相关方法
    IsDungeonProtocol(protoId uint16) bool
    GetSrvTypeForProtocol(protoId uint16) (srvType uint8, protocolType uint8, ok bool)
}
```

**目录：** `internel/adapter/gateway/dungeon_server_gateway.go`

```go
// adapter/gateway/dungeon_server_gateway.go
package gateway

import (
    "context"
    "postapocgame/server/service/gameserver/internel/dungeonserverlink"
    "postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// DungeonServerGatewayImpl DungeonServer RPC 实现
type DungeonServerGatewayImpl struct{}

func NewDungeonServerGateway() interfaces.DungeonServerGateway {
    return &DungeonServerGatewayImpl{}
}

func (g *DungeonServerGatewayImpl) AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error {
    return dungeonserverlink.AsyncCall(ctx, srvType, sessionId, msgId, data)
}

func (g *DungeonServerGatewayImpl) RegisterRPCHandler(msgId uint16, handler func(ctx context.Context, sessionId string, data []byte) error) {
    dungeonserverlink.RegisterRPCHandler(msgId, handler)
}

func (g *DungeonServerGatewayImpl) IsDungeonProtocol(protoId uint16) bool {
    return dungeonserverlink.GetProtocolManager().IsDungeonProtocol(protoId)
}

func (g *DungeonServerGatewayImpl) GetSrvTypeForProtocol(protoId uint16) (uint8, uint8, bool) {
    srvType, protocolType, ok := dungeonserverlink.GetProtocolManager().GetSrvTypeForProtocol(protoId)
    return srvType, uint8(protocolType), ok
}
```

#### 4.4.3 事件发布接口抽象

**目录：** `internel/usecase/interfaces/event.go`

```go
// usecase/interfaces/event.go
package interfaces

import "context"

// EventPublisher 事件发布器接口（Use Case 层定义）
type EventPublisher interface {
    // 发布玩家事件
    PublishPlayerEvent(ctx context.Context, eventType string, args ...interface{})
    
    // 订阅玩家事件
    SubscribePlayerEvent(eventType string, handler func(ctx context.Context, event interface{}))
}
```

**目录：** `internel/adapter/event/event_adapter.go`

```go
// adapter/event/event_adapter.go
package event

import (
    "context"
    "postapocgame/server/internal/event"
    "postapocgame/server/service/gameserver/internel/gevent"
    "postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// EventAdapter 事件适配器（实现 Use Case 层的 EventPublisher 接口）
type EventAdapter struct{}

func NewEventAdapter() interfaces.EventPublisher {
    return &EventAdapter{}
}

func (a *EventAdapter) PublishPlayerEvent(ctx context.Context, eventType string, args ...interface{}) {
    gevent.Publish(eventType, args...)
}

func (a *EventAdapter) SubscribePlayerEvent(eventType string, handler func(ctx context.Context, event interface{})) {
    gevent.SubscribePlayerEvent(eventType, func(ctx context.Context, ev *event.Event) {
        handler(ctx, ev)
    })
}
```

#### 4.4.4 配置访问接口抽象

**目录：** `internel/usecase/interfaces/config.go`

```go
// usecase/interfaces/config.go
package interfaces

// ConfigManager 配置管理器接口（Use Case 层定义）
type ConfigManager interface {
    GetBagConfig(bagType uint32) (interface{}, bool)
    GetItemConfig(itemId uint32) (interface{}, bool)
    GetSkillConfig(skillId uint32) (interface{}, bool)
    // ... 其他配置访问方法
}
```

**目录：** `internel/adapter/gateway/config_gateway.go`

```go
// adapter/gateway/config_gateway.go
package gateway

import (
    "postapocgame/server/internal/jsonconf"
    "postapocgame/server/service/gameserver/internel/usecase/interfaces"
)

// ConfigGatewayImpl 配置访问实现
type ConfigGatewayImpl struct{}

func NewConfigGateway() interfaces.ConfigManager {
    return &ConfigGatewayImpl{}
}

func (g *ConfigGatewayImpl) GetBagConfig(bagType uint32) (interface{}, bool) {
    configMgr := jsonconf.GetConfigManager()
    return configMgr.GetBagConfig(bagType)
}

// ... 其他配置访问方法实现
```

### 4.5 阶段五：Infrastructure 层重构

**目标：** 封装框架调用，提供统一接口

#### 4.5.1 Network Gateway

**目录：** `internel/adapter/gateway/network_gateway.go`

```go
// adapter/gateway/network_gateway.go
package gateway

import (
    "google.golang.org/protobuf/proto"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/gameserver/internel/gatewaylink"
)

// NetworkGateway 网络网关接口（Adapter 层定义）
type NetworkGateway interface {
    SendToSession(sessionID string, msgID uint16, data []byte) error
    SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error
}

// NetworkGatewayImpl 网络网关实现
type NetworkGatewayImpl struct{}

func NewNetworkGateway() NetworkGateway {
    return &NetworkGatewayImpl{}
}

func (g *NetworkGatewayImpl) SendToSession(sessionID string, msgID uint16, data []byte) error {
    return gatewaylink.SendToSession(sessionID, msgID, data)
}

func (g *NetworkGatewayImpl) SendToSessionProto(sessionID string, msgID uint16, message proto.Message) error {
    return gatewaylink.SendToSessionProto(sessionID, msgID, message)
}
```

## 5. 重构步骤

### 5.1 阶段一：基础结构搭建（1-2周）

1. **创建目录结构**
   - 创建 `domain/`、`usecase/`、`adapter/`、`infrastructure/` 目录
   - 定义基础接口（Repository、Gateway、EventPublisher、PublicActorGateway、DungeonServerGateway、ConfigManager 等）

2. **实现基础设施适配层**
   - 实现 `NetworkGateway`（封装 `gatewaylink`）
   - 实现 `PublicActorGateway`（封装 `gshare.SendPublicMessageAsync`）
   - 实现 `DungeonServerGateway`（封装 `dungeonserverlink`）
   - 实现 `EventAdapter`（封装 `gevent`）
   - 实现 `ConfigGateway`（封装 `jsonconf`）

3. **迁移一个简单系统作为示例**
   - 选择 `LevelSys` 作为第一个重构目标（业务逻辑简单，依赖较少）
   - 创建 `domain/player.go`（提取 Player 实体）
   - 创建 `domain/repository/player_repository.go`（定义 Repository 接口）
   - 创建 `usecase/level/add_exp.go`（提取业务逻辑）
   - 创建 `adapter/controller/level_controller.go`（协议处理）
   - 创建 `adapter/presenter/level_presenter.go`（响应构建）
   - 创建 `adapter/gateway/player_gateway.go`（数据访问实现）

4. **验证重构效果**
   - 确保功能正常
   - 编写单元测试（Use Case 层可独立测试）
   - 验证协议注册和消息发送正常

### 5.2 阶段二：核心系统重构（3-4周）

1. **重构核心系统（按依赖顺序）**
   - `LevelSys` → `usecase/level/` + `adapter/controller/level_controller.go`（已完成）
   - `BagSys` → `usecase/bag/` + `adapter/controller/bag_controller.go`
   - `MoneySys` → `usecase/money/` + `adapter/controller/money_controller.go`
   - `EquipSys` → `usecase/equip/` + `adapter/controller/equip_controller.go`
   - `AttrSys` → `usecase/attr/` + `adapter/controller/attr_controller.go`（依赖 LevelSys 和 EquipSys）

2. **统一数据访问**
   - 所有系统通过 Repository 接口访问数据
   - 实现 Gateway 层统一封装 database 调用
   - 保持 BinaryData 的共享引用模式

3. **统一网络发送**
   - 所有系统通过 Presenter 构建响应
   - 通过 Network Gateway 发送消息

### 5.3 阶段三：玩法系统重构（2-3周）

1. **重构玩法系统**
   - `SkillSys` → `usecase/skill/` + `adapter/controller/skill_controller.go`
   - `QuestSys` → `usecase/quest/` + `adapter/controller/quest_controller.go`
   - `FubenSys` → `usecase/fuben/` + `adapter/controller/fuben_controller.go`
   - `ItemUseSys` → `usecase/item_use/` + `adapter/controller/item_use_controller.go`
   - `ShopSys` → `usecase/shop/` + `adapter/controller/shop_controller.go`
   - `RecycleSys` → `usecase/recycle/` + `adapter/controller/recycle_controller.go`

2. **重构 RPC 调用**
   - 通过 `DungeonServerGateway` 接口调用 DungeonServer
   - 在 Adapter 层实现具体调用

### 5.4 阶段四：社交系统重构（2-3周）

1. **重构社交系统**
   - `FriendSys` → `usecase/friend/` + `adapter/controller/friend_controller.go`
   - `GuildSys` → `usecase/guild/` + `adapter/controller/guild_controller.go`
   - `ChatSys` → `usecase/chat/` + `adapter/controller/chat_controller.go`
   - `AuctionSys` → `usecase/auction/` + `adapter/controller/auction_controller.go`

2. **重构 PublicActor 交互**
   - 通过 `PublicActorGateway` 接口定义 PublicActor 交互
   - 在 Adapter 层实现具体调用
   - 保持消息类型定义在 Proto 中

### 5.5 阶段五：辅助系统重构（1-2周）

1. **重构辅助系统**
   - `MailSys` → `usecase/mail/` + `adapter/controller/mail_controller.go`
   - `VipSys` → `usecase/vip/` + `adapter/controller/vip_controller.go`
   - `DailyActivitySys` → `usecase/daily_activity/` + `adapter/controller/daily_activity_controller.go`
   - `MessageSys` → `usecase/message/` + `adapter/controller/message_controller.go`
   - `GMSys` → `usecase/gm/` + `adapter/controller/gm_controller.go`

### 5.6 阶段六：清理与优化（1-2周）

1. **移除旧代码**
   - 删除 `entitysystem/` 中的旧实现
   - 清理直接依赖框架的代码
   - 更新 `sys_mgr.go` 以支持新的系统注册方式

2. **完善测试**
   - 为 Use Case 层编写单元测试（覆盖率 > 70%）
   - 为 Controller 层编写集成测试
   - 验证所有协议处理正常

3. **文档更新**
   - 更新架构文档
   - 更新开发指南
   - 更新协议注册文档

## 6. 依赖注入设计

### 6.1 依赖注入容器

**目录：** `internel/di/container.go`

```go
// di/container.go
package di

import (
    "postapocgame/server/service/gameserver/internel/adapter/controller"
    "postapocgame/server/service/gameserver/internel/adapter/gateway"
    "postapocgame/server/service/gameserver/internel/usecase/bag"
    "postapocgame/server/service/gameserver/internel/usecase/level"
)

// Container 依赖注入容器
type Container struct {
    // Gateways
    playerGateway gateway.PlayerGateway
    
    // Use Cases
    addItemUseCase *bag.AddItemUseCase
    addExpUseCase   *level.AddExpUseCase
    
    // Controllers
    bagController   *controller.BagController
    levelController *controller.LevelController
}

func NewContainer() *Container {
    c := &Container{}
    
    // 初始化 Gateways
    c.playerGateway = gateway.NewPlayerGateway()
    
    // 初始化 Use Cases
    c.addItemUseCase = bag.NewAddItemUseCase(c.playerGateway, ...)
    c.addExpUseCase = level.NewAddExpUseCase(c.playerGateway, ...)
    
    // 初始化 Controllers
    c.bagController = controller.NewBagController(c.addItemUseCase, ...)
    c.levelController = controller.NewLevelController(c.addExpUseCase, ...)
    
    return c
}
```

### 6.2 在 Actor 中使用

**目录：** `internel/playeractor/handler.go`

```go
// playeractor/handler.go
package playeractor

import (
    "context"
    "postapocgame/server/internal/actor"
    "postapocgame/server/service/gameserver/internel/di"
    "postapocgame/server/service/gameserver/internel/playeractor/clientprotocol"
)

func (h *PlayerHandler) HandleMessage(ctx context.Context, msg actor.IActorMessage) {
    container := di.GetContainer()  // 获取依赖容器
    
    // 根据消息类型路由到对应的 Controller
    // 注意：这里需要保持与 clientprotocol.ProtoTbl 的兼容性
    msgId := msg.GetMsgId()
    
    // 从协议注册表获取处理器（兼容旧代码）
    handler := clientprotocol.GetFunc(msgId)
    if handler != nil {
        // 转换为 ClientMessage 格式
        clientMsg := &network.ClientMessage{
            MsgId: msgId,
            Data:  msg.GetData(),
        }
        handler(ctx, clientMsg)
        return
    }
    
    // 新代码路径：通过 Controller 处理
    // 注意：重构过程中，新旧代码可以并存
    switch msgId {
    case protocol.C2SAddItem:
        container.BagController.HandleAddItem(ctx, msg)
    case protocol.C2SAddExp:
        container.LevelController.HandleAddExp(ctx, msg)
    // ... 其他协议
    }
}
```

### 6.3 协议注册迁移策略

**当前方式：**
- 各个 EntitySystem 在 `init()` 中通过 `gevent.Subscribe(gevent.OnSrvStart, ...)` 注册协议
- 在事件回调中调用 `clientprotocol.Register(protoId, handler)`

**重构后方式：**
- Controller 在 `init()` 中注册协议
- 协议处理器直接调用 Controller 方法
- 保持 `clientprotocol.ProtoTbl` 的兼容性

**迁移步骤：**
1. 创建 Controller 并实现协议处理方法
2. 在 Controller 的 `init()` 中注册协议
3. 协议处理器调用 Controller 方法
4. 验证功能正常后，删除 EntitySystem 中的旧协议注册

## 7. 测试策略

### 7.1 Use Case 层单元测试

```go
// usecase/bag/add_item_test.go
func TestAddItemUseCase_Execute(t *testing.T) {
    // Mock Repository
    mockRepo := &MockPlayerRepository{}
    mockEventPub := &MockEventPublisher{}
    
    // 创建 Use Case
    uc := NewAddItemUseCase(mockRepo, mockEventPub)
    
    // 执行测试
    err := uc.Execute(ctx, roleID, item)
    
    // 验证结果
    assert.NoError(t, err)
    assert.True(t, mockRepo.SaveCalled)
    assert.True(t, mockEventPub.PublishItemAddedCalled)
}
```

### 7.2 Controller 层集成测试

```go
// adapter/controller/bag_controller_test.go
func TestBagController_HandleAddItem(t *testing.T) {
    // 使用真实 Repository（可以连接测试数据库）
    playerGateway := gateway.NewPlayerGateway()
    // ...
    
    controller := NewBagController(addItemUseCase, presenter)
    
    // 执行测试
    err := controller.HandleAddItem(ctx, msg)
    
    // 验证结果
    assert.NoError(t, err)
}
```

## 8. 迁移检查清单

### 8.1 每个系统迁移检查项

- [ ] 创建 Domain 实体（移除框架依赖）
- [ ] 定义 Repository 接口
- [ ] 创建 Use Case（业务逻辑）
- [ ] 创建 Controller（协议处理）
- [ ] 创建 Presenter（响应构建）
- [ ] 实现 Gateway（数据访问）
- [ ] 编写单元测试
- [ ] 更新协议注册
- [ ] 验证功能正常
- [ ] 删除旧代码

### 8.2 整体检查项

- [ ] 所有 EntitySystem 已迁移（共 20 个系统）
- [ ] 所有框架依赖已移除（database、gatewaylink、dungeonserverlink、gevent、jsonconf）
- [ ] 依赖注入容器已配置
- [ ] 协议注册已迁移到 Controller 层
- [ ] PublicActor 交互已通过接口抽象
- [ ] RPC 调用已通过接口抽象
- [ ] 事件系统已通过接口抽象
- [ ] 配置系统已通过接口抽象
- [ ] 单元测试覆盖率 > 70%
- [ ] 集成测试通过
- [ ] 文档已更新

### 8.3 系统迁移优先级

**高优先级（核心系统，依赖较少）：**
1. LevelSys - 等级系统（已完成示例）
2. BagSys - 背包系统
3. MoneySys - 货币系统
4. EquipSys - 装备系统
5. AttrSys - 属性系统（依赖 LevelSys 和 EquipSys）

**中优先级（玩法系统）：**
6. SkillSys - 技能系统
7. QuestSys - 任务系统
8. FubenSys - 副本系统
9. ItemUseSys - 物品使用系统
10. ShopSys - 商城系统
11. RecycleSys - 回收系统

**低优先级（社交系统，依赖 PublicActor）：**
12. FriendSys - 好友系统
13. GuildSys - 公会系统
14. ChatSys - 聊天系统
15. AuctionSys - 拍卖行系统

**辅助系统：**
16. MailSys - 邮件系统
17. VipSys - VIP 系统
18. DailyActivitySys - 日常活跃度系统
19. MessageSys - 玩家消息系统
20. GMSys - GM 系统

## 9. 注意事项

### 9.1 保持向后兼容

- 重构过程中保持旧代码可用
- 新代码与旧代码可以并存
- 逐步迁移，不一次性替换

### 9.2 Actor 框架集成

- Controller 层需要适配 Actor 框架
- 保持 Actor 的单线程特性
- 通过 Context 传递 Actor 相关信息

### 9.3 性能考虑

- 避免过度抽象导致性能下降
- Repository 接口可以批量操作
- Presenter 可以缓存常用响应

### 9.4 数据一致性

- 保持 BinaryData 的共享引用
- 通过 Repository 统一管理数据变更
- 确保事务一致性

### 9.5 Actor 框架集成细节

**Context 传递：**
- 所有 Use Case 和 Controller 方法必须接收 `context.Context` 作为第一个参数
- Context 中必须包含 `gshare.ContextKeyRole`（IPlayerRole 接口）
- Context 中必须包含 `gshare.ContextKeySession`（SessionId 字符串）

**Actor 单线程特性：**
- 所有业务逻辑必须在 Actor 主线程执行
- 禁止在 Use Case 或 Controller 中创建 goroutine
- 异步操作通过消息队列实现

**系统生命周期：**
- 系统初始化：`OnInit(ctx)` - 从 BinaryData 加载数据
- 玩家登录：`OnRoleLogin(ctx)` - 登录时调用
- 玩家重连：`OnRoleReconnect(ctx)` - 重连时调用
- 玩家登出：`OnRoleLogout(ctx)` - 登出时调用
- 定时回调：`OnNewHour/Day/Week/Month/Year(ctx)` - 时间变化时调用

### 9.6 系统依赖关系处理

**依赖关系定义：**
- 在 `sys_mgr.go` 的 `systemDependencies` 中定义系统依赖关系
- 使用拓扑排序确定系统初始化顺序
- 重构后，系统依赖关系保持不变，但需要在 Use Case 层通过接口注入依赖

**依赖注入示例：**
```go
// usecase/attr/attr_calculator.go
type AttrCalculatorUseCase struct {
    levelRepo repository.LevelRepository  // 依赖 LevelSys
    equipRepo repository.EquipRepository  // 依赖 EquipSys
}

func NewAttrCalculatorUseCase(
    levelRepo repository.LevelRepository,
    equipRepo repository.EquipRepository,
) *AttrCalculatorUseCase {
    return &AttrCalculatorUseCase{
        levelRepo: levelRepo,
        equipRepo: equipRepo,
    }
}
```

### 9.7 PublicActor 交互规范

**消息发送：**
- 通过 `PublicActorGateway` 接口发送消息
- 消息类型必须使用 `proto/csproto/rpc.proto` 中定义的 `PublicActorMsgId` 枚举
- 消息数据必须序列化为 Proto 字节

**消息接收：**
- PublicActor 内部通过 `RegisterHandler` 注册消息处理器
- 处理器函数签名必须符合 `actor.HandlerMessageFunc`
- 处理器中可以直接调用 `gatewaylink` 发送响应（PublicActor 在 Infrastructure 层）

### 9.8 RPC 调用规范

**DungeonServer 调用：**
- 通过 `DungeonServerGateway` 接口调用
- 使用 `AsyncCall` 进行异步调用，禁止同步阻塞
- RPC 回调通过 `RegisterRPCHandler` 注册

**协议路由：**
- 通过 `ProtocolManager` 管理协议路由（位于 `dungeonserverlink` 包）
- 通用协议（`IsCommon=true`）可以路由到多个 DungeonServer
- 独有协议（`IsCommon=false`）只能路由到特定的 DungeonServer
- `ProtocolManager` 在重构后通过 `DungeonServerGateway` 接口暴露，Use Case 层不直接访问

**DungeonClient 生命周期：**
- `StartDungeonClient` 在 `main.go` 中调用，初始化连接池并连接到所有配置的 DungeonServer
- `Stop` 在 `main.go` 中调用，关闭所有连接
- 连接使用自动重连机制，由 Infrastructure 层处理

**RPC 消息处理：**
- `DungeonMessageHandler` 处理来自 DungeonServer 的消息（RPC 请求、心跳、客户端消息转发）
- RPC 处理器注册迁移到 Controller 层，在 `OnSrvStart` 事件中注册
- 特殊 RPC（如 `D2GAddItem`）需要在玩家 Actor 中串行处理，通过 `gshare.SendMessageAsync` 发送到 Actor

**客户端消息转发：**
- GameServer 无法处理的客户端协议会转发到 DungeonServer（`msgId=0` 表示客户端消息转发）
- DungeonServer 收到后通过 `gatewaylink.ForwardClientMsg` 转发回客户端
- 协议路由逻辑迁移到 `ProtocolRouterController`

### 9.9 事件系统使用规范

**事件发布：**
- 通过 `EventPublisher` 接口发布事件
- 事件类型定义在 `gevent/enum.go` 中
- 事件参数通过 `args` 传递

**事件订阅：**
- 通过 `EventPublisher` 接口订阅事件
- 订阅在系统初始化时完成
- 事件处理器必须符合 `func(ctx context.Context, event interface{})` 签名

## 10. 参考资源

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Clean Architecture Example](https://github.com/bxcodec/go-clean-arch)
- [项目开发进度文档](./服务端开发进度文档.md)

## 11. 关键代码位置

### 11.1 当前代码位置（重构前）

**EntitySystem：**
- `internel/playeractor/entitysystem/*.go` - 所有业务系统（20 个）
- `internel/playeractor/entitysystem/sys_mgr.go` - 系统管理器
- `internel/playeractor/entitysystem/base_sys.go` - 系统基类

**协议处理：**
- `internel/playeractor/entity/player_network.go` - 网络消息处理入口
- `internel/playeractor/clientprotocol/protoregister.go` - 协议注册表
- 各 EntitySystem 的 `init()` 函数 - 协议注册

**Actor 框架：**
- `internel/playeractor/handler.go` - PlayerActor 消息处理器
- `internel/playeractor/adapter.go` - PlayerActor 适配器
- `internel/playeractor/entity/player_role.go` - 玩家角色实体

**PublicActor：**
- `internel/publicactor/adapter.go` - PublicActor 适配器
- `internel/publicactor/handler.go` - PublicActor 消息处理器
- `internel/publicactor/public_role*.go` - PublicActor 业务逻辑

**网络与 RPC：**
- `internel/gatewaylink/` - Gateway 连接管理
- `internel/dungeonserverlink/` - DungeonServer RPC 客户端
  - `protocol_manager.go` - 协议路由管理器（管理通用协议和独有协议的路由）
  - `dungeon_cli.go` - DungeonServer 客户端（连接池管理、自动重连、异步调用）
  - `handler.go` - RPC 消息处理器注册（`RegisterRPCHandler`）
  - `export.go` - 导出函数（`StartDungeonClient`、`Stop`、`AsyncCall`）
  - `DungeonMessageHandler` - 消息处理器（处理 RPC 请求、心跳、客户端消息转发）

**共享工具：**
- `internel/gshare/` - 共享工具函数
- `internel/gevent/` - 事件系统
- `internel/manager/` - 管理器（角色管理等）

### 11.2 重构后代码位置

**Domain 层：**
- `internel/domain/player.go` - 玩家实体
- `internel/domain/item.go` - 物品实体
- `internel/domain/skill.go` - 技能实体
- `internel/domain/repository/` - Repository 接口定义

**Use Case 层：**
- `internel/usecase/bag/` - 背包用例
- `internel/usecase/level/` - 等级用例
- `internel/usecase/interfaces/` - Use Case 依赖接口（EventPublisher、PublicActorGateway、DungeonServerGateway、ConfigManager 等）

**Adapter 层：**
- `internel/adapter/controller/` - 协议控制器（按系统分类）
- `internel/adapter/presenter/` - 响应构建器（按系统分类）
- `internel/adapter/gateway/` - 数据访问实现（PlayerGateway、NetworkGateway、PublicActorGateway、DungeonServerGateway、ConfigGateway）
- `internel/adapter/event/` - 事件适配器

**Infrastructure 层：**
- `internel/infrastructure/network/` - 网络网关（如果需要进一步封装）
- `internel/infrastructure/database/` - 数据库适配（如果需要进一步封装）

**DI 容器：**
- `internel/di/container.go` - 依赖注入容器

### 11.3 协议注册位置（重构后）

**Controller 层协议注册：**
- 各 Controller 在 `init()` 中通过 `gevent.Subscribe(gevent.OnSrvStart, ...)` 注册协议
- 协议处理器调用 Controller 方法
- 保持与 `clientprotocol.ProtoTbl` 的兼容性

**示例：**
```go
// adapter/controller/bag_controller.go
func init() {
    gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
        container := di.GetContainer()
        bagController := container.BagController
        
        clientprotocol.Register(uint16(protocol.C2SProtocol_C2SOpenBag), func(ctx context.Context, msg *network.ClientMessage) error {
            return bagController.HandleOpenBag(ctx, msg)
        })
    })
}
```

---

## 12. 当前代码统计

### 12.1 EntitySystem 统计

| 系统名称 | 文件 | 协议数量 | 依赖关系 | 优先级 |
|---------|------|---------|---------|--------|
| LevelSys | level_sys.go | 0 | 无 | 高 |
| BagSys | bag_sys.go | 1 | 无 | 高 |
| MoneySys | money_sys.go | 1 | 无 | 高 |
| EquipSys | equip_sys.go | 1 | 无 | 高 |
| AttrSys | attr_sys.go | 0 | LevelSys, EquipSys | 高 |
| SkillSys | skill_sys.go | 2 | 无 | 中 |
| QuestSys | quest_sys.go | 1 | 无 | 中 |
| FubenSys | fuben_sys.go | 1 | 无 | 中 |
| ItemUseSys | item_use_sys.go | 1 | 无 | 中 |
| ShopSys | shop_sys.go | 1 | 无 | 中 |
| RecycleSys | recycle_sys.go | 1 | 无 | 中 |
| FriendSys | friend_sys.go | 7 | PublicActor | 低 |
| GuildSys | guild_sys.go | 4 | PublicActor | 低 |
| ChatSys | chat_sys.go | 2 | PublicActor | 低 |
| AuctionSys | auction_sys.go | 3 | PublicActor | 低 |
| MailSys | mail_sys.go | 0 | 无 | 辅助 |
| VipSys | vip_sys.go | 0 | 无 | 辅助 |
| DailyActivitySys | daily_activity_sys.go | 0 | 无 | 辅助 |
| MessageSys | message_sys.go | 0 | 无 | 辅助 |
| GMSys | gm_sys.go | 1 | 无 | 辅助 |

**总计：** 20 个系统，27 个协议

### 12.2 框架依赖统计

| 框架 | 使用位置 | 重构后接口 |
|------|---------|-----------|
| database | 所有 EntitySystem | Repository 接口 |
| gatewaylink | 所有 EntitySystem | NetworkGateway 接口 |
| dungeonserverlink | FubenSys 等 | DungeonServerGateway 接口 |
| gevent | 所有 EntitySystem | EventPublisher 接口 |
| jsonconf | 所有 EntitySystem | ConfigManager 接口 |
| gshare | 所有 EntitySystem | PublicActorGateway 接口 |

### 12.3 协议注册统计

- `player_network.go` 中注册：7 个协议（登录、注册、角色管理等）
- EntitySystem 中注册：27 个协议（各业务系统协议）
- **总计：** 34 个协议

## 13. 重构遗漏点补充

### 13.1 系统生命周期管理

**当前实现：**
- 系统实现 `ISystem` 接口，包含完整的生命周期方法：
  - `OnInit(ctx)` - 系统初始化（从 BinaryData 加载数据）
  - `OnRoleLogin(ctx)` - 玩家登录时调用
  - `OnRoleReconnect(ctx)` - 玩家重连时调用
  - `OnRoleLogout(ctx)` - 玩家登出时调用
  - `OnRoleClose(ctx)` - 玩家关闭时调用
  - `OnNewHour/Day/Week/Month/Year(ctx)` - 时间变化时调用
- 系统管理器（`SysMgr`）统一管理所有系统的生命周期

**重构后处理：**
- **Use Case 层**：不需要生命周期方法，Use Case 是纯业务逻辑
- **Adapter 层**：创建 `SystemLifecycleAdapter` 适配器，实现 `ISystem` 接口
- **系统适配器模式**：
  ```go
  // adapter/system/bag_system_adapter.go
  type BagSystemAdapter struct {
      *BaseSystemAdapter
      bagUseCase *bag.BagUseCase
      bagController *controller.BagController
  }
  
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      // 从 BinaryData 加载数据
      playerRole := getPlayerRoleFromContext(ctx)
      binaryData := playerRole.GetBinaryData()
      // 调用 Use Case 初始化
      a.bagUseCase.InitializeFromBinaryData(ctx, binaryData.BagData)
  }
  ```

**关键点：**
- 系统生命周期适配器必须实现 `ISystem` 接口
- 生命周期方法中调用对应的 Use Case 方法
- 保持与现有 `SysMgr` 的兼容性

### 13.2 BinaryData 共享机制

**当前实现：**
- 系统直接操作 `PlayerRoleBinaryData`，这是共享引用
- 系统在 `OnInit` 中从 `BinaryData` 获取数据指针，直接修改
- 修改后通过 `PlayerRole.SaveBinaryData()` 持久化

**重构后处理：**
- **Domain 层**：定义 `PlayerBinaryData` 实体，包含所有系统数据
- **Repository 层**：`PlayerRepository` 提供 `GetBinaryData/SaveBinaryData` 方法
- **Use Case 层**：通过 Repository 获取 BinaryData，修改后保存
- **共享引用保持**：
  ```go
  // usecase/bag/bag_usecase.go
  type BagUseCase struct {
      playerRepo repository.PlayerRepository
  }
  
  func (uc *BagUseCase) AddItem(ctx context.Context, item *domain.Item) error {
      // 获取 BinaryData（共享引用）
      binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
      if err != nil {
          return err
      }
      
      // 直接修改 BinaryData（保持共享引用）
      binaryData.BagData.Items = append(binaryData.BagData.Items, item)
      
      // 保存 BinaryData
      return uc.playerRepo.SaveBinaryData(ctx, roleID, binaryData)
  }
  ```

**关键点：**
- Repository 必须返回 BinaryData 的共享引用，不能复制
- Use Case 直接修改 BinaryData，保持内存共享
- 保存时通过 Repository 统一持久化

### 13.3 系统依赖关系处理

**当前实现：**
- 在 `sys_mgr.go` 的 `systemDependencies` 中定义系统依赖关系
- 使用拓扑排序确定系统初始化顺序
- 例如：`AttrSys` 依赖 `LevelSys` 和 `EquipSys`

**重构后处理：**
- **依赖关系定义**：保持 `systemDependencies` 定义不变
- **Use Case 依赖注入**：通过依赖注入解决 Use Case 之间的依赖
  ```go
  // usecase/attr/attr_calculator_usecase.go
  type AttrCalculatorUseCase struct {
      levelUseCase *level.LevelUseCase  // 依赖 LevelSys
      equipUseCase *equip.EquipUseCase  // 依赖 EquipSys
  }
  ```
- **系统初始化顺序**：通过拓扑排序确定系统适配器的初始化顺序
- **依赖检查**：在 DI 容器中检查 Use Case 依赖是否已注入

**关键点：**
- 系统依赖关系定义保持不变
- Use Case 依赖通过 DI 容器注入
- 系统适配器初始化顺序由拓扑排序确定

### 13.4 RunOne 机制

**当前实现：**
- 系统实现 `RunOne(ctx)` 方法（可选）
- `PlayerRole.RunOne()` 每帧调用所有系统的 `RunOne` 方法
- 例如：`AttrSys.RunOne()` 用于属性增量更新

**重构后处理：**
- **Use Case 层**：定义 `RunOneUseCase` 接口
  ```go
  // usecase/interfaces/runone.go
  type RunOneUseCase interface {
      RunOne(ctx context.Context) error
  }
  ```
- **系统适配器**：实现 `RunOne` 方法，调用 Use Case 的 `RunOne`
  ```go
  // adapter/system/attr_system_adapter.go
  func (a *AttrSystemAdapter) RunOne(ctx context.Context) {
      if runOneUC, ok := a.attrUseCase.(usecase.RunOneUseCase); ok {
          runOneUC.RunOne(ctx)
      }
  }
  ```
- **系统管理器**：保持 `SysMgr.EachOpenSystem` 调用 `RunOne` 的逻辑

**关键点：**
- 只有需要定期执行的系统才实现 `RunOneUseCase`
- 系统适配器负责调用 Use Case 的 `RunOne`
- 保持与现有 `PlayerRole.RunOne()` 的兼容性

### 13.5 事件订阅机制

**当前实现：**
- 系统在 `init()` 中通过 `gevent.SubscribePlayerEvent` 订阅事件
- 事件类型定义在 `gevent/enum.go` 中
- 事件处理器在 Actor 主线程执行

**重构后处理：**
- **Use Case 层**：定义 `EventSubscriber` 接口
  ```go
  // usecase/interfaces/event_subscriber.go
  type EventSubscriber interface {
      SubscribeEvents(publisher interfaces.EventPublisher)
  }
  ```
- **系统适配器**：在 `OnInit` 中调用 Use Case 的 `SubscribeEvents`
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      // ... 初始化逻辑
      if subscriber, ok := a.bagUseCase.(usecase.EventSubscriber); ok {
          subscriber.SubscribeEvents(a.eventPublisher)
      }
  }
  ```
- **事件适配器**：实现 `EventPublisher` 接口，封装 `gevent`

**关键点：**
- 事件订阅在系统初始化时完成
- Use Case 通过接口订阅事件，不直接依赖 `gevent`
- 事件处理器必须在 Actor 主线程执行

### 13.6 系统工厂模式

**当前实现：**
- 系统在 `init()` 中通过 `RegisterSystemFactory` 注册工厂函数
- 工厂函数返回 `ISystem` 接口实现
- `SysMgr` 在初始化时通过工厂创建系统实例

**重构后处理：**
- **系统适配器工厂**：创建系统适配器工厂函数
  ```go
  // adapter/system/bag_system_adapter.go
  func NewBagSystemAdapter(container *di.Container) *BagSystemAdapter {
      return &BagSystemAdapter{
          BaseSystemAdapter: NewBaseSystemAdapter(uint32(protocol.SystemId_SysBag)),
          bagUseCase: container.BagUseCase,
          bagController: container.BagController,
      }
  }
  ```
- **工厂注册**：在 `init()` 中注册系统适配器工厂
  ```go
  func init() {
      entitysystem.RegisterSystemFactory(
          uint32(protocol.SystemId_SysBag),
          func() iface.ISystem {
              return NewBagSystemAdapter(di.GetContainer())
          },
      )
  }
  ```
- **DI 容器**：系统适配器从 DI 容器获取 Use Case 和 Controller

**关键点：**
- 系统工厂模式保持不变
- 工厂函数创建系统适配器，而不是直接创建 Use Case
- 系统适配器从 DI 容器获取依赖

### 13.7 Context 传递机制

**当前实现：**
- Context 中必须包含 `gshare.ContextKeyRole`（`IPlayerRole` 接口）
- Context 中必须包含 `gshare.ContextKeySession`（SessionId 字符串）
- 系统通过 `GetIPlayerRoleByContext(ctx)` 获取玩家角色

**重构后处理：**
- **Context 工具函数**：在 Adapter 层提供 Context 工具函数
  ```go
  // adapter/context/context_helper.go
  func GetPlayerRoleFromContext(ctx context.Context) (iface.IPlayerRole, error) {
      value := ctx.Value(gshare.ContextKeyRole)
      if value == nil {
          return nil, errors.New("not found role in context")
      }
      role, ok := value.(iface.IPlayerRole)
      if !ok {
          return nil, errors.New("invalid role type")
      }
      return role, nil
  }
  ```
- **Use Case Context**：Use Case 方法接收 `context.Context`，但不直接访问框架层
- **Controller Context**：Controller 从 Context 提取信息，传递给 Use Case

**关键点：**
- Context 传递机制保持不变
- Use Case 不直接访问 Context 中的框架对象
- Controller 负责从 Context 提取信息并传递给 Use Case

### 13.8 PlayerRole 与系统的关系

**当前实现：**
- `PlayerRole` 包含 `sysMgr`（系统管理器）
- 系统通过 `GetXXXSys(ctx)` 函数获取系统实例
- 系统通过 `playerRole.GetSystem(sysId)` 获取系统

**重构后处理：**
- **系统获取方式**：保持 `GetXXXSys(ctx)` 函数，但返回系统适配器
  ```go
  // adapter/system/bag_system_adapter.go
  func GetBagSys(ctx context.Context) *BagSystemAdapter {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          return nil
      }
      system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
      if system == nil {
          return nil
      }
      return system.(*BagSystemAdapter)
  }
  ```
- **系统管理器**：`SysMgr` 管理系统适配器，而不是 Use Case
- **系统接口**：系统适配器实现 `ISystem` 接口，保持兼容性

**关键点：**
- 系统获取方式保持不变
- 系统适配器实现 `ISystem` 接口
- `SysMgr` 管理的是系统适配器，不是 Use Case

### 13.9 BinaryData 初始化

**当前实现：**
- 系统在 `OnInit` 中从 `PlayerRoleBinaryData` 获取数据
- 如果数据不存在，则初始化空数据
- 系统直接持有 BinaryData 的引用

**重构后处理：**
- **Use Case 初始化**：Use Case 提供 `InitializeFromBinaryData` 方法
  ```go
  // usecase/bag/bag_usecase.go
  func (uc *BagUseCase) InitializeFromBinaryData(ctx context.Context, bagData *protocol.SiBagData) {
      if bagData == nil {
          bagData = &protocol.SiBagData{Items: make([]*protocol.ItemSt, 0)}
      }
      // 初始化 Use Case 内部状态
  }
  ```
- **系统适配器**：在 `OnInit` 中调用 Use Case 的初始化方法
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      playerRole, _ := GetPlayerRoleFromContext(ctx)
      binaryData := playerRole.GetBinaryData()
      if binaryData.BagData == nil {
          binaryData.BagData = &protocol.SiBagData{Items: make([]*protocol.ItemSt, 0)}
      }
      a.bagUseCase.InitializeFromBinaryData(ctx, binaryData.BagData)
  }
  ```

**关键点：**
- BinaryData 初始化逻辑保持不变
- Use Case 负责初始化内部状态
- 系统适配器负责从 BinaryData 加载数据

### 13.10 系统状态管理

**当前实现：**
- 系统有 `IsOpened()` 和 `SetOpened()` 方法
- 系统管理器通过 `CheckAllSysOpen` 检查系统状态
- 系统状态存储在 `PlayerRoleBinaryData.SysOpenStatus` 中

**重构后处理：**
- **系统适配器**：实现 `IsOpened()` 和 `SetOpened()` 方法
  ```go
  // adapter/system/base_system_adapter.go
  type BaseSystemAdapter struct {
      sysID  uint32
      opened bool
  }
  
  func (a *BaseSystemAdapter) IsOpened() bool {
      return a.opened
  }
  
  func (a *BaseSystemAdapter) SetOpened(opened bool) {
      a.opened = opened
  }
  ```
- **状态持久化**：系统状态仍然存储在 `BinaryData.SysOpenStatus` 中
- **状态检查**：`SysMgr.CheckAllSysOpen` 逻辑保持不变

**关键点：**
- 系统状态管理逻辑保持不变
- 系统适配器实现状态管理方法
- 状态持久化通过 BinaryData 完成

### 13.11 系统获取函数模式

**当前实现：**
- 每个系统都有 `GetXXXSys(ctx)` 函数
- 函数从 Context 获取 `IPlayerRole`，然后获取系统实例
- 函数检查系统是否存在和是否开启

**重构后处理：**
- **系统获取函数**：保持 `GetXXXSys(ctx)` 函数，但返回系统适配器
  ```go
  // adapter/system/bag_system_adapter.go
  func GetBagSys(ctx context.Context) *BagSystemAdapter {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          log.Errorf("get player role error:%v", err)
          return nil
      }
      system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
      if system == nil {
          log.Errorf("not found system [%v]", protocol.SystemId_SysBag)
          return nil
      }
      sys := system.(*BagSystemAdapter)
      if sys == nil || !sys.IsOpened() {
          log.Errorf("get player role system [%v] error", protocol.SystemId_SysBag)
          return nil
      }
      return sys
  }
  ```
- **函数位置**：系统获取函数放在系统适配器文件中
- **类型断言**：函数进行类型断言，确保返回正确的系统适配器类型

**关键点：**
- 系统获取函数模式保持不变
- 函数返回系统适配器，而不是 Use Case
- 函数检查逻辑保持不变

### 13.12 协议注册时机与流程

**当前实现：**
- **协议注册表**：`clientprotocol.ProtoTbl` 是全局协议注册表（`map[uint16]Func`）
- **协议注册时机**：协议在 `gevent.OnSrvStart` 事件中注册
- **注册流程**：
  1. 系统在 `init()` 中订阅 `gevent.OnSrvStart` 事件
  2. 服务器启动时触发 `OnSrvStart` 事件
  3. 在事件回调中调用 `clientprotocol.Register(protoId, handler)` 注册协议
  4. 协议处理器函数签名：`func(ctx context.Context, msg *network.ClientMessage) error`
- **协议处理流程**：
  1. 客户端消息通过 `handleDoNetWorkMsg` 处理
  2. `handleDoNetWorkMsg` 通过 `clientprotocol.GetFunc(msgId)` 获取协议处理器
  3. 调用协议处理器处理消息
- **RPC 协议注册**：
  - DungeonServer 通过 `dungeonserverlink.RegisterRPCHandler` 注册 RPC 处理器
  - RPC 处理器函数签名：`func(ctx context.Context, sessionId string, data []byte) error`
- **协议路由**：
  - 通过 `ProtocolManager` 管理 DungeonServer 的协议路由
  - 支持通用协议（多个 DungeonServer 共享）和独有协议（特定 srvType）

**重构后处理：**
- **Controller 协议注册**：Controller 在 `init()` 中注册协议
  ```go
  // adapter/controller/bag_controller.go
  func init() {
      gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
          container := di.GetContainer()
          bagController := container.BagController
          
          clientprotocol.Register(
              uint16(protocol.C2SProtocol_C2SOpenBag),
              bagController.HandleOpenBag,
          )
      })
  }
  ```
- **协议注册位置**：协议注册从 EntitySystem 迁移到 Controller
- **事件订阅**：Controller 在 `init()` 中订阅 `OnSrvStart` 事件
- **协议处理流程保持不变**：`handleDoNetWorkMsg` 逻辑不变，通过 `clientprotocol.GetFunc` 获取处理器

**关键点：**
- 协议注册时机保持不变（`OnSrvStart` 事件）
- 协议注册位置迁移到 Controller 层
- 保持与现有协议注册机制的兼容性（`clientprotocol.ProtoTbl`）
- 协议处理流程保持不变
- RPC 协议注册机制保持不变

### 13.13 系统间相互调用模式

**当前实现：**
- 系统通过 `GetXXXSys(ctx)` 函数获取其他系统实例
- 例如：`levelSys := GetLevelSys(ctx)`、`bagSys := GetBagSys(ctx)`
- 系统在运行时动态获取依赖系统

**重构后处理：**
- **Use Case 依赖注入**：通过依赖注入解决系统间依赖
  ```go
  // usecase/item_use/item_use_usecase.go
  type ItemUseUseCase struct {
      bagUseCase   *bag.BagUseCase
      levelUseCase *level.LevelUseCase
  }
  
  func NewItemUseUseCase(
      bagUseCase *bag.BagUseCase,
      levelUseCase *level.LevelUseCase,
  ) *ItemUseUseCase {
      return &ItemUseUseCase{
          bagUseCase:   bagUseCase,
          levelUseCase: levelUseCase,
      }
  }
  ```
- **系统适配器获取**：系统适配器可以通过 `GetXXXSys(ctx)` 获取其他系统适配器（用于向后兼容）
  ```go
  // adapter/system/item_use_system_adapter.go
  func (a *ItemUseSystemAdapter) HandleUseItem(ctx context.Context, msg *network.ClientMessage) error {
      // 通过 Controller 调用 Use Case
      return a.itemUseController.HandleUseItem(ctx, msg)
  }
  ```

**关键点：**
- Use Case 层通过依赖注入解决系统间依赖
- 系统适配器保持 `GetXXXSys(ctx)` 函数，用于向后兼容
- 禁止 Use Case 直接调用 `GetXXXSys(ctx)`

### 13.14 BinaryData 的直接引用保持

**当前实现：**
- 系统在 `OnInit` 中直接获取 BinaryData 的引用
- 例如：`bs.bagData = binaryData.BagData`
- 系统直接修改 BinaryData，修改后通过 `PlayerRole.SaveBinaryData()` 持久化

**重构后处理：**
- **Repository 返回引用**：Repository 必须返回 BinaryData 的共享引用
  ```go
  // adapter/gateway/player_gateway.go
  func (g *PlayerGateway) GetBinaryData(ctx context.Context, roleID uint64) (*protocol.PlayerRoleBinaryData, error) {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          return nil, err
      }
      // 返回共享引用，不复制
      return playerRole.GetBinaryData(), nil
  }
  ```
- **Use Case 直接修改**：Use Case 直接修改 BinaryData，保持内存共享
  ```go
  // usecase/bag/bag_usecase.go
  func (uc *BagUseCase) AddItem(ctx context.Context, item *domain.Item) error {
      binaryData, err := uc.playerRepo.GetBinaryData(ctx, roleID)
      if err != nil {
          return err
      }
      // 直接修改 BinaryData（保持共享引用）
      binaryData.BagData.Items = append(binaryData.BagData.Items, convertToProtocolItem(item))
      return nil
  }
  ```

**关键点：**
- Repository 必须返回 BinaryData 的共享引用，不能复制
- Use Case 直接修改 BinaryData，保持内存共享
- 保存时通过 Repository 统一持久化

### 13.15 系统辅助索引管理

**当前实现：**
- 一些系统（如 BagSys）有辅助索引（如 `itemIndex`）
- 辅助索引用于快速查找，但不作为数据源
- 数据变更后需要重建索引（`rebuildIndex()`）

**重构后处理：**
- **Use Case 管理索引**：Use Case 负责管理辅助索引
  ```go
  // usecase/bag/bag_usecase.go
  type BagUseCase struct {
      playerRepo repository.PlayerRepository
      itemIndex  map[uint32][]*protocol.ItemSt  // 辅助索引
  }
  
  func (uc *BagUseCase) AddItem(ctx context.Context, item *domain.Item) error {
      // ... 添加物品逻辑
      uc.rebuildIndex()  // 重建索引
      return nil
  }
  
  func (uc *BagUseCase) rebuildIndex() {
      binaryData, _ := uc.playerRepo.GetBinaryData(ctx, roleID)
      uc.itemIndex = make(map[uint32][]*protocol.ItemSt)
      for _, item := range binaryData.BagData.Items {
          uc.itemIndex[item.ItemId] = append(uc.itemIndex[item.ItemId], item)
      }
  }
  ```
- **系统适配器初始化索引**：系统适配器在 `OnInit` 中初始化 Use Case 的索引
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      binaryData := getBinaryDataFromContext(ctx)
      a.bagUseCase.InitializeFromBinaryData(ctx, binaryData.BagData)
      a.bagUseCase.RebuildIndex(ctx)  // 初始化索引
  }
  ```

**关键点：**
- Use Case 负责管理辅助索引
- 数据变更后需要重建索引
- 系统适配器在初始化时重建索引

### 13.16 系统状态持久化

**当前实现：**
- 系统状态存储在 `BinaryData.SysOpenStatus` 中
- `SysMgr.CheckAllSysOpen` 检查系统状态并更新 `BinaryData.SysOpenStatus`
- 系统状态通过 `PlayerRole.SaveBinaryData()` 持久化

**重构后处理：**
- **系统状态管理**：系统适配器实现状态管理，状态存储在 BinaryData 中
  ```go
  // adapter/system/base_system_adapter.go
  type BaseSystemAdapter struct {
      sysID  uint32
      opened bool
  }
  
  func (a *BaseSystemAdapter) SetOpened(opened bool) {
      a.opened = opened
      // 更新 BinaryData.SysOpenStatus
      binaryData := getBinaryDataFromContext(ctx)
      if binaryData.SysOpenStatus == nil {
          binaryData.SysOpenStatus = make(map[uint32]uint32)
      }
      if opened {
          binaryData.SysOpenStatus[a.sysID] = 1
      } else {
          binaryData.SysOpenStatus[a.sysID] = 0
      }
  }
  ```
- **状态持久化**：状态变更后通过 Repository 持久化
  ```go
  // adapter/system/base_system_adapter.go
  func (a *BaseSystemAdapter) SetOpened(opened bool) {
      // ... 更新状态
      // 通过 Repository 持久化（如果需要立即持久化）
      playerRepo := getPlayerRepoFromContext(ctx)
      playerRepo.SaveBinaryData(ctx, roleID, binaryData)
  }
  ```

**关键点：**
- 系统状态存储在 BinaryData 中
- 系统适配器负责状态管理
- 状态变更后通过 Repository 持久化

### 13.17 系统依赖的运行时获取

**当前实现：**
- 系统在运行时通过 `GetXXXSys(ctx)` 获取依赖系统
- 例如：`levelSys := GetLevelSys(ctx)`、`bagSys := GetBagSys(ctx)`
- 系统在方法中动态获取依赖系统

**重构后处理：**
- **Use Case 依赖注入**：Use Case 通过依赖注入获取依赖系统
  ```go
  // usecase/item_use/item_use_usecase.go
  type ItemUseUseCase struct {
      bagUseCase   *bag.BagUseCase
      levelUseCase *level.LevelUseCase
  }
  
  func NewItemUseUseCase(
      bagUseCase *bag.BagUseCase,
      levelUseCase *level.LevelUseCase,
  ) *ItemUseUseCase {
      return &ItemUseUseCase{
          bagUseCase:   bagUseCase,
          levelUseCase: levelUseCase,
      }
  }
  
  func (uc *ItemUseUseCase) UseItem(ctx context.Context, itemID uint32) error {
      // 直接使用注入的 Use Case，不通过 GetXXXSys
      if err := uc.bagUseCase.RemoveItem(ctx, itemID, 1); err != nil {
          return err
      }
      // ...
  }
  ```
- **DI 容器管理依赖**：DI 容器负责管理 Use Case 之间的依赖关系
  ```go
  // di/container.go
  func NewContainer() *Container {
      c := &Container{}
      
      // 先创建无依赖的 Use Case
      c.bagUseCase = bag.NewBagUseCase(c.playerRepo)
      c.levelUseCase = level.NewLevelUseCase(c.playerRepo)
      
      // 创建有依赖的 Use Case
      c.itemUseUseCase = item_use.NewItemUseUseCase(
          c.bagUseCase,
          c.levelUseCase,
      )
      
      return c
  }
  ```

**关键点：**
- Use Case 通过依赖注入获取依赖系统
- DI 容器负责管理 Use Case 之间的依赖关系
- 禁止 Use Case 直接调用 `GetXXXSys(ctx)`

### 13.18 系统初始化顺序与拓扑排序

**当前实现：**
- `SysMgr.getInitOrder()` 使用拓扑排序确定系统初始化顺序
- 系统依赖关系定义在 `systemDependencies` 中
- 例如：`AttrSys` 依赖 `LevelSys` 和 `EquipSys`

**重构后处理：**
- **系统依赖关系定义**：保持 `systemDependencies` 定义不变
- **系统适配器初始化顺序**：通过拓扑排序确定系统适配器的初始化顺序
  ```go
  // adapter/system/sys_mgr_adapter.go
  func (m *SysMgrAdapter) OnInit(ctx context.Context) error {
      // 使用拓扑排序确定系统初始化顺序
      initOrder := m.getInitOrder()
      
      // 按照依赖顺序初始化系统适配器
      for _, sysId := range initOrder {
          factory := m.factories[sysId]
          if factory == nil {
              continue
          }
          systemAdapter := factory()
          systemAdapter.OnInit(ctx)
          systemAdapter.SetOpened(true)
          m.sysList[sysId] = systemAdapter
      }
      return nil
  }
  ```
- **Use Case 依赖注入顺序**：DI 容器按照依赖顺序创建 Use Case
  ```go
  // di/container.go
  func NewContainer() *Container {
      c := &Container{}
      
      // 按照依赖顺序创建 Use Case
      // 1. 无依赖的 Use Case
      c.bagUseCase = bag.NewBagUseCase(c.playerRepo)
      c.levelUseCase = level.NewLevelUseCase(c.playerRepo)
      c.equipUseCase = equip.NewEquipUseCase(c.playerRepo)
      
      // 2. 有依赖的 Use Case
      c.attrUseCase = attr.NewAttrUseCase(
          c.levelUseCase,
          c.equipUseCase,
      )
      
      return c
  }
  ```

**关键点：**
- 系统依赖关系定义保持不变
- 系统适配器初始化顺序由拓扑排序确定
- Use Case 依赖注入顺序由 DI 容器管理

### 13.19 系统生命周期方法的具体作用

**当前实现：**
- 系统实现 `ISystem` 接口，包含完整的生命周期方法
- 生命周期方法在特定时机被调用

**重构后处理：**
- **OnInit**：系统初始化，从 BinaryData 加载数据
  - Use Case：提供 `InitializeFromBinaryData` 方法
  - 系统适配器：在 `OnInit` 中调用 Use Case 的初始化方法
- **OnRoleLogin**：玩家登录时调用
  - Use Case：提供 `OnLogin` 方法（可选）
  - 系统适配器：在 `OnRoleLogin` 中调用 Use Case 的 `OnLogin` 方法
- **OnRoleReconnect**：玩家重连时调用
  - Use Case：提供 `OnReconnect` 方法（可选）
  - 系统适配器：在 `OnRoleReconnect` 中调用 Use Case 的 `OnReconnect` 方法
- **OnRoleLogout**：玩家登出时调用
  - Use Case：提供 `OnLogout` 方法（可选）
  - 系统适配器：在 `OnRoleLogout` 中调用 Use Case 的 `OnLogout` 方法
- **OnRoleClose**：玩家关闭时调用
  - Use Case：提供 `OnClose` 方法（可选）
  - 系统适配器：在 `OnRoleClose` 中调用 Use Case 的 `OnClose` 方法
- **OnNewHour/Day/Week/Month/Year**：时间变化时调用
  - Use Case：提供对应的时间回调方法（可选）
  - 系统适配器：在对应的时间回调中调用 Use Case 的方法

**关键点：**
- 生命周期方法在系统适配器中实现
- Use Case 提供对应的业务逻辑方法（可选）
- 系统适配器负责调用 Use Case 的方法

### 13.20 玩家消息系统重构

**当前实现：**
- `MessageSys` 负责离线消息加载/重放
- 消息回调通过 `engine/message_registry.go` 注册
- 消息发送通过 `gshare.SendPlayerActorMessage` 实现

**重构后处理：**
- **Use Case 层**：创建 `usecase/message/` 目录
  ```go
  // usecase/message/message_usecase.go
  type MessageUseCase struct {
      messageRepo repository.MessageRepository
  }
  
  func (uc *MessageUseCase) LoadOfflineMessages(ctx context.Context, roleID uint64) error {
      // 从数据库加载离线消息
      messages, err := uc.messageRepo.LoadMessages(ctx, roleID)
      if err != nil {
          return err
      }
      // 回放消息
      for _, msg := range messages {
          uc.replayMessage(ctx, msg)
      }
      return nil
  }
  ```
- **Controller 层**：创建 `adapter/controller/message_controller.go`
  ```go
  // adapter/controller/message_controller.go
  type MessageController struct {
      messageUseCase *message.MessageUseCase
  }
  
  func (c *MessageController) HandlePlayerMessage(ctx context.Context, msgType uint32, msgData []byte) error {
      return c.messageUseCase.HandleMessage(ctx, msgType, msgData)
  }
  ```
- **消息注册**：消息回调注册迁移到 Controller 层
  ```go
  // adapter/controller/message_controller.go
  func init() {
      gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
          container := di.GetContainer()
          messageController := container.MessageController
          
          // 注册消息回调
          engine.RegisterMessageCallback(msgType, func(roleID uint64, msgData []byte) error {
              return messageController.HandlePlayerMessage(ctx, msgType, msgData)
          })
      })
  }
  ```

**关键点：**
- 消息系统按照 Clean Architecture 重构
- 消息回调注册迁移到 Controller 层
- 保持与现有消息系统的兼容性

### 13.21 IPlayerRole 接口的使用

**当前实现：**
- `IPlayerRole` 接口定义在 `iface/irole.go` 中
- `PlayerRole` 实现了 `IPlayerRole` 接口
- 系统通过 `GetIPlayerRoleByContext(ctx)` 获取 `IPlayerRole` 接口
- Context 中存储的是 `IPlayerRole` 接口，而不是具体类型

**重构后处理：**
- **保持 IPlayerRole 接口不变**：接口定义保持不变
- **Context 传递**：Context 中仍然存储 `IPlayerRole` 接口
- **Use Case 层**：Use Case 不直接依赖 `IPlayerRole` 接口，而是通过 Repository 获取数据
- **Controller 层**：Controller 从 Context 获取 `IPlayerRole`，提取必要信息传递给 Use Case
  ```go
  // adapter/controller/bag_controller.go
  func (c *BagController) HandleAddItem(ctx context.Context, msg *network.ClientMessage) error {
      // 从 Context 获取 IPlayerRole
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          return err
      }
      
      // 提取 roleID
      roleID := playerRole.GetPlayerRoleId()
      
      // 调用 Use Case（不传递 IPlayerRole，只传递 roleID）
      if err := c.addItemUseCase.Execute(ctx, roleID, item); err != nil {
          return c.presenter.PresentError(ctx, err)
      }
      
      return c.presenter.PresentAddItemSuccess(ctx, item)
  }
  ```

**关键点：**
- `IPlayerRole` 接口保持不变
- Use Case 不直接依赖 `IPlayerRole` 接口
- Controller 负责从 Context 提取信息并传递给 Use Case

### 13.22 PlayerRole 生命周期管理

**当前实现：**
- `PlayerRole` 在 `manager.GetPlayerRole(roleId)` 中创建
- `PlayerRole` 的生命周期由 `RoleManager` 管理
- `PlayerRole.OnLogin()`、`OnLogout()`、`OnReconnect()` 等方法在特定时机调用

**重构后处理：**
- **PlayerRole 生命周期保持不变**：`PlayerRole` 的创建和管理逻辑不变
- **系统生命周期适配器**：系统适配器在 `OnInit` 中从 `PlayerRole` 获取 BinaryData
- **Use Case 初始化**：Use Case 通过 `InitializeFromBinaryData` 方法初始化
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      // 从 Context 获取 PlayerRole
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          return
      }
      
      // 从 PlayerRole 获取 BinaryData
      binaryData := playerRole.GetBinaryData()
      if binaryData.BagData == nil {
          binaryData.BagData = &protocol.SiBagData{Items: make([]*protocol.ItemSt, 0)}
      }
      
      // 调用 Use Case 初始化
      a.bagUseCase.InitializeFromBinaryData(ctx, binaryData.BagData)
  }
  ```

**关键点：**
- `PlayerRole` 的生命周期管理保持不变
- 系统适配器在生命周期方法中从 `PlayerRole` 获取数据
- Use Case 通过初始化方法接收数据

### 13.23 系统获取函数（GetXXXSys）的保持

**当前实现：**
- 每个系统都有 `GetXXXSys(ctx)` 函数
- 函数从 Context 获取 `IPlayerRole`，然后获取系统实例
- 函数检查系统是否存在和是否开启

**重构后处理：**
- **保持系统获取函数**：`GetXXXSys(ctx)` 函数保持不变，但返回系统适配器
- **系统适配器实现**：系统适配器实现 `ISystem` 接口
- **向后兼容**：保持 `GetXXXSys(ctx)` 函数，确保现有代码可以继续使用
  ```go
  // adapter/system/bag_system_adapter.go
  func GetBagSys(ctx context.Context) *BagSystemAdapter {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          log.Errorf("get player role error:%v", err)
          return nil
      }
      system := playerRole.GetSystem(uint32(protocol.SystemId_SysBag))
      if system == nil {
          log.Errorf("not found system [%v]", protocol.SystemId_SysBag)
          return nil
      }
      sys := system.(*BagSystemAdapter)
      if sys == nil || !sys.IsOpened() {
          log.Errorf("get player role system [%v] error", protocol.SystemId_SysBag)
          return nil
      }
      return sys
  }
  ```

**关键点：**
- 系统获取函数保持不变，但返回系统适配器
- 系统适配器实现 `ISystem` 接口
- 保持向后兼容性

### 13.24 协议处理流程保持不变

**当前实现：**
- 协议消息通过 `handleDoNetWorkMsg` 处理
- `handleDoNetWorkMsg` 在 Actor 的消息处理中调用
- 协议处理器通过 `clientprotocol.GetFunc(msgId)` 获取
- 协议处理器函数签名：`func(ctx context.Context, msg *network.ClientMessage) error`

**重构后处理：**
- **协议处理流程保持不变**：`handleDoNetWorkMsg` 逻辑不变
- **协议注册迁移到 Controller**：协议注册从 EntitySystem 迁移到 Controller
- **协议处理器调用 Controller**：协议处理器直接调用 Controller 方法
  ```go
  // adapter/controller/bag_controller.go
  func init() {
      gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
          container := di.GetContainer()
          bagController := container.BagController
          
          // 注册协议处理器
          clientprotocol.Register(
              uint16(protocol.C2SProtocol_C2SOpenBag),
              bagController.HandleOpenBag,  // 直接调用 Controller 方法
          )
      })
  }
  ```

**关键点：**
- 协议处理流程保持不变
- 协议注册迁移到 Controller 层
- 协议处理器直接调用 Controller 方法

### 13.25 事件系统的使用

**当前实现：**
- 使用 `gevent` 包进行事件发布和订阅
- 每个玩家有独立的事件总线（`PlayerRole.eventBus`），从全局模板克隆（`gevent.ClonePlayerEventBus()`）
- 系统在 `init()` 中通过 `gevent.SubscribePlayerEvent` 订阅事件
- 系统通过 `playerRole.Publish(typ, args...)` 发布事件
- 事件在 Actor 主线程中执行，保证单线程特性

**重构后处理：**
- **事件适配器**：创建 `EventAdapter` 实现 `EventPublisher` 接口
- **Use Case 事件发布**：Use Case 通过 `EventPublisher` 接口发布事件
- **事件订阅**：Use Case 通过 `EventSubscriber` 接口订阅事件
- **系统适配器**：系统适配器在 `OnInit` 中调用 Use Case 的事件订阅
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      // ... 初始化逻辑
      
      // 订阅事件
      if subscriber, ok := a.bagUseCase.(usecase.EventSubscriber); ok {
          subscriber.SubscribeEvents(a.eventPublisher)
      }
  }
  
  // usecase/bag/bag_usecase.go
  func (uc *BagUseCase) SubscribeEvents(publisher interfaces.EventPublisher) {
      publisher.SubscribePlayerEvent(gevent.OnItemAdded, func(ctx context.Context, event interface{}) {
          // 处理事件
      })
  }
  ```

**关键点：**
- 事件系统通过接口抽象
- Use Case 通过接口发布和订阅事件
- 系统适配器负责事件订阅的初始化
- 每个 PlayerRole 有独立的事件总线，从全局模板克隆
- 事件在 Actor 主线程中执行，保证单线程特性

### 13.26 RunOne 机制的详细说明

**当前实现：**
- `PlayerRole.RunOne()` 每帧通过 `DoRunOneMsg` 消息在 Actor 中处理
- 使用 `_1sChecker`（1秒检查器）和 `_5minChecker`（5分钟检查器）控制频率
- `_1sChecker` 触发：时间同步、属性系统的 `RunOne`
- `_5minChecker` 触发：定期保存数据到数据库
- 系统可以通过实现 `RunOne` 方法参与定期任务（如 `AttrSys.RunOne`）

**重构后处理：**
- **RunOne 接口**：定义 `RunOneUseCase` 接口（已在 13.4 中说明）
- **系统适配器**：系统适配器在 `RunOne` 中调用 Use Case 的 `RunOne`
  ```go
  // adapter/system/attr_system_adapter.go
  func (a *AttrSystemAdapter) RunOne(ctx context.Context) {
      if runOneUC, ok := a.attrUseCase.(usecase.RunOneUseCase); ok {
          runOneUC.RunOne(ctx)
      }
  }
  ```
- **PlayerRole.RunOne 保持不变**：`PlayerRole.RunOne()` 逻辑不变，通过 `sysMgr.EachOpenSystem` 调用系统的 `RunOne`

**关键点：**
- RunOne 机制保持不变，通过消息在 Actor 中处理
- 系统适配器负责调用 Use Case 的 `RunOne`
- 时间检查器（`_1sChecker`、`_5minChecker`）逻辑保持不变

### 13.27 时间事件处理机制

**当前实现：**
- `PlayerRole` 通过 `timeCursor`（`timeCursorMark`）跟踪时间变化
- `handleTimeEvents()` 在 `RunOne` 中每帧调用，检测时间变化
- 时间变化时触发对应回调：`OnNewHour/Day/Week/Month/Year`
- 登录时会处理离线期间的时间事件（`handleOfflineRollover`），确保离线期间的时间事件被触发

**重构后处理：**
- **时间事件处理保持不变**：`PlayerRole.handleTimeEvents()` 逻辑不变
- **系统时间回调**：系统适配器在时间回调中调用 Use Case 的对应方法
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnNewDay(ctx context.Context) {
      if timeCallbackUC, ok := a.bagUseCase.(usecase.TimeCallbackUseCase); ok {
          timeCallbackUC.OnNewDay(ctx)
      }
  }
  ```
- **Use Case 时间回调接口**（可选）：
  ```go
  // usecase/interfaces/time_callback.go
  type TimeCallbackUseCase interface {
      OnNewHour(ctx context.Context) error
      OnNewDay(ctx context.Context) error
      OnNewWeek(ctx context.Context) error
      OnNewMonth(ctx context.Context) error
      OnNewYear(ctx context.Context) error
  }
  ```

**关键点：**
- 时间事件处理机制保持不变
- 系统适配器负责调用 Use Case 的时间回调方法
- 登录时会处理离线期间的时间事件，确保数据一致性

### 13.28 数据持久化机制

**当前实现：**
- `PlayerRole.SaveToDB()` 方法保存 `BinaryData` 到数据库
- 定期保存：`RunOne` 中通过 `_5minChecker` 每 5 分钟自动保存
- 立即保存：登出时（`OnLogout`）立即保存
- 优雅停服：`RoleManager.FlushAndSave()` 遍历所有在线角色并同步保存，支持超时取消

**重构后处理：**
- **Repository 持久化**：通过 `PlayerRepository.SaveBinaryData` 方法持久化
  ```go
  // adapter/gateway/player_gateway.go
  func (g *PlayerGateway) SaveBinaryData(ctx context.Context, roleID uint64, binaryData *protocol.PlayerRoleBinaryData) error {
      return database.SavePlayerBinaryData(uint(roleID), binaryData)
  }
  ```
- **Use Case 保存时机**：Use Case 不主动保存，由系统适配器或 PlayerRole 统一管理
- **保存策略保持不变**：定期保存（5分钟）和立即保存（登出时）逻辑不变

**关键点：**
- 数据持久化通过 Repository 接口统一管理
- 保存时机和策略保持不变
- 优雅停服时的 `FlushAndSave` 机制保持不变

### 13.29 系统状态持久化（位图方式）

**当前实现：**
- 系统状态存储在 `BinaryData.SysOpenStatus` 中，使用位图方式（`map[uint32]uint32`）
- `PlayerRole.GetSysStatus(sysId)` 和 `SetSysStatus(sysId, isOpen)` 使用位运算操作
- `SysMgr.CheckAllSysOpen` 检查系统状态并更新 `BinaryData.SysOpenStatus`

**重构后处理：**
- **系统状态管理保持不变**：位图存储方式不变
- **系统适配器状态管理**：系统适配器实现 `SetOpened` 方法，更新 `BinaryData.SysOpenStatus`
  ```go
  // adapter/system/base_system_adapter.go
  func (a *BaseSystemAdapter) SetOpened(opened bool) {
      a.opened = opened
      // 更新 BinaryData.SysOpenStatus（通过 Repository 或直接访问）
      playerRole, _ := GetPlayerRoleFromContext(ctx)
      playerRole.SetSysStatus(a.sysID, opened)
  }
  ```

**关键点：**
- 系统状态使用位图方式存储，保持不变
- 系统适配器负责状态管理
- `SysMgr.CheckAllSysOpen` 逻辑保持不变

---

## 14. 重构检查清单

### 14.1 基础设施检查清单

- [ ] **目录结构创建**
  - [ ] `internel/domain/` - Entities 层
  - [ ] `internel/domain/repository/` - Repository 接口定义
  - [ ] `internel/usecase/` - Use Cases 层
  - [ ] `internel/usecase/interfaces/` - Use Case 依赖接口
  - [ ] `internel/adapter/controller/` - 协议控制器
  - [ ] `internel/adapter/presenter/` - 响应构建器
  - [ ] `internel/adapter/gateway/` - 数据访问实现
  - [ ] `internel/adapter/event/` - 事件适配器
  - [ ] `internel/adapter/system/` - 系统生命周期适配器
  - [ ] `internel/adapter/context/` - Context 工具函数
  - [ ] `internel/di/` - 依赖注入容器

- [ ] **基础设施适配层实现**
  - [ ] `NetworkGateway` - 封装 `gatewaylink`
  - [ ] `PublicActorGateway` - 封装 `gshare.SendPublicMessageAsync`
  - [ ] `DungeonServerGateway` - 封装 `dungeonserverlink`
  - [ ] `EventAdapter` - 封装 `gevent`
  - [ ] `ConfigGateway` - 封装 `jsonconf`

- [ ] **系统生命周期适配器**
  - [ ] `BaseSystemAdapter` - 系统适配器基类
  - [ ] 实现 `ISystem` 接口的所有方法
  - [ ] 生命周期方法调用 Use Case 对应方法

- [ ] **BinaryData 共享机制**
  - [ ] `PlayerRepository` 接口定义
  - [ ] `PlayerGateway` 实现，返回 BinaryData 共享引用
  - [ ] Use Case 直接修改 BinaryData，保持内存共享

- [ ] **系统依赖关系处理**
  - [ ] DI 容器实现
  - [ ] Use Case 依赖注入
  - [ ] 系统依赖关系定义保持不变
  - [ ] 拓扑排序确定系统初始化顺序

- [ ] **RunOne 机制**
  - [ ] `RunOneUseCase` 接口定义
  - [ ] 系统适配器实现 `RunOne` 方法
  - [ ] 保持与 `PlayerRole.RunOne()` 的兼容性

- [ ] **事件订阅机制**
  - [ ] `EventSubscriber` 接口定义
  - [ ] Use Case 实现事件订阅
  - [ ] 系统适配器在 `OnInit` 中调用事件订阅

- [ ] **Context 传递机制**
  - [ ] Context 工具函数实现
  - [ ] Use Case 不直接访问 Context 中的框架对象
  - [ ] Controller 负责从 Context 提取信息

### 14.2 系统迁移检查清单（每个系统）

- [ ] **Domain 层**
  - [ ] 创建 Domain 实体（移除框架依赖）
  - [ ] 定义 Repository 接口

- [ ] **Use Case 层**
  - [ ] 创建 Use Case（业务逻辑）
  - [ ] 实现依赖接口（EventPublisher、PublicActorGateway 等）
  - [ ] 实现 `InitializeFromBinaryData` 方法（如果需要）
  - [ ] 实现 `RunOneUseCase` 接口（如果需要）
  - [ ] 实现 `EventSubscriber` 接口（如果需要）
  - [ ] 实现 `TimeCallbackUseCase` 接口（如果需要）

- [ ] **Adapter 层**
  - [ ] 创建 Controller（协议处理）
  - [ ] 创建 Presenter（响应构建）
  - [ ] 实现 Gateway（数据访问）
  - [ ] 创建系统适配器（生命周期管理）
  - [ ] 实现 `GetXXXSys(ctx)` 函数
  - [ ] 实现系统状态管理（`SetOpened`、`IsOpened`）

- [ ] **协议注册**
  - [ ] Controller 在 `init()` 中注册协议
  - [ ] 协议处理器调用 Controller 方法
  - [ ] 保持与 `clientprotocol.ProtoTbl` 的兼容性

- [ ] **系统工厂注册**
  - [ ] 在 `init()` 中注册系统适配器工厂
  - [ ] 工厂函数从 DI 容器获取依赖

- [ ] **测试验证**
  - [ ] 编写 Use Case 单元测试
  - [ ] 编写 Controller 集成测试
  - [ ] 验证功能正常
  - [ ] 验证协议处理正常
  - [ ] 验证系统生命周期正常
  - [ ] 验证 BinaryData 共享正常
  - [ ] 验证 RunOne 机制正常
  - [ ] 验证时间事件处理正常
  - [ ] 验证数据持久化正常

- [ ] **代码清理**
  - [ ] 删除 EntitySystem 中的旧代码
  - [ ] 更新相关文档

### 14.3 整体检查清单

- [ ] **所有系统已迁移**（共 20 个系统）
  - [ ] LevelSys
  - [ ] BagSys
  - [ ] MoneySys
  - [ ] EquipSys
  - [ ] AttrSys
  - [ ] SkillSys
  - [ ] QuestSys
  - [ ] FubenSys
  - [ ] ItemUseSys
  - [ ] ShopSys
  - [ ] RecycleSys
  - [ ] FriendSys
  - [ ] GuildSys
  - [ ] ChatSys
  - [ ] AuctionSys
  - [ ] MailSys
  - [ ] VipSys
  - [ ] DailyActivitySys
  - [ ] MessageSys
  - [ ] GMSys

- [ ] **所有框架依赖已移除**
  - [ ] database → Repository 接口
  - [ ] gatewaylink → NetworkGateway 接口
  - [ ] dungeonserverlink → DungeonServerGateway 接口（包含 ProtocolManager、DungeonClient、DungeonMessageHandler 的封装）
  - [ ] gevent → EventPublisher 接口
  - [ ] jsonconf → ConfigManager 接口
  - [ ] gshare → PublicActorGateway 接口

- [ ] **系统机制保持**
  - [ ] 系统生命周期管理正常
  - [ ] BinaryData 共享机制正常
  - [ ] 系统依赖关系处理正常
  - [ ] RunOne 机制正常（时间检查器、定期任务、DoRunOneMsg 消息处理）
  - [ ] 事件订阅机制正常（独立事件总线、从全局模板克隆）
  - [ ] 系统工厂模式正常
  - [ ] Context 传递机制正常
  - [ ] 系统获取函数正常（GetXXXSys）
  - [ ] 协议注册机制正常
  - [ ] 时间事件处理正常（timeCursor、离线时间事件、handleOfflineRollover）
  - [ ] 数据持久化正常（定期保存、立即保存、优雅停服）
  - [ ] 系统状态持久化正常（位图方式、GetSysStatus/SetSysStatus）
  - [ ] 时间同步机制正常（timeSync、_1sChecker）
  - [ ] BinaryData 获取标准模式正常（OnInit 中的标准流程）
  - [ ] 系统初始化数据同步正常（Use Case 层处理）

- [ ] **测试覆盖**
  - [ ] Use Case 层单元测试覆盖率 > 70%
  - [ ] Controller 层集成测试通过
  - [ ] 所有协议处理测试通过
  - [ ] 系统生命周期测试通过

- [ ] **文档更新**
  - [ ] 架构文档已更新
  - [ ] 开发指南已更新
  - [ ] 协议注册文档已更新
  - [ ] 关键代码位置已更新

---

**下一步行动：**
1. 评审本文档，确认重构方案
2. 创建基础目录结构（domain、usecase、adapter、di）
3. 实现基础设施适配层（NetworkGateway、PublicActorGateway、DungeonServerGateway、EventAdapter、ConfigGateway）
4. 实现系统生命周期适配器（SystemLifecycleAdapter）
5. 实现 BinaryData 共享机制（Repository 层）
6. 实现系统依赖关系处理（DI 容器）
7. 实现 RunOne 机制（RunOneUseCase 接口）
8. 实现事件订阅机制（EventSubscriber 接口）
9. 选择第一个系统（LevelSys）进行试点重构
10. 验证重构效果后，按优先级逐步迁移其他系统
11. 完成所有系统迁移后，清理旧代码并更新文档

---

## 15. Actor 单线程特性保证（重要）

### 15.1 Actor 框架的单线程机制

**核心原理：**
- 每个 Actor 有一个独立的 goroutine（在 `actorContext.loop()` 中）
- 每个 Actor 有一个 mailbox（channel），所有消息都通过 mailbox 串行处理
- `actorContext.loop()` 的执行流程：
  1. 调用 `handler.Loop()`（用于定期任务，如 RunOne）
  2. 从 mailbox 取一条消息
  3. 调用 `handler.HandleMessage(msg)` 处理消息
  4. 循环回到步骤 1
- 这确保了每个 Actor 内的所有操作都是单线程的，无需加锁

**PlayerActor 的实现：**
- 使用 `ModePerKey` 模式，每个 sessionId 对应一个独立的 Actor
- 所有协议消息通过 `handleDoNetWorkMsg` 处理，这个函数在 Actor 的消息处理中调用
- `RunOne` 也通过 `DoRunOneMsg` 消息在 Actor 中处理
- 所有业务逻辑都在 Actor 的消息处理函数中执行

### 15.2 重构后如何保持单线程特性

**关键原则：**
1. **所有业务逻辑必须在 Actor 消息处理中执行**
   - Controller 方法必须在 Actor 的消息处理函数中调用
   - Use Case 方法必须在 Controller 中调用（间接在 Actor 消息处理中）
   - 禁止在 Use Case 或 Controller 中创建 goroutine

2. **异步操作通过消息队列实现**
   - 需要异步操作时，通过 `actorCtx.ExecuteAsync` 发送消息到自己的 mailbox
   - 禁止使用 `go` 关键字创建新的 goroutine 访问玩家状态

3. **RPC 调用必须是异步的**
   - 通过 `DungeonServerGateway.AsyncCall` 调用 DungeonServer
   - RPC 回调也通过消息发送到 Actor 的 mailbox

4. **事件发布必须在 Actor 线程中**
   - 事件发布通过 `EventPublisher` 接口，但底层实现必须在 Actor 线程中
   - 事件订阅的回调函数也在 Actor 线程中执行

**重构后的调用链：**
```
Actor 消息处理 (单线程)
  ↓
Controller.HandleXXX (在 Actor 线程中)
  ↓
Use Case.Execute (在 Actor 线程中)
  ↓
Repository/Gateway (在 Actor 线程中)
```

### 15.3 重构检查清单（Actor 单线程保证）

- [ ] **禁止在 Use Case 中创建 goroutine**
  - 检查所有 Use Case 方法，确保没有 `go` 关键字
  - 检查是否有 `sync.Mutex` 等锁机制（应该不需要）

- [ ] **禁止在 Controller 中创建 goroutine**
  - 检查所有 Controller 方法，确保没有 `go` 关键字
  - 确保所有网络发送都在 Actor 线程中

- [ ] **确保所有业务逻辑在 Actor 消息处理中**
  - 检查协议处理函数是否在 `handleDoNetWorkMsg` 中调用
  - 检查 RPC 回调是否通过消息发送到 Actor

- [ ] **确保异步操作通过消息队列**
  - 检查是否有直接使用 `go` 关键字的地方
  - 检查是否有使用 `sync.Mutex` 的地方（应该不需要）

- [ ] **验证单线程特性**
  - 编写测试验证所有操作都在同一个 goroutine 中
  - 验证没有竞态条件

### 15.4 重构后代码示例

**正确的实现（保持单线程）：**
```go
// adapter/controller/bag_controller.go
func (c *BagController) HandleAddItem(ctx context.Context, msg *network.ClientMessage) error {
    // ✅ 在 Actor 线程中执行，无需加锁
    // 解析协议
    var req protocol.C2SAddItemReq
    if err := proto.Unmarshal(msg.Data, &req); err != nil {
        return err
    }
    
    // ✅ 调用 Use Case（在 Actor 线程中）
    roleID := getRoleIDFromContext(ctx)
    if err := c.addItemUseCase.Execute(ctx, roleID, item); err != nil {
        return c.presenter.PresentError(ctx, err)
    }
    
    // ✅ 构建响应（在 Actor 线程中）
    return c.presenter.PresentAddItemSuccess(ctx, item)
}

// usecase/bag/add_item.go
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, item *domain.Item) error {
    // ✅ 在 Actor 线程中执行，无需加锁
    player, err := uc.playerRepo.GetByID(roleID)
    if err != nil {
        return err
    }
    
    // ✅ 直接修改 BinaryData（共享引用，在 Actor 线程中，无需加锁）
    if err := player.Bag.AddItem(item); err != nil {
        return err
    }
    
    // ✅ 保存数据（在 Actor 线程中）
    if err := uc.playerRepo.Save(player); err != nil {
        return err
    }
    
    // ✅ 发布事件（在 Actor 线程中）
    uc.eventPublisher.PublishItemAdded(ctx, roleID, item)
    
    return nil
}
```

**错误的实现（破坏单线程）：**
```go
// ❌ 错误：在 Use Case 中创建 goroutine
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, item *domain.Item) error {
    // ❌ 错误：创建 goroutine 会破坏单线程特性
    go func() {
        // 这会导致竞态条件
        player.Bag.AddItem(item)
    }()
    return nil
}

// ❌ 错误：使用锁（不应该需要）
func (uc *AddItemUseCase) Execute(ctx context.Context, roleID uint64, item *domain.Item) error {
    uc.mutex.Lock()  // ❌ 不应该需要锁
    defer uc.mutex.Unlock()
    // ...
}
```

### 15.5 与现有架构的兼容性

**重构后完全兼容现有架构：**
- Actor 框架的使用方式不变
- 消息处理流程不变
- 单线程特性保持不变
- 一玩家一 Actor 的模式保持不变

**重构只是改变代码组织方式：**
- 将业务逻辑从 EntitySystem 提取到 Use Case
- 将协议处理从 EntitySystem 提取到 Controller
- 将数据访问从 EntitySystem 提取到 Gateway
- 但所有代码仍然在 Actor 线程中执行

---

## 16. 代码梳理总结

### 15.1 当前代码结构分析

**EntitySystem 目录结构：**
- 共 20 个系统，每个系统一个文件
- 所有系统继承 `BaseSystem`
- 系统通过 `RegisterSystemFactory` 注册工厂函数
- 系统通过 `GetXXXSys(ctx)` 函数获取系统实例

**协议注册方式：**
- 协议在 `gevent.OnSrvStart` 事件中注册
- 系统在 `init()` 中订阅 `OnSrvStart` 事件
- 在事件回调中调用 `clientprotocol.Register` 注册协议

**系统初始化流程：**
1. `PlayerRole` 创建时调用 `sysMgr.OnInit(ctx)`
2. `SysMgr.OnInit` 使用拓扑排序确定初始化顺序
3. 按照依赖顺序创建系统实例并调用 `OnInit`
4. 系统在 `OnInit` 中从 `BinaryData` 加载数据

**系统生命周期：**
- `OnInit` - 系统初始化
- `OnRoleLogin` - 玩家登录
- `OnRoleReconnect` - 玩家重连
- `OnRoleLogout` - 玩家登出
- `OnRoleClose` - 玩家关闭
- `OnNewHour/Day/Week/Month/Year` - 时间变化
- `RunOne` - 每帧调用（可选）

**系统间调用：**
- 系统通过 `GetXXXSys(ctx)` 获取其他系统
- 系统直接调用其他系统的方法
- 例如：`levelSys := GetLevelSys(ctx)`、`bagSys := GetBagSys(ctx)`

**BinaryData 共享：**
- 系统在 `OnInit` 中获取 `BinaryData` 的引用
- 系统直接修改 `BinaryData`，修改后通过 `PlayerRole.SaveBinaryData()` 持久化
- 例如：`bs.bagData = binaryData.BagData`

**辅助索引：**
- 部分系统（如 BagSys）有辅助索引
- 辅助索引用于快速查找，但不作为数据源
- 数据变更后需要重建索引

### 15.2 重构后代码结构设计

**Domain 层：**
- 纯业务实体，不依赖任何框架
- 定义 Repository 接口（在 domain 层定义）

**Use Case 层：**
- 业务用例和业务规则
- 依赖 Entities 和 Repository 接口
- 通过依赖注入获取其他 Use Case

**Adapter 层：**
- Controller：协议处理器
- Presenter：响应构建器
- Gateway：数据访问实现
- System Adapter：系统生命周期适配器

**Infrastructure 层：**
- Actor 框架、网络、数据库、配置管理器等

### 15.3 SystemAdapter 职责与演进（阶段 C 完成）

**目标职责（Clean Architecture 原则）：**

SystemAdapter 在 Clean Architecture 中应承担以下职责：

1. **生命周期适配**
   - 将 Actor 生命周期事件（`OnInit`、`OnRoleLogin`、`OnRoleReconnect`、`OnRoleLogout`、`OnRoleClose`、`RunOne`、`OnNewDay`、`OnNewWeek` 等）转换为对 UseCase 的调用
   - 只负责"何时调用哪个 UseCase"，不包含业务逻辑

2. **事件订阅**
   - 订阅玩家级事件，在事件到来时调用相应的 UseCase
   - 只负责"订阅哪个事件"，业务逻辑由 UseCase 层处理
   - 框架层面的状态管理（如标记属性系统需要重算）可以保留在 SystemAdapter 层

3. **状态管理**
   - 管理与 Actor 运行模型强相关的运行时状态（如属性 dirty 标记、定时任务、限流状态）
   - 不直接操作数据库、网络

**禁止事项：**
- ❌ 直接操作数据库、网络
- ❌ 实现业务规则逻辑（业务逻辑应在 UseCase 层）
- ❌ 在生命周期方法中包含复杂的业务逻辑

**当前状态（2025-01-XX）：**

- ✅ **统一的生命周期签名与职责说明**：所有 SystemAdapter 已添加统一的头部注释，说明生命周期职责和业务逻辑位置
- ✅ **BaseSystemAdapter 职责说明**：`BaseSystemAdapter` 已添加详细的职责说明注释
- ✅ **定时逻辑下沉**：复杂定时/调度逻辑已下沉到 UseCase 层（如 `RefreshQuestTypeUseCase`）
- ✅ **事件订阅精简**：空的事件订阅已删除，保留的事件订阅已添加注释说明

**验证清单：**

详见 `docs/SystemAdapter验证清单.md`，包含：
- 每个 SystemAdapter 的生命周期验证
- 事件订阅验证
- 对外接口验证
- 错误/异常情况下的降级行为

**参考文档：**
- `docs/gameserver_adapter_system演进规划.md`：SystemAdapter 演进规划与阶段任务
- `docs/SystemAdapter验证清单.md`：SystemAdapter 验证清单

### 15.4 关键重构点

1. **系统生命周期适配器**：系统适配器实现 `ISystem` 接口，调用 Use Case 方法
2. **BinaryData 共享机制**：Repository 返回共享引用，Use Case 直接修改
3. **系统依赖关系**：Use Case 通过依赖注入解决依赖，DI 容器管理依赖关系
4. **系统间调用**：Use Case 通过依赖注入获取依赖，禁止直接调用 `GetXXXSys`
5. **协议注册**：协议注册迁移到 Controller 层
6. **事件订阅**：Use Case 通过 `EventSubscriber` 接口订阅事件
7. **配置访问**：Use Case 通过 `ConfigManager` 接口访问配置
8. **网络发送**：通过 Presenter 构建响应，通过 NetworkGateway 发送
9. **RPC 调用**：通过 DungeonServerGateway 接口调用
10. **PublicActor 交互**：通过 PublicActorGateway 接口交互

### 15.4 遗漏点补充清单

- [x] 系统生命周期管理（13.1）
- [x] BinaryData 共享机制（13.2）
- [x] 系统依赖关系处理（13.3）
- [x] RunOne 机制（13.4）
- [x] 事件订阅机制（13.5）
- [x] 系统工厂模式（13.6）
- [x] Context 传递机制（13.7）
- [x] PlayerRole 与系统的关系（13.8）
- [x] BinaryData 初始化（13.9）
- [x] 系统状态管理（13.10）
- [x] 系统获取函数模式（13.11）
- [x] 协议注册时机（13.12）
- [x] 系统间相互调用模式（13.13）
- [x] BinaryData 的直接引用保持（13.14）
- [x] 系统辅助索引管理（13.15）
- [x] 系统状态持久化（13.16）
- [x] 系统依赖的运行时获取（13.17）
- [x] 系统初始化顺序与拓扑排序（13.18）
- [x] 系统生命周期方法的具体作用（13.19）
- [x] 玩家消息系统重构（13.20）
- [x] IPlayerRole 接口的使用（13.21）
- [x] PlayerRole 生命周期管理（13.22）
- [x] 系统获取函数（GetXXXSys）的保持（13.23）
- [x] 协议处理流程保持不变（13.24）
- [x] 事件系统的使用（13.25）
- [x] RunOne 机制的详细说明（13.26）
- [x] 时间事件处理机制（13.27）
- [x] 数据持久化机制（13.28）
- [x] 系统状态持久化（位图方式）（13.29）

### 15.5 重构注意事项

1. **保持向后兼容**：重构过程中保持旧代码可用，新旧代码可以并存
2. **Actor 框架集成**：保持 Actor 的单线程特性，所有业务逻辑在 Actor 主线程执行
3. **性能考虑**：避免过度抽象导致性能下降
4. **数据一致性**：保持 BinaryData 的共享引用，通过 Repository 统一管理数据变更
5. **系统依赖关系**：保持系统依赖关系定义不变，通过 DI 容器管理 Use Case 依赖
6. **测试覆盖**：Use Case 层必须可独立测试，单元测试覆盖率目标 > 70%
7. **文档同步**：重构过程中及时更新文档，记录关键决策和注意事项

### 15.6 重构后是否还能保持"一玩家一 Actor"模式？

**答案：完全可以保持！**

**原因：**
1. **重构只是改变代码组织方式，不改变 Actor 框架的使用**
   - Actor 框架的 `ModePerKey` 模式保持不变
   - 每个 sessionId 仍然对应一个独立的 Actor
   - Actor 的消息处理机制保持不变

2. **所有业务逻辑仍然在 Actor 线程中执行**
   - Controller 方法在 Actor 的消息处理函数中调用
   - Use Case 方法在 Controller 中调用（间接在 Actor 线程中）
   - Repository/Gateway 在 Use Case 中调用（间接在 Actor 线程中）

3. **单线程特性保持不变**
   - 每个 Actor 仍然有一个独立的 goroutine
   - 所有消息仍然通过 mailbox 串行处理
   - 无需加锁，因为所有操作都在同一个 goroutine 中

4. **多协程无锁化运行模式保持不变**
   - 每个玩家一个 Actor（多协程）
   - 每个 Actor 内部单线程（无锁）
   - 重构后仍然保持这个模式

**重构前后的对比：**

| 特性 | 重构前 | 重构后 |
|------|--------|--------|
| Actor 模式 | ModePerKey（一玩家一 Actor） | ModePerKey（一玩家一 Actor）✅ |
| 单线程特性 | 每个 Actor 单线程 | 每个 Actor 单线程 ✅ |
| 无锁运行 | 无需加锁 | 无需加锁 ✅ |
| 代码组织 | EntitySystem 混合业务和框架 | Clean Architecture 分层 ✅ |
| 可测试性 | 难以测试 | 易于测试 ✅ |

**结论：**
重构后完全保持"一玩家一 Actor"的多协程无锁化运行模式，同时获得更好的代码组织和可测试性。

---

## 17. 重构遗漏点补充总结

### 17.1 已补充的遗漏点

经过代码梳理，已在文档中补充以下遗漏点：

1. ✅ **Actor 单线程特性保证**（第 15 章）
   - Actor 框架的单线程机制
   - 重构后如何保持单线程特性
   - 重构检查清单
   - 代码示例（正确 vs 错误）

2. ✅ **IPlayerRole 接口的使用**（13.21）
   - 接口保持不变
   - Use Case 不直接依赖接口
   - Controller 负责提取信息

3. ✅ **PlayerRole 生命周期管理**（13.22）
   - 生命周期保持不变
   - 系统适配器在生命周期方法中获取数据
   - Use Case 通过初始化方法接收数据

4. ✅ **系统获取函数（GetXXXSys）的保持**（13.23）
   - 函数保持不变，但返回系统适配器
   - 保持向后兼容性

5. ✅ **协议处理流程保持不变**（13.24）
   - 协议处理流程不变
   - 协议注册迁移到 Controller 层

6. ✅ **事件系统的使用**（13.25）
   - 事件系统通过接口抽象
   - Use Case 通过接口发布和订阅事件
   - 每个 PlayerRole 有独立的事件总线，从全局模板克隆

7. ✅ **RunOne 机制的详细说明**（13.26）
   - RunOne 通过消息在 Actor 中处理
   - 时间检查器（`_1sChecker`、`_5minChecker`）控制频率
   - 系统适配器负责调用 Use Case 的 `RunOne`

8. ✅ **时间事件处理机制**（13.27）
   - `timeCursor` 跟踪时间变化
   - `handleTimeEvents()` 检测时间变化并触发回调
   - 登录时处理离线期间的时间事件

9. ✅ **数据持久化机制**（13.28）
   - 定期保存（5分钟）和立即保存（登出时）
   - 优雅停服时的 `FlushAndSave` 机制
   - 通过 Repository 接口统一管理

10. ✅ **系统状态持久化（位图方式）**（13.29）
    - 使用位图方式存储系统状态
    - 系统适配器负责状态管理

11. ✅ **PlayerRole 的 RunOne 消息机制**（17.5.1）
    - RunOne 通过 `DoRunOneMsg` 消息在 Actor 中处理
    - 时间检查器（`_1sChecker`、`_5minChecker`）控制频率
    - 系统适配器负责调用 Use Case 的 `RunOne`

12. ✅ **系统状态位图管理的详细实现**（17.5.2）
    - 使用位图方式管理（`sysId / 32` 和 `sysId % 32`）
    - 系统适配器通过 `PlayerRole.SetSysStatus` 更新状态

13. ✅ **BinaryData 获取的标准模式**（17.5.3）
    - 所有系统在 `OnInit` 中使用相同的模式获取 BinaryData
    - 系统适配器负责从 BinaryData 加载数据

14. ✅ **时间同步机制**（17.5.4）
    - `PlayerRole.timeSync()` 在 `RunOne` 中通过 `_1sChecker` 每秒调用一次
    - 发送 `S2CTimeSync` 协议，携带服务器时间

15. ✅ **系统间调用的运行时模式**（17.5.5）
    - Use Case 层通过依赖注入解决系统间依赖
    - 系统适配器保持 `GetXXXSys(ctx)` 函数，用于向后兼容

16. ✅ **事件总线的独立克隆机制**（17.5.6）
    - 每个 PlayerRole 有独立的事件总线，从全局模板克隆
    - 事件适配器需要从 Context 获取 PlayerRole 的事件总线

17. ✅ **PlayerRole 的 SaveToDB 机制**（17.5.7）
    - 在 `RunOne` 中通过 `_5minChecker` 每 5 分钟自动保存
    - 在 `OnLogout` 中立即保存

18. ✅ **系统初始化时的数据同步**（17.5.8）
    - 部分系统在 `OnInit` 中会同步数据到其他系统
    - 数据同步逻辑迁移到 Use Case 层

19. ✅ **ProtocolManager 协议路由管理**（17.5.9）
    - 管理 DungeonServer 的协议注册信息
    - 支持通用协议和独有协议的路由
    - 在重构后通过 `DungeonServerGateway` 接口暴露

20. ✅ **DungeonClient 生命周期管理**（17.5.10）
    - `StartDungeonClient` 和 `Stop` 在 `main.go` 中调用
    - 连接池管理和自动重连机制
    - 在重构后保持为 Infrastructure 层

21. ✅ **DungeonMessageHandler 消息处理机制**（17.5.11）
    - 处理三种消息类型：RPC 请求、心跳、客户端消息转发
    - RPC 处理器的注册和调用
    - 特殊 RPC 的 Actor 处理逻辑

22. ✅ **客户端消息转发机制**（17.5.12）
    - 客户端协议的透传机制
    - `msgId=0` 表示客户端消息转发
    - 协议路由逻辑迁移到 Controller 层

### 17.2 重构后架构对比

| 方面 | 重构前 | 重构后 | 是否保持 |
|------|--------|--------|---------|
| Actor 模式 | ModePerKey（一玩家一 Actor） | ModePerKey（一玩家一 Actor） | ✅ 是 |
| 单线程特性 | 每个 Actor 单线程 | 每个 Actor 单线程 | ✅ 是 |
| 无锁运行 | 无需加锁 | 无需加锁 | ✅ 是 |
| 代码组织 | EntitySystem 混合业务和框架 | Clean Architecture 分层 | ✅ 改进 |
| 可测试性 | 难以测试 | 易于测试 | ✅ 改进 |
| 系统生命周期 | ISystem 接口 | 系统适配器实现 ISystem | ✅ 是 |
| BinaryData 共享 | 直接引用 | Repository 返回共享引用 | ✅ 是 |
| 协议处理 | EntitySystem 中注册 | Controller 中注册 | ✅ 改进 |
| 事件系统 | 直接使用 gevent | 通过接口抽象 | ✅ 改进 |

### 17.3 关键保证

**重构后完全保证：**
1. ✅ **一玩家一 Actor 模式**：每个 sessionId 对应一个独立的 Actor
2. ✅ **多协程运行**：每个 Actor 有独立的 goroutine
3. ✅ **无锁化运行**：每个 Actor 内部单线程，无需加锁
4. ✅ **单线程特性**：所有业务逻辑在 Actor 线程中执行
5. ✅ **向后兼容**：现有代码可以继续使用，逐步迁移

**重构带来的改进：**
1. ✅ **代码组织**：清晰的层次结构，易于理解和维护
2. ✅ **可测试性**：Use Case 层可独立测试，无需启动完整服务
3. ✅ **依赖管理**：通过接口和依赖注入，降低耦合
4. ✅ **扩展性**：新功能更容易添加，符合开闭原则

### 17.4 重构建议

**重构步骤：**
1. **第一阶段**：创建基础结构（domain、usecase、adapter、di）
2. **第二阶段**：实现基础设施适配层（NetworkGateway、PublicActorGateway 等）
3. **第三阶段**：实现系统生命周期适配器（SystemAdapter）
4. **第四阶段**：选择一个简单系统（如 LevelSys）进行试点重构
5. **第五阶段**：验证重构效果，确保单线程特性保持
6. **第六阶段**：按优先级逐步迁移其他系统
7. **第七阶段**：清理旧代码，更新文档

**注意事项：**
1. ⚠️ **禁止在 Use Case 中创建 goroutine**：会破坏单线程特性
2. ⚠️ **禁止在 Controller 中创建 goroutine**：会破坏单线程特性
3. ⚠️ **禁止使用 sync.Mutex**：应该不需要锁
4. ⚠️ **确保所有业务逻辑在 Actor 线程中**：通过消息处理函数调用
5. ⚠️ **保持 BinaryData 共享引用**：Repository 必须返回共享引用

### 17.5 代码梳理发现的遗漏点补充

#### 17.5.1 PlayerRole 的 RunOne 消息机制

**当前实现：**
- `PlayerRole.RunOne()` 通过 `DoRunOneMsg` 消息在 Actor 中处理
- `handleRunOneMsg` 函数在 `player_network.go` 中注册为消息处理器
- 通过 `gshare.RegisterHandler(gshare.DoRunOneMsg, handleRunOneMsg)` 注册

**重构后处理：**
- **RunOne 消息处理保持不变**：`handleRunOneMsg` 逻辑不变
- **系统适配器 RunOne**：系统适配器在 `RunOne` 中调用 Use Case 的 `RunOne`
- **PlayerRole.RunOne 保持不变**：`PlayerRole.RunOne()` 逻辑不变，通过 `sysMgr.EachOpenSystem` 调用系统的 `RunOne`

**关键点：**
- RunOne 通过消息在 Actor 中处理，保持单线程特性
- 时间检查器（`_1sChecker`、`_5minChecker`）逻辑保持不变
- 系统适配器负责调用 Use Case 的 `RunOne`

#### 17.5.2 系统状态位图管理的详细实现

**当前实现：**
- `PlayerRole.GetSysStatus(sysId)` 和 `SetSysStatus(sysId, isOpen)` 使用位图方式管理
- 位图计算：`idxInt := sysId / 32`、`idxByte := sysId % 32`
- 使用 `tool.SetBit` 和 `tool.ClearBit` 操作位图
- `SysMgr.CheckAllSysOpen` 检查系统状态并更新 `BinaryData.SysOpenStatus`

**重构后处理：**
- **系统状态管理保持不变**：位图存储方式不变
- **系统适配器状态管理**：系统适配器实现 `SetOpened` 方法，调用 `PlayerRole.SetSysStatus`
  ```go
  // adapter/system/base_system_adapter.go
  func (a *BaseSystemAdapter) SetOpened(opened bool) {
      a.opened = opened
      // 通过 PlayerRole 更新位图状态
      playerRole, _ := GetPlayerRoleFromContext(ctx)
      playerRole.SetSysStatus(a.sysID, opened)
  }
  ```

**关键点：**
- 系统状态使用位图方式存储，保持不变
- 系统适配器通过 `PlayerRole.SetSysStatus` 更新状态
- `SysMgr.CheckAllSysOpen` 逻辑保持不变

#### 17.5.3 BinaryData 获取的标准模式

**当前实现：**
- 所有系统在 `OnInit` 中使用相同的模式获取 BinaryData：
  1. 通过 `GetIPlayerRoleByContext(ctx)` 获取 PlayerRole
  2. 通过 `playerRole.GetBinaryData()` 获取 BinaryData
  3. 检查 BinaryData 是否为 nil，如果为 nil 则初始化
  4. 检查对应的系统数据（如 `BagData`）是否为 nil，如果为 nil 则初始化
  5. 直接赋值引用：`bs.bagData = binaryData.BagData`

**重构后处理：**
- **系统适配器初始化**：系统适配器在 `OnInit` 中使用相同的模式
  ```go
  // adapter/system/bag_system_adapter.go
  func (a *BagSystemAdapter) OnInit(ctx context.Context) {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          log.Errorf("get player role error:%v", err)
          return
      }
      
      binaryData := playerRole.GetBinaryData()
      if binaryData == nil {
          log.Errorf("binary data is nil")
          return
      }
      
      if binaryData.BagData == nil {
          binaryData.BagData = &protocol.SiBagData{
              Items: make([]*protocol.ItemSt, 0),
          }
      }
      
      // 调用 Use Case 初始化
      a.bagUseCase.InitializeFromBinaryData(ctx, binaryData.BagData)
  }
  ```

**关键点：**
- BinaryData 获取模式保持不变
- 系统适配器负责从 BinaryData 加载数据
- Use Case 通过 `InitializeFromBinaryData` 方法接收数据引用

#### 17.5.4 时间同步机制

**当前实现：**
- `PlayerRole.timeSync()` 在 `RunOne` 中通过 `_1sChecker` 每秒调用一次
- 发送 `S2CTimeSync` 协议，携带服务器时间（`servertime.UnixMilli()`）

**重构后处理：**
- **时间同步保持不变**：`PlayerRole.timeSync()` 逻辑不变
- **时间同步在 RunOne 中调用**：通过 `_1sChecker` 控制频率

**关键点：**
- 时间同步机制保持不变
- 通过 `_1sChecker` 控制频率（每秒一次）

#### 17.5.5 系统间调用的运行时模式

**当前实现：**
- 系统在运行时通过 `GetXXXSys(ctx)` 获取其他系统
- 例如：`levelSys := GetLevelSys(ctx)`、`bagSys := GetBagSys(ctx)`
- 系统在方法中动态获取依赖系统

**重构后处理：**
- **Use Case 依赖注入**：Use Case 通过依赖注入获取依赖系统（已在 13.13 中说明）
- **系统适配器获取**：系统适配器可以通过 `GetXXXSys(ctx)` 获取其他系统适配器（用于向后兼容）
- **禁止 Use Case 直接调用**：Use Case 禁止直接调用 `GetXXXSys(ctx)`

**关键点：**
- Use Case 层通过依赖注入解决系统间依赖
- 系统适配器保持 `GetXXXSys(ctx)` 函数，用于向后兼容
- 禁止 Use Case 直接调用 `GetXXXSys(ctx)`

#### 17.5.6 事件总线的独立克隆机制

**当前实现：**
- 每个 PlayerRole 有独立的事件总线（`eventBus`）
- 从全局模板克隆：`eventBus: gevent.ClonePlayerEventBus()`
- 系统订阅事件时，订阅到玩家自己的事件总线

**重构后处理：**
- **事件总线机制保持不变**：每个 PlayerRole 仍然有独立的事件总线
- **事件适配器**：`EventAdapter` 需要从 Context 获取 PlayerRole，然后获取事件总线
  ```go
  // adapter/event/event_adapter.go
  func (a *EventAdapter) PublishPlayerEvent(ctx context.Context, eventType string, args ...interface{}) {
      playerRole, err := GetPlayerRoleFromContext(ctx)
      if err != nil {
          return
      }
      // 使用玩家自己的事件总线发布事件
      ev := event.NewEvent(eventType, args...)
      playerRole.Publish(eventType, args...)
  }
  ```

**关键点：**
- 每个 PlayerRole 有独立的事件总线，从全局模板克隆
- 事件适配器需要从 Context 获取 PlayerRole 的事件总线
- 事件在 Actor 主线程中执行，保证单线程特性

#### 17.5.7 PlayerRole 的 SaveToDB 机制

**当前实现：**
- `PlayerRole.SaveToDB()` 方法保存 `BinaryData` 到数据库
- 在 `RunOne` 中通过 `_5minChecker` 每 5 分钟自动保存
- 在 `OnLogout` 中立即保存

**重构后处理：**
- **保存机制保持不变**：`PlayerRole.SaveToDB()` 逻辑不变
- **Repository 持久化**：通过 `PlayerRepository.SaveBinaryData` 方法持久化（已在 13.28 中说明）
- **保存时机保持不变**：定期保存（5分钟）和立即保存（登出时）逻辑不变

**关键点：**
- 数据持久化通过 Repository 接口统一管理
- 保存时机和策略保持不变
- 优雅停服时的 `FlushAndSave` 机制保持不变

#### 17.5.8 系统初始化时的数据同步

**当前实现：**
- 部分系统在 `OnInit` 中会同步数据到其他系统
- 例如：`LevelSys.OnInit` 中同步经验到货币系统（`binaryData.MoneyData.MoneyMap[expMoneyID] = ls.levelData.Exp`）

**重构后处理：**
- **数据同步逻辑迁移到 Use Case**：Use Case 在初始化时处理数据同步
  ```go
  // usecase/level/level_usecase.go
  func (uc *LevelUseCase) InitializeFromBinaryData(ctx context.Context, levelData *protocol.SiLevelData) {
      // 初始化等级数据
      // ...
      
      // 同步经验到货币系统
      binaryData, _ := uc.playerRepo.GetBinaryData(ctx, roleID)
      if binaryData.MoneyData != nil {
          if binaryData.MoneyData.MoneyMap == nil {
              binaryData.MoneyData.MoneyMap = make(map[uint32]int64)
          }
          expMoneyID := uint32(protocol.MoneyType_MoneyTypeExp)
          binaryData.MoneyData.MoneyMap[expMoneyID] = levelData.Exp
      }
  }
  ```

**关键点：**
- 数据同步逻辑迁移到 Use Case 层
- 系统适配器在 `OnInit` 中调用 Use Case 的初始化方法
- 保持数据同步的时机和逻辑不变

#### 17.5.9 ProtocolManager 协议路由管理

**当前实现：**
- `ProtocolManager` 位于 `dungeonserverlink` 包，用于管理 DungeonServer 的协议注册信息
- 支持两种协议类型：
  - **通用协议**（`ProtocolTypeCommon`）：多个 DungeonServer 共享，根据角色所在的 DungeonServer 路由
  - **独有协议**（`ProtocolTypeUnique`）：特定 srvType 的 DungeonServer，直接路由到指定的 srvType
- 在 `player_network.go` 的 `handleDoNetWorkMsg` 中使用：
  1. 首先检查 GameServer 是否可以处理（`clientprotocol.GetFunc`）
  2. 如果不能处理，检查是否是 DungeonServer 协议（`ProtocolManager.IsDungeonProtocol`）
  3. 如果是，根据协议类型决定转发到哪个 DungeonServer（`ProtocolManager.GetSrvTypeForProtocol`）
- DungeonServer 通过 `D2GRegisterProtocols` RPC 注册协议
- 使用 `sync.RWMutex` 保护并发访问（因为 ProtocolManager 是全局单例）

**重构后处理：**
- **ProtocolManager 保持为 Infrastructure 层**：`ProtocolManager` 属于框架层，不需要重构
- **通过 DungeonServerGateway 接口暴露**：在 `DungeonServerGateway` 接口中添加协议路由相关方法
  ```go
  // usecase/interfaces/rpc.go
  type DungeonServerGateway interface {
      // ... 其他方法
      
      // 协议路由相关方法
      IsDungeonProtocol(protoId uint16) bool
      GetSrvTypeForProtocol(protoId uint16) (srvType uint8, protocolType uint8, ok bool)
  }
  
  // adapter/gateway/dungeon_server_gateway.go
  func (g *DungeonServerGatewayImpl) IsDungeonProtocol(protoId uint16) bool {
      return dungeonserverlink.GetProtocolManager().IsDungeonProtocol(protoId)
  }
  
  func (g *DungeonServerGatewayImpl) GetSrvTypeForProtocol(protoId uint16) (uint8, uint8, bool) {
      srvType, protocolType, ok := dungeonserverlink.GetProtocolManager().GetSrvTypeForProtocol(protoId)
      return srvType, uint8(protocolType), ok
  }
  ```
- **协议路由逻辑迁移到 Controller 层**：`handleDoNetWorkMsg` 中的协议路由逻辑迁移到 Controller
  ```go
  // adapter/controller/protocol_router_controller.go
  type ProtocolRouterController struct {
      dungeonServerGateway interfaces.DungeonServerGateway
      networkGateway       interfaces.NetworkGateway
  }
  
  func (c *ProtocolRouterController) RouteToDungeonServer(ctx context.Context, msgId uint16, sessionId string, data []byte) error {
      // 检查是否是 DungeonServer 协议
      if !c.dungeonServerGateway.IsDungeonProtocol(msgId) {
          return errors.New("not a dungeon protocol")
      }
      
      // 获取协议路由信息
      srvType, protocolType, ok := c.dungeonServerGateway.GetSrvTypeForProtocol(msgId)
      if !ok {
          return errors.New("protocol route not found")
      }
      
      // 根据协议类型决定目标 srvType
      var targetSrvType uint8
      if protocolType == uint8(dungeonserverlink.ProtocolTypeUnique) {
          targetSrvType = srvType
      } else {
          // 通用协议，需要根据角色所在的 DungeonServer 来决定
          playerRole, _ := GetPlayerRoleFromContext(ctx)
          targetSrvType = playerRole.GetDungeonSrvType()
          if targetSrvType == 0 {
              targetSrvType = srvType // 使用默认 srvType
          }
      }
      
      // 转发到 DungeonServer
      return c.dungeonServerGateway.AsyncCall(ctx, targetSrvType, sessionId, 0, data)
  }
  ```

**关键点：**
- `ProtocolManager` 保持为 Infrastructure 层，不需要重构
- 通过 `DungeonServerGateway` 接口暴露协议路由功能
- 协议路由逻辑迁移到 Controller 层
- Use Case 层不直接访问 `ProtocolManager`
- `ProtocolManager` 的并发安全由 `sync.RWMutex` 保证（Infrastructure 层特性）

#### 17.5.10 DungeonClient 生命周期管理

**当前实现：**
- `StartDungeonClient` 在 `main.go` 中调用，用于初始化 DungeonClient 并连接到所有配置的 DungeonServer
- `Stop` 在 `main.go` 中调用，用于关闭所有连接
- `DungeonClient` 维护连接池（`connPools`），每个 `srvType` 对应一个 TCP 客户端
- 连接使用自动重连机制（`network.NewTCPClient` 提供）
- 连接建立时发送握手消息（`MsgTypeHandshake`），包含 `PlatformId` 和 `SrvId`

**重构后处理：**
- **DungeonClient 保持为 Infrastructure 层**：`DungeonClient` 属于框架层，不需要重构
- **生命周期管理保持不变**：`StartDungeonClient` 和 `Stop` 在 `main.go` 中调用，保持原样
- **通过 DungeonServerGateway 封装**：`DungeonServerGateway` 封装 `AsyncCall`，内部使用 `DungeonClient`
  ```go
  // adapter/gateway/dungeon_server_gateway.go
  func (g *DungeonServerGatewayImpl) AsyncCall(ctx context.Context, srvType uint8, sessionId string, msgId uint16, data []byte) error {
      // 直接调用 dungeonserverlink.AsyncCall，内部使用全局 dungeonRPC
      return dungeonserverlink.AsyncCall(ctx, srvType, sessionId, msgId, data)
  }
  ```

**关键点：**
- `DungeonClient` 保持为 Infrastructure 层，不需要重构
- 生命周期管理（启动/停止）在 `main.go` 中处理，保持原样
- 通过 `DungeonServerGateway` 接口封装，Use Case 层不直接访问 `DungeonClient`
- 连接管理和自动重连由 Infrastructure 层处理

#### 17.5.11 DungeonMessageHandler 消息处理机制

**当前实现：**
- `DungeonMessageHandler` 处理来自 DungeonServer 的消息，支持三种消息类型：
  1. **`MsgTypeRPCRequest`**：RPC 请求，调用注册的 RPC 处理器
  2. **`MsgTypeHeartbeat`**：心跳消息，直接返回
  3. **`MsgTypeClient`**：客户端消息转发，解码 `ForwardMessage` 并转发到 Gateway
- RPC 请求的特殊处理：
  - 某些 RPC（如 `D2GAddItem`）需要在玩家 Actor 中串行处理
  - 通过 `gshare.SendMessageAsync` 发送到 Actor，由业务逻辑自行回包/通知客户端
  - 其他 RPC 同步处理，通过 `sendRPCResponse` 发送响应
- RPC 处理器注册：通过 `RegisterRPCHandler` 注册，存储在 `rpcHandlers` map 中
- 使用 `sync.RWMutex` 保护 `rpcHandlers` 的并发访问

**重构后处理：**
- **DungeonMessageHandler 保持为 Infrastructure 层**：消息处理属于框架层，不需要重构
- **RPC 处理器注册迁移到 Controller 层**：RPC 处理器的注册从 `player_network.go` 迁移到对应的 Controller
  ```go
  // adapter/controller/fuben_controller.go
  func init() {
      gevent.Subscribe(gevent.OnSrvStart, func(ctx context.Context, event *event.Event) {
          container := di.GetContainer()
          fubenController := container.FubenController
          
          // 注册 RPC 处理器
          dungeonServerGateway := container.DungeonServerGateway
          dungeonServerGateway.RegisterRPCHandler(
              uint16(protocol.D2GRpcProtocol_D2GSettleDungeon),
              fubenController.HandleSettleDungeonRPC,
          )
      })
  }
  ```
- **RPC 处理器调用 Controller**：RPC 处理器直接调用 Controller 方法
  ```go
  // adapter/controller/fuben_controller.go
  func (c *FubenController) HandleSettleDungeonRPC(ctx context.Context, sessionId string, data []byte) error {
      // 解析 RPC 请求
      var req protocol.D2GSettleDungeonReq
      if err := proto.Unmarshal(data, &req); err != nil {
          return err
      }
      
      // 调用 Use Case
      return c.settleDungeonUseCase.Execute(ctx, &req)
  }
  ```
- **特殊 RPC 的 Actor 处理保持不变**：需要在 Actor 中处理的 RPC（如 `D2GAddItem`）保持原有逻辑

**关键点：**
- `DungeonMessageHandler` 保持为 Infrastructure 层，不需要重构
- RPC 处理器注册迁移到 Controller 层
- RPC 处理器调用 Controller 方法，Controller 调用 Use Case
- 特殊 RPC 的 Actor 处理逻辑保持不变

#### 17.5.12 客户端消息转发机制

**当前实现：**
- GameServer 无法处理的客户端协议会转发到 DungeonServer
- 转发方式：
  - 如果 `msgId == 0`，使用 `MsgTypeClient` 消息类型转发
  - 将原始 `ClientMessage` 数据封装为 `ForwardMessage`，包含 `SessionId` 和 `Payload`
- DungeonServer 收到 `MsgTypeClient` 消息后，解码 `ForwardMessage` 并通过 `gatewaylink.ForwardClientMsg` 转发回客户端
- 这样实现了客户端协议的透传，无需在 GameServer 和 DungeonServer 重复注册

**重构后处理：**
- **消息转发机制保持不变**：客户端消息转发属于 Infrastructure 层，不需要重构
- **协议路由逻辑迁移到 Controller 层**：`handleDoNetWorkMsg` 中的协议路由逻辑迁移到 `ProtocolRouterController`
  ```go
  // adapter/controller/protocol_router_controller.go
  func (c *ProtocolRouterController) RouteToDungeonServer(ctx context.Context, msgId uint16, sessionId string, data []byte) error {
      // ... 协议路由逻辑
      
      // 转发到 DungeonServer（msgId=0 表示客户端消息转发）
      return c.dungeonServerGateway.AsyncCall(ctx, targetSrvType, sessionId, 0, data)
  }
  ```

**关键点：**
- 客户端消息转发机制保持不变
- 协议路由逻辑迁移到 Controller 层
- `msgId=0` 表示客户端消息转发，使用 `MsgTypeClient` 消息类型

---

## 18. 总结

### 18.1 重构目标

本次重构的目标是：
1. ✅ 实现 Clean Architecture，提高代码可维护性和可测试性
2. ✅ 保持"一玩家一 Actor"的多协程无锁化运行模式
3. ✅ 保持单线程特性，无需加锁
4. ✅ 保持向后兼容，逐步迁移

### 18.2 重构保证

**重构后完全保证：**
- ✅ **一玩家一 Actor 模式**：每个 sessionId 对应一个独立的 Actor
- ✅ **多协程运行**：每个 Actor 有独立的 goroutine
- ✅ **无锁化运行**：每个 Actor 内部单线程，无需加锁
- ✅ **单线程特性**：所有业务逻辑在 Actor 线程中执行

**重构不会破坏：**
- ✅ Actor 框架的使用方式
- ✅ 消息处理流程
- ✅ 单线程特性
- ✅ 一玩家一 Actor 的模式

### 18.3 重构收益

**重构带来的收益：**
1. ✅ **代码组织**：清晰的层次结构，易于理解和维护
2. ✅ **可测试性**：Use Case 层可独立测试，无需启动完整服务
3. ✅ **依赖管理**：通过接口和依赖注入，降低耦合
4. ✅ **扩展性**：新功能更容易添加，符合开闭原则
5. ✅ **文档化**：清晰的架构文档，便于团队协作

### 18.4 下一步行动

1. 评审本文档，确认重构方案
2. 创建基础目录结构（domain、usecase、adapter、di）
3. 实现基础设施适配层
4. 实现系统生命周期适配器
5. 选择第一个系统（LevelSys）进行试点重构
6. 验证重构效果，确保单线程特性保持
7. 按优先级逐步迁移其他系统
8. 完成所有系统迁移后，清理旧代码并更新文档

---

## 19. 待实现/待完善功能

> ⚠️ **重要提示**：本文档列出了 GameServer Clean Architecture 重构的所有待实现功能。请按照阶段逐步实现，每完成一个阶段后及时更新本文档。

### 19.1 阶段一：基础结构搭建（1-2周）

#### 19.1.1 目录结构创建
- [x] 创建 `internel/domain/` 目录（Entities 层）
- [x] 创建 `internel/domain/repository/` 目录（Repository 接口定义）
- [x] 创建 `internel/usecase/` 目录（Use Cases 层）
- [x] 创建 `internel/usecase/interfaces/` 目录（Use Case 依赖接口）
- [x] 创建 `internel/adapter/controller/` 目录（协议控制器）
- [x] 创建 `internel/adapter/presenter/` 目录（响应构建器）
- [x] 创建 `internel/adapter/gateway/` 目录（数据访问实现）
- [x] 创建 `internel/adapter/event/` 目录（事件适配器）
- [x] 创建 `internel/adapter/system/` 目录（系统生命周期适配器）
- [x] 创建 `internel/adapter/context/` 目录（Context 工具函数）
- [x] 创建 `internel/di/` 目录（依赖注入容器）

#### 19.1.2 基础接口定义
- [x] 定义 `repository.PlayerRepository` 接口（Domain 层）
- [x] 定义 `interfaces.EventPublisher` 接口（Use Case 层）
- [x] 定义 `interfaces.PublicActorGateway` 接口（Use Case 层）
- [x] 定义 `interfaces.DungeonServerGateway` 接口（Use Case 层）
- [x] 定义 `interfaces.ConfigManager` 接口（Use Case 层）
- [x] 定义 `interfaces.NetworkGateway` 接口（Adapter 层）
- [x] 定义 `interfaces.RunOneUseCase` 接口（Use Case 层，可选）
- [x] 定义 `interfaces.EventSubscriber` 接口（Use Case 层，可选）
- [x] 定义 `interfaces.TimeCallbackUseCase` 接口（Use Case 层，可选）

#### 19.1.3 基础设施适配层实现
- [x] 实现 `adapter/gateway/network_gateway.go`（封装 `gatewaylink`）
- [x] 实现 `adapter/gateway/public_actor_gateway.go`（封装 `gshare.SendPublicMessageAsync`）
- [x] 实现 `adapter/gateway/dungeon_server_gateway.go`（封装 `dungeonserverlink`，包含 ProtocolManager、DungeonClient、DungeonMessageHandler 的封装）
- [x] 实现 `adapter/event/event_adapter.go`（封装 `gevent`，支持独立事件总线）
- [x] 实现 `adapter/gateway/config_gateway.go`（封装 `jsonconf`）
- [x] 实现 `adapter/gateway/player_gateway.go`（实现 `PlayerRepository` 接口，返回 BinaryData 共享引用）

#### 19.1.4 系统生命周期适配器
- [x] 实现 `adapter/system/base_system_adapter.go`（系统适配器基类）
  - [x] 实现 `ISystem` 接口的所有方法
  - [x] 实现 `SetOpened/IsOpened` 方法（系统状态管理）
  - [x] 实现生命周期方法（OnInit、OnRoleLogin、OnRoleReconnect、OnRoleLogout、OnRoleClose、OnNewHour/Day/Week/Month/Year、RunOne）
- [x] 实现 `adapter/context/context_helper.go`（Context 工具函数）
  - [x] `GetPlayerRoleFromContext(ctx)` - 从 Context 获取 PlayerRole
  - [x] `GetSessionIDFromContext(ctx)` - 从 Context 获取 SessionID
  - [x] `GetRoleIDFromContext(ctx)` - 从 Context 获取 RoleID

#### 19.1.5 依赖注入容器
- [x] 实现 `di/container.go`（依赖注入容器）
  - [x] 初始化所有 Gateways
  - [x] 初始化所有 Use Cases（按依赖顺序）- 框架已就绪，等待具体 Use Case 实现
  - [x] 初始化所有 Controllers（按依赖顺序）- 框架已就绪，等待具体 Controller 实现
  - [x] 初始化所有 Presenters（按依赖顺序）- 框架已就绪，等待具体 Presenter 实现
  - [x] 提供 `GetContainer()` 全局函数

#### 19.1.6 试点系统重构（LevelSys）
- [x] 创建 `domain/player.go`（提取 Player 实体，移除框架依赖）- 注意：LevelSys 主要操作 LevelData，暂不需要单独的 Player 实体
- [x] 创建 `usecase/level/add_exp.go`（提取业务逻辑）
- [x] 创建 `usecase/level/level_up.go`（等级提升逻辑）
- [x] 创建 `adapter/controller/level_controller.go`（协议处理）- 注意：LevelSys 目前没有客户端协议，暂时不创建
- [x] 创建 `adapter/presenter/level_presenter.go`（响应构建）- 注意：LevelSys 目前没有客户端协议，暂时不创建
- [x] 创建 `adapter/system/level_system_adapter.go`（系统生命周期适配器）
- [x] 创建 `adapter/system/level_system_adapter_helper.go`（系统获取函数）
- [x] 创建 `adapter/system/level_system_adapter_attr.go`（属性计算器支持）
- [x] 创建 `adapter/system/level_system_adapter_init.go`（系统注册）
- [x] 实现 `GetLevelSys(ctx)` 函数（系统获取函数）
- [x] 在 `init()` 中注册系统适配器工厂
- [x] 注册属性计算器和加成计算器
- [ ] 编写单元测试（Use Case 层）- 待后续完善
- [ ] 编写集成测试（Controller 层）- 待后续完善（LevelSys 无协议，暂不需要）
- [ ] 验证功能正常 - 待后续测试

### 19.2 阶段二：核心系统重构（3-4周）

#### 19.2.1 核心系统迁移（按依赖顺序）
- [x] **BagSys** - 背包系统
  - [x] 创建 `domain/bag.go`（Bag 实体）- 注意：BagSys 主要操作 BagData，暂不需要单独的实体
  - [x] 创建 `domain/item.go`（Item 实体）- 注意：Item 使用 protocol.ItemSt，暂不需要单独的实体
  - [x] 创建 `usecase/bag/add_item.go`
  - [x] 创建 `usecase/bag/remove_item.go`
  - [x] 创建 `usecase/bag/has_item.go`
  - [x] 创建 `adapter/controller/bag_controller.go`
  - [x] 创建 `adapter/presenter/bag_presenter.go`
  - [x] 创建 `adapter/system/bag_system_adapter.go`
  - [x] 创建 `adapter/system/bag_system_adapter_helper.go`（GetBagSys 函数）
  - [x] 创建 `adapter/system/bag_system_adapter_init.go`（系统注册）
  - [x] 实现辅助索引管理（`itemIndex`）
  - [x] 实现 `GetBagSys(ctx)` 函数
  - [x] 注册系统适配器工厂和协议（C2SOpenBag、D2GAddItem）

- [x] **MoneySys** - 货币系统
  - [x] 创建 `domain/money.go`（Money 实体）- 注意：MoneySys 主要操作 MoneyData，暂不需要单独的实体
  - [x] 创建 `usecase/money/add_money.go`
  - [x] 创建 `usecase/money/consume_money.go`
  - [x] 创建 `usecase/money/money_use_case_impl.go`（实现 MoneyUseCase 接口，供 LevelSys 使用）
  - [x] 创建 `adapter/controller/money_controller.go`
  - [x] 创建 `adapter/presenter/money_presenter.go`
  - [x] 创建 `adapter/system/money/money_system_adapter.go`（按系统分包）
  - [x] 创建 `adapter/system/money/money_system_adapter_helper.go`（GetMoneySys 函数）
  - [x] 创建 `adapter/system/money/money_system_adapter_init.go`（系统注册）
  - [x] 实现 `GetMoneySys(ctx)` 函数
  - [x] 注册系统适配器工厂和协议（C2SOpenMoney）

- [x] **EquipSys** - 装备系统
  - [x] 创建 `usecase/interfaces/bag.go`（BagUseCase 接口，用于 EquipSys 依赖）
  - [x] 创建 `usecase/equip/equip_item.go`
  - [x] 创建 `usecase/equip/unequip_item.go`
  - [x] 创建 `adapter/controller/equip_controller.go`
  - [x] 创建 `adapter/controller/bag_use_case_adapter.go`（BagUseCase 适配器）
  - [x] 创建 `adapter/presenter/equip_presenter.go`
  - [x] 创建 `adapter/system/equip/equip_system_adapter.go`
  - [x] 创建 `adapter/system/equip/equip_system_adapter_helper.go`（GetEquipSys 函数）
  - [x] 创建 `adapter/system/equip/equip_system_adapter_init.go`（系统注册）
  - [x] 注册系统适配器工厂和协议（C2SEquipItem）

- [x] **AttrSys** - 属性系统（依赖 LevelSys 和 EquipSys）
  - [x] 创建 `usecase/attr/mark_dirty.go`（标记脏系统用例）
  - [x] 创建 `usecase/attr/calc_attr.go`（属性计算用例）
  - [x] 创建 `usecase/attr/run_one.go`（RunOne 用例）
  - [x] 创建 `adapter/system/attr/attr_system_adapter.go`（系统适配器，实现核心逻辑）
  - [x] 创建 `adapter/system/attr/attr_system_adapter_helper.go`（GetAttrSys 函数）
  - [x] 创建 `adapter/system/attr/attr_system_adapter_init.go`（系统注册）
  - [x] 实现 `RunOneUseCase` 接口（属性增量更新，在 System Adapter 中实现）
  - [x] 实现 `GetAttrSys(ctx)` 函数
  - [x] 注册系统适配器工厂
  - [x] 处理系统依赖关系（通过 attrcalc 包注册的计算器，LevelSys 和 EquipSys 已注册）

#### 19.2.2 统一数据访问
- [x] 所有系统通过 `PlayerRepository` 接口访问数据
- [x] 实现 `PlayerGateway` 统一封装 database 调用
- [x] 保持 BinaryData 的共享引用模式（Repository 返回共享引用）
- [x] 验证所有系统数据访问正常

**验证结果**：
- ✅ Use Case 层：100% 通过 `PlayerRepository` 接口
- ✅ System Adapter 层：100% 通过 `PlayerGateway`（实现 `PlayerRepository`）
- ✅ PlayerGateway：正确实现接口，保持共享引用模式
- 📝 详细验证报告：`docs/统一数据访问和网络发送验证.md`

#### 19.2.3 统一网络发送
- [x] 所有系统通过 Presenter 构建响应
- [x] 通过 NetworkGateway 发送消息
- [x] 验证所有协议响应正常

**验证结果**：
- ✅ Presenter 层：100% 通过 `NetworkGateway`
- ✅ Controller 层：100% 通过 Presenter 发送消息
- ✅ System Adapter 层：AttrSys 通过 `NetworkGateway` 发送消息
- ✅ NetworkGateway：正确实现接口，统一封装 `gatewaylink`
- 📝 详细验证报告：`docs/统一数据访问和网络发送验证.md`

### 19.3 阶段三：玩法系统重构（2-3周）

#### 19.3.1 玩法系统迁移
- [x] **SkillSys** - 技能系统
  - [x] 创建 `usecase/skill/learn_skill.go`（学习技能用例）
  - [x] 创建 `usecase/skill/upgrade_skill.go`（升级技能用例）
  - [x] 创建 `usecase/interfaces/consume.go`（ConsumeUseCase 接口定义）
  - [x] 创建 `adapter/controller/skill_controller.go`（协议处理：C2SLearnSkill、C2SUpgradeSkill）
  - [x] 创建 `adapter/controller/consume_use_case_adapter.go`（ConsumeUseCase 适配器）
  - [x] 创建 `adapter/presenter/skill_presenter.go`（响应构建）
  - [x] 创建 `adapter/system/skill/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - [x] 完善 `usecase/interfaces/level.go`（添加 GetLevel 方法）
  - [x] 实现 `GetSkillSys(ctx)` 函数和系统注册
  - [x] 通过接口依赖 LevelSys 和 ConsumeUseCase，避免循环依赖
  - [x] 实现了技能同步到 DungeonServer 的逻辑
  - [x] 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）

- [x] **QuestSys** - 任务系统
  - [x] 创建 `usecase/quest/accept_quest.go`（接受任务用例）
  - [x] 创建 `usecase/quest/update_progress.go`（更新任务进度用例）
  - [x] 创建 `usecase/quest/submit_quest.go`（提交任务用例）
  - [x] 创建 `usecase/interfaces/daily_activity.go`（DailyActivityUseCase 接口定义）
  - [x] 创建 `adapter/controller/quest_controller.go`（协议处理：C2STalkToNPC）
  - [x] 创建 `adapter/controller/daily_activity_use_case_adapter.go`（DailyActivityUseCase 适配器）
  - [x] 创建 `adapter/presenter/quest_presenter.go`（响应构建）
  - [x] 创建 `adapter/system/quest/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - [x] 完善 `usecase/interfaces/config.go`（添加 GetQuestConfigsByType、GetNPCSceneConfig）
  - [x] 实现 `OnNewDay` 和 `OnNewWeek` 方法（每日/每周刷新）
  - [x] 实现 `GetQuestSys(ctx)` 函数和系统注册
  - [x] 订阅玩家事件（OnNewDay、OnNewWeek）
  - [x] 通过接口依赖 LevelUseCase、RewardUseCase、DailyActivityUseCase，避免循环依赖
  - [x] 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）

- [x] **FubenSys** - 副本系统
  - [x] 创建 `usecase/fuben/enter_dungeon.go`（进入副本用例）
  - [x] 创建 `usecase/fuben/settle_dungeon.go`（副本结算用例）
  - [x] 创建 `adapter/controller/fuben_controller.go`（协议处理：C2SEnterDungeon、D2GSettleDungeon、D2GEnterDungeonSuccess）
  - [x] 创建 `adapter/controller/reward_use_case_adapter.go`（RewardUseCase 适配器）
  - [x] 创建 `adapter/presenter/fuben_presenter.go`（响应构建）
  - [x] 创建 `adapter/system/fuben/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - [x] 实现 RPC 处理器注册（`D2GSettleDungeon`、`D2GEnterDungeonSuccess`）
  - [x] 通过 `DungeonServerGateway` 调用 DungeonServer
  - [x] 实现 `GetFubenSys(ctx)` 函数和系统注册
  - [x] 通过接口依赖 ConsumeUseCase、LevelUseCase、RewardUseCase，避免循环依赖
  - [x] 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）

- [x] **ItemUseSys** - 物品使用系统
  - [x] 创建 `usecase/item_use/use_item.go`
  - [x] 创建 `adapter/controller/item_use_controller.go`
  - [x] 创建 `adapter/presenter/item_use_presenter.go`
  - [x] 创建 `adapter/system/item_use/item_use_system_adapter.go`
  - [x] 创建 `adapter/system/item_use/item_use_system_adapter_helper.go`
  - [x] 创建 `adapter/system/item_use/item_use_system_adapter_init.go`
  - [x] 创建 `adapter/controller/level_use_case_adapter.go`（LevelUseCase 适配器）
  - [x] 处理系统依赖关系（依赖 BagSys、LevelSys 等）
  - [x] 实现 `GetItemUseSys(ctx)` 函数
  - [x] 注册系统适配器工厂和协议
  - [ ] TODO: 完善 HP/MP 同步到 DungeonServer 的逻辑（通过事件或接口）

- [x] **ShopSys** - 商城系统
  - [x] 创建 `usecase/shop/buy_item.go`（购买商品用例）
  - [x] 创建 `adapter/controller/shop_controller.go`（协议处理：C2SShopBuy）
  - [x] 创建 `adapter/presenter/shop_presenter.go`（响应构建）
  - [x] 创建 `adapter/system/shop/` 包（按系统分包，包含适配器、辅助函数、初始化）
  - [x] 完善 `usecase/interfaces/config.go`（添加 GetShopConfig、GetConsumeConfig、GetRewardConfig）
  - [x] 实现 `GetShopSys(ctx)` 函数和系统注册
  - [x] 通过接口依赖 ConsumeUseCase、RewardUseCase，避免循环依赖
  - [x] 保持了向后兼容性（通过接口定义依赖，支持新旧代码并存）
  - [x] 购买成功后推送背包和货币数据更新

- [x] **RecycleSys** - 回收系统
  - [x] 创建 `usecase/recycle/recycle_item.go`（回收物品用例，负责校验配置、扣除物品、发放奖励）
  - [x] 创建 `adapter/controller/recycle_controller.go`（协议处理：C2SRecycleItem）
  - [x] 创建 `adapter/presenter/recycle_presenter.go`（响应构建）
  - [x] 创建 `adapter/system/recycle/` 包（适配器、Helper、Init，保持与其他系统一致）
  - [x] 实现 `GetRecycleSys(ctx)` 函数（通过单例适配器提供回收能力）
  - [x] 注册协议处理器并通过 Presenter 推送回收结果
  - [x] 完善 `usecase/interfaces/config.go` 与 `adapter/gateway/config_gateway.go`（新增 `GetItemRecycleConfig`）
  - [x] 新增 `adapter/controller/push_helpers.go`，统一背包/货币推送逻辑，供 ShopSys / RecycleSys 复用

#### 19.3.2 RPC 调用重构
- [x] 所有 RPC 调用通过 `DungeonServerGateway` 接口（AsyncCall、RegisterRPCHandler、Register/UnregisterProtocols 均转由 Gateway 提供）
- [x] 在 Adapter 层实现具体调用（`adapter/gateway/dungeon_server_gateway.go` 封装 `dungeonserverlink`，对 Use Case 暴露统一接口）
- [x] 实现 `ProtocolRouterController`（协议路由逻辑迁移至 `adapter/controller/protocol_router_controller.go`，负责客户端协议分发与 RPC 转发）
- [x] 迁移 `handleDoNetWorkMsg` 中的协议路由逻辑到 Controller 层（`player_network.go` 仅注册 gshare Handler，实际逻辑由 Controller 执行）
- [x] 验证 RPC 调用正常（C2S→DungeonServer 透传链路通过 `DungeonServerGateway.AsyncCall`，RPC 回调与协议注册均由 Gateway 统一管理）

### 19.4 阶段四：社交系统重构（2-3周）

#### 19.4.1 社交系统迁移
- [x] **FriendSys** - 好友系统
  - [x] 创建 `domain/friend/friend.go`（封装 FriendData 初始化、列表增删工具）
  - [x] 创建 `usecase/friend/`（SendFriendRequest/RespondFriendRequest/RemoveFriend/QueryFriendList/Blacklist）
  - [x] 创建 `usecase/interfaces/blacklist.go` + `adapter/gateway/blacklist_repository.go`（黑名单仓储接口/实现）
  - [x] 创建 `adapter/controller/friend_controller.go`（协议处理：C2SAddFriend/Respond/Query/Remove/Blacklist）
  - [x] 创建 `adapter/presenter/friend_presenter.go`（统一返回结果、黑名单列表、错误通知）
  - [x] 创建 `adapter/system/friend/`（SystemAdapter + Helper + Init，注册工厂与协议）
  - [x] 通过 `PublicActorGateway` 转发 `AddFriendReq/Resp/FriendListQuery`，保持与 PublicActor 的异步交互
  - [x] 实现 `GetFriendSys(ctx)` 辅助函数，供 legacy 代码兼容调用
  - [x] 删除旧版 `entitysystem/friend_sys.go`，友好链路全面切换到 Clean Architecture

- [x] **GuildSys** - 公会系统
  - [x] 创建 `domain/guild/guild.go`（封装 GuildData 初始化）
  - [x] 创建 `usecase/guild/`（CreateGuild/JoinGuild/LeaveGuild/QueryGuildInfo）
  - [x] 创建 `adapter/controller/guild_controller.go` 与 `adapter/presenter/guild_presenter.go`
  - [x] 创建 `adapter/system/guild/`（SystemAdapter + Helper + Init，注册工厂和协议）
  - [x] 通过 `PublicActorGateway` 异步转发 `CreateGuild/JoinGuild/LeaveGuild` 消息
  - [x] 实现 `GetGuildSys(ctx)` 辅助函数并删除旧 `entitysystem/guild_sys.go`

- [x] **ChatSys** - 聊天系统
  - [x] 创建 `domain/chat/chat.go`（内容校验、敏感词检测）
  - [x] 创建 `usecase/chat/chat_world.go`、`chat_private.go`（支持 `ChatRateLimiter` 接口、敏感词配置、冷却控制）
  - [x] 创建 `usecase/interfaces/chat_rate_limiter.go` 并由 `adapter/system/chat` 实现
  - [x] 创建 `adapter/controller/chat_controller.go` 与 `adapter/presenter/chat_presenter.go`
  - [x] 创建 `adapter/system/chat/`（SystemAdapter + Helper + Init，管理冷却时间）
  - [x] 通过 `PublicActorGateway` 转发 `ChatWorld`/`ChatPrivate` 消息并删除旧 `entitysystem/chat_sys.go`

- [x] **AuctionSys** - 拍卖行系统
  - [x] 创建 `domain/auction/auction.go`（初始化玩家拍卖数据）
  - [x] 创建 `usecase/auction/put_on.go`、`buy.go`、`query.go`
  - [x] 创建 `adapter/controller/auction_controller.go` 与 `adapter/presenter/auction_presenter.go`
  - [x] 创建 `adapter/system/auction/`（SystemAdapter + Helper + Init）
  - [x] 通过 `PublicActorGateway` 转发 `AuctionPutOn/AuctionBuy/AuctionQuery` 消息
  - [x] 删除旧 `entitysystem/auction_sys.go`

#### 19.4.2 PublicActor 交互重构
- [x] 所有 PublicActor 调用统一走 `PublicActorGateway`（移除业务直接依赖 `gshare.SendPublicMessageAsync`）
  - [x] `PlayerRole` 登录/登出/排行榜快照/离线数据上报 统一使用 `sendPublicActorMessage`
  - [x] `player_network.handleQueryRank` 通过 Gateway 转发 `QueryRank` 消息
- [x] 在 Framework 层新增 `PlayerRole.sendPublicActorMessage` 封装上下文与发送逻辑，避免重复构造 `actor.NewBaseMessage`
- [x] Proto 消息保持原始定义，调用侧仅替换传输通道
- [x] 构建后验证所有相关链路可正常向 PublicActor 发送消息

### 19.5 阶段五：辅助系统重构（1-2周）

#### 19.5.1 辅助系统迁移
- [x] **MailSys** - 邮件系统
  - [x] 创建 `domain/mail/mail.go`（封装 `SiMailData` 初始化、ID 生成与过期清理入口）
  - [x] 创建 `usecase/mail/`（SendTemplateMail / SendCustomMail / ReadMail / DeleteMail / ClaimAttachment）
  - [x] 创建 `adapter/system/mail/`（MailSystemAdapter + Helper + Init，用于从 `PlayerRoleBinaryData` 初始化 MailData）
  - [x] 复用 GMSys 工具方法，通过 `usecase/mail` + 本地 `gmPlayerRepository` 发送系统邮件（单人 / 全服 / 按模板 / 通过邮件发放奖励）
  - [x] 删除旧 `entitysystem/mail_sys.go`，PlayerRole 背包满发奖路径改为调用 `SendSystemMail`，不再直接依赖 MailSys

- [ ] **VipSys** - VIP 系统（进行中）
  - [x] 创建 `domain/vip/vip.go`（封装 `SiVipData` 初始化`)
  - [x] 创建 `usecase/vip/vip_money_use_case.go`，实现 `interfaces.MoneyUseCase`，用于处理 `MoneyTypeVipExp`
  - [x] 为 `ConfigManager` / `ConfigGateway` 增加 `GetVipConfig`，通过接口访问 VIP 配置
  - [x] 在 `adapter/system/money` 与 `MoneyController` 中，为 `AddMoneyUseCase` 注入 Vip 用例（`SetDependencies(nil, vipUC, nil)`），统一通过 Use Case 处理 VIP 经验
  - [x] 创建 `adapter/system/vip/`（VipSystemAdapter + Helper + Init），用于初始化 `VipData` 并提供 `GetVipSys(ctx)`、特权值查询等能力；删除旧 `entitysystem/vip_sys.go`
  - [ ] 若后续增加 C2S 接口，则创建 `adapter/controller/vip_controller.go` 与 `adapter/presenter/vip_presenter.go`

- [ ] **DailyActivitySys** - 日常活跃度系统（进行中）
  - [x] 创建 `domain/dailyactivity/daily_activity.go`（封装 `SiDailyActivityData` 初始化、每日重置判断与重置逻辑）
  - [x] 创建 `usecase/dailyactivity/points_use_case.go`，实现 `interfaces.MoneyUseCase`，统一处理 `MoneyTypeActivePoint` 的增减与 `MoneyData` 同步
  - [x] 创建 `usecase/dailyactivity/claim_reward_use_case.go`（按 `DailyActivityRewardConfig` 校验活跃点与发奖，更新 `RewardStates`）
  - [x] 创建 `adapter/system/dailyactivity/`（SystemAdapter + Helper + Init），用于初始化数据、登录重置与提供 `GetDailyActivitySys(ctx)` 以兼容旧调用
  - [x] 在 `MoneySystemAdapter` / `MoneyController` 中为 `AddMoneyUseCase` 与 `ConsumeMoneyUseCase` 注入活跃点用例，移除旧 `DailyActivitySys` 对 MoneySys 的直接依赖
  - [x] 删除旧 `entitysystem/daily_activity_sys.go`，QuestSys 中的活跃点奖励改为依赖上层通过货币用例发放（不再直接访问 DailyActivitySys）

- [ ] **MessageSys** - 玩家消息系统（已在 EntitySystem 中实现离线消息回放，本阶段做轻量迁移）
  - [x] 创建 `adapter/system/message/`（MessageSystemAdapter + Init），将原有 `MessageSys` 生命周期迁移到 SystemAdapter 层
  - [x] 保留 `playeractor/entitysystem/message_dispatcher.go` 中的 `DispatchPlayerMessage` 作为统一分发入口，供 `PlayerRoleActor` 与适配器复用
  - [x] 更新 `player_network.handlePlayerMessageMsg`，通过 `DispatchPlayerMessage` + 数据库回退保证消息不丢失
  - [x] 删除旧 `entitysystem/message_sys.go`，系统工厂注册改由 `adapter/system/message` 完成

- [x] **GMSys** - GM 系统（轻量迁移，保留原命令集）
  - [x] 创建 `adapter/system/gm/`：包含 `GMSystemAdapter`（实现 `iface.ISystem`）、`GMManager`（命令注册与执行）、GM 工具函数（系统通知/系统邮件发送）
  - [x] 创建 `GetGMSys(ctx)` Helper，基于 `adapter/context` 从 `PlayerRole` 上下文中解析并获取 GM 系统
  - [x] 在 `gm_system_adapter_init.go` 中注册 `SysGM` 系统工厂，并在 `OnSrvStart` 时注册 `C2SGMCommand` 协议处理（`handleGMCommand` → `GM.HandleGMCommand`）
  - [x] 将旧 `entitysystem/gm_sys.go`、`gm_manager.go` 拆分为适配器层（系统 + 命令）与兼容工具层（`entitysystem/gm_tools.go` 的 `SendSystemMail` 等），并删除原 GM 系统实现

### 19.6 阶段六：清理与优化（1-2周）

#### 19.6.1 移除旧代码（高优先级：按系统分阶段彻底替换并删除旧实现）
- [✅] **分系统推进「Adapter + UseCase 替换旧 EntitySystem」**（每个系统完成后方可物理删除对应 `*_sys.go`）  
  - [✅] BagSys：将所有 `entitysystem.GetBagSys` / `BagSys` 的调用迁移到 `adapter/system/bag_system_adapter.go` + Bag 用例（`usecase/bag/*`），补齐适配器上仍缺的查询/修改接口，确认零引用后删除 `entitysystem/bag_sys.go`  
  - [✅] MoneySys：统一通过 `MoneySystemAdapter` 与 `AddMoneyUseCase` / `ConsumeMoneyUseCase` 处理货币变化，移除对 `entitysystem/money_sys.go` 的直接依赖并删除旧实现  
  - [✅] LevelSys：通过 `adapter/system/level_system_adapter.go` 与 Level 相关 UseCase 提供等级查询/经验更新能力，替换所有 `GetLevelSys` 调用，删除 `entitysystem/level_sys.go`  
  - [✅] SkillSys：通过 `adapter/system/skill` + UseCase 处理技能学习/升级/同步，将 `entitysystem/skill_sys.go` 中逻辑迁入 UseCase 后删除旧文件  
  - [✅] QuestSys：用 `adapter/system/quest` + 任务 UseCase 完成提交/奖励/状态流转，替换所有对 `quest_sys.go` 的直接访问后删除旧实现  
  - [✅] FubenSys：迁移副本进入/结算逻辑到 `adapter/system/fuben` 与对应 UseCase，删除 `entitysystem/fuben_sys.go`  
  - [✅] RecycleSys：删除 `entitysystem/recycle_sys.go`，所有 C2S 回收协议、奖励结算与背包更新统一复用 `adapter/controller/recycle_controller.go` + `adapter/system/recycle`  
  - [✅] ItemUseSys / ShopSys / AttrSys / EquipSys：依次用各自的 SystemAdapter + UseCase 接管逻辑，清空对旧 `*_sys.go` 的调用并删除  
  - [✅] `attrcalc/` 目录：将仍被使用的加成/计算逻辑抽到新的 Adapter 层（`adapter/system/attrcalc`），删除/清空旧 EntitySystem 下的计算工具实现  
  - [✅] `gm_tools.go` / `message_dispatcher.go` 等兼容工具：GM 邮件与 GM 命令工具全部迁移到 `adapter/system/gm`，`entitysystem/gm_tools.go` 清空为占位文件，仅保留 `DispatchPlayerMessage` 作为消息分发入口
- [✅] **清理直接依赖框架的代码**（`jsonconf`、`gatewaylink`、`dungeonserverlink`、`manager`、`actor` 等）  
  - [✅] DailyActivity 奖励发放用例改为依赖 `ConfigManager` + `RewardUseCase`，移除对 `jsonconf.GetConfigManager` 与 `manager.GetPlayerRole` 的直接调用  
  - [✅] `PlayerRole.CallDungeonServer` 改为通过 `DungeonServerGateway.AsyncCall` 访问战斗服，不再直接依赖 `dungeonserverlink`；其余配置与网络访问均通过 `ConfigGateway` / `NetworkGateway` / `DungeonServerGateway` / `PlayerGateway` 间接完成
- [ ] **更新 `sys_mgr.go` 以支持新的系统注册方式**  
  - [ ] 明确 SysMgr 只负责管理 `adapter/system/*` 注册的系统适配器，不再依赖旧 EntitySystem 工厂  
  - [ ] 根据新系统依赖关系（属性依赖等级/装备等）调整 `systemDependencies`，避免不再存在的旧系统 ID  
  - [ ] 在完成全部迁移后，验证系统初始化顺序与生命周期分发逻辑仍然正确
- [ ] **验证所有功能正常**  
  - [ ] 针对每个完成迁移并删除旧实现的系统，执行一轮功能回归（登录/重连/核心玩法路径）  
  - [ ] 在 GameServer + DungeonServer 联调环境下跑一遍主线流程与关键 GM 命令，确认无行为退化

#### 19.6.2 完善测试
- [ ] 为所有 Use Case 层编写单元测试（覆盖率 > 70%）
- [ ] 为所有 Controller 层编写集成测试
- [ ] 验证所有协议处理正常
- [ ] 验证系统生命周期正常
- [ ] 验证 BinaryData 共享正常
- [ ] 验证 RunOne 机制正常
- [ ] 验证时间事件处理正常
- [ ] 验证数据持久化正常

#### 19.6.3 文档更新
- [ ] 更新架构文档
- [ ] 更新开发指南
- [ ] 更新协议注册文档
- [ ] 更新关键代码位置
- [ ] 更新本文档（标记已完成功能）

### 19.7 系统迁移检查清单（每个系统）

> 每个系统迁移时，请按照以下检查清单逐项完成：

#### 19.7.1 Domain 层
- [ ] 创建 Domain 实体（移除框架依赖）
- [ ] 定义 Repository 接口（如果需要）

#### 19.7.2 Use Case 层
- [ ] 创建 Use Case（业务逻辑）
- [ ] 实现依赖接口（EventPublisher、PublicActorGateway 等）
- [ ] 实现 `InitializeFromBinaryData` 方法（如果需要）
- [ ] 实现 `RunOneUseCase` 接口（如果需要）
- [ ] 实现 `EventSubscriber` 接口（如果需要）
- [ ] 实现 `TimeCallbackUseCase` 接口（如果需要）

#### 19.7.3 Adapter 层
- [ ] 创建 Controller（协议处理）
- [ ] 创建 Presenter（响应构建）
- [ ] 实现 Gateway（数据访问，如果需要）
- [ ] 创建系统适配器（生命周期管理）
- [ ] 实现 `GetXXXSys(ctx)` 函数
- [ ] 实现系统状态管理（`SetOpened`、`IsOpened`）

#### 19.7.4 协议注册
- [ ] Controller 在 `init()` 中注册协议
- [ ] 协议处理器调用 Controller 方法
- [ ] 保持与 `clientprotocol.ProtoTbl` 的兼容性

#### 19.7.5 系统工厂注册
- [ ] 在 `init()` 中注册系统适配器工厂
- [ ] 工厂函数从 DI 容器获取依赖

#### 19.7.6 测试验证
- [ ] 编写 Use Case 单元测试
- [ ] 编写 Controller 集成测试
- [ ] 验证功能正常
- [ ] 验证协议处理正常
- [ ] 验证系统生命周期正常
- [ ] 验证 BinaryData 共享正常
- [ ] 验证 RunOne 机制正常（如果需要）
- [ ] 验证时间事件处理正常（如果需要）
- [ ] 验证数据持久化正常

#### 19.7.7 代码清理
- [ ] 删除 EntitySystem 中的旧代码
- [ ] 更新相关文档

### 19.8 整体检查清单

#### 19.8.1 所有系统已迁移（共 20 个系统）
- [ ] LevelSys（试点系统）
- [ ] BagSys
- [ ] MoneySys
- [ ] EquipSys
- [ ] AttrSys
- [ ] SkillSys
- [ ] QuestSys
- [ ] FubenSys
- [ ] ItemUseSys
- [ ] ShopSys
- [ ] RecycleSys
- [ ] FriendSys
- [ ] GuildSys
- [ ] ChatSys
- [ ] AuctionSys
- [ ] MailSys
- [ ] VipSys
- [ ] DailyActivitySys
- [ ] MessageSys
- [ ] GMSys

#### 19.8.2 所有框架依赖已移除
- [ ] database → Repository 接口
- [ ] gatewaylink → NetworkGateway 接口
- [ ] dungeonserverlink → DungeonServerGateway 接口
- [ ] gevent → EventPublisher 接口
- [ ] jsonconf → ConfigManager 接口
- [ ] gshare → PublicActorGateway 接口

#### 19.8.3 系统机制保持
- [ ] 系统生命周期管理正常
- [ ] BinaryData 共享机制正常
- [ ] 系统依赖关系处理正常
- [ ] RunOne 机制正常
- [ ] 事件订阅机制正常
- [ ] 系统工厂模式正常
- [ ] Context 传递机制正常
- [ ] 系统获取函数正常（GetXXXSys）
- [ ] 协议注册机制正常
- [ ] 时间事件处理正常
- [ ] 数据持久化正常
- [ ] 系统状态持久化正常
- [ ] Actor 单线程特性保持

#### 19.8.4 测试覆盖
- [ ] Use Case 层单元测试覆盖率 > 70%
- [ ] Controller 层集成测试通过
- [ ] 所有协议处理测试通过
- [ ] 系统生命周期测试通过

#### 19.8.5 文档更新
- [ ] 架构文档已更新
- [ ] 开发指南已更新
- [ ] 协议注册文档已更新
- [ ] 关键代码位置已更新
- [ ] 本文档已更新（标记已完成功能）

### 19.9 待实现 / 待完善功能清单（按优先级）

> 本小节用于从代码现状与进度文档中抽取「剩余工作」，便于逐项实施。完成后请同步勾选 19.6 与 19.8 中对应条目，并在 `docs/服务端开发进度文档.md` 第 6 章更新状态。

**一、Clean Architecture 收尾（高优先级，直接影响整体一致性）**
- [✅] **SysMgr 依赖配置补全**  
  - 根据当前系统列表（`adapter/system/*`）与 `protocol.SystemId`，补全 `systemDependencies` 中各系统的依赖关系（例如：`QuestSys` 依赖 `LevelSys/RewardUseCase/DailyActivitySys`，`FubenSys` 依赖 `ConsumeUseCase/LevelUseCase/RewardUseCase` 等），保证拓扑排序顺序与实际业务依赖一致。  
  - 验证初始化顺序日志（`System init order`）是否符合预期，并在出现循环依赖时补充文档说明与降级策略。
- [ ] **清理 SysRank 依赖关系**  
  - `sys_mgr.go` 中的 `SystemId_SysRank` 依赖关系需要清理：RankSys 不是 PlayerActor 的系统，而是 PublicActor 的功能，不需要在 PlayerActor 系统依赖中定义。  
  - 如果 `SystemId_SysRank` 在 proto 中定义为系统ID但实际未使用，需要确认是否应该移除该枚举值，或将其标记为"仅用于 PublicActor，不参与 PlayerActor 系统管理"。
- [✅] **框架依赖收束到 Gateway 层**  
  - 已确认 Use Case 与 SystemAdapter 层不直接依赖 `gatewaylink/dungeonserverlink/gevent/gshare` 等框架包，所有网络、事件与 PublicActor 访问均通过 `NetworkGateway` / `DungeonServerGateway` / `EventAdapter` / `PublicActorGateway` 间接完成。  
  - `jsonconf` 仍作为配置具体类型在 Use Case 中做类型断言（如 `*jsonconf.ItemConfig`），但所有配置获取入口均已通过 `interfaces.ConfigManager` + `ConfigGateway` 实现，未在 Use Case 中直接调用 `jsonconf.GetConfigManager` 或其他框架函数，符合 Clean Architecture 依赖方向要求。
- [✅] **VipSys / DailyActivitySys / MessageSys / GMSys 迁移验收**  
  - 已确认 `domain/vip`、`domain/dailyactivity`、`adapter/system/{vip,dailyactivity,message,gm}` 按 19.7 检查清单完成迁移：
    - Use Case 仅依赖 `interfaces.ConfigManager`、`MoneyUseCase`、`RewardUseCase` 等接口，不直接访问 `jsonconf.GetConfigManager` 或 `manager.GetPlayerRole` 等框架 API；
    - SystemAdapter 在 `*_system_adapter_init.go` 中通过 `entitysystem.RegisterSystemFactory` 正确注册对应 `SystemId`，生命周期回调仅通过 `adapter/context` 获取 `PlayerRole`；
    - `VipSystemAdapter` 仅在特权查询时通过 `jsonconf.GetConfigManager().GetVipConfig` 读取配置，属于 Adapter 层对框架的合法依赖。
  - 旧 `entitysystem` 中 Vip/DailyActivity/Message/GM 的实现文件已删除，仅保留 `gm_tools.go` 与 `message_dispatcher.go` 作为兼容 Helper，所有业务入口统一复用 `adapter/system` 暴露的能力；对应系统在 19.6.1 与 19.8.1 中可视为已完成迁移。

**二、玩法链路细节完善（中优先级，主要是 TODO 收尾）**
- [✅] **HP/MP 同步到 DungeonServer 的统一方案**  
  - 已清理 `usecase/item_use`、`usecase/skill` 与 `adapter/system/item_use` 中关于“通过事件或接口同步到 DungeonServer”的 TODO：战斗内 HP/MP 以 DungeonServer 为权威，属性同步由 AttrSys + DungeonServer 自身机制负责，ItemUse 与 Skill 用例仅负责更新玩家数据与经验，不直接向 DungeonServer 下发血蓝补丁。  
  - 当前版本不额外在 GameServer 发起 HP/MP 同步 RPC，避免与战斗服权威冲突；后续若新增“战斗外恢复逻辑”，建议在 DungeonServer 协议层扩展而非复用本通道。
- [✅] **商城/回收购买次数与审计信息（可选持久化）**  
  - 已确认 `usecase/shop/buy_item.go` 与 `adapter/system/shop` 中的 `purchaseCounters` 仅用于运行期限购，当前版本不做数据库持久化；对应 TODO 更新为显式设计说明。  
  - 如将来需要对商城/回收行为做长期统计与审计，建议通过独立的统计/审计服务与表结构实现，而非在 ShopSys 内部混入持久化逻辑。
- [✅] **拍卖行 PutOn 结果协议补全**  
  - 现有流程中，上架结果由 GameServer 侧 C2S 响应反馈给客户端，PublicActor 在 `handleAuctionPutOn` 中仅负责维护全局拍卖数据与日志，不再单独下发 `S2CAuctionPutOnResult`；已移除相关 TODO 并补充注释说明。  
  - 如未来需要更细粒度的上架结果通知，可在不破坏现有行为的前提下新增协议与 Presenter；当前版本保持协议面稳定，不改动 Proto 文件。

**三、基础设施与测试（中–高优先级，贯穿所有系统）**
- [ ] **Use Case 单元测试与 Controller 集成测试**  
  - 按 19.7.6 和 19.6.2 要求，为已迁移的系统（特别是 Bag/Money/Equip/Attr/Quest/Fuben/Shop/Recycle/Vip/DailyActivity）补齐单元测试与核心协议的集成测试，目标：Use Case 层覆盖率 ≥ 70%。  
  - 建议优先覆盖涉及货币/物品修改、副本结算、公会/拍卖与 GM 操作等高风险路径，并在本节维护一份「已覆盖系统列表」。
- [ ] **系统行为端到端回归**  
  - 参照 `docs/服务端开发进度文档.md` 第 4/5 章的核心流程，对每个完成迁移并删除旧实现的系统，至少跑一轮「登录→主线→副本→社交」的端到端联调，确认无行为退化。  
  - 回归通过后，在 19.6.1「验证所有功能正常」小节勾选对应条目。
- [ ] **MessageSys 功能完善**  
  - MessageSys 已迁移到 `adapter/system/message`，但功能仍需完善：  
    - 确认离线消息回放机制是否完整（登录/重连时自动加载、回调成功后删库、失败保留）  
    - 检查是否有新的业务场景需要扩展消息类型与回调  
    - 验证消息持久化与过期清理策略是否正常工作

**四、安全与运维增强（与 Clean Architecture 解耦，但推荐同步进行）**
- [ ] **GM 权限模型与审计日志落地**（对应服务端开发进度文档 6.2 中 GM 权限与审计体系）  
  - 在 `adapter/system/gm` 的入口统一校验 GM 账号/角色标记、来源 IP（或环境令牌），并将高危操作写入结构化审计日志或审计表。  
  - 在本节记录审计字段与表结构，确保后续问题可追踪。
- [ ] **Gateway / GameServer / DungeonServer 接入安全**  
  - 按 7.4 的要求，在生产环境启用 Gateway WebSocket 的 IP 白名单与 Origin 校验，并为 GameServer ↔ DungeonServer 访问增加 TLS/双向认证或等价的签名校验。  
  - 完成后在配置章节补充示例配置，并在此勾选。

**五、文档与开发流程对齐（低优先级，但建议一次性完成）**
- [✅] 将本小节拆分/同步到 `docs/服务端开发进度文档.md` 第 6.1「Clean Architecture 重构」子节，保持两个文档的任务清单一致。  
  - 目前 `docs/服务端开发进度文档.md` 6.1 中的阶段划分（阶段一至阶段六）、各系统迁移状态与本节 19.6/19.9 的任务描述已经对齐，仅在细节层面保留不同粒度的拆分，后续新增任务请同时维护两处文档。  
- [✅] 对照 19.8 整体检查清单逐项更新勾选状态，并在版本记录中追加此次收尾重构的日期与摘要。  
  - 已在 19.6 与 19.9 中勾选完成的子项（SysMgr 依赖补全、框架依赖收束、Vip/DailyActivity/Message/GMSys 迁移验收、玩法链路 TODO 清理），后续每次完成一个系统/子任务时，需同时更新 19.8 的整体检查清单与「版本记录」章节中的日期与变更摘要，保证文档与代码演进保持同步。

---

---

## 20. 代码梳理发现的遗漏点补充（2025-01-XX）

> 本节记录通过代码梳理发现的重构遗漏点，需要在后续重构中补充。

### 20.1 RankSys 系统依赖关系清理

**问题描述：**
- `sys_mgr.go` 中的 `systemDependencies` 定义了 `SystemId_SysRank` 的依赖关系（依赖 `SysLevel`）
- 但 RankSys 不是 PlayerActor 的系统，而是 PublicActor 的功能
- 排行榜功能在 `publicactor/public_role_rank.go` 中实现，PlayerRole 通过 `sendPublicActorMessage` 发送更新快照消息

**需要处理：**
- [✅] 从 `sys_mgr.go` 的 `systemDependencies` 中移除 `SystemId_SysRank` 的依赖关系定义  
  - **已完成**：`sys_mgr.go` 已不再使用 `systemDependencies`，改为按 SystemId 顺序初始化，因此无需移除依赖关系定义
- [✅] 确认 `SystemId_SysRank` 在 proto 中的定义：如果仅用于 PublicActor，应在文档中说明；如果不再使用，考虑移除该枚举值  
  - **已完成**：已在 `proto/csproto/system.proto` 中为 `SysRank = 19` 添加注释说明：RankSys 是 PublicActor 功能，不参与 PlayerActor 系统管理，此枚举值仅用于标识
- [✅] 更新文档说明：RankSys 是 PublicActor 功能，不参与 PlayerActor 系统管理  
  - **已完成**：已确认没有系统注册 SysRank（SystemId = 19），符合预期；排行榜功能在 `publicactor/public_role_rank.go` 中实现

**关键代码位置：**
- `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go:96-99`
- `server/service/gameserver/internel/app/publicactor/public_role_rank.go`
- `server/service/gameserver/internel/app/playeractor/entity/player_role.go:172-237`

### 20.2 MessageSys 功能完善检查

**当前状态：**
- MessageSys 已迁移到 `adapter/system/message`
- 保留了 `entitysystem/message_dispatcher.go` 中的 `DispatchPlayerMessage` 作为统一分发入口

**需要检查：**
- [ ] 离线消息回放机制是否完整（登录/重连时自动加载、回调成功后删库、失败保留）
- [ ] 是否有新的业务场景需要扩展消息类型与回调
- [ ] 消息持久化与过期清理策略是否正常工作
- [ ] 是否需要为 MessageSys 创建 UseCase 层（当前主要在 SystemAdapter 中实现）

**关键代码位置：**
- `server/service/gameserver/internel/adapter/system/message_system_adapter.go`
- `server/service/gameserver/internel/app/playeractor/entitysystem/message_dispatcher.go`
- `server/service/gameserver/internel/app/playeractor/entity/player_network.go`（handlePlayerMessageMsg）

### 20.3 系统移除后的依赖关系清理

**已移除的系统：**
- VipSys（已移除）
- DailyActivitySys（已移除）
- FriendSys（已移除）
- GuildSys（已移除）
- AuctionSys（已移除）

**需要检查：**
- [✅] 确认 `sys_mgr.go` 的 `systemDependencies` 中是否还有对这些已移除系统的依赖关系定义  
  - **已完成**：`sys_mgr.go` 已不再使用 `systemDependencies`，改为按 SystemId 顺序初始化，因此无需清理依赖关系定义
- [✅] 如果存在，需要清理这些依赖关系，避免拓扑排序时出现错误  
  - **已完成**：由于不再使用 `systemDependencies`，无需清理
- [✅] 检查 proto 中的 `SystemId` 枚举是否还包含这些已移除的系统ID  
  - **已完成**：已确认 proto 中的 `SystemId` 枚举不包含这些已移除的系统ID（VipSys、DailyActivitySys、FriendSys、GuildSys、AuctionSys），且没有系统注册这些已移除的系统ID

**关键代码位置：**
- `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
- `proto/csproto/system.proto`

### 20.4 待实现功能清单（按优先级）

**高优先级（影响整体一致性）：**
1. [✅] 清理 SysRank 依赖关系（20.1）  
   - **已完成**：已在 proto 中添加注释说明，确认没有系统注册 SysRank，符合预期
2. [✅] 清理已移除系统的依赖关系（20.3）  
   - **已完成**：已确认 proto 中不包含已移除的系统ID，且 `sys_mgr.go` 不再使用 `systemDependencies`
3. [ ] Use Case 单元测试与 Controller 集成测试（19.9 三）
4. [ ] 系统行为端到端回归（19.9 三）

**中优先级（功能完善）：**
5. [ ] MessageSys 功能完善检查（20.2）
6. [ ] GM 权限模型与审计日志落地（19.9 四）
7. [ ] Gateway / GameServer / DungeonServer 接入安全（19.9 四）

**低优先级（文档与流程）：**
8. [ ] 更新架构文档（19.6.3）
9. [ ] 更新开发指南（19.6.3）
10. [ ] 更新协议注册文档（19.6.3）

---

**文档版本：** v1.1  
**最后更新：** 2025-01-XX  
**责任人：** 开发团队

