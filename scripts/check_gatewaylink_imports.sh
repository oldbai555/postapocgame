#!/bin/bash

# 检查 gatewaylink 导入的脚本
# 禁止除白名单外的包引用 gatewaylink

set -e

GAMESERVER_DIR="server/service/gameserver"

# 允许的文件列表（白名单）
# PlayerActor 相关文件允许使用 gatewaylink
ALLOWED_FILES=(
    "${GAMESERVER_DIR}/internel/app/playeractor/entity/player_network.go"
    "${GAMESERVER_DIR}/internel/app/playeractor/entity/player_role.go"
    "${GAMESERVER_DIR}/internel/adapter/gateway/network_gateway.go"
    "${GAMESERVER_DIR}/internel/adapter/gateway/session_gateway.go"
    "${GAMESERVER_DIR}/internel/infrastructure/gatewaylink/handler.go"
    "${GAMESERVER_DIR}/internel/infrastructure/gatewaylink/sender.go"
    "${GAMESERVER_DIR}/internel/infrastructure/gatewaylink/export.go"
    "${GAMESERVER_DIR}/internel/app/engine/server.go"
    "${GAMESERVER_DIR}/internel/infrastructure/dungeonserverlink/dungeon_cli.go"
)

echo "Checking gatewaylink imports..."

# 查找所有引用 gatewaylink 的文件
VIOLATIONS=$(grep -r "postapocgame/server/service/gameserver/internel/infrastructure/gatewaylink" \
    "${GAMESERVER_DIR}" --include="*.go" -l 2>/dev/null || true)

if [ -z "$VIOLATIONS" ]; then
    echo "✅ No gatewaylink imports found"
    exit 0
fi

# 检查是否有违规文件
UNAUTHORIZED=""
while IFS= read -r file; do
    IS_ALLOWED=false
    for allowed_file in "${ALLOWED_FILES[@]}"; do
        # 使用相对路径比较
        if [[ "$file" == *"$allowed_file"* ]] || [[ "$file" == "$allowed_file" ]]; then
            IS_ALLOWED=true
            break
        fi
    done
    if [ "$IS_ALLOWED" = false ]; then
        UNAUTHORIZED="${UNAUTHORIZED}${file}\n"
    fi
done <<< "$VIOLATIONS"

if [ -n "$UNAUTHORIZED" ]; then
    echo "❌ Unauthorized gatewaylink imports found:"
    echo -e "$UNAUTHORIZED"
    echo ""
    echo "Only the following files are allowed to import gatewaylink:"
    printf '%s\n' "${ALLOWED_FILES[@]}"
    exit 1
else
    echo "✅ All gatewaylink imports are in allowed files"
    exit 0
fi

