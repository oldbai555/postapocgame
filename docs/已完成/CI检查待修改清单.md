# CI 检查待修改清单

**生成时间**：2025-12-04  
**检查工具**：go vet, staticcheck, gatewaylink 导入检查

---

## ✅ 检查通过项

### 1. gatewaylink 导入检查
- **状态**：✅ 通过
- **说明**：所有 gatewaylink 导入都在白名单文件中，无违规情况

### 2. staticcheck 检查
- **状态**：✅ 通过
- **说明**：0 个问题，所有代码质量问题已修复

### 3. go vet 检查
- **状态**：✅ 通过
- **说明**：0 个警告，所有代码问题已修复

### 4. 编译检查
- **状态**：✅ 通过
- **说明**：所有代码编译通过，无错误

---

## ✅ 所有问题已修复

**最终验证时间**：2025-12-04  
**验证结果**：
- ✅ `staticcheck`：0 个问题
- ✅ `go vet`：0 个警告
- ✅ `gatewaylink` 导入检查：通过
- ✅ 编译检查：通过

**详细报告**：详见 `docs/staticcheck检查报告_2025-12-04.md`

---

## 📋 历史问题记录（已全部修复）

### 1. go vet 警告：非常量格式字符串 ✅ 已修复

go vet 检测到 3 处使用了非常量格式字符串，这可能导致潜在的格式化错误。

**修复状态**：✅ 已完成（2025-12-04）

#### 问题 1：`chat_private.go:54` ✅ 已修复

**文件**：`server/service/gameserver/internel/usecase/chat/chat_private.go`  
**行号**：54  
**问题**：`customerr.NewError(reason)` 中 `reason` 是变量，不是常量格式字符串

**修复前代码**：
```go
if ok, reason := chatdomain.ValidateContent(content, privateChatMaxRunes); !ok {
    return customerr.NewError(reason)
}
```

**修复后代码**：
由于 `customerr.NewError` 支持格式化参数（`format string, args ...interface{}`），应该使用格式化字符串：

```go
if ok, reason := chatdomain.ValidateContent(content, privateChatMaxRunes); !ok {
    return customerr.NewError("聊天内容验证失败: %s", reason)
}
```

这样 `reason` 作为参数传入，而不是作为格式字符串本身。

---

#### 问题 2：`chat_world.go:47` ✅ 已修复

**文件**：`server/service/gameserver/internel/usecase/chat/chat_world.go`  
**行号**：47  
**问题**：`customerr.NewError(reason)` 中 `reason` 是变量，不是常量格式字符串

**修复前代码**：
```go
if ok, reason := chatdomain.ValidateContent(content, worldChatMaxRunes); !ok {
    return customerr.NewError(reason)
}
```

**修复后代码**：
```go
if ok, reason := chatdomain.ValidateContent(content, worldChatMaxRunes); !ok {
    return customerr.NewError("聊天内容验证失败: %s", reason)
}
```

---

#### 问题 3：`buff_sys.go:43` ✅ 已修复

**文件**：`server/service/gameserver/internel/app/dungeonactor/entitysystem/buff_sys.go`  
**行号**：43  
**问题**：`customerr.NewErrorByCode` 中使用了 `fmt.Sprintf`，导致格式字符串不是常量

**修复前代码**：
```go
return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), fmt.Sprintf("buff %d not found", buffId))
```

**修复后代码**：
```go
return customerr.NewErrorByCode(int32(protocol.ErrorCode_Internal_Error), "buff %d not found", buffId)
```

**额外修复**：移除了未使用的 `fmt` 包导入。

---

## 📝 修复状态

### ✅ 已完成修复（2025-12-04）
1. ✅ **buff_sys.go:43** - 已修复为使用格式化参数，并移除了未使用的 `fmt` 导入
2. ✅ **chat_private.go:54** - 已修复格式字符串问题
3. ✅ **chat_world.go:47** - 已修复格式字符串问题

### 验证结果
- ✅ **go vet**：所有警告已消除，检查通过
- ✅ **gatewaylink 导入检查**：通过
- ✅ **staticcheck**：版本不匹配问题已解决（2025-12-04），已从源码使用 Go 1.24.10 重新编译，现在可以正常工作并检测代码质量问题

---

## 🔍 检查工具说明

### go vet
- **用途**：Go 官方静态分析工具
- **检查项**：常见错误、格式字符串问题、未使用的变量等
- **状态**：发现 3 个警告

### staticcheck
- **状态**：✅ 已更新（2025-12-04）
- **说明**：已从源码使用 Go 1.24.10 重新编译 staticcheck，版本不匹配问题已解决
- **当前版本**：2025.1.1 (0.6.1)，使用 Go 1.24.10 编译

### gatewaylink 导入检查
- **状态**：✅ 通过
- **说明**：所有导入都在白名单中

---

## ✅ 修复后验证

修复完成后，已运行以下命令验证：

```powershell
# Windows
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1
```

**验证结果**：
- ✅ **go vet**：通过，无警告
- ✅ **gatewaylink 导入检查**：通过
- ✅ **staticcheck**：版本不匹配问题已解决，已使用 Go 1.24.10 重新编译，现在可以正常工作

所有代码问题和工具问题已修复完成！

---

## ✅ 最新修复（2025-12-04 第二轮）

### 1. 错误处理一致性 ✅ 已修复

**问题**：代码中存在 `fmt.Errorf` 和 `errors.New` 的混用，不符合统一错误处理规范。

**修复内容**：
- ✅ 将所有 `fmt.Errorf` 替换为 `customerr.NewError`（支持格式化参数）
- ✅ 修复的文件：
  - `server/service/gameserver/internel/adapter/system/gm_tools.go`
  - `server/service/gameserver/internel/app/playeractor/entitysystem/sys_mgr.go`
  - `server/service/gameserver/internel/app/engine/config.go`
  - `server/service/gameserver/internel/infrastructure/dungeonserverlink/dungeon_cli.go`

**修复示例**：
```go
// 修复前
return fmt.Errorf("dungeon service not connected: srvType=%d", srvType)

// 修复后
return customerr.NewError("dungeon service not connected: srvType=%d", srvType)
```

---

### 2. SA1012 - nil Context 问题 ✅ 已修复

**问题**：代码中多处传递 `nil` 作为 Context，staticcheck 建议使用 `context.TODO()`。

**修复内容**：
- ✅ 将所有 `pr.WithContext(nil)` 替换为 `pr.WithContext(context.TODO())`
- ✅ 将所有 `sendPublicActorMessage(nil, ...)` 替换为 `sendPublicActorMessage(context.TODO(), ...)`
- ✅ 修复的文件：
  - `server/service/gameserver/internel/app/playeractor/entity/player_role.go`（14 处）
  - `server/service/gameserver/internel/app/playeractor/entity/player_role_economy.go`（2 处）
  - `server/service/gameserver/internel/core/gshare/message_sender.go`（1 处）

---

### 3. SA4023 - 永远不会为真的比较 ✅ 已修复

**问题**：代码中对永远不会为 nil 的返回值进行了 nil 检查。

**修复内容**：
- ✅ 移除了 `GetAttrSys() == nil` 检查（`NewBaseEntity` 中总是初始化）
- ✅ 移除了 `GetFightSys() == nil` 检查
- ✅ 移除了 `GetMoveSys() == nil` 检查（在 `actor_msg.go` 中）
- ✅ 修复的文件：
  - `server/service/gameserver/internel/app/dungeonactor/entity/entity.go`
  - `server/service/gameserver/internel/app/dungeonactor/entity/monster.go`
  - `server/service/gameserver/internel/app/dungeonactor/entity/rolest.go`
  - `server/service/gameserver/internel/app/dungeonactor/fuben/actor_msg.go`

---

### 4. S1040 - 冗余的类型断言 ✅ 已修复

**问题**：`GetBySession` 返回的已经是 `iface.IEntity` 类型，不需要再进行类型断言。

**修复内容**：
- ✅ 移除了所有对 `GetBySession` 返回值的冗余类型断言
- ✅ 修复的文件：
  - `server/service/gameserver/internel/app/dungeonactor/entity/rolest.go`
  - `server/service/gameserver/internel/app/dungeonactor/entitysystem/drop_sys.go`
  - `server/service/gameserver/internel/app/dungeonactor/entitysystem/fight_sys.go`
  - `server/service/gameserver/internel/app/dungeonactor/entitysystem/move_sys.go`（6 处）

**修复示例**：
```go
// 修复前
entityAny, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
if !ok || entityAny == nil {
    return nil
}
entity, ok := entityAny.(iface.IEntity)
if !ok {
    return nil
}

// 修复后
entity, ok := entitymgr.GetEntityMgr().GetBySession(sessionId)
if !ok || entity == nil {
    return nil
}
```

---

### 5. S1023 - 冗余的 return 语句 ✅ 已修复

**问题**：函数末尾的冗余 `return` 语句。

**修复内容**：
- ✅ 移除了 `sys_mgr.go:119` 的冗余 `return` 语句
- ✅ 移除了 `player_role.go:428` 的冗余 `return` 语句

---

### 6. SA4003 - 无效的数值比较 ✅ 已修复

**问题**：`uint32 >= math.MaxUint32` 永远不会为真，因为 `math.MaxUint32` 是 `uint32` 的最大值。

**修复内容**：
- ✅ 将 `idxMap[et] >= math.MaxUint32` 改为 `idxMap[et] == math.MaxUint32`
- ✅ 将 `magicMap[et] >= math.MaxUint16` 改为 `magicMap[et] == math.MaxUint16`
- ✅ 修复了重复递增的逻辑错误
- ✅ 修复的文件：
  - `server/service/gameserver/internel/app/dungeonactor/entitymgr/handle.go`

---

## ✅ 已修复问题

### 1. SA1029 - Context Key 类型问题 ✅ 已修复（2025-12-04）

**问题**：不应该使用内置类型 `string` 作为 context key，应该定义自定义类型以避免冲突。

**修复内容**：
- ✅ 在 `gshare/protocol.go` 中定义 `ContextKey` 类型
- ✅ 将所有字符串 key 替换为自定义类型 `gshare.ContextKeySession` 和 `gshare.ContextKeyRole`
- ✅ 修复了约 20 处违规用法，涉及以下文件：
  - `dungeon_item_controller.go`
  - `move_controller.go`（4 处）
  - `revive_controller.go`
  - `skill_controller.go`
  - `dungeonactor/adapter.go`
  - 以及其他使用 `context.WithValue` 的文件

**规范文档**：已创建 `docs/Context_Key规范.md`，定义了 Context Key 的使用规范。

**验证结果**：
- ✅ 编译通过
- ✅ `staticcheck` 检查通过，无 SA1029 警告

---

### 2. U1000 - 未使用的函数/变量 ✅ 已修复（2025-12-04）

**问题**：部分函数和变量未被使用。

**修复内容**：
- ✅ 删除 `player_role_economy.go` 中未使用的 `rollbackInfo` 字段：
  - 第一个 `rollbackInfo` 结构体：删除 `itemID`, `count` 字段
  - 第二个 `rollbackInfo` 结构体：删除 `itemID`, `key`, `count`, `snapshot` 字段
  - 这些字段是预留的，但当前代码只处理货币回滚，不需要物品回滚字段
- ✅ 删除 `public_role.go` 中未使用的 `lastCleanOfflineMessagesTime` 字段
  - 该字段定义了但从未使用，可能是预留功能，当前离线消息清理逻辑已在 `GetOfflineMessages` 中实现

**验证结果**：
- ✅ 编译通过
- ✅ `staticcheck` 检查通过，0 个 U1000 警告

**说明**：
- 未使用的函数（如 `flushAOIChanges`, `sendClientProtoViaPlayerActor` 等）已在 `docs/未使用函数分析报告.md` 中分析并修复
- `offlineDataFlushIntervalMs` 常量已正确使用，不在 U1000 列表中

---

## 📝 修复统计

### ✅ 已修复（2025-12-04 第一轮）
- ✅ **错误处理一致性**：4 个文件，约 10 处修复
- ✅ **SA1012 - nil Context**：3 个文件，17 处修复
- ✅ **SA4023 - 永远不会为真的比较**：4 个文件，4 处修复
- ✅ **S1040 - 冗余的类型断言**：4 个文件，9 处修复
- ✅ **S1023 - 冗余的 return 语句**：2 个文件，2 处修复
- ✅ **SA4003 - 无效的数值比较**：1 个文件，1 处修复

### ✅ 已修复（2025-12-04 第二轮）
- ✅ **SA1029 - Context Key 类型**：约 20 处，已定义规范并修复
  - 定义了 `ContextKey` 类型
  - 创建了 `docs/Context_Key规范.md` 规范文档
  - 修复了所有直接使用字符串作为 context key 的问题

### ✅ 已修复（2025-12-04 第三轮）
- ✅ **U1000 - 未使用的函数/变量**：已全部修复
  - ✅ 删除 `player_role_economy.go` 中未使用的 `rollbackInfo` 字段（`itemID`, `count`, `key`, `snapshot`）
  - ✅ 删除 `public_role.go` 中未使用的 `lastCleanOfflineMessagesTime` 字段
  - ✅ 验证 `offlineDataFlushIntervalMs` 常量已正确使用（不在 U1000 列表中）

### ✅ 已修复（2025-12-04 第四轮）
- ✅ **未使用函数修复**：详见 `docs/未使用函数分析报告.md`
  - ✅ 修复了 AOI 系统相关函数的使用（`flushAOIChanges`, `notifyAppear`, `notifyDisappear` 等）
  - ✅ 修复了离线数据刷新功能（`flushOfflineDataIfNeeded`）
  - ✅ 替换了 `sendClientProtoViaPlayerActor` 的使用（7 处）
  - ✅ 处理了 `spawnScene2Monsters` 作为 fallback
  - ✅ 删除了未使用的 `logWarn` 函数

### 📊 总计
- **修复文件数**：20+ 个文件
- **修复问题数**：70+ 处
- **创建规范文档**：3 份
- **最终状态**：✅ 所有 staticcheck 和 go vet 检查通过

---

## 📌 注意事项

1. **格式字符串安全**：使用变量作为格式字符串可能导致格式化错误或安全问题，建议使用明确的格式化参数
2. **错误处理一致性**：✅ 已统一使用 `customerr.NewError` 和 `customerr.NewErrorByCode`（支持格式化参数）
3. **Context 使用规范**：✅ 已统一使用 `context.TODO()` 替代 `nil`
4. **代码审查**：修复后建议进行代码审查，确保错误消息清晰且有意义
5. **Context Key 规范**：✅ 已定义 `docs/Context_Key规范.md`，所有 Context Key 必须使用自定义类型
6. **未使用代码清理**：✅ 所有 U1000 问题已修复，代码更加简洁
7. **未使用函数分析**：详见 `docs/未使用函数分析报告.md`，所有报告中的函数已修复或使用
8. **代码质量**：✅ 所有 staticcheck 和 go vet 检查通过，代码质量达到标准

---

## 🎉 修复完成总结

**最终验证时间**：2025-12-04  
**验证工具**：`staticcheck`, `go vet`, `gatewaylink` 导入检查

### ✅ 最终验证结果

```bash
# 运行完整 CI 检查
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1
```

**输出结果**：
```
✅ go vet passed
✅ staticcheck passed
✅ All gatewaylink imports are in allowed files
✅ All CI checks completed!
```

### 📊 修复成果

- **修复文件数**：20+ 个文件
- **修复问题数**：70+ 处
- **创建规范文档**：3 份
  - `docs/Context_Key规范.md`
  - `docs/未使用函数分析报告.md`
  - `docs/staticcheck检查报告_2025-12-04.md`
- **代码质量提升**：
  - ✅ 统一的错误处理方式
  - ✅ 规范的 Context 使用
  - ✅ 类型安全的 Context Key
  - ✅ 简洁的代码结构（无未使用代码）
  - ✅ 正确的逻辑实现

### 🔗 相关文档

- `docs/CI检查待修改清单.md` - 本文档
- `docs/Context_Key规范.md` - Context Key 使用规范
- `docs/未使用函数分析报告.md` - 未使用函数分析
- `docs/staticcheck检查报告_2025-12-04.md` - 完整检查报告

---

**所有 CI 检查问题已修复完成！代码质量达到标准。** ✅

