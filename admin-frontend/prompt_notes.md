# Vue3 前端开发 Cursor Prompt

## 系统级长期提示（放入 Cursor 项目设置）

```text
你是 admin-system-vue3 的"前端开发与文档协作助手"，需要：
1）按现有代码/文档开发或重构；2）完成功能后同步文档。

权威文档：
- docs/vue3实现方案.md        —— 前端实现方案：架构设计、功能模块、goctl 协同方案、实现步骤、项目结构。
- docs/前端开发进度.md         —— 开发进度追踪：已完成功能、待实现功能、技术决策记录。

架构原则（必须遵守）：
- 分层架构：Page → Component → Store(Pinia) → API → Backend
- 组件设计：高内聚低耦合，支持 Props/Slots/Emits，完整 TypeScript 类型
- 状态管理：Pinia 模块化，避免过度嵌套，使用计算属性处理派生状态
- API 调用：在 services 层统一处理，统一错误处理，拦截器在 utils/request.ts
- 路由设计：模块化定义，完整权限验证，懒加载优化
- 前后端协作：使用 go-zero 的 goctl 工具从 .api 文件生成 TypeScript 代码，确保类型一致性

开发规范：
- TypeScript 严格模式，完整的类型定义
- 组件命名：PascalCase，文件名与组件名一致
- 样式管理：SCSS + BEM 命名，使用变量管理主题色，避免全局污染
- API 请求：统一使用 Axios，拦截器处理 Token 和错误
- 权限控制：路由级 + 组件级 + 按钮级（v-permission 指令）
- 代码质量：ESLint + Prettier，无 console/debugger（生产环境）

Element Plus 使用规范：
- 仅使用 Tailwind 核心工具类，不使用自定义类（无编译器）
- 表单验证：使用 Element Plus Form 组件 + 自定义规则
- 表格操作：支持排序、筛选、分页、行内编辑
- 对话框：统一使用 Dialog/Drawer 组件

关键目录结构：
- src/api/           - API 接口定义（按模块分文件）
  - generated/       - goctl 自动生成的 TS 代码（不要手动修改）
  - index.ts         - 手动封装的 API（基于 generated/ 二次封装）
- src/components/    - 通用组件库（common/business/layout）
- src/pages/         - 页面组件（按功能模块分目录）
- src/stores/        - Pinia 状态管理（模块化）
- src/services/      - 业务服务层（封装复杂业务逻辑）
- src/utils/         - 工具函数（http/auth/storage/format等）
- src/router/        - 路由配置（模块化 + 守卫）
- src/types/         - TypeScript 类型定义
  - generated/       - 从后端 API 生成的类型（不要手动修改）
  - index.ts         - 手动定义的类型

goctl 代码生成规范：
- 后端更新 .api 文件后，使用 goctl api ts 生成 TypeScript 代码
- 生成命令：goctl api ts -api {file}.api -dir ./frontend/src/api/generated
- 生成的代码存放在 src/api/generated/ 和 src/types/generated/
- **禁止手动修改 generated/ 目录下的代码**
- 在 src/api/ 手动封装生成的代码，添加错误处理、拦截器等
- 类型定义优先使用 generated/ 中的类型，避免重复定义

文档维护规则：
- 实现方案文档（vue3实现方案.md）：仅在架构调整或新增模块时更新
- 进度文档（前端开发进度.md）：
  * 新功能完成后补充到"已完成功能"
  * 待实现功能用 [ ] 标记，进行中用 [⏳]，已完成用 [✅]
  * 技术决策/架构调整记录到"技术决策记录"
  * 关键文件路径补充到"关键代码位置"
- 修改文档必须真实读写文件，回复时简述改动（≤5行），不整篇粘贴

开发工作流：
1. 接到任务先读 vue3实现方案.md 对应章节（含 goctl 协同），了解设计原则
2. 查看 前端开发进度.md 确认当前进度和依赖关系
3. 等待后端提供 .api 文件，使用 goctl 生成 TypeScript 代码
4. 基于生成的类型和 API 函数进行二次封装
5. 开发时遵守架构原则和开发规范
6. 完成后更新进度文档，确保代码可编译、可运行
7. 提交代码前运行 lint 检查，修复所有警告
```

---

## 单次任务模板

### 1. 功能开发 + 同步文档

```text
阅读以下文档和代码后，完成功能并同步文档：
- 文档：docs/vue3实现方案.md、docs/前端开发进度.md
- 代码范围：{示例：src/pages/system/user/、src/api/user.ts}

目标：
- {用 2～3 句话描述要实现的功能，例如：实现用户管理页面，包含列表查询、新增、编辑、删除、启用禁用功能}

技术要求：
- 遵守分层架构：Page → Component → Store → API
- TypeScript 严格类型，ESLint 无错误
- 表单验证完整，错误提示友好
- 响应式设计，支持移动端适配
- 权限控制到按钮级别

文档更新：
- 前端开发进度.md：
  * "已完成功能"补充本次实现的功能说明
  * "待实现功能"更新状态（✅/⏳/ [ ]）
  * 如有技术决策记录到"技术决策记录"
  * 关键文件补充到"关键代码位置"
- vue3实现方案.md：仅在架构调整时更新

完成后用 ≤5 行总结：
1. 实现了哪些功能
2. 修改了哪些文件
3. 更新了哪些文档
```

### 2. 组件开发 + 封装

```text
开发通用组件并补充文档：
- 文档：docs/vue3实现方案.md、docs/前端开发进度.md
- 目标路径：{示例：src/components/business/UserSelector.vue}

组件需求：
- {描述组件功能，例如：用户选择器组件，支持单选/多选，搜索过滤，部门筛选}

技术要求：
- Props/Emits 完整类型定义
- 支持 v-model 双向绑定
- 提供合理的默认值和插槽
- 完整的 JSDoc 注释
- 使用 Element Plus 组件二次封装

文档更新：
- 在"已完成功能"或"通用组件"章节补充组件说明
- 包含：组件名、功能、主要 Props、使用示例
```

### 3. API 接口对接

```text
对接后端 API 并实现前端调用：
- 文档：docs/vue3实现方案.md、docs/前端开发进度.md
- 后端 API 文件：{示例：api/user.api}
- 前端目标：{示例：src/api/user.ts、src/types/user.ts}

开发步骤：
1. 从后端获取最新的 .api 文件
2. 使用 goctl 生成 TypeScript 代码：
   goctl api ts -api {file}.api -dir ./src/api/generated
3. 在 src/api/ 手动封装生成的代码：
   - 导入 generated/ 中的函数和类型
   - 添加统一的错误处理
   - 添加 loading 状态管理
   - 添加请求/响应拦截器（如需要）
4. 导出类型到 src/types/ 供全局使用
5. 在页面/组件中使用封装后的 API

示例封装：
```typescript
// src/api/user.ts
import { login, getUserList } from './generated/user'
import type { LoginRequest, LoginResponse } from './generated/user'
import { handleApiError } from '@/utils/error'

export const userApi = {
  async login(data: LoginRequest): Promise<LoginResponse> {
    try {
      return await login(data)
    } catch (error) {
      handleApiError(error)
      throw error
    }
  },
  getUserList
}

export type { LoginRequest, LoginResponse }
```

要求：
- **禁止手动修改 generated/ 目录下的代码**
- 类型定义优先使用 generated/ 中的类型
- 统一错误处理，友好的错误提示
- 完整的 TypeScript 类型约束
- 拦截器处理 Token 和通用错误

文档更新：
- 补充到"已完成功能"的 API 对接部分
- 记录 API 文件路径和生成命令
- 记录封装规范（如有特殊处理）
```

### 4. 页面重构/优化

```text
重构优化现有页面：
- 文档：docs/vue3实现方案.md、docs/前端开发进度.md
- 目标文件：{示例：src/pages/system/user/index.vue}

重构目标：
- {描述重构原因和目标，例如：优化表格性能，拆分复杂组件，改进状态管理}

要求：
- 保持功能不变
- 优化代码结构和性能
- 改进类型定义
- 增强可维护性

文档更新：
- 在"技术决策记录"补充重构说明
- 包含：重构原因、方案、效果
```

### 5. 架构讨论（不改代码）

```text
阅读 docs/vue3实现方案.md、docs/前端开发进度.md，分析以下架构问题并给方案：

背景：
- {描述问题，例如：状态管理混乱，组件职责不清，API 调用分散等}

期望输出：
- 明确各层职责边界
- 给出可落地的重构方案
- 评估改动范围和影响
- 提供迁移建议

输出格式：
- 问题分析（当前架构问题）
- 解决方案（具体改进措施）
- 实施步骤（分阶段实施计划）
- 风险评估（可能的影响和应对）
```

### 6. 只更新进度文档

```text
根据以下变化更新 docs/前端开发进度.md，其他文档不动：
- 完成功能：{列出已完成的功能}
- 新增任务：{列出新增的待实现功能}
- 技术决策：{列出需要记录的技术决策}

要求：
- 使用 [ ] / [⏳] / [✅] 管理 TODO
- 已完成移到"已完成功能"，TODO 中标记 ✅ 或删除
- 完成后简述改动（≤3行）
```

---

## goctl 代码生成专用模板

### 生成 TypeScript 代码

```text
从后端 .api 文件生成 TypeScript 代码：
- API 文件：{示例：api/user.api}
- 生成位置：src/api/generated/
- 生成命令：goctl api ts -api {file}.api -dir ./src/api/generated

执行步骤：
1. 确保已安装 goctl 工具（go install github.com/zeromicro/go-zero/tools/goctl@latest）
2. 在项目根目录执行生成命令
3. 检查生成的文件（类型定义、API 函数）
4. 在 src/api/ 创建同名文件进行二次封装
5. 导出类型到 src/types/

注意事项：
- generated/ 目录应添加到 .gitignore（可选，取决于团队规范）
- 如果选择提交 generated/ 代码，便于 code review 和版本对比
- 每次后端更新 .api 文件后需重新生成

文档更新：
- 记录生成命令和文件路径
- 更新"API 对接"部分的进度
```

### 批量生成多个模块

```text
批量生成多个模块的 TypeScript 代码：
- API 文件列表：{列出所有 .api 文件}
- 生成脚本：scripts/generate-api.sh

创建生成脚本：
```bash
#!/bin/bash
# scripts/generate-api.sh

API_DIR="../backend/api"
OUTPUT_DIR="./src/api/generated"

# 生成 user 模块
goctl api ts -api $API_DIR/user.api -dir $OUTPUT_DIR

# 生成 role 模块
goctl api ts -api $API_DIR/role.api -dir $OUTPUT_DIR

# 生成 permission 模块
goctl api ts -api $API_DIR/permission.api -dir $OUTPUT_DIR

echo "TypeScript 代码生成完成！"
```

使用方式：
```bash
chmod +x scripts/generate-api.sh
./scripts/generate-api.sh
```

文档更新：
- 在"开发工具"章节记录生成脚本
- 更新所有模块的 API 对接状态
```

### 处理生成代码冲突

```text
处理 goctl 生成的代码与现有代码冲突：

问题场景：
- {描述冲突情况，如：类型定义冲突、函数签名不匹配}

解决方案：
1. 检查后端 .api 文件定义是否正确
2. 重新生成 TypeScript 代码
3. 调整封装层代码适配新的类型定义
4. 如果是 goctl 生成问题，考虑：
   - 自定义 goctl 模板
   - 提交 issue 到 go-zero 仓库
   - 使用 AST 后处理生成的代码

文档更新：
- 如遇到通用问题，记录到"技术决策记录"
- 包含问题描述、解决方案、影响范围
```

---

## 常见场景快速提示

### 新增 CRUD 页面
```text
实现 {模块名} 的完整 CRUD 页面：
- 后端 API：{示例：api/user.api}
- 生成 TS 代码：goctl api ts -api api/user.api -dir ./src/api/generated
- 封装 API：src/api/{module}.ts
- 页面实现：src/pages/system/{module}/
  - 列表：查询、分页、排序、筛选
  - 新增：表单验证、提交
  - 编辑：数据回显、更新
  - 删除：确认提示、批量删除
- 权限：按钮级权限控制

遵守标准架构，完成后更新进度文档。
```

### 对接完整模块 API
```text
对接 {模块名} 的所有后端接口：
- 后端 API 文件：{示例：api/user.api}
- 生成命令：goctl api ts -api api/user.api -dir ./src/api/generated
- 封装位置：src/api/{module}.ts
- 类型导出：src/types/{module}.ts（re-export generated 类型）
- 更新进度文档的 API 对接部分

禁止手动修改 generated/ 目录，所有自定义逻辑在封装层实现。
```

### 开发业务组件
```text
开发业务组件：{组件名}
- 功能：{描述}
- 位置：src/components/business/{ComponentName}.vue
- 要求：Props/Emits 类型完整，支持插槽，有使用示例
- 更新进度文档的组件部分
```

---

## 维护说明

- 本 prompt 应与 docs/vue3实现方案.md、docs/前端开发进度.md 配合使用
- 文档结构调整时同步更新本 prompt
- 可根据项目实际情况新增场景模板
- 团队成员可基于此 prompt 自定义个人工作流