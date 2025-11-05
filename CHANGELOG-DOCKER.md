# Docker 配置更新日誌

## 🎉 新增功能

### Docker 完整支援

現在整個專案可以通過 Docker 一鍵啟動，無需安裝 Go 和 Node.js！

## 📁 新增文件清單

### 1. Docker 配置文件

| 文件 | 說明 |
|------|------|
| `docker-compose.yml` | Docker 容器編排配置 |
| `backend/Dockerfile` | 後端 Docker 鏡像構建文件 |
| `backend/.dockerignore` | 後端 Docker 忽略文件 |
| `frontend/Dockerfile` | 前端 Docker 鏡像構建文件 |
| `frontend/nginx.conf` | 前端 Nginx 配置 |
| `frontend/.dockerignore` | 前端 Docker 忽略文件 |

### 2. 啟動腳本

| 文件 | 說明 |
|------|------|
| `docker-start.sh` | Docker 一鍵啟動腳本 |
| `docker-stop.sh` | Docker 停止腳本 |
| `docker-verify.sh` | Docker 環境驗證腳本 |
| `quick-start.sh` | 智能選擇啟動方式 |

### 3. 文檔

| 文件 | 說明 |
|------|------|
| `DOCKER.md` | Docker 詳細使用指南 |
| `QUICKSTART.md` | 快速開始指南 |
| `README.md` | 更新了 Docker 相關說明 |

## 🚀 快速使用

### 方式 1：自動選擇（最簡單）

```bash
./quick-start.sh
```

### 方式 2：Docker 啟動（推薦）

```bash
./docker-start.sh
```

### 方式 3：本地開發

```bash
./start.sh
```

## 🏗️ Docker 架構

### 服務構成

```yaml
services:
  backend:              # Go 後端（Signaling + SFU）
    - Port: 8080
    - Image: ~50MB
    - Health Check: ✅
    
  frontend:             # Vue 3 前端（Nginx）
    - Port: 5173 (映射到 80)
    - Image: ~30MB
    - Health Check: ✅
```

### 網路配置

```
webrtc-network (bridge)
├── backend (webrtc-backend)
└── frontend (webrtc-frontend)
```

## 📊 性能優化

### 鏡像大小優化

| 服務 | 優化前 | 優化後 | 方法 |
|------|--------|--------|------|
| Backend | ~1GB | ~50MB | 多階段構建 + Alpine |
| Frontend | ~500MB | ~30MB | 多階段構建 + Nginx |

### 構建優化

- ✅ 使用 `.dockerignore` 減少構建上下文
- ✅ 利用 Docker 層緩存
- ✅ 分離依賴安裝和代碼複製

## 🛠️ 常用命令速查

```bash
# 啟動
./docker-start.sh

# 停止
./docker-stop.sh

# 查看日誌
docker-compose logs -f

# 查看狀態
docker-compose ps

# 重啟
docker-compose restart

# 重新構建
docker-compose build --no-cache

# 完全清理
docker-compose down --rmi all --volumes

# 進入容器
docker exec -it webrtc-backend sh
```

## 🎯 驗收測試

### ✅ 已測試場景

- [x] 本機啟動（localhost）
- [x] 局域網訪問（多設備）
- [x] Docker 容器編排
- [x] 健康檢查
- [x] 日誌輸出
- [x] 容器重啟
- [x] 資源清理

### ⏳ 待測試場景

- [ ] 生產環境部署
- [ ] HTTPS 配置
- [ ] 負載測試
- [ ] 多節點部署

## 📝 技術細節

### 後端 Dockerfile

```dockerfile
# 多階段構建
FROM golang:1.21-alpine AS builder
# ... 構建階段

FROM alpine:latest
# ... 運行階段
```

**特點**：
- 使用 Alpine Linux（體積小）
- 多階段構建（只保留編譯產物）
- 無需 Go 運行時

### 前端 Dockerfile

```dockerfile
# 多階段構建
FROM node:18-alpine AS builder
# ... 構建階段

FROM nginx:alpine
# ... 運行階段
```

**特點**：
- 使用 Nginx 服務靜態文件
- SPA 路由支援
- Gzip 壓縮
- 資源緩存優化

## 🌐 訪問地址

### 本機訪問

```
主播端: http://localhost:5173/broadcaster
觀眾端: http://localhost:5173/viewer
後端API: http://localhost:8080/health
```

### 局域網訪問

```
主播端: http://<你的IP>:5173/broadcaster
觀眾端: http://<你的IP>:5173/viewer
```

查詢 IP：`ipconfig getifaddr en0`（macOS）

## 🔧 自定義配置

### 修改端口

編輯 `docker-compose.yml`:

```yaml
services:
  backend:
    ports:
      - "9000:8080"  # 使用 9000 端口
  
  frontend:
    ports:
      - "8000:80"    # 使用 8000 端口
```

### 添加環境變數

編輯 `docker-compose.yml`:

```yaml
services:
  backend:
    environment:
      - PORT=8080
      - LOG_LEVEL=debug
      - CUSTOM_VAR=value
```

## 🎉 總結

現在你可以：

1. ✅ **一鍵啟動**：無需配置環境
2. ✅ **環境一致**：Docker 容器保證一致性
3. ✅ **快速部署**：適合演示和測試
4. ✅ **易於維護**：標準化容器管理

---

**版本**：Docker v1.0  
**更新日期**：2025-11-06  
**狀態**：✅ 已完成並測試

