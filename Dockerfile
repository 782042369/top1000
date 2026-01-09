# 构建阶段：使用多阶段构建最小化生产镜像
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

# 构建优化的Go应用（完全静态）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath -o main ./cmd/top1000

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

# 最终生产阶段：使用Alpine优化版本
FROM alpine:3.19
WORKDIR /app

# 仅安装必需的包
RUN apk --no-cache add ca-certificates wget tzdata && \
    rm -rf /var/cache/apk/*

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 从 service-builder 阶段复制所有必要文件
COPY --from=service-builder --chown=appuser:appgroup /app/main ./main
COPY --from=web-builder --chown=appuser:appgroup /app/web-dist ./web-dist

# 设置用户权限
USER appuser

# 设置时区
ENV TZ=Asia/Shanghai

# 声明端口
ENV PORT=7066
EXPOSE 7066

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:7066/health || exit 1

CMD ["./main"]
