# CI 检查脚本快速使用指南

## 🚀 快速开始（30秒）

### Windows 用户

1. 打开 PowerShell
2. 进入项目根目录：`cd C:\bgg\postapocgame`
3. 运行检查：`powershell -ExecutionPolicy Bypass -File scripts\ci_check.ps1`

### Linux/Mac 用户

1. 打开终端
2. 进入项目根目录：`cd /path/to/postapocgame`
3. 运行检查：`bash scripts/ci_check.sh`

## 📋 检查内容

脚本会自动检查：

1. ✅ **go vet** - Go 官方静态分析
2. ✅ **staticcheck** - 更强大的静态分析（自动安装）
3. ✅ **gatewaylink 导入规则** - 禁止违规导入

## ⚠️ 重要说明

- **go vet 和 staticcheck 的警告不会阻止脚本**：这些工具可能会发现代码问题，但脚本会继续执行
- **gatewaylink 导入违规会阻止脚本**：如果发现违规导入，脚本会退出，必须修复后才能继续

## 🔍 查看详细说明

更多信息请查看：[scripts/README.md](README.md)

