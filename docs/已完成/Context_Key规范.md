# Context Key 使用规范

**创建时间**：2025-12-04  
**适用范围**：`server/service/gameserver` 目录

---

## 📋 规范说明

### 问题背景

Go 的 `context.WithValue` 接受 `interface{}` 类型的 key，但使用内置类型（如 `string`）作为 key 存在以下问题：

1. **类型冲突风险**：不同包可能使用相同的字符串值作为 key，导致值被意外覆盖
2. **静态检查警告**：`staticcheck` 的 SA1029 规则会检测并警告此类用法
3. **代码可维护性**：字符串 key 难以追踪和重构

### 解决方案

**使用自定义类型作为 Context Key**，确保类型安全且避免冲突。

---

## ✅ 规范定义

### 1. Context Key 类型定义

所有 Context Key 应在 `server/service/gameserver/internel/core/gshare/protocol.go` 中统一定义：

```go
package gshare

// ContextKey 类型用于定义 Context 的 key，避免使用字符串导致的冲突
type ContextKey string

const (
	// ContextKeyRole 用于在 Context 中存储玩家角色对象
	ContextKeyRole ContextKey = "playerRole"
	// ContextKeySession 用于在 Context 中存储 Session ID
	ContextKeySession ContextKey = "playerRoleSession"
)
```

### 2. 使用规范

#### ✅ 正确用法

```go
import "postapocgame/server/service/gameserver/internel/core/gshare"

// 设置 Context 值
ctx := context.WithValue(ctx, gshare.ContextKeySession, sessionID)
ctx = context.WithValue(ctx, gshare.ContextKeyRole, playerRole)

// 获取 Context 值
sessionID, ok := ctx.Value(gshare.ContextKeySession).(string)
if !ok {
    // 处理错误
}
```

#### ❌ 错误用法

```go
// ❌ 错误：直接使用字符串作为 key
ctx := context.WithValue(ctx, "session", sessionID)
ctx := context.WithValue(ctx, "playerRole", playerRole)

// ❌ 错误：使用未定义的类型
type myKey string
ctx := context.WithValue(ctx, myKey("session"), sessionID)
```

---

## 📝 新增 Context Key 的流程

### 步骤 1：在 `gshare/protocol.go` 中定义

```go
const (
	// ContextKeyXXX 用于在 Context 中存储 XXX
	ContextKeyXXX ContextKey = "xxx"
)
```

### 步骤 2：添加注释说明

- 说明该 key 的用途
- 说明存储的值类型
- 说明使用场景

### 步骤 3：导出并使用

- 通过 `gshare.ContextKeyXXX` 访问
- 确保所有使用处都使用统一的 key

---

## 🔍 检查与验证

### 静态检查

运行 `staticcheck` 检查是否有违规用法：

```bash
# Windows
cd server
staticcheck ./service/gameserver/... | Select-String -Pattern "SA1029"

# Linux/Mac
cd server
staticcheck ./service/gameserver/... | grep "SA1029"
```

### CI 集成

在 CI 脚本中已包含 `staticcheck` 检查，会自动检测 SA1029 违规。

---

## 📊 已定义的 Context Key

| Key | 类型 | 用途 | 定义位置 |
|-----|------|------|----------|
| `ContextKeyRole` | `iface.IPlayerRole` | 存储玩家角色对象 | `gshare/protocol.go` |
| `ContextKeySession` | `string` | 存储 Session ID | `gshare/protocol.go` |

---

## ⚠️ 注意事项

1. **不要直接使用字符串**：所有 Context Key 必须使用 `gshare.ContextKeyXXX` 常量
2. **类型断言**：从 Context 获取值时，必须进行类型断言并检查 `ok`
3. **统一管理**：所有 Context Key 应在 `gshare/protocol.go` 中统一定义
4. **向后兼容**：修改现有 Context Key 时，需要考虑向后兼容性

---

## 🔗 相关文件

- `server/service/gameserver/internel/core/gshare/protocol.go` - Context Key 定义
- `docs/CI检查待修改清单.md` - CI 检查清单

---

## 📌 修复历史

**2025-12-04**：
- ✅ 定义 `ContextKey` 类型
- ✅ 将所有字符串 key 替换为自定义类型
- ✅ 修复所有 SA1029 警告（约 20 处）
- ✅ 建立规范文档

