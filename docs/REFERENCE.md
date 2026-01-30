# 脚本命令参考

本文档汇总了 Top1000 项目的所有可用脚本命令。

**最后更新:** 2026-01-30

## 后端命令（Go）

### 位置
`server/` 目录

### 开发命令

| 命令 | 说明 |
|------|------|
| `air` | 启动 Air 热重载开发服务器 |
| `go run ./cmd/top1000/main.go` | 直接运行应用 |
| `go build -o ./tmp/main ./cmd/top1000` | 构建二进制文件到 tmp 目录 |
| `go build -o top1000 ./cmd/top1000` | 构建二进制文件到当前目录 |

### 测试命令

| 命令 | 说明 |
|------|------|
| `go test ./...` | 运行所有测试 |
| `go test -v ./...` | 运行测试（详细输出） |
| `go test -v ./internal/storage` | 运行特定包的测试 |
| `go test -cover ./...` | 运行测试并显示覆盖率 |
| `go test -coverprofile=coverage.out ./...` | 生成覆盖率报告 |
| `go tool cover -html=coverage.out` | 在浏览器中查看覆盖率 |

### 依赖管理

| 命令 | 说明 |
|------|------|
| `go mod download` | 下载依赖 |
| `go mod tidy` | 整理依赖（移除未使用的） |
| `go mod verify` | 验证依赖完整性 |
| `go mod why <package>` | 解释为什么需要某个依赖 |

### 工具命令

| 命令 | 说明 |
|------|------|
| `go install github.com/air-verse/air@latest` | 安装 Air 热重载工具 |
| `go install github.com/swaggo/swag/cmd/swag@latest` | 安装 Swagger 文档生成器 |

## 前端命令（TypeScript/Vite）

### 位置
`web/` 目录

### package.json 脚本

| 脚本 | 命令 | 说明 |
|------|------|------|
| `dev` | `vite` | 启动开发服务器（默认 5173 端口） |
| `build` | `vite build --emptyOutDir` | 生产构建（输出到 dist/） |
| `lint` | `eslint --fix --concurrency=auto` | 代码检查和自动修复 |
| `preview` | `vite preview` | 预览构建产物 |

### 依赖管理

| 命令 | 说明 |
|------|------|
| `pnpm install` | 安装依赖 |
| `pnpm install <package>` | 安装指定包 |
| `pnpm install -D <package>` | 安装开发依赖 |
| `pnpm update` | 更新依赖 |
| `pnpm outdated` | 检查过期依赖 |

### Node 版本要求

```json
{
  "node": ">=24.3.0",
  "pnpm": ">=10.12.4"
}
```

## Docker 命令

### 构建和运行

| 命令 | 说明 |
|------|------|
| `docker build -t top1000 .` | 构建镜像 |
| `docker build -t top1000:latest --no-cache .` | 无缓存构建 |
| `docker run -d -p 7066:7066 top1000` | 运行容器 |
| `docker-compose up -d` | 使用 Compose 启动服务 |
| `docker-compose up -d --build` | 强制重新构建 |

### 容器管理

| 命令 | 说明 |
|------|------|
| `docker ps` | 查看运行中的容器 |
| `docker-compose ps` | 查看 Compose 服务状态 |
| `docker logs -f top1000-iyuu` | 查看容器日志 |
| `docker-compose logs -f` | 查看所有服务日志 |
| `docker-compose restart` | 重启服务 |
| `docker-compose down` | 停止并删除容器 |
| `docker exec -it top1000-iyuu sh` | 进入容器（注：scratch 镜像无 shell） |

### 镜像管理

| 命令 | 说明 |
|------|------|
| `docker images` | 列出本地镜像 |
| `docker rmi top1000` | 删除镜像 |
| `docker pull 782042369/top1000-iyuu:latest` | 拉取官方镜像 |

## Redis 命令

### 连接和测试

| 命令 | 说明 |
|------|------|
| `redis-cli` | 连接本地 Redis |
| `redis-cli -h <host> -p 6379 -a <password>` | 连接远程 Redis |
| `redis-cli ping` | 测试连接 |
| `redis-cli INFO` | 查看 Redis 信息 |

### 数据操作

| 命令 | 说明 |
|------|------|
| `KEYS *` | 列出所有 key |
| `GET top1000:data` | 获取 Top1000 数据 |
| `GET sites:data` | 获取站点数据 |
| `DEL top1000:data` | 删除 Top1000 数据 |
| `TTL top1000:data` | 查看过期时间 |
| `EXPIRE top1000:data 86400` | 设置过期时间（秒） |

### 服务器操作

| 命令 | 说明 |
|------|------|
| `BGSAVE` | 后台保存数据到磁盘 |
| `FLUSHALL` | 清空所有数据（危险） |
| `INFO memory` | 查看内存使用 |

## 快捷命令别名

### 推荐的 shell 别名

```bash
# ~/.bashrc 或 ~/.zshrc

# Top1000 项目别名
alias top-dev='cd /Users/yanghongxuan/study/my-projects/top1000/server && air'
alias top-web='cd /Users/yanghongxuan/study/my-projects/top1000/web && pnpm dev'
alias top-build='cd /Users/yanghongxuan/study/my-projects/top1000/web && pnpm build'
alias top-logs='docker-compose logs -f top1000'
alias top-restart='docker-compose restart top1000'
alias top-status='docker-compose ps'
```

## 常用组合命令

### 完整开发流程

```bash
# 后端开发
cd server && air

# 前端开发（新终端）
cd web && pnpm dev
```

### 完整构建流程

```bash
# 构建前端
cd web && pnpm build

# 构建后端
cd server && go build -o top1000 ./cmd/top1000
```

### Docker 部署流程

```bash
# 构建并启动
docker-compose down && docker-compose up -d --build

# 查看日志
docker-compose logs -f
```

## 相关文档

- [开发工作流](./CONTRIB.md)
- [运维手册](./RUNBOOK.md)
- [部署文档](../DEPLOYMENT.md)
