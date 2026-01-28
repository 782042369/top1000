# internal/config - 配置管理

[根目录](../../CLAUDE.md) > [internal](../) > **config**

## 模块快照

**职责**：环境变量加载、配置验证、单例模式管理

**关键文件**：`config.go`

**核心函数**：
- `Load()` - 加载配置（单例）
- `Get()` - 获取配置实例
- `Validate()` - 配置验证

## 环境变量

| 变量名 | 是否必需 | 默认值 | 说明 |
|--------|----------|--------|------|
| `REDIS_ADDR` | 必需 | - | Redis 地址 |
| `REDIS_PASSWORD` | 必需 | - | Redis 密码 |
| `REDIS_DB` | 可选 | `0` | Redis 数据库编号 |
| `IYUU_SIGN` | 可选 | - | IYUU API 签名 |

## 常量配置

```go
const (
    DefaultPort        = "7066"
    DefaultWebDistDir  = "./web-dist"
    DefaultAPIURL      = "https://api.iyuu.cn/top1000.php"
    DefaultDataExpire  = 24 * time.Hour  // 数据过期阈值
    DefaultRedisDB     = 0
    DefaultRedisKey    = "top1000:data"
    DefaultSitesKey    = "top1000:sites"
    DefaultSitesExpire = 24 * time.Hour
)
```

## 配置验证

验证逻辑：
- `REDIS_ADDR` 不能为空
- `REDIS_PASSWORD` 不能为空
- 失败时返回所有缺失字段（一次性报告）

## 数据结构

```go
type Config struct {
    RedisAddr     string  // Redis 地址
    RedisPassword string  // Redis 密码
    RedisDB       int     // Redis 数据库编号
    IYYUSign      string  // IYUU 签名
}
```

## 测试

无测试文件。

**建议**：添加配置加载测试、验证逻辑测试。

---

*文档生成时间：2026-01-28 13:08:52*
