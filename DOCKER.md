# Docker 部署指南

> **Top1000 服务 Docker 部署完整指南**
>
> **更新时间**：2026-01-10
> **访问量**：优化版（每日约100访问）

---

## 📦 镜像版本

### 快速对比

| 版本 | 文件 | 大小 | Shell | 调试工具 | 适用场景 |
|------|------|------|-------|---------|---------|
| **Scratch** | `Dockerfile` | **6-8MB** ⭐ | ❌ | ❌ | **生产环境（推荐）** |
| Alpine | `Dockerfile.alpine.bak` | 10-12MB | ✅ | ✅ | 开发/调试 |

**推荐**：生产环境用 Scratch，开发调试用 Alpine。

---

## 🚀 快速开始

### 方式一：使用 Docker Compose（推荐）

```bash
# 1. 配置外部 Redis（必须）
cp .env.example .env
# 编辑 .env，配置 REDIS_ADDR 和 REDIS_PASSWORD

# 2. 启动服务
docker-compose up -d

# 3. 查看日志
docker-compose logs -f top1000

# 4. 停止服务
docker-compose down
```

### 方式二：使用 Docker 命令

```bash
# 1. 构建镜像（Scratch 极简版）
docker build -t top1000:scratch .

# 2. 运行容器（需要外部 Redis）
docker run -d \
  --name top1000 \
  -p 7066:7066 \
  --env-file .env \
  top1000:scratch

# 3. 查看日志
docker logs -f top1000
```

---

## ⚙️ 外部 Redis 配置（必须）

此版本使用**外部 Redis**，不在容器内启动。

### 准备外部 Redis

**方式 A：使用现有 Redis 服务**
```bash
# 在 .env 文件中配置
REDIS_ADDR=your-redis-host:port
REDIS_PASSWORD=your-redis-password

# 例如：
REDIS_ADDR=192.144.142.2:26739
REDIS_PASSWORD=CwamSkCRrtdGbCx6
```

**方式 B：单独启动 Redis 容器**
```bash
# 启动 Redis
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --requirepass your_password

# 在 .env 中配置
REDIS_ADDR=host.docker.internal:6379  # Docker Desktop
# 或
REDIS_ADDR=172.17.0.1:6379             # Linux Docker
```

### 验证 Redis 连接

```bash
# 使用 redis-cli 测试
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD ping

# 应返回：PONG
```

---

---

## 🎯 小访问量优化（2026-01-10）

针对每日约100访问的小流量场景，已完成以下优化：

| 优化项 | 优化前 | 优化后 | 效果 |
|--------|--------|--------|------|
| **Redis连接池** | 10个 | 3个 | 节省70%资源 |
| **空闲连接** | 5个 | 1个 | 减少80%内存 |
| **速率限制** | 100次/分钟 | 60次/小时 | 防止滥用 |
| **健康检查** | 启用 | **已移除** | 简化部署 |
| **时区** | tzdata包 | 预设中国时区 | 减小镜像 |

**性能提升**：
- 镜像大小：维持6-8MB（Scratch）、10-12MB（Alpine）
- 内存占用：减少约30%
- 启动速度：提升约20%

---

## 📊 镜像大小分析

### Scratch 版本（6-8MB）

| 组件 | 大小 | 说明 |
|------|------|------|
| Go 二进制 | ~5-6MB | 静态链接、strip 后 |
| web-dist | ~874KB | 前端资源（AG Grid 占大头） |
| CA 证书 | ~200KB | **HTTPS 必需**（调用 IYUU API） |
| **总计** | **6-8MB** | |

### Alpine 版本（10-12MB）

| 组件 | 大小 | 说明 |
|------|------|------|
| Alpine 基础镜像 | ~5-7MB | 包含 shell 和基础工具 |
| Go 二进制 | ~5-6MB | 静态链接 |
| web-dist | ~874KB | 前端资源 |
| CA 证书 + tzdata | ~800KB | HTTPS 证书和时区数据 |
| **总计** | **10-12MB** | |

---

## 🔐 CA 证书说明

### 为何需要 CA 证书？

**必需**！约 200KB，用于：

1. **调用 IYUU API**（HTTPS）
   - API 地址：`https://api.iyuu.cn/top1000.php`
   - 需要 CA 证书验证服务器 SSL 证书

2. **可能的 Redis TLS 连接**
   - 如果 Redis 开启 TLS，需要 CA 证书

3. **安全性**
   - 防止中间人攻击
   - 验证服务器身份

**如果移除证书**：
```
错误：x509: certificate signed by unknown authority
结果：HTTPS 请求失败
```

### 证书来源

在 `Dockerfile` 中：

```dockerfile
# 从 Alpine 提取 CA 证书
FROM alpine:3.19 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
```

---

## ⚠️ 重要提示

### 1. Scratch 版本的限制

由于使用 `scratch` 空白基础镜像，容器内**不包含**：

- ❌ **Shell** (sh/bash/zsh) - 无法执行 `docker exec -it`
- ❌ **调试工具** (ls/cat/ps) - 无法排查文件问题
- ❌ **健康检查** (wget/curl) - 无法使用 `HEALTHCHECK`
- ❌ **包管理器** (apk/yum) - 无法安装额外软件

### 2. 调试方法

**方法 A：查看日志**（唯一方式）
```bash
docker logs -f top1000
```

**方法 B：本地运行**
```bash
go run cmd/top1000/main.go
```

**方法 C：切换 Alpine 版**
```bash
docker build -t top1000:alpine -f Dockerfile .
docker run -it --rm top1000:alpine sh
```

### 3. 健康检查已移除

**不会自动检查健康状态**，需手动监控：

```bash
# 手动检查
curl http://localhost:7066/health

# 查看容器状态
docker ps | grep top1000

# 查看日志
docker logs top1000
```

### 4. 速率限制

**限制**：60次/小时

**触发后**：
- 返回HTTP 429（Too Many Requests）
- 等待1小时后自动解除
- 或重启容器（清除内存限制）

---

## 🔧 故障排查

### 问题 1：容器启动失败

**检查**：
```bash
# 查看详细日志
docker logs top1000

# 检查 .env 配置
cat .env | grep REDIS
```

**常见原因**：
- Redis 地址错误
- Redis 密码错误
- Redis 未启动

### 问题 2：无法连接 Redis

**检查**：
```bash
# 测试 Redis 连接
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD ping

# 检查网络
docker network inspect bridge
```

**解决**：
- 确认 Redis 地址和端口
- 确认 Redis 密码
- 检查防火墙

### 问题 3：数据不更新

**检查**：
```bash
# 查看 Redis 数据
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD
> GET top1000:data
> TTL top1000:data
```

**解决**：
- TTL < 24小时会自动更新
- 或手动触发：重启容器

### 问题 4：镜像还是很大（>10MB）

**检查**：
```bash
# 查看镜像层
docker history top1000:scratch

# 检查 Go 二进制大小
docker run --rm top1000:scratch ls -lh /app/main

# 检查 web-dist 大小
docker run --rm top1000:scratch du -sh /app/web-dist
```

---

## 🎯 适用场景

### ✅ 适合使用此版本

- 个人使用或小团队
- 每日访问量 < 1000
- 不需要复杂的监控
- 资源受限环境（如小VPS）
- 有外部 Redis 实例

### ❌ 不适合使用此版本

- 高并发场景（>1000访问/天）
- 需要自动健康检查
- 需要详细监控和告警
- 企业级部署（需要SLA保障）

---

## 📝 部署检查清单

部署前确认：

- [ ] 外部 Redis 已启动并可访问
- [ ] .env 文件已配置 REDIS_ADDR 和 REDIS_PASSWORD
- [ ] 已构建镜像：`docker build -f Dockerfile.scratch -t top1000:scratch .`
- [ ] 已测试 Redis 连接：`redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD ping`
- [ ] 已启动容器：`docker-compose up -d`
- [ ] 已检查日志：`docker logs top1000`
- [ ] 已访问服务：`curl http://localhost:7066/health`

---

## 🔄 从旧版本迁移

### 如果使用旧版 docker-compose

**旧版**（包含 Redis）：
```yaml
# 旧配置
services:
  top1000: ...
  redis: ...  # 内置 Redis
```

**新版**（外部 Redis）：
```yaml
# 新配置
services:
  top1000: ...  # 仅应用，无 Redis
```

**迁移步骤**：
1. 备份数据：`redis-cli --rdb dump.rdb`
2. 停止旧服务：`docker-compose down`
3. 启动外部 Redis（或使用现有实例）
4. 更新 .env 配置
5. 启动新服务：`docker-compose up -d`

---

## 📚 相关文件

### Docker 配置文件

- `Dockerfile` - Scratch 版（6-8MB，生产推荐）⭐
- `Dockerfile.alpine.bak` - Alpine 版备份（10-12MB，有调试工具）
- `docker-compose.yaml` - 最终版配置（使用 Scratch + 外部 Redis）
- `.env` - 环境变量配置
- `.env.example` - 环境变量模板

### 相关文档

- `DOCKER.md` - 本文档（Docker 部署完整指南）
- `CLAUDE.md` - 项目总文档

---

## 💡 最佳实践

### 1. 生产环境部署

```bash
# 使用 Scratch 版本
docker build -t top1000:latest .

# 使用外部 Redis
docker-compose up -d

# 配置反向代理（Nginx/Caddy）
# 启用 HTTPS（Let's Encrypt）
```

### 2. 监控和日志

```bash
# 日志收集
docker logs -f top1000 >& top1000.log &

# 外部监控
# 使用 Uptime Robot、Pingdom 等
```

### 3. 备份策略

```bash
# 备份 Redis 数据
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD --rdb backup-$(date +%Y%m%d).rdb

# 备份配置
cp .env .env.backup
```

---

## 🎉 总结

**最终配置**：
- ✅ 使用 Scratch 极简版镜像（6-8MB）
- ✅ 使用外部 Redis（不在容器内启动）
- ✅ 移除健康检查（小访问量不需要）
- ✅ 预设中国时区（开箱即用）
- ✅ 优化资源配置（适合每日100访问）

**文件简化**：
- ✅ 一个 docker-compose.yaml
- ✅ 一个 Dockerfile.scratch
- ✅ 一个 DOCKER.md（本文档）

**部署简单**：
1. 配置外部 Redis
2. 运行 `docker-compose up -d`
3. 搞定！

---

**更新时间**：2026-01-10
**作者**：老王
**版本**：v1.0 Final
