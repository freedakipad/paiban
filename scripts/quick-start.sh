#!/bin/bash
# PaiBan 快速启动脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "   🗓️  PaiBan 排班引擎 - 快速启动"
echo "=========================================="
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ 未安装 Go${NC}"
    echo "  请安装 Go 1.23+: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
echo -e "${GREEN}✓ Go $GO_VERSION${NC}"

# 编译
echo "编译中..."
mkdir -p bin
go build -o bin/paiban cmd/server/main.go
echo -e "${GREEN}✓ 编译成功${NC}"

# 检查端口
if lsof -i :7012 &> /dev/null; then
    echo -e "${YELLOW}! 端口 7012 已被占用，尝试停止旧进程...${NC}"
    pkill -f "bin/paiban" 2>/dev/null || true
    sleep 2
fi

# 启动
echo "启动服务..."
./bin/paiban &
PID=$!
sleep 2

# 验证
if curl -s http://localhost:7012/health > /dev/null; then
    echo ""
    echo -e "${GREEN}✓ 服务已启动 (PID: $PID)${NC}"
    echo ""
    echo "=========================================="
    echo "  服务地址:"
    echo "    API:      http://localhost:7012"
    echo "    健康检查: http://localhost:7012/health"
    echo ""
    echo "  前端控制台:"
    echo "    cd frontend && python3 -m http.server 8888"
    echo "    访问: http://localhost:8888"
    echo ""
    echo "  停止服务:"
    echo "    pkill -f 'bin/paiban'"
    echo "=========================================="
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    exit 1
fi
