#!/bin/bash

set -e

echo "=========================================="
echo "🚀 WebRTC 直播系統 - 快速啟動"
echo "=========================================="
echo ""

# 檢測 Docker 是否可用
HAS_DOCKER=false
if command -v docker &> /dev/null && command -v docker-compose &> /dev/null; then
    if docker info > /dev/null 2>&1; then
        HAS_DOCKER=true
    fi
fi

# 檢測本地開發環境
HAS_GO=false
HAS_NODE=false
if command -v go &> /dev/null; then
    HAS_GO=true
fi
if command -v node &> /dev/null && command -v npm &> /dev/null; then
    HAS_NODE=true
fi

echo "📊 環境檢測："
echo "  Docker:   $([ "$HAS_DOCKER" = true ] && echo "✅" || echo "❌")"
echo "  Go:       $([ "$HAS_GO" = true ] && echo "✅" || echo "❌")"
echo "  Node.js:  $([ "$HAS_NODE" = true ] && echo "✅" || echo "❌")"
echo ""

# 選擇啟動方式
if [ "$HAS_DOCKER" = true ]; then
    echo "✨ 推薦使用 Docker 方式啟動（無需安裝 Go/Node.js）"
    echo ""
    read -p "選擇啟動方式 [1: Docker, 2: 本地開發]: " choice
    
    if [ "$choice" = "1" ] || [ -z "$choice" ]; then
        echo ""
        echo "🐋 使用 Docker 啟動..."
        ./docker-start.sh
    elif [ "$choice" = "2" ]; then
        if [ "$HAS_GO" = true ] && [ "$HAS_NODE" = true ]; then
            echo ""
            echo "💻 使用本地開發方式啟動..."
            ./start.sh
        else
            echo ""
            echo "❌ 本地開發環境不完整"
            echo "   需要安裝 Go 1.21+ 和 Node.js 18+"
            exit 1
        fi
    else
        echo "❌ 無效選擇"
        exit 1
    fi
elif [ "$HAS_GO" = true ] && [ "$HAS_NODE" = true ]; then
    echo "💻 使用本地開發方式啟動..."
    ./start.sh
else
    echo "❌ 錯誤：無可用的啟動方式"
    echo ""
    echo "請選擇以下方式之一："
    echo "  1. 安裝 Docker: https://docs.docker.com/get-docker/"
    echo "  2. 安裝 Go (1.21+) 和 Node.js (18+)"
    exit 1
fi

