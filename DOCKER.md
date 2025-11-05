# Docker 部署指南

> 使用 Docker 一鍵啟動 WebRTC 直播系統，無需安裝 Go 和 Node.js

## 🎯 優勢

- ✅ **一鍵啟動**：無需配置開發環境
- ✅ **環境隔離**：不污染本機環境
- ✅ **一致性**：所有環境行為一致
- ✅ **易部署**：可直接部署到生產環境

## 📋 前置需求

### 安裝 Docker

**macOS**:
```bash
# 下載並安裝 Docker Desktop
# https://www.docker.com/products/docker-desktop

# 或使用 Homebrew
brew install --cask docker
```

**Linux**:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose

# 啟動 Docker
sudo systemctl start docker
sudo systemctl enable docker
```

**Windows**:
- 下載並安裝 [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)

### 驗證安裝

```bash
docker --version
# Docker version 24.0.0 或更高

docker-compose --version
# Docker Compose version v2.20.0 或更高
```

## 🚀 快速開始

### 1️⃣ 驗證配置（可選）

```bash
./docker-verify.sh
```

### 2️⃣ 啟動服務

```bash
./docker-start.sh
```

首次啟動會自動：
1. 構建後端 Docker 鏡像（~2-3 分鐘）
2. 構建前端 Docker 鏡像（~3-5 分鐘）
3. 啟動容器
4. 顯示訪問地址

### 3️⃣ 訪問應用

```
主播端: http://localhost:5173/broadcaster
觀眾端: http://localhost:5173/viewer
後端API: http://localhost:8080/health
```

### 4️⃣ 停止服務

```bash
./docker-stop.sh
```

## 📊 服務架構

```
┌─────────────────────────────────────┐
│         Docker Network              │
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │   Backend    │  │  Frontend   │ │
│  │   (Go)       │  │  (Nginx)    │ │
│  │   Port 8080  │  │  Port 80    │ │
│  └──────────────┘  └─────────────┘ │
│                                     │
└─────────────────────────────────────┘
         ↓                    ↓
    Host:8080           Host:5173
```

## 🛠️ 常用命令

### 查看日誌

```bash
# 所有服務
docker-compose logs -f

# 後端日誌
docker-compose logs -f backend

# 前端日誌
docker-compose logs -f frontend

# 最後 100 行
docker-compose logs --tail=100
```

### 查看狀態

```bash
# 容器狀態
docker-compose ps

# 資源使用
docker stats webrtc-backend webrtc-frontend
```

### 重啟服務

```bash
# 重啟所有服務
docker-compose restart

# 重啟特定服務
docker-compose restart backend
docker-compose restart frontend
```

### 進入容器

```bash
# 進入後端容器
docker exec -it webrtc-backend sh

# 進入前端容器
docker exec -it webrtc-frontend sh
```

### 清理資源

```bash
# 停止並刪除容器
docker-compose down

# 同時刪除鏡像
docker-compose down --rmi all

# 同時刪除 volumes
docker-compose down --volumes

# 完全清理
docker-compose down --rmi all --volumes
```

## 🔧 自定義配置

### 修改端口

編輯 `docker-compose.yml`:

```yaml
services:
  backend:
    ports:
      - "8080:8080"  # 改為 "9000:8080" 使用 9000 端口
  
  frontend:
    ports:
      - "5173:80"    # 改為 "8000:80" 使用 8000 端口
```

### 環境變數

編輯 `docker-compose.yml`:

```yaml
services:
  backend:
    environment:
      - PORT=8080
      - LOG_LEVEL=debug
      # 添加更多環境變數
```

### 資源限制

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

## 🐛 故障排除

### 1. 端口被佔用

```bash
# 檢查端口
lsof -i :8080
lsof -i :5173

# 停止佔用進程或修改 docker-compose.yml 端口
```

### 2. 容器無法啟動

```bash
# 查看詳細錯誤
docker-compose logs

# 清理並重新構建
docker-compose down --rmi all
./docker-start.sh
```

### 3. 前端無法連接後端

```bash
# 檢查網路
docker network ls
docker network inspect webrtc-demo_webrtc-network

# 檢查後端健康狀態
docker exec webrtc-backend wget -O- http://localhost:8080/health
```

### 4. 構建失敗

```bash
# 查看構建日誌
docker-compose build --no-cache --progress=plain

# 單獨構建服務
docker-compose build backend
docker-compose build frontend
```

### 5. Docker daemon 未運行

```bash
# macOS: 啟動 Docker Desktop
open -a Docker

# Linux: 啟動 Docker 服務
sudo systemctl start docker
```

## 📈 生產環境部署

### 1. 使用環境變數文件

創建 `.env`:

```env
BACKEND_PORT=8080
FRONTEND_PORT=5173
LOG_LEVEL=info
```

更新 `docker-compose.yml`:

```yaml
services:
  backend:
    ports:
      - "${BACKEND_PORT}:8080"
```

### 2. 使用 HTTPS

需要添加反向代理（Nginx/Traefik）和 SSL 證書。

### 3. 持久化數據

```yaml
services:
  backend:
    volumes:
      - backend-data:/app/data

volumes:
  backend-data:
```

### 4. 健康檢查

已內建健康檢查：

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 💡 效能優化

### 1. 多階段構建

已使用多階段構建減少鏡像大小：

- 後端：從 ~1GB 減少到 ~50MB
- 前端：從 ~500MB 減少到 ~30MB

### 2. 使用 .dockerignore

已配置忽略不必要的文件，加快構建速度。

### 3. 緩存優化

```bash
# 利用 Docker 緩存加速構建
docker-compose build

# 完全重新構建
docker-compose build --no-cache
```

## 📚 更多資源

- [Docker 官方文檔](https://docs.docker.com/)
- [Docker Compose 文檔](https://docs.docker.com/compose/)
- [最佳實踐](https://docs.docker.com/develop/dev-best-practices/)

---

**提示**：首次構建可能需要 5-10 分鐘，後續啟動只需幾秒鐘。

