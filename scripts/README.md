# go-zero 代码生成脚本

本目录包含用于生成 go-zero 代码的便捷脚本。

## 脚本列表

### 1. generate-sql.sh - SQL 脚本生成工具

用于快速生成新功能模块的初始化 SQL 脚本。

**使用方法：**
```bash
# 可在任何目录下运行，脚本会自动定位项目目录
./scripts/generate-sql.sh -group <group> -name <name>
```

**参数：**
- `-group <group>`: 功能组名（必需，如 `user`, `file`）
- `-name <name>`: 功能名称（必需，如 `用户管理`, `文件管理`）

**示例：**
```bash
# 生成用户管理模块的 SQL
./scripts/generate-sql.sh -group user -name 用户管理

# 生成文件管理模块的 SQL
./scripts/generate-sql.sh -group file -name 文件管理
```

**功能说明：**
- 生成的 SQL 文件位于 `admin-server/db/` 目录下
- 文件名格式：`init_<group>.sql`
- 主键为自增，不需要手动赋值
- 临时目录已在 `data.sql` 中初始化（id=9）
- 生成的菜单默认归类在临时目录下
- 包含以下内容：
  - 菜单数据（主菜单 + 新增/编辑/删除按钮）
  - 权限数据（list/create/update/delete）
  - 接口数据（GET/POST/PUT/DELETE）
  - 权限-菜单关联数据
  - 权限-接口关联数据
- **同时生成 `.api` 文件内容**，可直接复制追加到 `admin-server/api/admin.api`

**输出内容：**
1. **建表 SQL 文件**：生成到 `admin-server/db/create_table_<group>.sql`
   - 包含默认字段：`id`（主键自增）、`created_at`、`updated_at`、`deleted_at`
   - 表名使用 `{{.Group}}`（可根据需要手动修改表名）
   - 包含主键和 `deleted_at` 索引
2. **初始化 SQL 文件**：生成到 `admin-server/db/init_<group>.sql`
   - 包含菜单、权限、接口等初始化数据
   - 菜单路径：`/temp/<group>`（临时目录下）
   - 前端组件路径：`temp/<GroupUpper>List`
3. **.api 文件**：生成到 `admin-server/api/<group>.api.temp`，包含：
   - 类型定义（Item、ListReq、ListResp、CreateReq、UpdateReq）
   - 服务定义（@server 块，包含 List/Create/Update/Delete 四个接口）
4. **Vue 页面文件**：生成到 `admin-frontend/src/views/temp/<GroupUpper>List.vue`
   - 使用 `D2Table` 组件
   - 包含搜索、列表、新增、编辑、删除功能
   - 自动调用生成的 API（`<group>Api.list/create/update/delete`）

**使用步骤：**
1. 执行脚本生成建表SQL、初始化SQL、.api文件和Vue页面
2. **先执行建表SQL**：将 `create_table_<group>.sql` 在数据库中执行（或手动添加到 `admin-server/db/tables.sql`）
3. **再执行初始化SQL**：将 `init_<group>.sql` 在数据库中执行
4. **追加 .api 内容**：将 `<group>.api.temp` 文件的内容追加到 `admin-server/api/admin.api` 的：
   - 类型定义部分：追加到 `type (` 块内（在 `)` 之前）
   - 服务定义部分：追加到文件末尾
5. **生成前端 TypeScript 代码**：执行 `./scripts/generate-ts.sh` 生成前端 API 代码
6. **前端页面已生成**：Vue 页面已生成到 `admin-frontend/src/views/temp/<GroupUpper>List.vue`，可直接使用
7. 追加完成后，可以删除 `<group>.api.temp` 文件

**注意事项：**
- 菜单、按钮、接口的启用状态默认为 1（启用），可根据需要修改
- 菜单默认归类在临时目录下，可在菜单管理中调整
- 生成的 SQL 文件需要在数据库中执行
- .api 内容需要手动复制追加到 `admin-server/api/admin.api`

**技术实现：**
- 使用 Golang 编写，通过模板文件生成 SQL、.api 和 Vue 页面
- 建表 SQL 模板文件：`scripts/sqlgen/templates/create_table.sql.tpl`
- 初始化 SQL 模板文件：`scripts/sqlgen/templates/init_module.sql.tpl`
- .api 模板文件：`scripts/sqlgen/templates/init_module.api.tpl`
- Vue 页面模板文件：`scripts/sqlgen/templates/list_page.vue.tpl`
- 如需修改生成逻辑，只需修改模板文件即可

**注意事项：**
- 建表SQL中的表名使用 `{{.Group}}`，可根据实际需要手动修改（如添加 `admin_` 前缀等）
- 建表SQL只包含默认字段，业务字段需要手动添加
- 建议将建表SQL添加到 `admin-server/db/tables.sql` 中统一管理

### 2. generate-model.sh
从 SQL DDL 文件生成 Model 代码（使用自定义模板支持软删除、时间戳字段、分页和分片查询）

**使用方法：**
```bash
# 可在任何目录下运行，脚本会自动定位项目目录
./scripts/generate-model.sh db/init.sql

# 或使用相对路径
./scripts/generate-model.sh init.sql
```

**选项：**
- `-c, --cache`: 启用缓存（默认启用）
- `-d, --dir DIR`: 指定输出目录（默认: internal/model）
- `-h, --help`: 显示帮助信息

**示例：**
```bash
# 基本用法
./scripts/generate-model.sh 002_init_rbac.sql

# 禁用缓存
./scripts/generate-model.sh 002_init_rbac.sql --no-cache

# 指定输出目录
./scripts/generate-model.sh 002_init_rbac.sql -d internal/model/custom
```

### 3. generate-api.sh
从 .api 文件生成 API Handler 代码骨架

**使用方法：**
```bash
# 可在任何目录下运行，脚本会自动定位项目目录
./scripts/generate-api.sh user.api

# 或使用相对路径
./scripts/generate-api.sh api/user.api
```

**示例：**
```bash
# 生成用户管理 API Handler
./scripts/generate-api.sh user.api

# 生成角色管理 API Handler
./scripts/generate-api.sh role.api
```

### 4. generate-ts.sh
从 .api 文件生成前端 TypeScript 代码

**使用方法：**
```bash
# 可在任何目录下运行，脚本会自动定位项目目录
# 默认使用 admin-server/api/admin.api
./scripts/generate-ts.sh

# 或指定 API 文件
./scripts/generate-ts.sh admin.api
./scripts/generate-ts.sh api/admin.api
```

**示例：**
```bash
# 使用默认 admin.api 生成 TypeScript 代码
./scripts/generate-ts.sh

# 指定其他 API 文件
./scripts/generate-ts.sh user.api
```

**注意事项：**
- 生成的代码在 `admin-frontend/src/api/generated/` 目录
- **禁止手动修改 generated/ 目录下的文件**
- 在 `src/api/` 中二次封装（错误处理、拦截器集成、统一返回类型）
- 如果生成的路径包含 `/auth` 前缀（如 `/api/v1/auth/login`），需要在封装时修正路径（去掉 `/auth`，改为 `/api/v1/login`）

## 前置要求

1. **安装 goctl**：
   ```bash
   go install github.com/zeromicro/go-zero/tools/goctl@latest
   # 确保 GOPATH/bin 在 PATH 中
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **初始化模板**（首次使用）：
   ```bash
   cd admin-server
   goctl template init --home .template
   ```

3. **设置脚本执行权限**（Linux/Mac）：
   ```bash
   chmod +x scripts/*.sh
   ```

## 注意事项

### Model 生成
- 使用自定义模板（`--home .template`），支持：
  - 统一时间戳字段（`created_at`、`updated_at`、`deleted_at`，int64类型，秒级时间戳）
  - 软删除功能
  - 分页查询方法（`FindPage`）
  - 分片查询方法（`FindChunk`）
- 生成的 Model 包含完整的 CRUD 操作方法
- 确保 SQL 文件中的表包含 `created_at`、`updated_at`、`deleted_at` 字段（BIGINT类型，默认值0）
- **推荐使用 `db/init.sql` 作为初始化SQL文件**（在未上线时只维护这一份）

### API Handler 生成
- 生成的 Handler 代码在 `internal/handler/` 目录
- Types 定义会生成临时文件，需要手动合并到 `internal/types/types.go`
- 生成的代码需要手动改造以使用 Service 层和统一响应格式

### TypeScript 代码生成
- 生成的代码在 `admin-frontend/src/api/generated/` 目录
- **禁止手动修改 generated/ 目录下的文件**
- 在 `src/api/` 中二次封装（错误处理、拦截器集成、统一返回类型）
- 注意路径修正：go-zero 会根据 group 名称添加路径前缀，如果后端路由没有对应前缀，需要在封装时修正

## 完整工作流示例

### 1. 更新数据库初始化文件
```sql
-- db/init.sql（在未上线时只维护这一份初始化SQL）
CREATE TABLE IF NOT EXISTS `new_table` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(64) NOT NULL,
  `created_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(秒级时间戳)',
  `updated_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新时间(秒级时间戳)',
  `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(秒级时间戳,0表示未删除)',
  PRIMARY KEY (`id`),
  KEY `idx_new_table_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. 生成 Model 代码
```bash
# 可在任何目录下运行
./scripts/generate-model.sh db/init.sql
```

### 3. 定义 API 文件
```go
// api/new_module.api
syntax = "v1"

service new-module-api {
    @handler NewModuleList
    get /new-modules returns (NewModuleListResp)
}
```

### 4. 生成 Handler 代码
```bash
./scripts/generate-api.sh new_module.api
```

### 5. 生成前端 TypeScript 代码
```bash
# 生成前端 TypeScript 代码
./scripts/generate-ts.sh
```

### 6. 手动完善
- 合并 types 到 `internal/types/types.go`
- 实现 Service 层业务逻辑
- 改造 Handler 使用 Service 和统一响应格式
- 在 `src/api/` 中封装前端 API（修正路径、错误处理、拦截器集成）

## 故障排查

### goctl 未找到
```bash
# 检查是否在 PATH 中
which goctl  # Linux/Mac
Get-Command goctl  # Windows

# 如果未安装，执行：
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 确保 GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin  # Linux/Mac
$env:PATH += ";$(go env GOPATH)\bin"  # Windows PowerShell
```

### 模板目录不存在
```bash
cd admin-server
goctl template init --home .template
```

### 权限问题（Linux/Mac）
```bash
chmod +x scripts/*.sh
```

