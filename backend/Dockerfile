# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 複製依賴文件
COPY go.mod go.sum ./
RUN go mod download

# 複製源代碼
COPY . .

# 編譯
RUN go build -o bin/server .

# Run stage
FROM alpine:latest

WORKDIR /app

# 從 builder 複製編譯好的二進制
COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]

