# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# 複製依賴文件
COPY package*.json ./
RUN npm ci

# 複製源代碼
COPY . .

# 編譯生產版本
RUN npm run build

# Production stage
FROM nginx:alpine

# 複製編譯好的靜態文件
COPY --from=builder /app/dist /usr/share/nginx/html

# 複製自定義 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

