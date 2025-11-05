# 🚀 快速開始指南

## 選擇啟動方式

### 🎯 方式 1：自動選擇（最簡單）

```bash
./quick-start.sh
```

腳本會自動檢測你的環境並推薦最佳啟動方式。

---

### 🐋 方式 2：Docker（推薦）

**適合**：
- ✅ 快速測試和演示
- ✅ 不想安裝 Go/Node.js
- ✅ 確保環境一致性

**步驟**：

```bash
# 1. 驗證環境（可選）
./docker-verify.sh

# 2. 啟動
./docker-start.sh

# 3. 訪問
open http://localhost:5173/broadcaster

# 4. 停止
./docker-stop.sh
```

📖 詳細說明：[DOCKER.md](./DOCKER.md)

---

### 💻 方式 3：本地開發

**適合**：
- ✅ 需要調試代碼
- ✅ 需要即時熱更新
- ✅ 了解 Go 和 Node.js

**前置需求**：
- Go 1.21+
- Node.js 18+

**步驟**：

```bash
# 1. 安裝依賴（首次）
cd frontend && npm install && cd ..
cd backend && go mod download && cd ..

# 2. 啟動
./start.sh

# 3. 訪問
open http://localhost:5173/broadcaster

# 4. 停止
./stop.sh
```

---

## 📱 測試功能

### 本機測試

1. **主播端**：`http://localhost:5173/broadcaster`
   - 點擊「開始直播」
   - 授權攝像頭/麥克風

2. **觀眾端**：`http://localhost:5173/viewer`
   - 點擊「加入直播」
   - 觀看直播

### 局域網測試（多設備）

1. 查看本機 IP：
   ```bash
   # macOS
   ipconfig getifaddr en0
   
   # Linux
   hostname -I | awk '{print $1}'
   ```

2. 確保所有設備連接同一 WiFi

3. 用手機/平板訪問：
   ```
   http://你的IP:5173/broadcaster
   http://你的IP:5173/viewer
   ```

---

## 🎯 常見場景

### 場景 1：快速演示給別人看

```bash
./docker-start.sh
# 分享: http://你的IP:5173/viewer
```

### 場景 2：開發新功能

```bash
./start.sh
# 修改代碼，自動熱更新
```

### 場景 3：測試多人觀看

```bash
# 開啟多個瀏覽器視窗
# 1 個主播 + N 個觀眾
```

---

## 🛠️ 故障排除

### 無法啟動？

```bash
# 檢查端口佔用
lsof -i :8080
lsof -i :5173

# 殺掉佔用進程
kill -9 <PID>
```

### 看不到畫面？

1. F12 打開開發者工具查看錯誤
2. 重新整理頁面
3. 重啟服務

### Docker 構建慢？

首次構建需要下載依賴，需要 5-10 分鐘。
後續啟動只需幾秒鐘。

---

## 📚 更多資源

- **主文檔**：[README.md](./README.md)
- **Docker 指南**：[DOCKER.md](./DOCKER.md)
- **技術架構**：見 README 中的流程圖

---

## 💡 小提示

- 使用 Chrome 瀏覽器測試效果最佳
- 開發時使用本地方式（熱更新）
- 演示時使用 Docker 方式（穩定）
- 記得允許瀏覽器訪問攝像頭/麥克風

---

**祝你使用愉快！** 🎉

