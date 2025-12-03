# SystemAdapter 系统开启检查优化方案

## 1. 问题描述

### 1.1 当前问题

在 Clean Architecture 重构后，当前的调用流程是：
```
Controller -> UseCase
```

但是 `sys_mgr.go` 的目的是控制功能的开启与关闭，而 Controller 直接调用 UseCase 时，并没有经过 System 的开启/关闭检查。如果系统未开启，UseCase 仍然会被执行，这不符合预期。

### 1.2 期望流程

期望的流程应该是：
```
Controller -> [检查 System 是否开启] -> UseCase
```

如果系统未开启，应该直接返回错误，不执行 UseCase。

### 1.3 职责边界问题

目前 UseCase 和 System 之间的职责有点暧昧：
- **SystemAdapter**：负责系统生命周期适配、事件订阅、状态管理（包括系统开启/关闭状态）
- **UseCase**：负责纯业务逻辑，不应该感知系统开启/关闭状态（这是框架层面的职责）

## 2. 解决方案

### 2.1 方案概述

**在 Controller 层添加系统开启状态检查**，在调用 UseCase 之前，先通过 SystemAdapter 的 helper 函数检查系统是否开启。

**理由**：
1. **符合 Clean Architecture 原则**：Controller 层负责协议处理和框架层面的检查（如系统开启状态），UseCase 层保持纯业务逻辑
2. **职责清晰**：系统开启/关闭是框架层面的状态管理，应该在 Controller 层处理
3. **向后兼容**：现有的 SystemAdapter helper 函数（如 `GetBagSys(ctx)`）已经实现了开启状态检查，可以直接复用

### 2.2 实现方式

#### 方式一：在 Controller 中直接检查（推荐）

在每个 Controller 的方法中，在调用 UseCase 之前，先通过 SystemAdapter helper 函数检查系统是否开启：

```go
// 示例：bag_controller.go
func (c *BagController) HandleOpenBag(ctx context.Context, msg *network.ClientMessage) error {
    // 1. 检查系统是否开启
    bagSys := system.GetBagSys(ctx)
    if bagSys == nil {
        // 系统未开启，返回错误
        return customerr.New("背包系统未开启")
    }
    
    // 2. 系统已开启，继续执行 UseCase
    // ... 原有逻辑
}
```

**优点**：
- 实现简单，直接复用现有的 helper 函数
- 职责清晰，Controller 负责框架层面的检查
- 不需要修改 UseCase 层

**缺点**：
- 需要在每个 Controller 方法中重复添加检查代码

#### 方式二：创建统一的 SystemValidator（可选优化）

创建一个 `SystemValidator` 接口和实现，统一处理系统开启检查：

```go
// adapter/system/system_validator.go
package system

import (
    "context"
    "postapocgame/server/internal/protocol"
    "postapocgame/server/pkg/customerr"
)

// SystemValidator 系统验证器接口
type SystemValidator interface {
    // CheckSystemEnabled 检查系统是否开启
    CheckSystemEnabled(ctx context.Context, sysId uint32) error
}

// systemValidatorImpl 系统验证器实现
type systemValidatorImpl struct{}

// NewSystemValidator 创建系统验证器
func NewSystemValidator() SystemValidator {
    return &systemValidatorImpl{}
}

// CheckSystemEnabled 检查系统是否开启
func (v *systemValidatorImpl) CheckSystemEnabled(ctx context.Context, sysId uint32) error {
    playerRole, err := adaptercontext.GetPlayerRoleFromContext(ctx)
    if err != nil {
        return customerr.Wrap(err)
    }
    
    sys := playerRole.GetSystem(sysId)
    if sys == nil {
        return customerr.New("系统不存在: %d", sysId)
    }
    
    if !sys.IsOpened() {
        return customerr.New("系统未开启: %d", sysId)
    }
    
    return nil
}
```

在 Controller 中使用：

```go
// 示例：bag_controller.go
type BagController struct {
    // ... 原有字段
    systemValidator system.SystemValidator
}

func NewBagController() *BagController {
    // ...
    return &BagController{
        // ...
        systemValidator: system.NewSystemValidator(),
    }
}

func (c *BagController) HandleOpenBag(ctx context.Context, msg *network.ClientMessage) error {
    // 检查系统是否开启
    if err := c.systemValidator.CheckSystemEnabled(ctx, uint32(protocol.SystemId_SysBag)); err != nil {
        return err
    }
    
    // 继续执行 UseCase
    // ...
}
```

**优点**：
- 统一了系统检查逻辑，便于维护
- 可以扩展其他检查逻辑（如权限检查等）

**缺点**：
- 增加了抽象层，可能过度设计

### 2.3 推荐方案

**推荐使用方式一（在 Controller 中直接检查）**，理由：
1. 实现简单，直接复用现有的 helper 函数
2. 代码清晰，每个 Controller 方法中都能看到系统检查逻辑
3. 符合 Clean Architecture 原则，Controller 层负责框架层面的检查

如果后续发现多个 Controller 有重复的系统检查逻辑，再考虑提取为统一的 SystemValidator。

## 3. 实施步骤

### 3.1 阶段一：为所有 Controller 添加系统检查

为每个 Controller 的方法添加系统开启检查：

1. **BagController**：`HandleOpenBag`、`HandleAddItem`
2. **MoneyController**：`HandleOpenMoney`
3. **EquipController**：`HandleEquipItem`
4. **SkillController**：`HandleLearnSkill`、`HandleUpgradeSkill`
5. **QuestController**：`HandleTalkToNPC`
6. **FubenController**：`HandleEnterDungeon`、`HandleSettleDungeon`
7. **ItemUseController**：`HandleUseItem`
8. **ShopController**：`HandleShopBuy`
9. **RecycleController**：`HandleRecycleItem`
10. **ChatController**：`HandleWorldChat`、`HandlePrivateChat`
11. **MailController**：相关方法
12. **GMController**：`HandleGMCommand`（GM 系统本身可能不需要检查，但可以统一处理）

### 3.2 阶段二：统一错误处理

为系统未开启的情况定义统一的错误码和错误消息：

```go
// pkg/customerr/errors.go
var (
    ErrSystemNotEnabled = customerr.New("系统未开启")
    ErrSystemNotFound   = customerr.New("系统不存在")
)
```

### 3.3 阶段三：补充测试

为每个 Controller 添加系统未开启场景的测试用例。

## 4. 代码示例

### 4.1 BagController 示例

```go
// HandleOpenBag 处理打开背包请求
func (c *BagController) HandleOpenBag(ctx context.Context, msg *network.ClientMessage) error {
    // 1. 检查系统是否开启
    bagSys := system.GetBagSys(ctx)
    if bagSys == nil {
        sessionID, _ := adaptercontext.GetSessionIDFromContext(ctx)
        // 返回错误响应
        resp := &protocol.S2COpenBagResp{
            Success: false,
            Message: "背包系统未开启",
        }
        return c.presenter.SendOpenBagResult(ctx, sessionID, resp)
    }
    
    // 2. 系统已开启，继续执行原有逻辑
    sessionID, err := adaptercontext.GetSessionIDFromContext(ctx)
    if err != nil {
        return err
    }
    
    roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
    if err != nil {
        return err
    }
    
    // ... 原有逻辑
}
```

### 4.2 FubenController 示例

```go
// HandleEnterDungeon 处理进入副本请求
func (c *FubenController) HandleEnterDungeon(ctx context.Context, msg *network.ClientMessage) error {
    // 1. 检查系统是否开启
    fubenSys := system.GetFubenSys(ctx)
    if fubenSys == nil {
        sessionID, _ := adaptercontext.GetSessionIDFromContext(ctx)
        resp := &protocol.S2CEnterDungeonResp{
            Success: false,
            Message: "副本系统未开启",
        }
        return c.presenter.SendEnterDungeonResult(ctx, sessionID, resp)
    }
    
    // 2. 系统已开启，继续执行 UseCase
    // ... 原有逻辑
}
```

## 5. 职责边界优化

### 5.1 SystemAdapter 职责

**SystemAdapter 应该负责**：
1. 生命周期适配（OnInit/RunOne/OnNewDay 等）
2. 事件订阅和分发
3. 管理与 Actor 运行模型强相关的运行时状态
4. **系统开启/关闭状态管理**（通过 `IsOpened()` 和 `SetOpened()`）

### 5.2 Controller 职责

**Controller 应该负责**：
1. 协议解析和参数验证
2. **框架层面的检查**（包括系统开启状态检查）
3. 调用 UseCase 执行业务逻辑
4. 调用 Presenter 构建响应

### 5.3 UseCase 职责

**UseCase 应该负责**：
1. 纯业务逻辑（不感知系统开启/关闭状态）
2. 业务规则校验
3. 数据操作（通过 Repository 接口）

## 6. 注意事项

1. **不要将系统检查逻辑下沉到 UseCase**：UseCase 应该保持纯业务逻辑，不应该感知系统开启/关闭状态
2. **统一错误处理**：系统未开启时，应该返回统一的错误码和错误消息
3. **向后兼容**：现有的 SystemAdapter helper 函数（如 `GetBagSys(ctx)`）已经实现了开启状态检查，可以直接复用
4. **测试覆盖**：为每个 Controller 添加系统未开启场景的测试用例

## 7. 后续优化（可选）

如果发现多个 Controller 有重复的系统检查逻辑，可以考虑：

1. **创建统一的 SystemValidator**（见方式二）
2. **在 BaseController 中添加系统检查方法**：
   ```go
   type BaseController struct {
       systemValidator system.SystemValidator
   }
   
   func (c *BaseController) CheckSystemEnabled(ctx context.Context, sysId uint32) error {
       return c.systemValidator.CheckSystemEnabled(ctx, sysId)
   }
   ```
3. **使用中间件模式**：在协议路由层统一添加系统检查（需要评估是否过度设计）

## 8. 总结

通过在 Controller 层添加系统开启状态检查，可以实现：
- ✅ Controller -> [检查 System 是否开启] -> UseCase 的流程
- ✅ 职责边界清晰：Controller 负责框架层面的检查，UseCase 保持纯业务逻辑
- ✅ 符合 Clean Architecture 原则
- ✅ 实现简单，直接复用现有的 helper 函数

