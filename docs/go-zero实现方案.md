# Go-Zero 后端详细实现方案

## 目录
1. [项目整体架构](#项目整体架构)
2. [核心功能模块](#核心功能模块)
3. [详细实现步骤](#详细实现步骤)
4. [项目结构](#项目结构)
5. [技术栈选择](#技术栈选择)

## 项目整体架构

### 架构设计原则
- 高内聚、低耦合，便于后续二次开发
- 模块化设计，支持功能模块的独立扩展
- 遵循 DDD（领域驱动设计）思想
- 使用分层架构（API层、业务层、数据层）

### 分层架构
```
┌─────────────────────────────────────┐
│         API Layer (HTTP/RPC)        │
├─────────────────────────────────────┤
│         Service Layer               │
├─────────────────────────────────────┤
│         Repository Layer            │
├─────────────────────────────────────┤
│         Model/Entity Layer          │
├─────────────────────────────────────┤
│         Database Layer (MySQL)      │
└─────────────────────────────────────┘
```
> 数据访问层统一使用 goctl model 生成的 sqlx + cache 代码，不再保留手写 GORM 版本；Model 生成使用自定义模板，内置 `created_at/updated_at/deleted_at`、软删除、分页与分片查询。

## 核心功能模块

### 1. 认证与授权模块
**功能需求：**
- 用户登录/注册/登出
- JWT Token 管理（生成、验证、刷新）
- 密码加密存储（bcrypt）
- 登录日志记录
- 会话管理

**关键技术点：**
- JWT 令牌双令牌方案（Access Token + Refresh Token）
- 中间件实现权限验证
- 黑名单机制处理登出

### 2. 用户权限管理模块
**功能需求：**
- 用户(User)管理：增删改查、启用禁用、批量操作
- 角色(Role)管理：创建、编辑、删除、权限分配
- 权限(Permission)管理：权限定义、权限分配、权限验证
- 部门(Department)管理：部门树结构、部门成员管理
- 用户-角色关联管理
- 角色-权限关联管理

**关键技术点：**
- RBAC（基于角色的访问控制）模型
- 权限验证中间件
- 缓存优化权限查询（Redis）
- 树形结构数据处理

### 3. 菜单与资源管理模块
**功能需求：**
- 菜单树管理（支持无限级联）
- 菜单权限绑定
- 操作按钮权限管理
- API 资源定义与管理
- 用户获取个性化菜单列表

**关键技术点：**
- 树形结构递归处理
- 权限与菜单的动态关联
- 菜单缓存策略

### 4. 系统配置管理模块
**功能需求：**
- 系统参数配置：应用名称、logo、主题色、超时设置等
- 配置分组管理
- 配置值的动态读写
- 配置变更审计日志
- 配置的热更新（无需重启应用）

**关键技术点：**
- 配置持久化
- Redis 缓存配置提高查询性能
- 事件驱动配置变更通知

### 5. 系统日志模块
**功能需求：**
- 操作日志：记录用户的所有增删改操作
- 登录日志：记录登录时间、IP、设备等信息
- 系统日志：应用运行日志、错误日志
- 日志查询与分页
- 日志导出功能（Excel/CSV）
- 日志定期清理机制

**关键技术点：**
- 异步日志处理（使用 channel/goroutine）
- 日志分级存储
- 日志查询优化（加索引）
- 定时任务清理历史日志

### 6. 数据字典模块
**功能需求：**
- 字典类型管理
- 字典项目管理
- 字典数据缓存
- 字典数据查询接口（供前端使用）

**关键技术点：**
- 字典数据缓存策略
- 类型安全的字典值处理

### 7. 文件管理模块
**功能需求：**
- 文件上传（支持多文件上传、断点续传）
- 文件下载
- 文件删除
- 文件列表查询
- 文件类型校验
- 存储策略支持（本地存储、OSS、S3等）

**关键技术点：**
- 文件上传中间件
- 断点续传实现
- 存储接口抽象（策略模式）
- 文件预处理（压缩、缩略图等）

### 8. 接口文档与测试模块
**功能需求：**
- Swagger/OpenAPI 自动生成
- 接口版本管理
- API 限流与分级

**关键技术点：**
- go-zero 自带的 API 文档生成
- 限流中间件实现

## 详细实现步骤

### 第一阶段：基础工程搭建
1. 初始化 go-zero 项目
2. 配置环境变量与配置文件管理
3. 数据库初始化与迁移
4. 公共工具函数库搭建

### 第二阶段：认证与授权
1. 实现用户登录/注册功能
2. JWT 令牌生成与验证
3. 权限验证中间件开发
4. 用户-角色-权限三层关联模型实现

### 第三阶段：核心业务模块
1. 用户管理模块完整实现
2. 角色管理模块完整实现
3. 权限管理模块完整实现
4. 部门管理模块完整实现

### 第四阶段：支撑功能模块
1. 菜单管理模块实现
2. 系统配置管理模块实现
3. 数据字典模块实现
4. 文件上传模块实现

### 第五阶段：日志与监控
1. 操作日志系统实现
2. 登录日志系统实现
3. 系统监控与健康检查
4. 错误追踪与上报

### 第六阶段：优化与扩展
1. 缓存策略优化（Redis）
2. 数据库查询优化（索引、分页）
3. 接口文档完善
4. 单元测试与集成测试

## 项目结构

```
admin-system/
├── api/                          # API 定义文件
│   ├── user.api
│   ├── role.api
│   ├── permission.api
│   ├── menu.api
│   ├── config.api
│   └── file.api
├── cmd/                          # 应用入口
│   └── api/
│       └── main.go
├── internal/                     # 内部实现
│   ├── config/                   # 配置模块
│   │   └── config.go
│   ├── middleware/               # 中间件
│   │   ├── auth.go
│   │   ├── log.go
│   │   └── cors.go
│   ├── handler/                  # 路由处理器
│   │   ├── user_handler.go
│   │   ├── role_handler.go
│   │   ├── permission_handler.go
│   │   └── ...
│   ├── service/                  # 业务逻辑层
│   │   ├── user_service.go
│   │   ├── role_service.go
│   │   ├── permission_service.go
│   │   └── ...
│   ├── model/                    # 数据模型
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   └── ...
│   ├── repository/               # 数据访问层
│   │   ├── user_repo.go
│   │   ├── role_repo.go
│   │   ├── permission_repo.go
│   │   └── ...
│   ├── logic/                    # 核心业务逻辑
│   │   ├── user_logic.go
│   │   └── ...
│   ├── types/                    # 自动生成的类型定义
│   │   └── types.go
│   └── svc/                      # 服务上下文
│       └── service_context.go
├── pkg/                          # 公共工具包
│   ├── utils/
│   │   ├── crypto.go             # 加密工具
│   │   ├── response.go           # 统一响应格式
│   │   ├── error.go              # 错误处理
│   │   └── logger.go
│   ├── jwt/
│   │   ├── token.go
│   │   └── claims.go
│   ├── cache/
│   │   └── redis.go
│   └── storage/
│       ├── local.go
│       ├── oss.go
│       └── s3.go
├── db/                           # 数据库相关
│   ├── migrations/               # 数据库迁移脚本
│   │   ├── 001_init_user.sql
│   │   ├── 002_init_role.sql
│   │   └── ...
│   └── seeds/                    # 初始化数据
│       └── init_data.sql
├── etc/                          # 配置文件
│   └── api.yaml
├── test/                         # 测试文件
│   ├── user_test.go
│   └── ...
├── go.mod
├── go.sum
└── README.md
```

## 技术栈选择

### 核心框架
- **go-zero**: 微服务框架，高性能、易扩展
- **gorm**: ORM 框架，支持多种数据库
- **MySQL 8.0+**: 关系型数据库

### 认证授权
- **JWT**: 令牌认证
- **bcrypt**: 密码加密

### 缓存与会话
- **Redis 6.0+**: 缓存与会话存储
- **go-redis**: Redis 客户端

### 日志与监控
- **zap**: 高性能日志库
- **promethues**: 监控指标

### 其他工具
- **goctl**: go-zero 代码生成工具
- **sqlc**: SQL 代码生成（可选）
- **migrate**: 数据库迁移工具

### 开发工具
- **Air**: Hot reload 工具
- **TestUtil**: 单元测试工具

## 关键实现要点

### 1. API 设计规范
```
认证相关：
- POST /api/v1/auth/login      - 登录
- POST /api/v1/auth/logout     - 登出
- POST /api/v1/auth/refresh    - 刷新令牌

用户相关：
- GET  /api/v1/users           - 用户列表
- POST /api/v1/users           - 创建用户
- GET  /api/v1/users/:id       - 获取用户详情
- PUT  /api/v1/users/:id       - 更新用户
- DELETE /api/v1/users/:id     - 删除用户

遵循 RESTful 设计规范
```

### 2. 错误处理统一
- 定义统一的错误码体系
- 所有错误通过标准化的格式返回
- 包含错误追踪 ID 便于问题排查

### 3. 权限验证流程
- 中间件层验证 Token 合法性
- Handler 层获取当前用户信息
- Service 层进行业务权限校验
- 支持细粒度权限控制

### 4. 数据库设计要点
- 设计合理的索引策略
- **数据库初始化规范**：
  - 在未上线时，`admin-server/db/` 目录只维护一份初始化SQL文件（`db/init.sql`）
  - 所有表必须包含 `created_at`、`updated_at`、`deleted_at` 字段（BIGINT类型，秒级时间戳，默认值0）
  - 上线后如需增量变更，再使用 migration 脚本
- **统一时间戳字段规范**：
  - 所有 DB 模型必须包含三个字段：`created_at`、`updated_at`、`deleted_at`
  - 类型统一为 `int64`（数据库字段类型为 BIGINT），存储秒级时间戳（Unix timestamp）
  - `created_at`：创建时间，创建时自动设置
  - `updated_at`：更新时间，更新时自动设置
  - `deleted_at`：删除时间，用于软删除（0 表示未删除，>0 表示已软删除）
- **软删除实现**：
  - GORM 默认不支持 int64 类型的软删除，需要手动实现
  - Repository 层查询时自动过滤 `deleted_at = 0` 的记录
  - 删除操作通过设置 `deleted_at` 为当前时间戳实现，而非真正删除
- **分页和分片查询**：
  - go-zero 生成的 Model 包含 `FindPage` 方法（分页查询）和 `FindChunk` 方法（分片查询）
  - `FindPage`：适用于列表查询场景，返回数据列表和总数
  - `FindChunk`：适用于大数据量分批处理场景，基于 lastId 进行分片查询
- 考虑数据库分表策略（大数据量场景）

### 5. 缓存策略
- 热数据缓存（用户、权限、菜单等）
- 缓存预热机制
- 缓存失效与更新策略
- 缓存穿透、击穿、雪崩防护

### 6. 性能优化
- 数据库连接池配置
- Redis 连接池配置
- 分页查询强制限制
- N+1 查询问题解决

### 7. 安全考虑
- SQL 注入防护（使用参数化查询）
- XSS 防护（数据验证与转义）
- CSRF 防护（token 验证）
- 敏感信息加密存储
- API 速率限制

## 依赖包清单

```go
// 核心依赖
github.com/zeromicro/go-zero
gorm.io/gorm
gorm.io/driver/mysql

// 认证授权
github.com/golang-jwt/jwt/v4
golang.org/x/crypto

// 缓存（统一使用 go-zero stores/redis 组件，避免直接依赖 go-redis/v9）
github.com/zeromicro/go-zero/core/stores/redis

// 日志
go.uber.org/zap
go.uber.org/zap/zapcore

// 配置管理
github.com/spf13/viper

// 工具库
github.com/google/uuid
github.com/spf13/cobra
```

### 8. 常量与枚举管理
- 系统级固定枚举（如通用状态 `"ok"/"error"`、Redis Key 前缀、固定路径 `/api/v1/ping`、限流提示文案等）统一在 `admin-server/internal/consts` 包中集中维护，禁止在业务代码中直接硬编码字符串。
- 业务可配置枚举（需要运营/管理员通过界面维护的选项，例如订单状态、通知类型等）统一走「数据字典」方案（参见 `docs/后端开发进度.md` 中的数据字典规范）。
- 新增模块时，先判断是**系统常量**还是**业务字典**：
  - **系统常量**：和基础设施/协议/技术实现强绑定、不会通过数据库动态调整的值 → 放入 `consts` 包。
  - **业务字典**：和业务含义强相关，需要被业务/运营调整或在多模块复用的值 → 使用字典表 + 字典接口。
- 常量命名使用 `PascalCase`，按功能分组（状态、Redis、限流、路径等）并添加注释，确保语义清晰，便于后续人员理解和全局检索。

## 开发建议

1. **优先使用 go-zero 工具进行代码生成**：
   - **Model 代码生成**：使用 `goctl model mysql ddl` 从 SQL 文件生成 Model 代码
     - **推荐使用脚本**：`./scripts/generate-model.sh db/init.sql`（可在任何目录运行）
     - 或直接命令：`goctl model mysql ddl -src db/init.sql -dir internal/model -c --home .template`
     - 使用项目自定义模板（`admin-server/.template`），支持：
       - 统一时间戳字段（`created_at`、`updated_at`、`deleted_at`，int64类型，秒级时间戳）
       - 软删除功能
       - 分页查询方法（`FindPage`）
       - 分片查询方法（`FindChunk`）
     - 生成的 Model 包含完整的 CRUD 操作方法，可直接使用
   - **API Handler 代码生成**：使用 `goctl api go` 从 `.api` 文件生成 Handler 代码骨架
   - **前端 TypeScript 代码生成**：使用 `goctl api ts` 从 `.api` 文件生成前端类型和 API 代码
   - **原则**：能使用 go-zero 工具生成的代码，一律使用工具生成；只有在业务需要特殊处理时，才进行自定义开发

2. **自定义模板说明**：
   - 项目已配置 go-zero 自定义模板（`admin-server/.template`），支持：
     - 统一时间戳字段（`created_at`、`updated_at`、`deleted_at`，int64 类型，秒级时间戳）
     - 软删除功能（查询自动过滤 `deleted_at = 0`，删除操作更新 `deleted_at` 而非真正删除）
     - 分页查询方法（`FindPage(ctx, page, pageSize)`）：返回数据列表和总数
     - 分片查询方法（`FindChunk(ctx, limit, lastId)`）：基于 lastId 的分片查询，适用于大数据量分批处理
   - 使用 `goctl` 命令时，必须指定 `--home .template` 参数以使用自定义模板
   - 模板文件说明详见 `admin-server/.template/README.md`

3. **Types 统一维护约定**：`admin-server/internal/types/types.go` 由人工统一维护，禁止被 goctl 覆盖；从 `.api` 重新生成代码时，只参考生成的临时 types 内容，按需手工合并后丢弃生成文件

4. **编写详细的接口文档**：便于前后端联调

5. **建立规范的 commit 历史**：便于版本管理和回溯

6. **编写单元测试和集成测试**：确保代码质量

7. **建立 API 版本管理机制**：支持平滑升级

8. **定期进行代码审查**：保证代码质量一致

9. **使用数据库迁移工具管理 schema**：避免手动修改数据库

## 后续扩展建议

- **OA系统扩展**：在此基础上添加流程引擎、审批模块
- **SCRM系统扩展**：添加客户管理、销售漏斗、营销自动化
- **SaaS 系统扩展**：添加租户隔离、多租户支持、计费模块
- **监控告警**：集成 Prometheus + Grafana 进行系统监控
- **消息队列**：集成 RabbitMQ/Kafka 处理异步任务
- **搜索引擎**：集成 Elasticsearch 实现全文搜索
