# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o insight-hub ./cmd/insight-hub

# Stage 2: Runtime
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 从构建阶段复制二进制文件
COPY --from=builder /build/insight-hub .
COPY --from=builder /build/config.postgres.yaml ./config.yaml

# 创建数据目录
RUN mkdir -p /data

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/v1/health || exit 1

# 运行
ENTRYPOINT ["./insight-hub"]
CMD ["-config", "/app/config.yaml"]
