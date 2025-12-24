# go-zero 自定义模板说明

本目录包含 go-zero 的自定义模板，已针对项目需求进行了修改，支持统一时间戳字段和软删除功能。

## 模板修改说明

### 1. 统一时间戳字段支持
- 所有生成的 Model 结构体自动包含三个字段：
  - `CreatedAt int64` - 创建时间（秒级时间戳）
  - `UpdatedAt int64` - 更新时间（秒级时间戳）
  - `DeletedAt int64` - 删除时间（秒级时间戳，0 表示未删除）

### 2. 软删除支持
- 所有查询方法自动添加 `deleted_at = 0` 条件，过滤已删除记录
- `Delete` 方法实现软删除：更新 `deleted_at` 为当前时间戳，而非真正删除
- `Update` 方法自动更新 `updated_at` 字段
- `Insert` 方法自动设置 `created_at` 和 `updated_at` 字段

### 3. 修改的模板文件
- `model/types.tpl` - 添加时间戳字段到结构体，添加分页和分片查询接口
- `model/find-one.tpl` - 查询时过滤已删除记录
- `model/find-one-by-field.tpl` - 按字段查询时过滤已删除记录
- `model/find-one-by-field-extra-method.tpl` - 主键查询时过滤已删除记录
- `model/delete.tpl` - 实现软删除逻辑
- `model/insert.tpl` - 自动设置创建时间和更新时间
- `model/update.tpl` - 自动更新 updated_at 并过滤已删除记录
- `model/customized.tpl` - 实现分页查询（FindPage）和分片查询（FindChunk）方法

### 4. 分页和分片查询支持
- **FindPage(ctx, page, pageSize)** - 分页查询
  - 参数：page（页码，从1开始）、pageSize（每页数量，最大100）
  - 返回：数据列表、总数、错误
  - 自动过滤已删除记录，按 id 倒序排列
  
- **FindChunk(ctx, limit, lastId)** - 分片查询
  - 参数：limit（每次查询数量，最大100）、lastId（上次查询的最后一条记录ID，0表示第一次查询）
  - 返回：数据列表、下次查询的lastId（0表示无更多数据）、错误
  - 自动过滤已删除记录，按 id 正序排列
  - 适用于大数据量分批处理场景

## 使用方法

### 从 SQL DDL 文件生成 Model

```bash
# 使用自定义模板生成 Model（推荐使用 scripts/generate-model.sh）
./scripts/generate-model.sh db/init.sql

# 或直接使用 goctl 命令
goctl model mysql ddl -src db/init.sql -dir internal/model -c -style gozero --home .template

# 参数说明：
# -src: SQL DDL 文件路径（推荐使用 db/init.sql）
# -dir: 生成的 Model 代码输出目录
# -c: 启用缓存（可选）
# -style: 代码风格（gozero）
# --home: 指定自定义模板目录
```

### 从数据库生成 Model

```bash
# 从数据库直接生成 Model
goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="table_name" -dir internal/model -c --home .template
```

## 注意事项

1. **数据库字段要求**：确保数据库表包含 `created_at`、`updated_at`、`deleted_at` 字段，类型为 `BIGINT`，默认值为 `0`
2. **时间戳格式**：所有时间戳字段使用秒级 Unix 时间戳（int64）
3. **软删除查询**：生成的查询方法会自动过滤 `deleted_at != 0` 的记录
4. **与 GORM 的兼容性**：go-zero 生成的 Model 基于 sqlx，如需与 GORM 配合使用，需要适配层

## 模板维护

如需修改模板：
1. 编辑 `.template/model/` 目录下的模板文件
2. 使用 `goctl model` 命令时指定 `--home .template` 参数
3. 模板修改后，重新生成 Model 代码即可应用新模板

