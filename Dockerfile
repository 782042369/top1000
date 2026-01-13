# ============================================
# 极简版 Dockerfile - Scratch 基础镜像
# 目标镜像大小：4-5MB（每日100访问量优化版）
# ============================================
# 警告：此镜像不包含 shell 和任何调试工具
# 如需调试，请使用 Dockerfile（Alpine版）
# 已移除健康检查（小访问量不需要）
# 时区默认为中国时区（Asia/Shanghai）
#
# 优化措施：
# 1. 前端：仅导入实际使用的 AG Grid 模块（减少16.5%）
# 2. Go 二进制：UPX 压缩（减少60%）
# 3. 基础镜像：Scratch 空镜像
# ============================================

# 阶段一：构建 service (Go版本)
FROM golang:1.25-alpine AS service-builder
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && \
    go mod verify

# 复制源代码
COPY cmd ./cmd
COPY internal ./internal

# 安装 UPX 压缩工具
RUN apk add --no-cache upx

# 构建完全静态的Go应用（极致优化）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static' -buildid=" \
    -trimpath \
    -o main ./cmd/top1000 && \
    chmod +x main

# 使用 UPX 压缩二进制文件（减少50-70%体积）
# --best: 最佳压缩比
# --lzma: 使用LZMA算法（压缩率更高）
RUN upx --best --lzma main

# 阶段二：构建 web
FROM node:24-alpine AS web-builder
WORKDIR /app

# 安装 pnpm
RUN npm install -g pnpm@10

# 优先复制包管理文件以利用构建缓存
COPY web/package.json web/pnpm-lock.yaml ./web/

# 安装依赖
RUN cd web && pnpm install --frozen-lockfile --prod=false

# 复制源代码
COPY web ./web/

# 执行构建，输出到 web-dist 目录
RUN cd web && pnpm build

# 阶段三：准备 CA 证书（从 Alpine 提取）
FROM alpine:3.19 AS certs
RUN apk --no-cache add ca-certificates

# ============================================
# 最终生产阶段：使用 Scratch（空镜像）
# ============================================
FROM scratch
WORKDIR /app

# 从 certs 阶段复制 CA 证书（HTTPS 必需）
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从 service-builder 阶段复制 Go 二进制
COPY --from=service-builder /app/main ./main

# 从 web-builder 阶段复制前端文件
COPY --from=web-builder /app/web-dist ./web-dist

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
# - 已移除健康检查（每日100访问量不需要）
# ============================================

CMD ["./main"]
