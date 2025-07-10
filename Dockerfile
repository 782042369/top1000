# 构建阶段：使用多阶段构建最小化生产镜像
# 阶段一：构建 web 和 service
FROM node:24-alpine AS builder
WORKDIR /app

# 安装 pnpm 并配置缓存
RUN npm i -g pnpm@10.12.4 && \
    pnpm config set store-dir /root/.pnpm-store

# 优先复制包管理文件以利用构建缓存
COPY web/package.json web/pnpm-lock.yaml ./web/
COPY service/package.json service/pnpm-lock.yaml service/scripts ./service/

# 安装所有依赖（包括devDependencies）
RUN cd web && pnpm install --frozen-lockfile && \
    cd ../service && pnpm install --frozen-lockfile

# 复制源代码
COPY web ./web/
COPY service ./service/

# 执行构建
RUN cd web && pnpm build && \
    cd ../service && pnpm build

# 生产阶段：仅安装生产依赖
FROM node:24-alpine AS production-deps

WORKDIR /app

RUN npm i -g pnpm@10.12.4 && \
    pnpm add @vercel/nft@0.24.4 fs-extra@11.2.0 --save-prod

COPY --from=builder /app /app

RUN cd service && \
    node ./scripts/minify-docker.cjs && \
    rm -rf ./node_modules ./scripts && \
    mv ./app-minimal/node_modules ./ && \
    rm -rf ./app-minimal


# 最终生产阶段
FROM node:24-alpine
WORKDIR /app

# 创建非特权用户
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 从各阶段复制必要文件
COPY --from=production-deps --chown=appuser:appgroup /app/node_modules ./node_modules/
COPY --from=builder --chown=appuser:appgroup /app/service/dist ./dist/
COPY --from=builder --chown=appuser:appgroup /app/service/public ./public/

# 设置用户权限
USER appuser

EXPOSE 7066
CMD ["node", "/app/dist/app.js"]
