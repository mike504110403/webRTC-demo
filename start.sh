#!/bin/bash

# WebRTC 直播系統啟動腳本

set -e  # 遇到錯誤立即退出

echo "========================================"
echo "  WebRTC 直播系統 - 啟動腳本"
echo "========================================"
echo ""

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 檢查是否在專案根目錄
if [ ! -f "README.md" ]; then
    echo -e "${RED}錯誤：請在專案根目錄執行此腳本${NC}"
    exit 1
fi

# 1. 檢查前端依賴
echo -e "${YELLOW}[1/4] 檢查前端依賴...${NC}"
if [ ! -d "frontend/node_modules" ]; then
    echo "   前端依賴未安裝，正在安裝..."
    cd frontend
    npm install
    cd ..
    echo -e "   ${GREEN}✓ 前端依賴安裝完成${NC}"
else
    echo -e "   ${GREEN}✓ 前端依賴已存在${NC}"
fi

# 2. 檢查後端依賴
echo ""
echo -e "${YELLOW}[2/4] 檢查後端依賴...${NC}"
cd backend
if ! go mod download 2>/dev/null; then
    echo -e "   ${RED}✗ 後端依賴下載失敗${NC}"
    exit 1
fi
echo -e "   ${GREEN}✓ 後端依賴已就緒${NC}"
cd ..

# 3. 啟動後端
echo ""
echo -e "${YELLOW}[3/4] 啟動後端服務...${NC}"
cd backend
echo "   啟動 Signaling Server + 內建 SFU..."
go run main.go &
BACKEND_PID=$!
cd ..

# 等待後端啟動
echo "   等待後端啟動..."
sleep 3

# 檢查後端是否啟動成功
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo -e "   ${RED}✗ 後端啟動失敗${NC}"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi
echo -e "   ${GREEN}✓ 後端服務已啟動 (PID: $BACKEND_PID)${NC}"
echo "   訪問: http://localhost:8080/health"

# 4. 啟動前端
echo ""
echo -e "${YELLOW}[4/4] 啟動前端服務...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

# 等待前端啟動
echo "   等待前端啟動..."
sleep 3

echo ""
echo -e "${GREEN}========================================"
echo "  ✓ 所有服務已啟動！"
echo "========================================${NC}"
echo ""
echo "服務地址："
echo "  • 前端:     http://localhost:5173"
echo "  • 主播端:   http://localhost:5173/broadcaster"
echo "  • 觀眾端:   http://localhost:5173/viewer"
echo "  • 後端API:  http://localhost:8080"
echo ""
echo "進程 ID："
echo "  • 後端 PID: $BACKEND_PID"
echo "  • 前端 PID: $FRONTEND_PID"
echo ""
echo -e "${YELLOW}按 Ctrl+C 停止所有服務${NC}"
echo ""

# 捕獲 Ctrl+C 信號
trap "echo ''; echo '正在停止服務...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; echo '已停止'; exit" INT TERM

# 保持腳本運行
wait

