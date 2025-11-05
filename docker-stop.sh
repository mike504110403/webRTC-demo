#!/bin/bash

set -e

echo "=========================================="
echo "🛑 停止 WebRTC 直播系統"
echo "=========================================="
echo ""

# 停止容器
echo "🛑 停止容器..."
docker-compose down

echo ""
echo "✅ 服務已停止"
echo ""
echo "💡 如需清理所有資源（包括鏡像和 volumes）："
echo "   docker-compose down --rmi all --volumes"
echo ""

