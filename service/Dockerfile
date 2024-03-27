# 阶段 1: 构建阶段
FROM node:16-alpine as builder

# 环境变量
ENV NODE_ENV=production
# 设置时区
ENV TZ="Asia/Shanghai"

WORKDIR /app

# 复制 package.json 和锁文件s
COPY package.json pnpm-lock.yaml ./

# 设置环境变量
RUN npm config set registry https://registry.npmmirror.com

# 安装 pnpm 并安装生产依赖
RUN npm install -g pnpm && pnpm install --prod

# 复制应用的其余部分
COPY ./server.json ./
COPY ./tsconfig.json ./
COPY ./static ./
COPY ./src ./

# 阶段 2: 生产阶段
FROM node:16-alpine as production

# 设置时区
ENV TZ="Asia/Shanghai"

ENV NODE_ENV=production
WORKDIR /app

# 只安装 pm2 用于运行应用
RUN npm install pm2 -g

# 从构建阶段复制构建的产物和必要的生产依赖
COPY --from=builder ./server.json ./
COPY --from=builder ./tsconfig.json ./
COPY --from=builder ./static ./
COPY --from=builder ./src ./

# 使用 PM2 运行应用
CMD ["pm2-runtime", "start", "server.json"]

# 暴露端口
EXPOSE 7066
