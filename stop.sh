#!/bin/bash

# WebRTC 直播系統停止腳本

echo "正在停止 WebRTC 直播系統..."

# 停止 Go 後端
pkill -f "go run main.go" 2>/dev/null
pkill -f "backend/main.go" 2>/dev/null

# 停止前端
pkill -f "vite" 2>/dev/null
pkill -f "npm run dev" 2>/dev/null

# 停止佔用端口的進程（備用方案）
lsof -ti:8080 | xargs kill -9 2>/dev/null  # 後端
lsof -ti:5173 | xargs kill -9 2>/dev/null  # 前端

echo "✓ 所有服務已停止"

