# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 复制go mod文件（go.sum可能不存在，使用通配符）
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建存储目录
RUN mkdir -p /root/storage/originals /root/storage/thumbnails /root/storage/temp

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]

