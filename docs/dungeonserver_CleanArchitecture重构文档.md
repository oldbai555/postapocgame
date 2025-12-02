# DungeonServer Clean Architecture 重构文档

更新时间：2025-01-XX  
责任人：开发团队

## 1. 文档目的

本文档旨在将 `server/service/dungeonserver` 按照 Clean Architecture（清洁架构）原则进行重构，实现业务逻辑与框架解耦，提高代码可测试性、可维护性和可扩展性。

## 2. 当前架构问题分析

### 2.1 依赖方向混乱

**问题描述：**
- EntitySystem（业务逻辑层）直接依赖 `gameserverlink`、`clientprotocol`、`network` 等框架层
- 业务逻辑与协议处理、网络发送、RPC 调用混在一起
- 内层（业务逻辑）依赖外层（框架），违反了依赖倒置原则

**典型示例：**

```go
// entitysystem/move_sys.go - 直接依赖网络和协议
func (ms *MoveSys) HandleStartMove(ctx context.Context, msg *network.ClientMessage) {
    // ❌ 直接解析协议
    var req protocol.C2SStartMoveReq
    proto.Unmarshal(msg.Data, &req)
    
    // ❌ 直接发送网络消息
    gameserverlink.BroadcastToScene(...)
}

// entitysystem/fight_sys.go - 直接依赖网络
func (s *FightSys) handleUseSkill(entity iface.IEntity, msg *network.ClientMessage) {
    // ❌ 直接处理协议和网络
    gameserverlink.BroadcastToScene(...)
}
```

### 2.2 业务逻辑与框架耦合

**问题描述：**
- EntitySystem 中混入了协议解析、网络发送、RPC 调用等框架代码
- 业务逻辑无法独立测试，必须启动完整的 Actor 框架和网络服务
- 系统之间通过直接调用而非接口交互

### 2.3 数据访问层缺失

**问题描述：**
- 没有数据访问接口抽象，EntitySystem 直接调用配置管理器
- 无法进行单元测试（无法 mock 配置）
- 数据访问逻辑分散在各个系统中

### 2.4 接口适配层不清晰

**问题描述：**
- 协议处理、响应构建、RPC 调用等适配逻辑混在业务逻辑中
- 没有明确的 Controller/Presenter 层
- 协议变更会影响业务逻辑代码

## 3. Clean Architecture 分层设计

### 3.1 分层结构

```
┌─────────────────────────────────────────────────────────┐
│  Frameworks & Drivers (框架层)                          │
│  - Actor 框架                                           │
│  - 网络层 (gameserverlink, network)                     │
│  - 配置管理器 (jsonconf)                                │
│  - 事件总线 (event)                                     │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Interface Adapters (接口适配层)                         │
│  - Controllers: 协议处理器                               │
│  - Presenters: 响应构建器                               │
│  - Gateways: 配置访问实现                                │
│  - RPC Adapters: RPC 调用适配器                         │
│  - Event Adapters: 事件适配器                           │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Use Cases (用例层)                                      │
│  - 业务用例: MoveEntity, CastSkill, ApplyBuff 等         │
│  - 业务规则: 移动校验、技能计算、伤害结算等              │
└─────────────────────────────────────────────────────────┘
                          ↑
┌─────────────────────────────────────────────────────────┐
│  Entities (实体层)                                       │
│  - 业务实体: Entity, Skill, Buff, Scene 等               │
│  - 值对象: Position, Damage, Effect 等                 │
└─────────────────────────────────────────────────────────┘
```

### 3.2 目录结构设计

```
server/service/dungeonserver/
├── internel/
│   ├── domain/                    # Entities 层
│   │   ├── entity.go              # 实体基类
│   │   ├── position.go            # 位置值对象
│   │   ├── skill.go               # 技能实体
│   │   ├── buff.go                # Buff 实体
│   │   ├── scene.go               # 场景实体
│   │   └── ...
│   │
│   ├── usecase/                   # Use Cases 层
│   │   ├── move/                  # 移动用例
│   │   │   ├── start_move.go
│   │   │   ├── update_move.go
│   │   │   └── end_move.go
│   │   ├── fight/                 # 战斗用例
│   │   │   ├── cast_skill.go
│   │   │   ├── calculate_damage.go
│   │   │   └── apply_effect.go
│   │   ├── buff/                  # Buff 用例
│   │   │   ├── add_buff.go
│   │   │   └── remove_buff.go
│   │   └── ...
│   │
│   ├── adapter/                   # Interface Adapters 层
│   │   ├── controller/           # 协议控制器
│   │   │   ├── move_controller.go
│   │   │   ├── fight_controller.go
│   │   │   └── ...
│   │   ├── presenter/            # 响应构建器
│   │   │   ├── move_presenter.go
│   │   │   └── ...
│   │   ├── gateway/              # 配置访问实现
│   │   │   ├── config_gateway.go
│   │   │   └── ...
│   │   ├── rpc/                  # RPC 调用适配器
│   │   │   ├── gameserver_rpc.go
│   │   │   └── ...
│   │   └── event/                # 事件适配器
│   │       └── event_adapter.go
│   │
│   ├── infrastructure/           # Frameworks & Drivers 层
│   │   ├── actor/                 # Actor 适配
│   │   ├── network/               # 网络适配
│   │   ├── config/                # 配置适配
│   │   └── event/                 # 事件适配
│   │
│   └── ... (保留现有目录用于过渡)
```

## 4. 重构方案

### 4.1 阶段一：Entities 层重构

**目标：** 提取纯业务实体，移除所有框架依赖

#### 4.1.1 创建 Domain 实体

**目录：** `internel/domain/`

**示例：Position 值对象**

```go
// domain/position.go
package domain

// Position 位置值对象（纯业务对象，不依赖任何框架）
type Position struct {
    X int32 // 格子坐标 X
    Y int32 // 格子坐标 Y
}

// ToPixel 转换为像素坐标
func (p Position) ToPixel() (int32, int32) {
    return p.X*128 + 64, p.Y*128 + 64
}

// Distance 计算距离（格子距离）
func (p Position) Distance(other Position) float64 {
    dx := float64(p.X - other.X)
    dy := float64(p.Y - other.Y)
    return math.Sqrt(dx*dx + dy*dy)
}

// IsValid 检查位置是否有效
func (p Position) IsValid() bool {
    return p.X >= 0 && p.Y >= 0
}
```

**示例：Entity 实体**

```go
// domain/entity.go
package domain

// Entity 实体基类（纯业务对象）
type Entity struct {
    ID       uint64
    Type     uint32
    Position Position
    HP       int64
    MP       int64
    MaxHP    int64
    MaxMP    int64
}

// IsAlive 判断是否存活
func (e *Entity) IsAlive() bool {
    return e.HP > 0
}

// TakeDamage 受到伤害（纯业务逻辑）
func (e *Entity) TakeDamage(damage int64) {
    e.HP -= damage
    if e.HP < 0 {
        e.HP = 0
    }
}

// Heal 治疗（纯业务逻辑）
func (e *Entity) Heal(amount int64) {
    e.HP += amount
    if e.HP > e.MaxHP {
        e.HP = e.MaxHP
    }
}
```

#### 4.1.2 定义 Repository 接口

**目录：** `internel/domain/repository/`

```go
// domain/repository/config_repository.go
package repository

import "postapocgame/server/service/dungeonserver/internel/domain"

// ConfigRepository 配置访问接口（定义在 domain 层）
type ConfigRepository interface {
    GetSkillConfig(skillID uint32) (*domain.SkillConfig, error)
    GetMonsterConfig(monsterID uint32) (*domain.MonsterConfig, error)
    GetMapConfig(mapID uint32) (*domain.MapConfig, error)
    // ... 其他配置访问
}
```

### 4.2 阶段二：Use Cases 层重构

**目标：** 实现业务用例，依赖 Entities 和 Repository 接口

#### 4.2.1 创建 Use Case

**目录：** `internel/usecase/`

**示例：StartMove Use Case**

```go
// usecase/move/start_move.go
package move

import (
    "context"
    "postapocgame/server/service/dungeonserver/internel/domain"
    "postapocgame/server/service/dungeonserver/internel/domain/repository"
)

// StartMoveUseCase 开始移动用例
type StartMoveUseCase struct {
    sceneRepo repository.SceneRepository
    // 可以注入其他依赖，如事件发布器接口
    eventPublisher EventPublisher
}

func NewStartMoveUseCase(
    sceneRepo repository.SceneRepository,
    eventPublisher EventPublisher,
) *StartMoveUseCase {
    return &StartMoveUseCase{
        sceneRepo:     sceneRepo,
        eventPublisher: eventPublisher,
    }
}

// Execute 执行开始移动用例
func (uc *StartMoveUseCase) Execute(ctx context.Context, entity *domain.Entity, dest Position) error {
    // 1. 验证目标位置
    scene, err := uc.sceneRepo.GetSceneByEntity(entity)
    if err != nil {
        return err
    }
    
    if !scene.IsWalkable(dest) {
        return ErrInvalidDestination
    }
    
    // 2. 执行业务逻辑（纯业务，不依赖框架）
    entity.StartMove(dest)
    
    // 3. 发布事件（通过接口）
    uc.eventPublisher.PublishMoveStarted(ctx, entity.ID, dest)
    
    return nil
}
```

**示例：CastSkill Use Case**

```go
// usecase/fight/cast_skill.go
package fight

import (
    "context"
    "postapocgame/server/service/dungeonserver/internel/domain"
    "postapocgame/server/service/dungeonserver/internel/domain/repository"
)

// CastSkillUseCase 释放技能用例
type CastSkillUseCase struct {
    configRepo repository.ConfigRepository
    eventPublisher EventPublisher
}

func NewCastSkillUseCase(
    configRepo repository.ConfigRepository,
    eventPublisher EventPublisher,
) *CastSkillUseCase {
    return &CastSkillUseCase{
        configRepo:    configRepo,
        eventPublisher: eventPublisher,
    }
}

// Execute 执行释放技能用例
func (uc *CastSkillUseCase) Execute(ctx context.Context, caster *domain.Entity, skillID uint32, targetPos Position) (*domain.SkillCastResult, error) {
    // 1. 获取技能配置
    skillConfig, err := uc.configRepo.GetSkillConfig(skillID)
    if err != nil {
        return nil, err
    }
    
    // 2. 验证技能释放条件（纯业务逻辑）
    if !caster.HasSkill(skillID) {
        return nil, ErrSkillNotLearned
    }
    
    if caster.MP < int64(skillConfig.ManaCost) {
        return nil, ErrNotEnoughMP
    }
    
    // 3. 计算技能效果（纯业务逻辑）
    result := caster.CastSkill(skillID, targetPos, skillConfig)
    
    // 4. 发布事件
    uc.eventPublisher.PublishSkillCast(ctx, caster.ID, skillID, result)
    
    return result, nil
}
```

#### 4.2.2 定义 Use Case 依赖接口

**目录：** `internel/usecase/interfaces/`

```go
// usecase/interfaces/event_publisher.go
package interfaces

// EventPublisher 事件发布器接口（Use Case 层定义）
type EventPublisher interface {
    PublishMoveStarted(ctx context.Context, entityID uint64, dest Position)
    PublishMoveEnded(ctx context.Context, entityID uint64, pos Position)
    PublishSkillCast(ctx context.Context, casterID uint64, skillID uint32, result interface{})
    PublishDamageDealt(ctx context.Context, attackerID, targetID uint64, damage int64)
    // ... 其他事件
}
```

### 4.3 阶段三：Interface Adapters 层重构

**目标：** 实现协议处理、数据访问、RPC 适配

#### 4.3.1 Controllers（协议控制器）

**目录：** `internel/adapter/controller/`

```go
// adapter/controller/move_controller.go
package controller

import (
    "context"
    "postapocgame/server/internal/network"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/dungeonserver/internel/adapter/presenter"
    "postapocgame/server/service/dungeonserver/internel/domain"
    "postapocgame/server/service/dungeonserver/internel/usecase/move"
)

// MoveController 移动协议控制器
type MoveController struct {
    startMoveUseCase *move.StartMoveUseCase
    updateMoveUseCase *move.UpdateMoveUseCase
    endMoveUseCase    *move.EndMoveUseCase
    presenter         *presenter.MovePresenter
}

func NewMoveController(
    startMoveUseCase *move.StartMoveUseCase,
    updateMoveUseCase *move.UpdateMoveUseCase,
    endMoveUseCase *move.EndMoveUseCase,
    presenter *presenter.MovePresenter,
) *MoveController {
    return &MoveController{
        startMoveUseCase:  startMoveUseCase,
        updateMoveUseCase: updateMoveUseCase,
        endMoveUseCase:    endMoveUseCase,
        presenter:         presenter,
    }
}

// HandleStartMove 处理开始移动协议
func (c *MoveController) HandleStartMove(ctx context.Context, entity *domain.Entity, msg *network.ClientMessage) error {
    // 1. 解析协议
    var req protocol.C2SStartMoveReq
    if err := proto.Unmarshal(msg.Data, &req); err != nil {
        return err
    }
    
    // 2. 转换为 Domain 对象
    destPos := domain.Position{
        X: argsdef.PixelCoordToTile(req.DestPx).X,
        Y: argsdef.PixelCoordToTile(req.DestPy).Y,
    }
    
    // 3. 调用 Use Case
    if err := c.startMoveUseCase.Execute(ctx, entity, destPos); err != nil {
        return c.presenter.PresentError(ctx, err)
    }
    
    // 4. 构建响应并广播
    return c.presenter.PresentMoveStarted(ctx, entity, destPos)
}
```

#### 4.3.2 Presenters（响应构建器）

**目录：** `internel/adapter/presenter/`

```go
// adapter/presenter/move_presenter.go
package presenter

import (
    "context"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/dungeonserver/internel/adapter/gateway/network"
    "postapocgame/server/service/dungeonserver/internel/domain"
)

// MovePresenter 移动响应构建器
type MovePresenter struct {
    networkGateway network.Gateway
    sceneGateway   SceneGateway
}

func NewMovePresenter(networkGateway network.Gateway, sceneGateway SceneGateway) *MovePresenter {
    return &MovePresenter{
        networkGateway: networkGateway,
        sceneGateway:   sceneGateway,
    }
}

// PresentMoveStarted 构建开始移动响应并广播
func (p *MovePresenter) PresentMoveStarted(ctx context.Context, entity *domain.Entity, dest domain.Position) error {
    // 1. 构建响应协议
    destPx, destPy := dest.ToPixel()
    resp := &protocol.S2CStartMove{
        EntityHdl: entity.ID,
        MoveData: &protocol.MoveData{
            DestPx: destPx,
            DestPy: destPy,
        },
    }
    
    // 2. 广播到场景
    scene := p.sceneGateway.GetSceneByEntity(entity)
    return p.networkGateway.BroadcastToScene(scene.ID, protocol.S2CProtocol_S2CStartMove, resp)
}
```

#### 4.3.3 Gateways（配置访问实现）

**目录：** `internel/adapter/gateway/`

```go
// adapter/gateway/config_gateway.go
package gateway

import (
    "postapocgame/server/internal/jsonconf"
    "postapocgame/server/service/dungeonserver/internel/domain"
    "postapocgame/server/service/dungeonserver/internel/domain/repository"
)

// ConfigGateway 配置访问实现（实现 domain 层的 Repository 接口）
type ConfigGateway struct {
    configMgr *jsonconf.ConfigManager
}

func NewConfigGateway() repository.ConfigRepository {
    return &ConfigGateway{
        configMgr: jsonconf.GetConfigManager(),
    }
}

func (g *ConfigGateway) GetSkillConfig(skillID uint32) (*domain.SkillConfig, error) {
    // 调用 jsonconf 获取配置
    cfg, ok := g.configMgr.GetSkillConfig(skillID)
    if !ok {
        return nil, ErrConfigNotFound
    }
    
    // 转换为 Domain 对象
    return convertToDomainSkillConfig(cfg), nil
}
```

#### 4.3.4 RPC Adapters（RPC 调用适配器）

**目录：** `internel/adapter/rpc/`

```go
// adapter/rpc/gameserver_rpc.go
package rpc

import (
    "context"
    "postapocgame/server/service/dungeonserver/internel/usecase/interfaces"
    "postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

// GameServerRPCAdapter GameServer RPC 适配器（实现 Use Case 层的接口）
type GameServerRPCAdapter struct{}

func NewGameServerRPCAdapter() interfaces.GameServerRPC {
    return &GameServerRPCAdapter{}
}

func (a *GameServerRPCAdapter) CallGameServer(ctx context.Context, sessionID string, msgID uint16, data []byte) error {
    // 调用 gameserverlink
    return gameserverlink.CallGameServer(ctx, sessionID, msgID, data)
}
```

### 4.4 阶段四：Infrastructure 层重构

**目标：** 封装框架调用，提供统一接口

#### 4.4.1 Network Gateway

**目录：** `internel/infrastructure/network/`

```go
// infrastructure/network/gateway.go
package network

import (
    "postapocgame/server/internal/protocol"
    "postapocgame/server/service/dungeonserver/internel/gameserverlink"
)

// Gateway 网络网关接口（Adapter 层定义）
type Gateway interface {
    BroadcastToScene(sceneID uint32, msgID uint16, data interface{}) error
    SendToEntity(entityID uint64, msgID uint16, data interface{}) error
}

// NetworkGateway 网络网关实现
type NetworkGateway struct{}

func NewNetworkGateway() Gateway {
    return &NetworkGateway{}
}

func (g *NetworkGateway) BroadcastToScene(sceneID uint32, msgID uint16, data interface{}) error {
    // 调用 gameserverlink 广播
    return gameserverlink.BroadcastToScene(sceneID, msgID, data)
}
```

## 5. 重构步骤

### 5.1 阶段一：基础结构搭建（1-2周）

1. **创建目录结构**
   - 创建 `domain/`、`usecase/`、`adapter/`、`infrastructure/` 目录
   - 定义基础接口（Repository、Gateway、EventPublisher 等）

2. **迁移一个简单系统作为示例**
   - 选择 `MoveSys` 作为第一个重构目标
   - 创建 `domain/position.go`（提取位置值对象）
   - 创建 `usecase/move/`（提取业务逻辑）
   - 创建 `adapter/controller/move_controller.go`（协议处理）
   - 创建 `adapter/gateway/config_gateway.go`（配置访问）

3. **验证重构效果**
   - 确保功能正常
   - 编写单元测试（Use Case 层可独立测试）

### 5.2 阶段二：核心系统重构（3-4周）

1. **重构核心系统**
   - `MoveSys` → `usecase/move/` + `adapter/controller/move_controller.go`
   - `FightSys` → `usecase/fight/` + `adapter/controller/fight_controller.go`
   - `AttrSys` → `usecase/attr/` + `adapter/controller/attr_controller.go`
   - `BuffSys` → `usecase/buff/` + `adapter/controller/buff_controller.go`

2. **统一配置访问**
   - 所有系统通过 ConfigRepository 接口访问配置
   - 实现 ConfigGateway 统一封装 jsonconf 调用

3. **统一网络发送**
   - 所有系统通过 Presenter 构建响应
   - 通过 Network Gateway 发送消息

### 5.3 阶段三：AI 和场景系统重构（2-3周）

1. **重构 AI 系统**
   - `AISys` → `usecase/ai/` + `adapter/controller/ai_controller.go`
   - AI 逻辑通过组合调用移动用例实现

2. **重构场景系统**
   - `SceneSt` → `domain/scene.go` + `usecase/scene/`
   - 场景管理逻辑提取到 Use Case 层

### 5.4 阶段四：清理与优化（1-2周）

1. **移除旧代码**
   - 删除 `entitysystem/` 中的旧实现
   - 清理直接依赖框架的代码

2. **完善测试**
   - 为 Use Case 层编写单元测试
   - 为 Controller 层编写集成测试

3. **文档更新**
   - 更新架构文档
   - 更新开发指南

## 6. 依赖注入设计

### 6.1 依赖注入容器

**目录：** `internel/di/container.go`

```go
// di/container.go
package di

import (
    "postapocgame/server/service/dungeonserver/internel/adapter/controller"
    "postapocgame/server/service/dungeonserver/internel/adapter/gateway"
    "postapocgame/server/service/dungeonserver/internel/usecase/move"
    "postapocgame/server/service/dungeonserver/internel/usecase/fight"
)

// Container 依赖注入容器
type Container struct {
    // Gateways
    configGateway gateway.ConfigGateway
    networkGateway gateway.NetworkGateway
    
    // Use Cases
    startMoveUseCase *move.StartMoveUseCase
    castSkillUseCase *fight.CastSkillUseCase
    
    // Controllers
    moveController  *controller.MoveController
    fightController *controller.FightController
}

func NewContainer() *Container {
    c := &Container{}
    
    // 初始化 Gateways
    c.configGateway = gateway.NewConfigGateway()
    c.networkGateway = gateway.NewNetworkGateway()
    
    // 初始化 Use Cases
    c.startMoveUseCase = move.NewStartMoveUseCase(c.configGateway, ...)
    c.castSkillUseCase = fight.NewCastSkillUseCase(c.configGateway, ...)
    
    // 初始化 Controllers
    c.moveController = controller.NewMoveController(c.startMoveUseCase, ...)
    c.fightController = controller.NewFightController(c.castSkillUseCase, ...)
    
    return c
}
```

### 6.2 在 Actor 中使用

```go
// fuben/actor_msg.go
func handleDoNetWorkMsg(msg actor.IActorMessage) {
    container := di.GetContainer()  // 获取依赖容器
    
    entity := getEntityFromContext(msg.GetContext())
    cliMsg := decodeClientMessage(msg.GetData())
    
    switch cliMsg.MsgId {
    case protocol.C2SProtocol_C2SStartMove:
        container.MoveController.HandleStartMove(msg.GetContext(), entity, cliMsg)
    case protocol.C2SProtocol_C2SUseSkill:
        container.FightController.HandleUseSkill(msg.GetContext(), entity, cliMsg)
    }
}
```

## 7. 测试策略

### 7.1 Use Case 层单元测试

```go
// usecase/move/start_move_test.go
func TestStartMoveUseCase_Execute(t *testing.T) {
    // Mock Repository
    mockSceneRepo := &MockSceneRepository{}
    mockEventPub := &MockEventPublisher{}
    
    // 创建 Use Case
    uc := NewStartMoveUseCase(mockSceneRepo, mockEventPub)
    
    // 执行测试
    entity := &domain.Entity{ID: 1}
    dest := domain.Position{X: 10, Y: 20}
    err := uc.Execute(ctx, entity, dest)
    
    // 验证结果
    assert.NoError(t, err)
    assert.True(t, entity.IsMoving())
    assert.True(t, mockEventPub.PublishMoveStartedCalled)
}
```

### 7.2 Controller 层集成测试

```go
// adapter/controller/move_controller_test.go
func TestMoveController_HandleStartMove(t *testing.T) {
    // 使用真实 Repository（可以连接测试配置）
    configGateway := gateway.NewConfigGateway()
    // ...
    
    controller := NewMoveController(startMoveUseCase, presenter)
    
    // 执行测试
    err := controller.HandleStartMove(ctx, entity, msg)
    
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
- [ ] 实现 Gateway（配置访问）
- [ ] 编写单元测试
- [ ] 更新协议注册
- [ ] 验证功能正常
- [ ] 删除旧代码

### 8.2 整体检查项

- [ ] 所有 EntitySystem 已迁移
- [ ] 所有框架依赖已移除
- [ ] 依赖注入容器已配置
- [ ] 单元测试覆盖率 > 70%
- [ ] 集成测试通过
- [ ] 文档已更新

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

### 9.4 实时性要求

- DungeonServer 是实时战斗服务器
- 业务逻辑必须高效
- 避免在 Use Case 层进行不必要的抽象

### 9.5 与 GameServer 的交互

- RPC 调用通过 Adapter 层封装
- 保持异步 RPC 的特性
- 错误处理统一在 Adapter 层

## 10. 参考资源

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Clean Architecture Example](https://github.com/bxcodec/go-clean-arch)
- [GameServer Clean Architecture 重构文档](./gameserver_CleanArchitecture重构文档.md)
- [项目开发进度文档](./服务端开发进度文档.md)

## 11. 关键代码位置（重构后）

### 11.1 Domain 层
- `internel/domain/entity.go` - 实体基类
- `internel/domain/position.go` - 位置值对象
- `internel/domain/skill.go` - 技能实体
- `internel/domain/repository/` - Repository 接口定义

### 11.2 Use Case 层
- `internel/usecase/move/` - 移动用例
- `internel/usecase/fight/` - 战斗用例
- `internel/usecase/buff/` - Buff 用例
- `internel/usecase/interfaces/` - Use Case 依赖接口

### 11.3 Adapter 层
- `internel/adapter/controller/` - 协议控制器
- `internel/adapter/presenter/` - 响应构建器
- `internel/adapter/gateway/` - 配置访问实现
- `internel/adapter/rpc/` - RPC 调用适配器
- `internel/adapter/event/` - 事件适配器

### 11.4 Infrastructure 层
- `internel/infrastructure/network/` - 网络网关
- `internel/infrastructure/config/` - 配置适配

### 11.5 DI 容器
- `internel/di/container.go` - 依赖注入容器

---

**下一步行动：**
1. 评审本文档，确认重构方案
2. 创建基础目录结构
3. 选择第一个系统（建议 MoveSys）进行试点重构
4. 验证重构效果后，逐步迁移其他系统

