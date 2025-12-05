# CI 检查脚本使用说明

本目录包含用于代码质量检查的 CI 脚本，包括静态分析和禁止导入规则检查。

## 脚本列表

### 1. `ci_check.sh` / `ci_check.ps1`
主 CI 检查脚本，集成以下检查：
- `go vet`：Go 官方静态分析工具
- `staticcheck`：Go 静态分析工具（自动安装）
- gatewaylink 导入检查：禁止除白名单外的包引用 gatewaylink

### 2. `check_gatewaylink_imports.sh` / `check_gatewaylink_imports.ps1`
独立的 gatewaylink 导入检查脚本，可单独运行。

## 快速开始

### 最简单的使用方式

**Windows:**
```powershell
# 在项目根目录运行
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1
```

**Linux/Mac:**
```bash
# 在项目根目录运行
bash scripts/ci_check.sh
```

## 详细使用方法

### Linux / Mac 环境

```bash
# 方式 1：运行完整 CI 检查（推荐）
cd /path/to/postapocgame
bash scripts/ci_check.sh

# 方式 2：只检查 gatewaylink 导入
bash scripts/check_gatewaylink_imports.sh

# 方式 3：在项目根目录直接运行（脚本会自动定位）
cd /path/to/postapocgame
./scripts/ci_check.sh
```

### Windows 环境

```powershell
# 方式 1：运行完整 CI 检查（推荐）
cd C:\path\to\postapocgame
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1

# 方式 2：只检查 gatewaylink 导入
powershell -ExecutionPolicy Bypass -File scripts\check_gatewaylink_imports.ps1

# 方式 3：如果已设置执行策略，可以直接运行
.\scripts\ci_check.ps1
```

### 执行结果说明

脚本会输出以下信息：

- ✅ **绿色**：检查通过
- ⚠️ **黄色**：发现问题但继续执行（如 go vet 或 staticcheck 发现警告）
- ❌ **红色**：检查失败，脚本退出（如 gatewaylink 导入违规）

**注意**：
- `go vet` 和 `staticcheck` 发现的问题不会导致脚本退出，只会显示警告
- `gatewaylink` 导入检查失败会导致脚本退出，必须修复后才能继续

### 设置 PowerShell 执行策略（可选）

如果不想每次都使用 `-ExecutionPolicy Bypass`，可以设置执行策略：

```powershell
# 以管理员身份运行 PowerShell，然后执行：
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 之后就可以直接运行脚本：
.\scripts\ci_check.ps1
```

## 集成到 Git Hooks

### Pre-commit Hook（推荐）

在提交代码前自动运行检查，避免提交有问题的代码。

#### Linux / Mac

创建 `.git/hooks/pre-commit`：

```bash
#!/bin/bash

# 运行 CI 检查
bash scripts/ci_check.sh

# 如果检查失败，阻止提交
if [ $? -ne 0 ]; then
    echo "❌ CI checks failed. Please fix the issues before committing."
    exit 1
fi
```

然后设置执行权限：
```bash
chmod +x .git/hooks/pre-commit
```

#### Windows

创建 `.git/hooks/pre-commit`（PowerShell 脚本）：

```powershell
# 运行 CI 检查
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1

# 如果检查失败，阻止提交
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ CI checks failed. Please fix the issues before committing." -ForegroundColor Red
    exit 1
}
```

或者创建 `.git/hooks/pre-commit.bat`：

```batch
@echo off
powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1
if %errorlevel% neq 0 (
    echo ❌ CI checks failed. Please fix the issues before committing.
    exit 1
)
```

## 集成到 CI/CD 流程

### GitHub Actions 示例

创建 `.github/workflows/ci.yml`：

```yaml
name: CI Checks

on:
  pull_request:
    paths:
      - 'server/service/gameserver/**'
      - 'scripts/**'
  push:
    branches:
      - main

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run CI checks
        working-directory: server
        run: |
          cd ..
          bash scripts/ci_check.sh
```

### GitLab CI 示例

创建 `.gitlab-ci.yml`：

```yaml
stages:
  - check

ci-checks:
  stage: check
  image: golang:1.24
  script:
    - bash scripts/ci_check.sh
  only:
    - merge_requests
    - main
```

### Jenkins Pipeline 示例

```groovy
pipeline {
    agent any
    
    stages {
        stage('CI Checks') {
            steps {
                sh 'bash scripts/ci_check.sh'
            }
        }
    }
}
```

## 检查项说明

### 1. go vet
Go 官方静态分析工具，检查常见错误：
- 格式字符串错误
- 未使用的变量
- 其他常见问题

### 2. staticcheck
更强大的静态分析工具，检查：
- 未使用的代码
- 潜在的错误
- 性能问题
- 代码风格问题

**注意**：如果未安装，脚本会自动安装到 `$GOPATH/bin` 或 `$HOME/go/bin`。

### 3. gatewaylink 导入检查
检查是否有违规的 `gatewaylink` 导入。

**白名单文件**（允许导入 gatewaylink）：
- `server/service/gameserver/internel/app/playeractor/entity/player_network.go`
- `server/service/gameserver/internel/app/playeractor/entity/player_role.go`
- `server/service/gameserver/internel/adapter/gateway/network_gateway.go`
- `server/service/gameserver/internel/adapter/gateway/session_gateway.go`
- `server/service/gameserver/internel/infrastructure/gatewaylink/*`（所有文件）
- `server/service/gameserver/internel/app/engine/server.go`
- `server/service/gameserver/internel/infrastructure/dungeonserverlink/dungeon_cli.go`

**其他文件**：禁止直接导入 `gatewaylink`，应通过 `NetworkGateway` 接口使用。

## 常见问题

### Q: 脚本显示 "go vet found issues" 怎么办？
A: 这是正常的。`go vet` 和 `staticcheck` 可能会发现一些代码问题（如格式字符串警告），这些不会阻止脚本继续执行。你可以：
1. 查看具体问题并修复
2. 如果问题不影响功能，可以暂时忽略
3. 脚本会继续执行其他检查项

### Q: staticcheck 未找到怎么办？
A: 脚本会自动安装。如果自动安装失败，可以手动安装：
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

确保 `$GOPATH/bin` 或 `$HOME/go/bin` 在 PATH 中。

### Q: 如何跳过某个检查？
A: 可以修改脚本，注释掉不需要的检查项，或者直接运行单独的检查脚本：
```bash
# 只运行 go vet
go vet ./server/service/gameserver/...

# 只运行 staticcheck
staticcheck ./server/service/gameserver/...

# 只检查 gatewaylink 导入
bash scripts/check_gatewaylink_imports.sh
```

### Q: 如何添加新的白名单文件？
A: 编辑 `scripts/check_gatewaylink_imports.sh` 和 `scripts/check_gatewaylink_imports.ps1`，在 `ALLOWED_FILES` 数组中添加新文件路径。

### Q: Windows 下脚本执行失败？
A: 确保使用 PowerShell 运行，并使用 `-ExecutionPolicy Bypass` 参数，或者设置执行策略。

### Q: 如何查看详细的错误信息？
A: 脚本会输出详细的错误信息。如果检查失败，会显示：
- `go vet` 的错误位置和描述
- `staticcheck` 的问题列表
- 违规的 gatewaylink 导入文件列表

## 最佳实践

1. **提交前检查**：使用 pre-commit hook 在提交前自动运行检查
2. **PR 前检查**：在创建 Pull Request 前运行完整检查
3. **CI 集成**：在 CI/CD 流程中集成检查，确保所有代码都通过检查
4. **定期检查**：定期运行检查，及时发现和修复问题

## 相关文档

- 详细规则说明：`docs/过度设计整改清单.md`
- 架构文档：`docs/服务端开发进度文档_full.md`

