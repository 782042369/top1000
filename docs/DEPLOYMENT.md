# Top1000 部署文档

本文档介绍如何部署 Top1000 应用。

## 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [Docker 部署](#docker-部署)
- [手动部署](#手动部署)
- [环境变量配置](#环境变量配置)
- [健康检查](#健康检查)
- [故障排查](#故障排查)

## 环境要求

### Docker 部署

- Docker 20.10+
- Docker Compose 2.0+

### 手动部署

- Go 1.25+
- Redis 7.0+

## 快速开始

### 使用 Docker Compose（推荐）

1. **克隆仓库**

```bash
git clone <repository-url>
cd top1000
```

2. **配置环境变量**

```bash
cp .env.example .env
# 编辑 .env 文件，设置必要的配置
```

3. **启动服务**

```bash
docker-compose up -d
```

4. **查看日志**

```bash
docker-compose logs -f top1000
```

5. **访问应用**

打开浏览器访问：`http://localhost:7066`

API 文档：`http://localhost:7066/swagger/`

## Docker 部署

### 使用 Dockerfile

1. **构建镜像**

```bash
docker build -t top1000:latest .
```

2. **运行容器**

```bash
docker run -d \
  --name top1000 \
  -p 7066:7066 \
  -e REDIS_ADDR=redis:6379 \
  -e REDIS_PASSWORD=your_password \
  -e IYUU_SIGN=your_sign \
  top1000:latest
```

### 使用 Docker Compose

Docker Compose 是最简单的部署方式，它会自动启动 Redis 和应用服务。

**启动服务**

```bash
docker-compose up -d
```

**停止服务**

```bash
docker-compose down
```

**查看服务状态**

```bash
docker-compose ps
```

**查看日志**

```bash
# 查看所有日志
docker-compose logs

# 实时查看日志
docker-compose logs -f

# 查看特定服务的日志
docker-compose logs -f top1000
docker-compose logs -f redis
```

**重启服务**

```bash
docker-compose restart
```

**更新服务**

```bash
git pull
docker-compose down
docker-compose build
docker-compose up -d
```

## 手动部署

### 1. 安装依赖

**macOS**

```bash
brew install go redis
```

**Ubuntu/Debian**

```bash
sudo apt update
sudo apt install golang redis-server
```

### 2. 启动 Redis

```bash
# 启动 Redis 服务
redis-server

# 或者使用 systemd（Linux）
sudo systemctl start redis
sudo systemctl enable redis
```

### 3. 配置环境变量

创建 `.env` 文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
IYUU_SIGN=your_iyuu_sign
```

### 4. 构建应用

```bash
cd server
go mod download
go build -o top1000 ./cmd/top1000
```

### 5. 运行应用

```bash
./top1000
```

应用将在 `http://localhost:7066` 启动。

## 环境变量配置

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `REDIS_ADDR` | 是 | - | Redis 地址（格式：host:port） |
| `REDIS_PASSWORD` | 是 | - | Redis 密码 |
| `REDIS_DB` | 否 | 0 | Redis 数据库编号 |
| `IYUU_SIGN` | 否 | - | IYUU API 签名（用于获取站点列表） |

### 获取 IYUU_SIGN

1. 访问 [IYUU 官网](https://iyuu.cn/)
2. 注册账号
3. 在个人中心获取 API 签名
4. 将签名填入 `.env` 文件

**注意**：IYUU_SIGN 是可选的，如果不配置，`/sites.json` 接口将不可用。

## 健康检查

应用提供以下健康检查端点：

- **Top1000 数据**：`http://localhost:7066/top1000.json`
- **站点列表**：`http://localhost:7066/sites.json`（需要配置 IYUU_SIGN）
- **Swagger UI**：`http://localhost:7066/swagger/`

### 检查脚本

```bash
#!/bin/bash

# 检查 Top1000 数据
curl -f http://localhost:7066/top1000.json || echo "❌ Top1000 API 不可用"

# 检查站点列表
curl -f http://localhost:7066/sites.json || echo "❌ Sites API 不可用"

echo "✅ 健康检查完成"
```

## 故障排查

### 问题 1：Redis 连接失败

**症状**

```
❌ Redis连接失败: dial tcp: connection refused
```

**解决方案**

1. 检查 Redis 是否运行：

```bash
redis-cli ping
```

2. 检查 Redis 配置：

```bash
# 检查 .env 文件
cat .env | grep REDIS

# 检查 Docker Compose 配置
docker-compose config
```

3. 确保 Redis 密码正确

### 问题 2：数据未更新

**症状**

Top1000 数据不是最新的

**解决方案**

1. 检查数据时间：

```bash
curl http://localhost:7066/top1000.json | jq '.time'
```

2. 查看应用日志：

```bash
docker-compose logs -f top1000 | grep "爬虫"
```

3. 手动触发更新（重启应用）：

```bash
docker-compose restart top1000
```

### 问题 3：IYUU 站点列表不可用

**症状**

```json
{"error": "未配置IYUU_SIGN环境变量"}
```

**解决方案**

1. 确保已配置 `IYUU_SIGN` 环境变量
2. 检查签名是否正确
3. 访问 [IYUU 官网](https://iyuu.cn/) 重新获取签名

### 问题 4：Docker 容器启动失败

**症状**

```bash
docker-compose up -d
# 容器退出
```

**解决方案**

1. 查看容器日志：

```bash
docker-compose logs top1000
```

2. 检查容器状态：

```bash
docker-compose ps
```

3. 检查 Redis 健康状态：

```bash
docker-compose ps redis
```

### 问题 5：端口冲突

**症状**

```
Error: listen tcp :7066: bind: address already in use
```

**解决方案**

1. 检查端口占用：

```bash
lsof -i :7066
```

2. 停止占用端口的进程，或修改端口：

```yaml
# docker-compose.yml
ports:
  - "7067:7066"  # 使用 7067 端口
```

## 性能优化

### Redis 配置

对于生产环境，建议优化 Redis 配置：

```conf
# redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### 应用配置

1. **增加连接池大小**（修改 `server/internal/storage/redis_store.go`）

```go
poolSize: 10  // 默认 3
minIdleConns: 5  // 默认 1
```

2. **调整超时时间**

```go
dialTimeout: 5 * time.Second
readTimeout: 3 * time.Second
writeTimeout: 3 * time.Second
```

## 监控和日志

### 查看应用日志

```bash
# Docker Compose
docker-compose logs -f top1000

# 手动部署
tail -f /var/log/top1000/app.log
```

### 日志级别

应用日志级别：

- ✅ 成功操作
- ⚠️ 警告信息
- ❌ 错误信息
- 📊 数据操作
- 🔍 爬虫操作

## 安全建议

1. **修改默认端口**：避免使用默认端口
2. **配置防火墙**：仅允许必要的端口访问
3. **使用强密码**：Redis 和 IYUU_SIGN 都应使用强密码
4. **定期更新**：保持依赖和系统更新
5. **备份 Redis 数据**：定期备份 Redis 数据

## 许可证

MIT License

## 支持

如有问题，请提交 [Issue](https://github.com/your-repo/issues)
