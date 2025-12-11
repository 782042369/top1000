# 构建阶段：使用多阶段构建最小化生产镜像
# 阶段一：构建 service (Go版本)
FROM golang:1.25-alpine as service-builder
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY cmd ./cmd
COPY internal ./internal

# 构建优化的Go应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -trimpath -o main ./cmd/top1000

# 阶段二：构建 web
FROM node:24-alpine as web-builder
WORKDIR /app

# 安装 pnpm 并配置缓存
RUN npm i -g pnpm@^10 && \
    pnpm config set store-dir /root/.pnpm-store

# 优先复制包管理文件以利用构建缓存
COPY web/package.json web/pnpm-lock.yaml ./web/

# 安装所有依赖（包括devDependencies）
RUN cd web && pnpm install --frozen-lockfile

# 复制源代码
COPY web ./web/

# 执行构建，输出到 web-dist 目录
RUN cd web && pnpm build

# 最终生产阶段
FROM alpine:latest
WORKDIR /app

# 安装ca-certificates以支持HTTPS请求
RUN apk --no-cache add ca-certificates

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 从 service-builder 阶段复制所有必要文件
COPY --from=service-builder --chown=appuser:appgroup /app/main ./main
COPY --from=web-builder --chown=appuser:appgroup /app/web-dist ./web-dist

# 创建 public 目录并设置正确的所有权
RUN mkdir -p ./public && chown appuser:appgroup ./public

# 设置用户权限
USER appuser

# 声明端口
ENV PORT=7066

EXPOSE 7066

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:7066/ || exit 1

CMD ["./main"]
