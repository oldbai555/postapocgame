#!/bin/bash

# go-zero TypeScript 代码生成脚本
# 从 .api 文件生成前端 TypeScript 代码
# 支持在任何目录下运行，自动定位项目目录

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# admin-server 目录（scripts的父目录）
ADMIN_SERVER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# 项目根目录（admin-server的父目录）
PROJECT_ROOT="$(cd "$ADMIN_SERVER_DIR/.." && pwd)"
ADMIN_FRONTEND_DIR="${PROJECT_ROOT}/admin-frontend"
API_DIR="${ADMIN_SERVER_DIR}/api"
GENERATED_DIR="${ADMIN_FRONTEND_DIR}/src/api/generated"

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
    echo -e "${GREEN}go-zero TypeScript 代码生成工具${NC}"
    echo ""
    echo "用法:"
    echo "  $0 [api_file]"
    echo ""
    echo "参数:"
    echo "  api_file          API 文件路径（可选，默认使用 admin-server/api/admin.api）"
    echo "                    相对于 api/ 或绝对路径"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                # 使用默认 admin.api"
    echo "  $0 admin.api      # 指定 API 文件"
    echo "  $0 api/admin.api  # 相对路径"
    echo "  $0 /absolute/path/to/admin.api  # 绝对路径"
    echo ""
    echo "注意:"
    echo "  - 生成的 TypeScript 代码在 admin-frontend/src/api/generated/ 目录"
    echo "  - 禁止手动修改 generated/ 目录下的文件"
    echo "  - 在 src/api/ 中二次封装（错误处理、拦截器集成、统一返回类型）"
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

# 如果没有指定 API 文件，使用默认的 admin.api
if [ -z "$API_FILE" ]; then
    API_FILE="admin.api"
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

# 检查前端目录是否存在
if [ ! -d "$ADMIN_FRONTEND_DIR" ]; then
    echo -e "${RED}错误: 前端目录不存在: ${ADMIN_FRONTEND_DIR}${NC}"
    exit 1
fi

# 显示配置信息
echo -e "${GREEN}=== go-zero TypeScript 代码生成 ===${NC}"
echo "项目根目录:  $PROJECT_ROOT"
echo "API 文件:    $API_FILE_PATH"
echo "输出目录:    ${GENERATED_DIR}"
echo "goctl 路径: $GOCTL_BIN"
echo ""

# 确认执行
read -p "确认生成 TypeScript 代码? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

# 确保输出目录存在
mkdir -p "$GENERATED_DIR"

# 执行生成命令
echo -e "${GREEN}正在生成 TypeScript 代码...${NC}"
cd "$ADMIN_SERVER_DIR"

"$GOCTL_BIN" api ts -api "$API_FILE_PATH" -dir "$GENERATED_DIR"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ TypeScript 代码生成成功!${NC}"
    echo -e "${YELLOW}注意:${NC}"
    echo -e "  - 生成的代码在: ${GENERATED_DIR}"
    echo -e "  - 禁止手动修改 generated/ 目录下的文件"
    echo -e "  - 在 src/api/ 中二次封装（错误处理、拦截器集成、统一返回类型）"
    echo -e "  - 如果路径包含 /auth 前缀，需要在封装时修正路径（去掉 /auth）"
else
    echo -e "${RED}✗ TypeScript 代码生成失败${NC}"
    exit 1
fi

