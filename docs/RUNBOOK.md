# 运维手册 (RUNBOOK)

本文档描述 Top1000 应用的运维操作和故障处理流程。

**最后更新:** 2026-01-30

## 目录

- [系统概览](#系统概览)
- [部署架构](#部署架构)
- [日常运维](#日常运维)
- [监控告警](#监控告警)
- [故障处理](#故障处理)
- [性能优化](#性能优化)
- [备份恢复](#备份恢复)

## 系统概览

### 应用架构

```
                    ┌─────────────────┐
                    │   Nginx/Caddy   │
                    │   (可选反向代理) │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Top1000 App    │
                    │  (Port: 7066)   │
                    │  - Fiber Server │
                    │  - 静态文件服务  │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │     Redis       │
                    │  (Port: 6379)   │
                    └─────────────────┘
```

### 关键指标

| 指标 | 正常值 | 告警阈值 |
|------|--------|----------|
| 应用响应时间 | < 200ms | > 1s |
| Redis 连接数 | 1-10 | > 50 |
| 内存使用 | < 100MB | > 200MB |
| CPU 使用 | < 10% | > 50% |
| 磁盘使用 | < 80% | > 90% |

## 部署架构

### Docker 部署（推荐）

```yaml
# docker-compose.yaml
services:
  top1000:
    image: 782042369/top1000-iyuu:latest
    ports:
      - "7066:7066"
    environment:
      - REDIS_ADDR=host.docker.internal:6379
      - REDIS_PASSWORD=your_password
    restart: always
```

### 环境变量

| 变量 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `PORT` | 否 | 7066 | 应用监听端口 |
| `REDIS_ADDR` | 是 | - | Redis 地址 |
| `REDIS_PASSWORD` | 是 | - | Redis 密码 |
| `REDIS_DB` | 否 | 0 | Redis 数据库编号 |
| `IYUU_SIGN` | 否 | - | IYUU API 签名 |
| `TZ` | 否 | UTC | 时区设置 |

## 日常运维

### 启动和停止

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看状态
docker-compose ps
```

### 日志查看

```bash
# 实时日志
docker-compose logs -f top1000

# 最近 100 行
docker-compose logs --tail=100 top1000

# 带时间戳
docker-compose logs -t top1000
```

### 健康检查

```bash
# 检查 Top1000 数据 API
curl -f http://localhost:7066/top1000.json | jq '.time'

# 检查站点列表 API
curl -f http://localhost:7066/sites.json | jq '.'

# 检查容器状态
docker inspect top1000-iyuu | jq '.[0].State.Health'
```

### 数据更新检查

```bash
# 查看数据更新时间
curl http://localhost:7066/top1000.json | jq -r '.time'

# 查看 Redis 中的 TTL
redis-cli -a your_password TTL top1000:data
```

## 监控告警

### 关键日志模式

```bash
# 正常启动
✅ Redis连接成功
✅ 静态文件服务已启用
[爬虫] 预加载成功，已存入Redis

# 异常情况
❌ Redis连接失败
❌ 保存数据失败
[爬虫] 预加载失败
```

### 日志关键字

| 关键字 | 含义 | 处理动作 |
|--------|------|----------|
| `Redis连接失败` | Redis 不可达 | 检查 Redis 服务 |
| `数据过期` | 数据需要更新 | 等待自动刷新 |
| `预加载失败` | 启动时获取数据失败 | 检查网络连接 |
| `保存数据失败` | Redis 写入失败 | 检查 Redis 磁盘空间 |

## 故障处理

### 问题 1：应用无法启动

**症状**

```
Error: listen tcp :7066: bind: address already in use
```

**诊断**

```bash
# 查找占用进程
lsof -i :7066

# 或使用 netstat
netstat -tuln | grep 7066
```

**解决**

```bash
# 停止占用进程
kill -9 <PID>

# 或修改端口
export PORT=7067
docker-compose up -d
```

### 问题 2：Redis 连接失败

**症状**

```
❌ Redis连接失败: dial tcp: connection refused
```

**诊断**

```bash
# 测试 Redis 连接
redis-cli -h <host> -p 6379 -a <password> ping

# 检查 Redis 状态
systemctl status redis
# 或
docker ps | grep redis
```

**解决**

```bash
# 启动 Redis
sudo systemctl start redis
# 或
docker start redis

# 检查防火墙
sudo ufw allow 6379
```

### 问题 3：数据不更新

**症状**

Top1000 数据时间戳过旧

**诊断**

```bash
# 检查数据时间
curl -s http://localhost:7066/top1000.json | jq -r '.time'

# 检查日志
docker-compose logs top1000 | grep "爬虫"

# 手动触发更新（重启应用）
docker-compose restart top1000
```

**解决**

数据会在 24 小时后自动过期刷新。如需立即更新：

```bash
# 删除 Redis 中的数据，下次请求会自动刷新
redis-cli -a <password> DEL top1000:data

# 重启应用触发预加载
docker-compose restart top1000
```

### 问题 4：容器反复重启

**症状**

```bash
docker-compose ps
# Restarting (1) X seconds ago
```

**诊断**

```bash
# 查看日志
docker-compose logs top1000

# 检查退出码
docker inspect top1000-iyuu | jq '.[0].State.ExitCode'
```

**常见原因**

1. Redis 配置错误
2. 环境变量缺失
3. 端口冲突

**解决**

检查 `.env` 文件配置是否正确，确保必需的环境变量已设置。

### 问题 5：内存使用过高

**症状**

容器内存使用持续增长

**诊断**

```bash
# 查看资源使用
docker stats top1000-iyuu

# 查看 Redis 内存使用
redis-cli -a <password> INFO memory
```

**解决**

```bash
# 设置 Redis 最大内存
redis-cli -a <password> CONFIG SET maxmemory 256mb
redis-cli -a <password> CONFIG SET maxmemory-policy allkeys-lru

# 重启应用释放内存
docker-compose restart top1000
```

## 性能优化

### Redis 优化

```bash
# /etc/redis/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### 应用优化

当前应用已针对低访问量优化（约 100 访问/天）：

- 数据缓存 24 小时
- 静态资源长期缓存
- 懒加载数据更新

### 并发控制

应用内置并发保护：

```go
// 防止同时更新
if lock.IsUpdating() {
    return // 跳过本次更新
}
```

## 备份恢复

### Redis 备份

```bash
# 手动备份
redis-cli -a <password> BGSAVE

# 备份文件位置
# Linux: /var/lib/redis/dump.rdb
# Docker: 在数据卷中
```

### 自动备份脚本

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/redis"

docker exec redis redis-cli --rdb /data/dump_$DATE.rdb
docker cp redis:/data/dump_$DATE.rdb $BACKUP_DIR/

# 保留最近 7 天的备份
find $BACKUP_DIR -name "dump_*.rdb" -mtime +7 -delete
```

### Redis 恢复

```bash
# 停止 Redis
docker-compose stop redis

# 恢复备份文件
cp backup.rdb /var/lib/redis/dump.rdb

# 启动 Redis
docker-compose start redis
```

## 扩展和升级

### 版本升级

```bash
# 拉取最新镜像
docker pull 782042369/top1000-iyuu:latest

# 停止旧容器
docker-compose down

# 启动新版本
docker-compose up -d
```

### 验证升级

```bash
# 检查服务状态
docker-compose ps

# 测试 API
curl http://localhost:7066/top1000.json

# 检查日志
docker-compose logs -f --tail=50 top1000
```

## 安全建议

1. **网络隔离**
   - Redis 不对外暴露
   - 使用 Docker 网络隔离

2. **访问控制**
   - 配置 Redis 密码
   - 限制容器网络访问

3. **定期更新**
   - 及时更新镜像
   - 关注安全公告

4. **日志审计**
   - 定期检查异常访问
   - 保留关键操作日志

## 相关文档

- [部署文档](./DEPLOYMENT.md)
- [开发工作流](./CONTRIB.md)
- [项目文档](../CLAUDE.md)

## 联系支持

如有问题，请提交 Issue 或联系维护者。

---

**文档版本:** 1.0.0
**最后更新:** 2026-01-30
