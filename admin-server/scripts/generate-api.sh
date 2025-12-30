#!/bin/bash

# go-zero API Handler 代码生成脚本
# 从 .api 文件生成 Handler 代码骨架
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
API_DIR="${ADMIN_SERVER_DIR}/api"

# 解析 goctl 路径（优先环境变量 GOCTL_BIN，其次 PATH，再次 GOPATH/bin/goctl）
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
    echo "或手动设置环境变量 GOCTL_BIN 指向 goctl 可执行文件"
    exit 1
fi

# 显示使用说明
usage() {
    echo -e "${GREEN}go-zero API Handler 代码生成工具${NC}"
    echo ""
    echo "用法:"
    echo "  $0 <api_file>"
    echo ""
    echo "参数:"
    echo "  api_file          API 文件路径（相对于 api/ 或绝对路径）"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 user.api"
    echo "  $0 api/user.api"
    echo "  $0 /absolute/path/to/user.api"
    echo ""
    echo "注意:"
    echo "  - 生成的 Handler 代码在 internal/handler/ 目录"
    echo "  - Types 定义会生成临时文件，需要手动合并到 internal/types/types.go"
    echo ""
}

# 解析参数
API_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        *)
            if [ -z "$API_FILE" ]; then
                API_FILE="$1"
            else
                echo -e "${RED}错误: 未知参数: $1${NC}"
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

# 检查是否提供了 API 文件
if [ -z "$API_FILE" ]; then
    echo -e "${RED}错误: 请指定 API 文件${NC}"
    usage
    exit 1
fi

# 解析 API 文件路径
if [[ "$API_FILE" == /* ]]; then
    # 绝对路径
    API_FILE_PATH="$API_FILE"
else
    # 相对路径，尝试多个位置
    if [ -f "${API_DIR}/${API_FILE}" ]; then
        API_FILE_PATH="${API_DIR}/${API_FILE}"
    elif [ -f "${ADMIN_SERVER_DIR}/${API_FILE}" ]; then
        API_FILE_PATH="${ADMIN_SERVER_DIR}/${API_FILE}"
    elif [ -f "${PROJECT_ROOT}/${API_FILE}" ]; then
        API_FILE_PATH="${PROJECT_ROOT}/${API_FILE}"
    elif [ -f "$API_FILE" ]; then
        API_FILE_PATH="$API_FILE"
    else
        echo -e "${RED}错误: 找不到文件: ${API_FILE}${NC}"
        echo "尝试过的路径:"
        echo "  - ${API_DIR}/${API_FILE}"
        echo "  - ${ADMIN_SERVER_DIR}/${API_FILE}"
        echo "  - ${PROJECT_ROOT}/${API_FILE}"
        echo "  - ${API_FILE} (当前目录)"
        exit 1
    fi
fi

# 检查 API 文件是否存在
if [ ! -f "$API_FILE_PATH" ]; then
    echo -e "${RED}错误: API 文件不存在: ${API_FILE_PATH}${NC}"
    exit 1
fi

# 显示配置信息
echo -e "${GREEN}=== go-zero API Handler 代码生成 ===${NC}"
echo "项目根目录:  $PROJECT_ROOT"
echo "API 文件:    $API_FILE_PATH"
echo "输出目录:    ${ADMIN_SERVER_DIR}"
echo "goctl 路径: $GOCTL_BIN"
echo ""

# 确认执行
read -p "确认生成 Handler 代码? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 执行生成命令
echo -e "${GREEN}正在生成 Handler 代码...${NC}"
cd "$ADMIN_SERVER_DIR"

"$GOCTL_BIN" api go -api "$API_FILE_PATH" -dir .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Handler 代码生成成功!${NC}"
    echo -e "${YELLOW}注意: 请检查生成的 types 文件，按需合并到 internal/types/types.go${NC}"
else
    echo -e "${RED}✗ Handler 代码生成失败${NC}"
    exit 1
fi
