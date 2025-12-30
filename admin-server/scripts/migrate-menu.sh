#!/bin/bash

# 菜单迁移脚本
# 用于将前端菜单从一个目录迁移到另一个目录
# 支持迁移 SQL 文件和前端 Vue 组件

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 项目根目录（scripts的父目录）
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 显示使用说明
usage() {
    echo -e "${GREEN}菜单迁移脚本${NC}"
    echo ""
    echo "用法:"
    echo "  $0 -old-path <old_path> -new-path <new_path> -old-component <old_component> -new-component <new_component> [-sql-file <sql_file>]"
    echo ""
    echo "参数:"
    echo "  -old-path <old_path>          旧菜单路径（如 /temp/chat）"
    echo "  -new-path <new_path>          新菜单路径（如 /chatroom/chat）"
    echo "  -old-component <old_component> 旧组件路径（如 temp/ChatList）"
    echo "  -new-component <new_component> 新组件路径（如 chatroom/ChatList）"
    echo "  -sql-file <sql_file>          可选，SQL 文件路径（如 admin-server/db/migrations/init_chat.sql）"
    echo ""
    echo "选项:"
    echo "  -h, --help                    显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -old-path /temp/chat -new-path /chatroom/chat -old-component temp/ChatList -new-component chatroom/ChatList -sql-file admin-server/db/migrations/init_chat.sql"
    echo ""
    echo "功能:"
    echo "  1. 更新 SQL 文件中的菜单路径和组件路径"
    echo "  2. 移动前端 Vue 组件文件"
    echo "  3. 更新组件文件中的导入路径（如果有）"
    echo ""
}

# 检查参数
if [ $# -eq 0 ]; then
    usage
    exit 0
fi

OLD_PATH=""
NEW_PATH=""
OLD_COMPONENT=""
NEW_COMPONENT=""
SQL_FILE=""

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -old-path)
            OLD_PATH="$2"
            shift 2
            ;;
        -new-path)
            NEW_PATH="$2"
            shift 2
            ;;
        -old-component)
            OLD_COMPONENT="$2"
            shift 2
            ;;
        -new-component)
            NEW_COMPONENT="$2"
            shift 2
            ;;
        -sql-file)
            SQL_FILE="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}错误: 未知参数: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# 检查必需参数
if [ -z "$OLD_PATH" ] || [ -z "$NEW_PATH" ] || [ -z "$OLD_COMPONENT" ] || [ -z "$NEW_COMPONENT" ]; then
    echo -e "${RED}错误: 必须提供 -old-path、-new-path、-old-component、-new-component 参数${NC}"
    usage
    exit 1
fi

# 显示配置信息
echo -e "${GREEN}=== 菜单迁移工具 ===${NC}"
echo "项目根目录:  $PROJECT_ROOT"
echo "旧菜单路径:  $OLD_PATH"
echo "新菜单路径:  $NEW_PATH"
echo "旧组件路径:  $OLD_COMPONENT"
echo "新组件路径:  $NEW_COMPONENT"
if [ -n "$SQL_FILE" ]; then
    echo "SQL 文件:    $SQL_FILE"
fi
echo ""

# 确认执行
read -p "确认执行迁移? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 1. 更新 SQL 文件（如果提供了）
if [ -n "$SQL_FILE" ]; then
    SQL_FILE_PATH="$PROJECT_ROOT/$SQL_FILE"
    if [ -f "$SQL_FILE_PATH" ]; then
        echo -e "${GREEN}正在更新 SQL 文件...${NC}"
        # 使用 sed 替换路径和组件
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s|$OLD_PATH|$NEW_PATH|g" "$SQL_FILE_PATH"
            sed -i '' "s|$OLD_COMPONENT|$NEW_COMPONENT|g" "$SQL_FILE_PATH"
        else
            # Linux
            sed -i "s|$OLD_PATH|$NEW_PATH|g" "$SQL_FILE_PATH"
            sed -i "s|$OLD_COMPONENT|$NEW_COMPONENT|g" "$SQL_FILE_PATH"
        fi
        echo -e "${GREEN}✓ SQL 文件更新成功${NC}"
    else
        echo -e "${YELLOW}警告: SQL 文件不存在: $SQL_FILE_PATH${NC}"
    fi
fi

# 2. 移动前端组件文件
OLD_COMPONENT_FILE="$PROJECT_ROOT/admin-frontend/src/views/$OLD_COMPONENT.vue"
NEW_COMPONENT_DIR="$PROJECT_ROOT/admin-frontend/src/views/$(dirname $NEW_COMPONENT)"
NEW_COMPONENT_FILE="$PROJECT_ROOT/admin-frontend/src/views/$NEW_COMPONENT.vue"

if [ -f "$OLD_COMPONENT_FILE" ]; then
    echo -e "${GREEN}正在移动前端组件文件...${NC}"
    # 创建新目录（如果不存在）
    mkdir -p "$NEW_COMPONENT_DIR"
    # 移动文件
    mv "$OLD_COMPONENT_FILE" "$NEW_COMPONENT_FILE"
    echo -e "${GREEN}✓ 前端组件文件移动成功${NC}"
    
    # 如果旧目录为空，删除它
    OLD_COMPONENT_DIR="$PROJECT_ROOT/admin-frontend/src/views/$(dirname $OLD_COMPONENT)"
    if [ -d "$OLD_COMPONENT_DIR" ] && [ -z "$(ls -A "$OLD_COMPONENT_DIR")" ]; then
        rmdir "$OLD_COMPONENT_DIR"
        echo -e "${GREEN}✓ 已删除空目录: $OLD_COMPONENT_DIR${NC}"
    fi
else
    echo -e "${YELLOW}警告: 前端组件文件不存在: $OLD_COMPONENT_FILE${NC}"
fi

echo -e "${GREEN}✓ 迁移完成!${NC}"
echo -e "${YELLOW}注意:${NC}"
echo -e "  - 请检查 SQL 文件是否正确更新"
echo -e "  - 请检查前端组件文件是否正确移动"
echo -e "  - 如果组件中有硬编码的路径，请手动更新"
echo -e "  - 执行 SQL 文件以更新数据库中的菜单路径"

