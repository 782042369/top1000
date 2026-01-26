# ============================================
# 极简版 Dockerfile - Scratch 基础镜像
# 目标镜像大小：4-5MB
# ============================================
# 优化措施：
# 1. 前端：仅导入实际使用的 AG Grid 模块（减少16.5%）
# 2. Go 二进制：UPX 压缩（减少60%）
# 3. 基础镜像：Scratch 空镜像
# ============================================

# 阶段一：构建 service (Go版本)
FROM golang:1.25-alpine AS service-builder
WORKDIR /app

LABEL stage="service-builder"

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN echo "📦 下载 Go 依赖..." && \
    go mod download && \
    echo "✅ 验证依赖完成" && \
    go mod verify

# 复制源代码（从  目录）
COPY cmd ./cmd
COPY internal ./internal

# 安装 UPX 压缩工具
RUN echo "🔧 安装 UPX 压缩工具..." && \
    apk add --no-cache upx

# 构建完全静态的Go应用（极致优化）
RUN echo "🔨 开始构建 Go 应用..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static' -buildid=" \
    -trimpath \
    -o main ./cmd/top1000 && \
    chmod +x main && \
    echo "✅ Go 应用构建完成"

# 使用 UPX 压缩二进制文件（减少50-70%体积）
RUN echo "🗜️  使用 UPX 压缩二进制文件..." && \
    upx --best --lzma main && \
    echo "✅ UPX 压缩完成"

# 阶段二：构建 web
FROM node:24-alpine AS web-builder
WORKDIR /app

LABEL stage="web-builder"

# 安装 pnpm
RUN echo "📦 安装 pnpm..." && \
    npm install -g pnpm@10 && \
    echo "✅ pnpm 安装完成"

# 优先复制包管理文件以利用构建缓存
COPY web/package.json web/pnpm-lock.yaml ./web/

# 安装依赖
RUN echo "📦 安装前端依赖..." && \
    cd web && \
    pnpm install --frozen-lockfile --prod=false && \
    echo "✅ 前端依赖安装完成"

# 复制源代码
COPY web ./web/

# 执行构建，输出到 web/dist 目录
RUN echo "🔨 开始构建前端..." && \
    cd web && \
    pnpm build && \
    echo "✅ 前端构建完成" && \
    echo "📁 构建产物位置: web/dist"

# 阶段三：准备 CA 证书（从 Alpine 提取）
FROM alpine:3.19 AS certs
RUN echo "🔒 准备 CA 证书..." && \
    apk --no-cache add ca-certificates && \
    echo "✅ CA 证书准备完成"

# ============================================
# 最终生产阶段：使用 Scratch（空镜像）
# ============================================
FROM scratch
WORKDIR /app

LABEL stage="production"

# 从 certs 阶段复制 CA 证书（HTTPS 必需）
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从 service-builder 阶段复制 Go 二进制
COPY --from=service-builder /app/main ./main

# 从 web-builder 阶段复制前端文件（注意路径：web/dist）
COPY --from=web-builder /app/web/dist ./web-dist

# 设置环境变量（时区默认为中国）
ENV PORT=7066
ENV TZ=Asia/Shanghai

# 声明端口
EXPOSE 7066

# ============================================
# 注意：Scratch 镜像不包含 shell，因此：
# - 无法使用 HEALTHCHECK（没有 wget/curl）
# - 无法进入容器调试（没有 sh/bash）
# - 推荐使用外部健康检查（如 Kubernetes livenessProbe）
# ============================================

CMD ["./main"]
