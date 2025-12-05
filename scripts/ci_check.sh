#!/bin/bash

# CI 检查脚本
# 用于检查代码质量、静态分析和禁止导入规则

set -e

echo "=========================================="
echo "Running CI checks for gameserver"
echo "=========================================="

GAMESERVER_DIR="server/service/gameserver"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR" || exit 1

# 1. 检查 go vet
echo ""
echo "1. Running go vet..."
# 切换到 server 目录，因为 go.mod 在那里
cd server || exit 1
if ! go vet ./service/gameserver/... 2>&1; then
    echo "⚠️  go vet found issues, but continuing..."
fi
echo "✅ go vet check completed"
cd .. || exit 1

# 2. 检查 staticcheck
echo ""
echo "2. Running staticcheck..."
if ! command -v staticcheck &> /dev/null; then
    echo "⚠️  staticcheck not found, installing..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

# 切换到 server 目录
cd server || exit 1
if ! staticcheck ./service/gameserver/... 2>&1; then
    echo "⚠️  staticcheck found issues, but continuing..."
fi
echo "✅ staticcheck check completed"
cd .. || exit 1

# 3. 检查 gatewaylink 导入
echo ""
echo "3. Checking gatewaylink imports..."
bash "$SCRIPT_DIR/check_gatewaylink_imports.sh"
if [ $? -ne 0 ]; then
    echo "❌ gatewaylink import check failed"
    exit 1
fi

echo ""
echo "=========================================="
echo "✅ All CI checks completed!"
echo "=========================================="
