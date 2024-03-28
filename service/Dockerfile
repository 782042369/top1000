# 阶段 1: 构建阶段
FROM node:18-alpine as builder

# 环境变量
ENV NODE_ENV=production
# 设置时区
ENV TZ="Asia/Shanghai"

WORKDIR /app

# 复制 package.json 和锁文件
COPY package.json pnpm-lock.yaml ./

# 设置环境变量
RUN npm config set registry https://registry.npmmirror.com

# 安装 pnpm 并安装生产依赖
# 合并 RUN 命令以减少镜像层
RUN npm install -g pnpm && pnpm install --prod

# 复制应用的其余部分
COPY . .

# 阶段 2: 生产阶段
FROM node:18-alpine as production

# 设置时区
ENV TZ="Asia/Shanghai"

ENV NODE_ENV=production
WORKDIR /app

# 从构建阶段复制构建的产物和必要的生产依赖
COPY --from=builder /app .

CMD ["npx", "ts-node", "./src/app.ts"]

# 暴露端口
EXPOSE 7066
