#!/bin/bash

set -e

echo "=========================================="
echo "🚀 WebRTC 直播系統 - Docker 啟動"
echo "=========================================="
echo ""

# 檢查 Docker 是否安裝
if ! command -v docker &> /dev/null; then
    echo "❌ 錯誤：未找到 Docker"
    echo "   請先安裝 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ 錯誤：未找到 docker-compose"
    echo "   請先安裝 docker-compose"
    exit 1
fi

# 檢查 Docker daemon 是否運行
if ! docker info > /dev/null 2>&1; then
    echo "❌ 錯誤：Docker daemon 未運行"
    echo "   請先啟動 Docker Desktop 或 Docker 服務"
    exit 1
fi

echo "✅ Docker 環境檢查通過"
echo ""

# 停止舊容器（如果存在）
echo "🧹 清理舊容器..."
docker-compose down 2>/dev/null || true
echo ""

# 構建鏡像
echo "🔨 構建 Docker 鏡像..."
docker-compose build --no-cache
echo ""

# 啟動服務
echo "🚀 啟動服務..."
docker-compose up -d
echo ""

# 等待服務啟動
echo "⏳ 等待服務啟動..."
sleep 5

# 檢查服務狀態
echo "📊 檢查服務狀態..."
docker-compose ps
echo ""

# 顯示日誌
echo "📋 服務日誌（最後 20 行）："
docker-compose logs --tail=20
echo ""

# 獲取本機 IP
LOCAL_IP=$(ipconfig getifaddr en0 2>/dev/null || echo "無法獲取")

echo "=========================================="
echo "✅ 服務啟動成功！"
echo "=========================================="
echo ""
echo "📱 訪問地址："
echo ""
echo "  本機訪問："
echo "    主播端: http://localhost:5173/broadcaster"
echo "    觀眾端: http://localhost:5173/viewer"
echo "    後端API: http://localhost:8080/health"
echo ""
if [ "$LOCAL_IP" != "無法獲取" ]; then
    echo "  局域網訪問："
    echo "    主播端: http://$LOCAL_IP:5173/broadcaster"
    echo "    觀眾端: http://$LOCAL_IP:5173/viewer"
    echo ""
fi
echo "=========================================="
echo ""
echo "💡 常用命令："
echo "  查看日誌:   docker-compose logs -f"
echo "  停止服務:   docker-compose down"
echo "  重啟服務:   docker-compose restart"
echo "  查看狀態:   docker-compose ps"
echo ""

