# SystemAdapter Code Review 清单

## 文档目的

本文档用于 Code Review 时检查 SystemAdapter 是否符合 Clean Architecture 原则，防止业务逻辑退化到适配层。

## 检查清单

### 1. 职责边界检查

- [ ] **SystemAdapter 是否只负责生命周期适配？**
  - 检查点：SystemAdapter 中的方法是否只做"何时调用哪个 UseCase"的调度
  - 禁止：在 SystemAdapter 中实现业务规则逻辑（如校验、计算、状态转换等）

- [ ] **业务逻辑是否在 UseCase 层？**
  - 检查点：所有业务逻辑（校验、计算、状态变更）是否都在 UseCase 层实现
  - 禁止：在 SystemAdapter 中直接操作数据库、网络或配置

- [ ] **状态管理是否合理？**
  - 检查点：SystemAdapter 中的状态是否只与 Actor 运行模型相关（如 dirty 标记、定时任务）
  - 禁止：在 SystemAdapter 中存储业务数据状态

### 2. 代码质量检查

- [ ] **是否使用了 servertime？**
  - 检查点：所有时间相关操作是否使用 `servertime.Now()` 而非 `time.Now()`
  - 禁止：直接调用 `time.Now()`、`time.Since()` 等标准库接口

- [ ] **是否通过接口访问依赖？**
  - 检查点：SystemAdapter 是否通过接口（Repository、Gateway）访问外部依赖
  - 禁止：直接 import 其他 SystemAdapter 或框架层代码

- [ ] **是否有未使用的字段或方法？**
  - 检查点：SystemAdapter 中是否有遗留的、未使用的字段或方法
  - 建议：删除或标记为废弃

### 3. 注释和文档检查

- [ ] **头部注释是否完整？**
  - 检查点：SystemAdapter 头部注释是否包含：
    - 生命周期职责说明
    - 业务逻辑位置说明
    - 防退化说明（禁止编写业务规则逻辑）

- [ ] **方法注释是否清晰？**
  - 检查点：每个方法是否有清晰的注释说明其职责
  - 建议：对于对外接口，说明使用场景和注意事项

### 4. 防退化机制检查

- [ ] **是否包含防退化说明？**
  - 检查点：SystemAdapter 头部注释是否包含防退化说明
  - 要求：明确标注"禁止编写业务规则逻辑，只允许调用 UseCase 与管理生命周期"

- [ ] **是否有业务逻辑下沉到 UseCase？**
  - 检查点：如果发现 SystemAdapter 中有业务逻辑，是否已重构到 UseCase 层
  - 要求：所有业务逻辑必须在 UseCase 层实现

## 常见问题与处理

### 问题 1：SystemAdapter 中包含业务逻辑

**示例**：
```go
// ❌ 错误：在 SystemAdapter 中实现业务逻辑
func (a *BagSystemAdapter) AddItem(ctx context.Context, itemID uint32, count uint32) error {
    // 业务逻辑：检查物品配置、堆叠规则、容量检查
    itemConfig := getItemConfig(itemID)
    if itemConfig == nil {
        return errors.New("item not found")
    }
    // ...
}

// ✅ 正确：在 SystemAdapter 中只调用 UseCase
func (a *BagSystemAdapter) AddItem(ctx context.Context, itemID uint32, count uint32) error {
    roleID, err := adaptercontext.GetRoleIDFromContext(ctx)
    if err != nil {
        return err
    }
    return a.addItemUseCase.Execute(ctx, roleID, itemID, count, 0)
}
```

**处理方式**：
1. 将业务逻辑提取到 UseCase 层
2. SystemAdapter 只负责调用 UseCase

### 问题 2：SystemAdapter 中直接操作数据库或网络

**示例**：
```go
// ❌ 错误：在 SystemAdapter 中直接操作数据库
func (a *BagSystemAdapter) SaveBagData(ctx context.Context) error {
    db := database.GetDB()
    return db.Save(bagData)
}

// ✅ 正确：通过 Repository 接口操作
func (a *BagSystemAdapter) SaveBagData(ctx context.Context) error {
    roleID, _ := adaptercontext.GetRoleIDFromContext(ctx)
    return a.playerRepo.SaveBinaryData(ctx, roleID, binaryData)
}
```

**处理方式**：
1. 通过 Repository/Gateway 接口访问外部依赖
2. 禁止直接操作数据库、网络或配置

### 问题 3：SystemAdapter 中存储业务数据状态

**示例**：
```go
// ❌ 错误：在 SystemAdapter 中存储业务数据状态
type BagSystemAdapter struct {
    items map[uint32]*Item  // 业务数据状态
}

// ✅ 正确：业务数据存储在 BinaryData 中，SystemAdapter 只维护辅助索引
type BagSystemAdapter struct {
    itemIndex map[uint32][]*protocol.ItemSt  // 辅助索引，用于快速查找
}
```

**处理方式**：
1. 业务数据存储在 BinaryData 中
2. SystemAdapter 只维护与 Actor 运行模型相关的状态（如辅助索引、dirty 标记）

## 参考文档

- `docs/gameserver_CleanArchitecture重构文档.md`：Clean Architecture 架构说明
- `docs/gameserver_adapter_system演进规划.md`：SystemAdapter 演进规划
- `server/service/gameserver/internel/adapter/system/base_system_adapter.go`：BaseSystemAdapter 职责说明

## 检查流程

1. **提交前自检**：开发者提交代码前，对照本清单自行检查
2. **Code Review**：Reviewer 对照本清单逐项检查
3. **发现问题**：如果发现问题，要求开发者重构到 UseCase 层
4. **记录问题**：将常见问题记录到本文档的"常见问题与处理"章节

## 更新记录

- 2025-01-XX：创建 Code Review 清单文档

