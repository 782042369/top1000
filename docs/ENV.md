# 环境变量配置参考

本文档列出了 Top1000 项目的所有环境变量及其用途。

**最后更新:** 2026-01-30

## 必需环境变量

### REDIS_ADDR

Redis 服务器地址，用于连接 Redis 数据存储。

| 属性 | 值 |
|------|-----|
| 类型 | `string` |
| 必需 | 是 |
| 示例 | `localhost:6379` 或 `redis:6379` |

```bash
REDIS_ADDR=localhost:6379
```

### REDIS_PASSWORD

Redis 服务器的认证密码。

| 属性 | 值 |
|------|-----|
| 类型 | `string` |
| 必需 | 是 |
| 示例 | `your_secure_password` |

```bash
REDIS_PASSWORD=your_secure_password
```

**注意**: 如果 Redis 没有设置密码，可以设置为空字符串。

## 可选环境变量

### REDIS_DB

Redis 数据库编号，用于在同一个 Redis 实例中隔离不同环境的数据。

| 属性 | 值 |
|------|-----|
| 类型 | `number` |
| 必需 | 否 |
| 默认值 | `0` |
| 范围 | 0-15 |

```bash
REDIS_DB=0
```

**建议的数据库分配**:
- `0` - 开发环境
- `1` - 测试环境
- `2` - 生产环境

### IYUU_SIGN

IYUU API 签名，用于获取站点列表数据。

| 属性 | 值 |
|------|-----|
| 类型 | `string` |
| 必需 | 否 |
| 功能 | 启用 `/sites.json` API 端点 |

```bash
IYUU_SIGN=your_iyuu_sign_here
```

**获取方式**:
1. 访问 [IYUU 官网](https://iyuu.cn/)
2. 注册账号
3. 在个人中心获取 API 签名

### PORT

应用监听端口。

| 属性 | 值 |
|------|-----|
| 类型 | `number` |
| 必需 | 否 |
| 默认值 | `7066` |

```bash
PORT=7066
```

### TZ

容器时区设置。

| 属性 | 值 |
|------|-----|
| 类型 | `string` |
| 必需 | 否 |
| 默认值 | `UTC` |
| 常用值 | `Asia/Shanghai`, `Asia/Hong_Kong` |

```bash
TZ=Asia/Shanghai
```

## 环境配置示例

### 开发环境 (.env.development)

```bash
# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# IYUU 配置
IYUU_SIGN=dev_sign_here

# 应用配置
PORT=7066
TZ=Asia/Shanghai
```

### 生产环境 (.env.production)

```bash
# Redis 配置
REDIS_ADDR=redis.production.internal:6379
REDIS_PASSWORD=strong_secure_password_here
REDIS_DB=2

# IYUU 配置
IYUU_SIGN=production_sign_here

# 应用配置
PORT=7066
TZ=Asia/Shanghai
```

### Docker Compose 环境

```yaml
# docker-compose.yaml
services:
  top1000:
    env_file:
      - .env
    environment:
      - PORT=7066
      - TZ=Asia/Shanghai
```

## Redis 连接字符串格式

### 标准 TCP 连接

```bash
REDIS_ADDR=hostname:port
REDIS_ADDR=localhost:6379
REDIS_ADDR=192.168.1.100:6379
```

### Docker 容器间连接

```bash
REDIS_ADDR=redis:6379
REDIS_ADDR=host.docker.internal:6379  # 从容器访问宿主机
```

### Unix Socket 连接

Go Redis 客户端也支持 Unix Socket：

```bash
# 需要修改代码支持
REDIS_ADDR=/var/run/redis/redis.sock
```

## 安全建议

### 密码强度

- 至少 16 个字符
- 包含大小写字母、数字、特殊字符
- 不要使用常见密码或字典词汇

### 密码生成

```bash
# 使用 openssl 生成随机密码
openssl rand -base64 24

# 使用 /dev/urandom
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 24 | head -n 1
```

### 环境变量存储

- 不要将 `.env` 文件提交到版本控制
- 使用 `.env.example` 作为模板
- 生产环境使用密钥管理服务（如 AWS Secrets Manager）

## 故障排查

### Redis 连接测试

```bash
# 测试连接
redis-cli -h $REDIS_ADDR -a $REDIS_PASSWORD ping

# 测试端口连通性
telnet $REDIS_HOST 6379
```

### 环境变量调试

```bash
# 显示所有环境变量
env | grep -E "REDIS|IYUU|PORT"

# 在 Docker 中检查环境变量
docker exec top1000-iyuu env | sort
```

## 相关文档

- [脚本命令参考](./REFERENCE.md)
- [运维手册](./RUNBOOK.md)
- [部署文档](../DEPLOYMENT.md)
