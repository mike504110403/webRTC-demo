#!/bin/bash

set -e

echo "=========================================="
echo "🔍 驗證 Docker 配置"
echo "=========================================="
echo ""

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 檢查函數
check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✅ $1 已安裝${NC}"
        return 0
    else
        echo -e "${RED}❌ $1 未安裝${NC}"
        return 1
    fi
}

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✅ 檔案存在: $1${NC}"
        return 0
    else
        echo -e "${RED}❌ 檔案不存在: $1${NC}"
        return 1
    fi
}

# 1. 檢查必要工具
echo "📦 檢查必要工具..."
check_command docker || exit 1
check_command docker-compose || exit 1
echo ""

# 2. 檢查 Docker daemon
echo "🐋 檢查 Docker daemon..."
if docker info > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker daemon 運行中${NC}"
else
    echo -e "${RED}❌ Docker daemon 未運行${NC}"
    exit 1
fi
echo ""

# 3. 檢查必要文件
echo "📄 檢查必要文件..."
check_file "docker-compose.yml" || exit 1
check_file "backend/Dockerfile" || exit 1
check_file "backend/main.go" || exit 1
check_file "backend/go.mod" || exit 1
check_file "frontend/Dockerfile" || exit 1
check_file "frontend/nginx.conf" || exit 1
check_file "frontend/package.json" || exit 1
echo ""

# 4. 驗證 docker-compose.yml 語法
echo "🔧 驗證 docker-compose.yml..."
if docker-compose config > /dev/null 2>&1; then
    echo -e "${GREEN}✅ docker-compose.yml 語法正確${NC}"
else
    echo -e "${RED}❌ docker-compose.yml 語法錯誤${NC}"
    docker-compose config
    exit 1
fi
echo ""

# 5. 檢查端口佔用
echo "🔌 檢查端口..."
check_port() {
    if lsof -i :$1 > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  端口 $1 已被佔用${NC}"
        lsof -i :$1
        return 1
    else
        echo -e "${GREEN}✅ 端口 $1 可用${NC}"
        return 0
    fi
}

check_port 8080
check_port 5173
echo ""

# 6. 測試構建（可選）
echo "🏗️  測試構建（可選，按 Ctrl+C 跳過）..."
read -t 10 -p "是否測試構建 Docker 鏡像？(y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "開始構建..."
    docker-compose build || exit 1
    echo -e "${GREEN}✅ 構建成功${NC}"
fi
echo ""

echo "=========================================="
echo -e "${GREEN}✅ 驗證完成！${NC}"
echo "=========================================="
echo ""
echo "💡 下一步："
echo "   ./docker-start.sh    # 啟動服務"
echo ""

