# 构建阶段：使用多阶段构建最小化生产镜像
# 阶段一：构建 web 和 service
FROM node:24-alpine AS builder
WORKDIR /app

RUN npm i pnpm@10.11.0 -g

# 优先复制包管理文件以利用构建缓存
COPY web/package.json web/pnpm-lock.yaml ./web/
COPY service/package.json service/pnpm-lock.yaml ./service/

# 安装项目依赖（web & service）
RUN cd web && pnpm install
RUN cd service && pnpm install

# 复制源代码（此时依赖已缓存，代码变更不会触发重新安装依赖）
COPY web ./web/
COPY service ./service/

# 执行构建
RUN cd web && pnpm build
RUN cd service && pnpm build

# -------------------------------------------
# 生产阶段：创建最小化生产镜像
FROM node:24-alpine

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段仅复制必要文件
COPY --from=builder --chown=appuser:appgroup \
    /app/service/dist ./dist/
COPY --from=builder --chown=appuser:appgroup \
    /app/service/package.json \
    /app/service/pnpm-lock.yaml ./
COPY --from=builder --chown=appuser:appgroup \
    /app/service/public ./public/

# 安装生产依赖（自动使用 corepack）
RUN npm i pnpm@10.11.0 -g && \
    pnpm install --frozen-lockfile

# 设置用户权限
USER appuser

EXPOSE 7066
CMD ["node", "/app/dist/app.js"]
