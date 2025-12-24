#!/bin/bash

# SQL 脚本生成工具
# 用于快速生成新功能模块的初始化 SQL 脚本
# 支持在任何目录下运行，自动定位项目目录

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
SQLGEN_DIR="${PROJECT_ROOT}/scripts/sqlgen"
OUTPUT_DIR="${PROJECT_ROOT}/admin-server/db"

# 显示使用说明
usage() {
    echo -e "${GREEN}SQL 脚本生成工具${NC}"
    echo ""
    echo "用法:"
    echo "  $0 -group <group> -name <name>"
    echo ""
    echo "参数:"
    echo "  -group <group>    功能组名（必需，如 user, file）"
    echo "  -name <name>      功能名称（必需，如 用户管理, 文件管理）"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -group user -name 用户管理"
    echo "  $0 -group file -name 文件管理"
    echo ""
    echo "注意:"
    echo "  - 生成的 SQL 文件在 admin-server/db/ 目录下"
    echo "  - 文件名格式: init_<group>.sql"
    echo "  - 主键为自增，不需要手动赋值"
    echo "  - 菜单默认归类在临时目录下"
    echo "  - 包含菜单、权限、接口及关联关系"
    echo ""
}

# 检查参数
if [ $# -eq 0 ]; then
    usage
    exit 0
fi

GROUP=""
NAME=""

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -group)
            GROUP="$2"
            shift 2
            ;;
        -name)
            NAME="$2"
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
if [ -z "$GROUP" ] || [ -z "$NAME" ]; then
    echo -e "${RED}错误: 必须提供 -group 和 -name 参数${NC}"
    usage
    exit 1
fi

# 检查 sqlgen 目录是否存在
if [ ! -d "$SQLGEN_DIR" ]; then
    echo -e "${RED}错误: sqlgen 目录不存在: ${SQLGEN_DIR}${NC}"
    exit 1
fi

# 检查模板文件是否存在
TEMPLATE_FILE="${SQLGEN_DIR}/templates/init_module.sql.tpl"
if [ ! -f "$TEMPLATE_FILE" ]; then
    echo -e "${RED}错误: 模板文件不存在: ${TEMPLATE_FILE}${NC}"
    exit 1
fi

# 显示配置信息
echo -e "${GREEN}=== SQL 脚本生成工具 ===${NC}"
echo "项目根目录:  $PROJECT_ROOT"
echo "功能组名:    $GROUP"
echo "功能名称:    $NAME"
echo "输出目录:    ${OUTPUT_DIR}"
echo "输出文件:    ${OUTPUT_DIR}/init_${GROUP}.sql"
echo ""

# 确认执行
read -p "确认生成 SQL 脚本? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 确保输出目录存在
mkdir -p "$OUTPUT_DIR"

# 编译并运行 Go 程序
echo -e "${GREEN}正在生成 SQL 脚本...${NC}"
cd "$SQLGEN_DIR"

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go 未安装${NC}"
    echo "请安装 Go: https://golang.org/dl/"
    exit 1
fi

# 编译 Go 程序
go build -o sqlgen main.go

# 运行程序
./sqlgen -group "$GROUP" -name "$NAME" -output "$OUTPUT_DIR" -template "${SQLGEN_DIR}/templates"

# 清理编译产物
rm -f sqlgen

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ SQL 脚本生成成功!${NC}"
    echo -e "${YELLOW}注意:${NC}"
    echo -e "  - 生成的 SQL 文件: ${OUTPUT_DIR}/init_${GROUP}.sql"
    echo -e "  - 请在数据库中执行该 SQL 文件"
    echo -e "  - 菜单默认归类在临时目录下，可在菜单管理中调整"
    echo -e "  - 可根据需要修改菜单、按钮、接口的启用状态"
else
    echo -e "${RED}✗ SQL 脚本生成失败${NC}"
    exit 1
fi

