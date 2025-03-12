# 使用更小的 Alpine 版本作为构建基础镜像
FROM node:18-alpine as web-builder
WORKDIR /app/web
COPY web/package*.json ./
# 利用 Alpine 的 pnpm 安装并启用缓存清理
RUN corepack enable && \
    pnpm install --frozen-lockfile && \
    pnpm cache clean --force
COPY web .
RUN pnpm build

FROM node:18-alpine as service-builder
WORKDIR /app/service
COPY service/package*.json ./
RUN corepack enable && \
    pnpm install --frozen-lockfile --prod && \
    pnpm cache clean --force
COPY service .
RUN pnpm build

# 最终生产镜像
FROM node:18-alpine

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制必要文件
COPY --from=service-builder --chown=appuser:appgroup \
    /app/service/dist ./dist/
COPY --from=service-builder --chown=appuser:appgroup \
    /app/service/package.json ./package.json

# 单独复制 pnpm 相关文件
COPY --from=service-builder /app/service/pnpm-lock.yaml ./

# 安装生产依赖并清理缓存
RUN corepack enable && \
    pnpm install --frozen-lockfile --prod && \
    pnpm cache clean --force

# 设置用户权限
USER appuser

EXPOSE 7066
CMD ["node", "/app/dist/app.js"]
