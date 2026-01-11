# PaiBan 排班引擎服务 Dockerfile
# 多阶段构建，优化镜像大小
# 最终镜像约 20MB

# ================================
# Stage 1: 依赖缓存阶段
# ================================
FROM golang:1.21-alpine AS deps

WORKDIR /app

# 只复制依赖文件以利用Docker缓存
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ================================
# Stage 2: 构建阶段
# ================================
FROM golang:1.21-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata upx

WORKDIR /app

# 从依赖阶段复制已下载的模块
COPY --from=deps /go/pkg/mod /go/pkg/mod

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 构建二进制文件（静态链接、优化大小）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -trimpath \
    -o /app/paiban \
    ./cmd/server

# 使用UPX压缩二进制文件（可选，进一步减小体积）
RUN upx --best --lzma /app/paiban || true

# ================================
# Stage 3: 运行阶段（使用distroless更安全）
# ================================
FROM gcr.io/distroless/static-debian12:nonroot AS production

# 复制时区信息
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Shanghai

# 工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/paiban /app/paiban

# 复制配置文件
COPY configs/ /app/configs/

# 暴露端口
EXPOSE 7012

# 使用非root用户（distroless:nonroot 默认UID 65532）
USER nonroot:nonroot

# 启动命令
ENTRYPOINT ["/app/paiban"]

# ================================
# Stage 3-alt: Alpine运行阶段（备选，包含shell便于调试）
# ================================
FROM alpine:3.19 AS alpine-runtime

# 安装运行时依赖（最小化）
RUN apk add --no-cache ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 paiban && \
    adduser -u 1000 -G paiban -s /sbin/nologin -D paiban

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/paiban /app/paiban

# 复制配置文件
COPY configs/ /app/configs/

# 设置权限
RUN chown -R paiban:paiban /app && \
    chmod +x /app/paiban

# 切换到非 root 用户
USER paiban

# 暴露端口
EXPOSE 7012

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7012/health || exit 1

# 启动命令
ENTRYPOINT ["/app/paiban"]

# 默认使用distroless生产镜像
FROM production

