#!/bin/bash

# go-zero Model 代码生成脚本
# 使用自定义模板从 SQL DDL 文件生成 Model 代码
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
ADMIN_SERVER_DIR="${PROJECT_ROOT}"
TEMPLATE_DIR="${ADMIN_SERVER_DIR}/.template"
MIGRATIONS_DIR="${ADMIN_SERVER_DIR}/db/migrations"
MODEL_DIR="${ADMIN_SERVER_DIR}/internal/model"

# 解析 goctl 路径（优先环境变量 GOCTL_BIN，其次 PATH，再尝试 GOPATH/bin/goctl）
GOCTL_BIN="${GOCTL_BIN:-}"
if [ -z "$GOCTL_BIN" ]; then
    GOCTL_BIN="$(command -v goctl 2>/dev/null || true)"
fi
if [ -z "$GOCTL_BIN" ]; then
    GOPATH_BIN="$(go env GOPATH 2>/dev/null)/bin/goctl"
    if [ -x "$GOPATH_BIN" ]; then
        GOCTL_BIN="$GOPATH_BIN"
    fi
fi
if [ -z "$GOCTL_BIN" ] || [ ! -x "$GOCTL_BIN" ]; then
    echo -e "${RED}错误: goctl 未安装或不可执行${NC}"
    echo "请运行: go install github.com/zeromicro/go-zero/tools/goctl@latest"
    echo "或设置环境变量 GOCTL_BIN 指向 goctl 可执行文件"
    exit 1
fi

# 检查模板目录是否存在
if [ ! -d "$TEMPLATE_DIR" ]; then
    echo -e "${RED}错误: 模板目录不存在: ${TEMPLATE_DIR}${NC}"
    echo "请先运行: cd ${ADMIN_SERVER_DIR} && goctl template init --home .template"
    exit 1
fi

# 显示使用说明
usage() {
    echo -e "${GREEN}go-zero Model 代码生成工具${NC}"
    echo ""
    echo "用法:"
    echo "  $0 <migration_file> [options]"
    echo ""
    echo "参数:"
    echo "  migration_file    迁移文件路径（相对于 db/migrations/ 或绝对路径）"
    echo ""
    echo "选项:"
    echo "  -c, --cache        启用缓存（默认: 启用）"
    echo "  -d, --dir DIR      输出目录（默认: internal/model）"
    echo "  -h, --help         显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 init.sql"
    echo "  $0 db/init.sql"
    echo "  $0 /absolute/path/to/init.sql"
    echo "  $0 init.sql -c"
    echo "  $0 init.sql -d internal/model/custom"
    echo ""
}

# 解析参数
MIGRATION_FILE=""
ENABLE_CACHE="-c"
OUTPUT_DIR="$MODEL_DIR"

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--cache)
            ENABLE_CACHE="-c"
            shift
            ;;
        -d|--dir)
            OUTPUT_DIR="$2"
            # 如果是相对路径，转换为绝对路径
            if [[ "$OUTPUT_DIR" != /* ]]; then
                OUTPUT_DIR="${ADMIN_SERVER_DIR}/${OUTPUT_DIR}"
            fi
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            if [ -z "$MIGRATION_FILE" ]; then
                MIGRATION_FILE="$1"
            else
                echo -e "${RED}错误: 未知参数: $1${NC}"
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

# 检查是否提供了迁移文件
if [ -z "$MIGRATION_FILE" ]; then
    echo -e "${RED}错误: 请指定迁移文件${NC}"
    usage
    exit 1
fi

# 解析迁移文件路径
if [[ "$MIGRATION_FILE" == /* ]]; then
    # 绝对路径
    SQL_FILE="$MIGRATION_FILE"
else
    # 相对路径，尝试多个位置
    if [ -f "${MIGRATIONS_DIR}/${MIGRATION_FILE}" ]; then
        SQL_FILE="${MIGRATIONS_DIR}/${MIGRATION_FILE}"
    elif [ -f "${ADMIN_SERVER_DIR}/${MIGRATION_FILE}" ]; then
        SQL_FILE="${ADMIN_SERVER_DIR}/${MIGRATION_FILE}"
    elif [ -f "${PROJECT_ROOT}/${MIGRATION_FILE}" ]; then
        SQL_FILE="${PROJECT_ROOT}/${MIGRATION_FILE}"
    elif [ -f "$MIGRATION_FILE" ]; then
        SQL_FILE="$MIGRATION_FILE"
    else
        echo -e "${RED}错误: 找不到文件: ${MIGRATION_FILE}${NC}"
        echo "尝试过的路径:"
        echo "  - ${MIGRATIONS_DIR}/${MIGRATION_FILE}"
        echo "  - ${ADMIN_SERVER_DIR}/${MIGRATION_FILE}"
        echo "  - ${PROJECT_ROOT}/${MIGRATION_FILE}"
        echo "  - ${MIGRATION_FILE} (当前目录)"
        exit 1
    fi
fi

# 检查 SQL 文件是否存在
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${RED}错误: SQL 文件不存在: ${SQL_FILE}${NC}"
    exit 1
fi

# 显示配置信息
echo -e "${GREEN}=== go-zero Model 代码生成 ===${NC}"
echo "项目根目录:  $PROJECT_ROOT"
echo "SQL 文件:    $SQL_FILE"
echo "输出目录:    $OUTPUT_DIR"
echo "模板目录:    $TEMPLATE_DIR"
echo "goctl 路径:  $GOCTL_BIN"
echo "缓存:        $([ -n "$ENABLE_CACHE" ] && echo "启用" || echo "禁用")"
echo ""

# 确认执行
read -p "确认生成 Model 代码? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 执行生成命令
echo -e "${GREEN}正在生成 Model 代码...${NC}"
cd "$ADMIN_SERVER_DIR"

"$GOCTL_BIN" model mysql ddl \
    -src "$SQL_FILE" \
    -dir "$OUTPUT_DIR" \
    $ENABLE_CACHE \
    -style gozero \
    --home "$TEMPLATE_DIR"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Model 代码生成成功!${NC}"
    echo -e "${GREEN}输出目录: ${OUTPUT_DIR}${NC}"
else
    echo -e "${RED}✗ Model 代码生成失败${NC}"
    exit 1
fi
