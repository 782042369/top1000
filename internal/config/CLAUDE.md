# 配置管理

> 读取环境变量的模块

---

## 模块功能

**从环境变量读取配置，启动时验证配置**

核心功能：
1. 读取环境变量（Redis密码、端口等）
2. 提供默认值（部分配置有默认值）
3. 启动时验证配置（Redis地址和密码必须填写）

---

## 配置项

### 服务器配置

```go
Port string  // HTTP端口，默认7066
```

### 前端配置

```go
WebDistDir string  // 前端构建目录，默认./web-dist
```

### API配置

```go
Top1000APIURL string  // 数据源API，默认IYUU的
```

### 缓存配置

```go
CacheDuration      time.Duration  // 静态文件缓存，默认24小时
DataExpireDuration time.Duration  // 数据过期检测，默认24小时
```

### Redis配置（重点）

```go
RedisEnabled   bool    // 是否启用Redis，默认true
RedisAddr      string  // Redis地址（**必须配置**）
RedisPassword  string  // Redis密码（**必须配置**）
RedisDB        int     // 数据库编号，默认0
RedisKeyPrefix string  // 键前缀，默认top1000:
```

---

## 怎么用

### 获取配置

```go
cfg := config.Get()
log.Printf("端口: %s", cfg.Port)
log.Printf("Redis: %s", cfg.RedisAddr)
```

**单例模式**：全局只有一个配置对象，首次调用时初始化。

---

### 验证配置（新增功能）

```go
if err := config.Validate(); err != nil {
    log.Fatalf("❌ 配置验证失败: %v", err)
}
```

**验证规则**：
- 如果`RedisEnabled=true`，必须配置`REDIS_ADDR`
- 如果`RedisEnabled=true`，必须配置`REDIS_PASSWORD`
- 验证失败会**直接fatal退出**

**使用场景**：程序启动时（server.go里调用）

---

## 环境变量读取

### 辅助函数

```go
// 读字符串
getEnv(key, defaultValue string) string

// 读整数
getEnvInt(key string, defaultValue int) int

// 读时间（如24h、30m）
getEnvDuration(key string, defaultValue time.Duration) time.Duration
```

### 使用示例

```go
Port: getEnv("PORT", "7066"),  // 读PORT，没有就用7066
RedisAddr: getEnv("REDIS_ADDR", ""),  // 读REDIS_ADDR，空字符串（强制配置）
```

---

## 配置验证细节

### Validate函数做了啥

```go
func Validate() error {
    if appConfig.RedisEnabled {
        // 检查Redis地址
        if appConfig.RedisAddr == "" {
            return fmt.Errorf("REDIS_ADDR 环境变量未设置")
        }

        // 检查Redis密码
        if appConfig.RedisPassword == "" {
            return fmt.Errorf("REDIS_PASSWORD 环境变量未设置")
        }
    }
    return nil
}
```

### 为啥要验证

**以前**：Redis密码硬编码在代码里
```go
RedisAddr:    getEnv("REDIS_ADDR", "192.144.142.2:26739"),
RedisPassword: getEnv("REDIS_PASSWORD", "CwamSkCRrtdGbCx6"),
```
**问题**：不安全，密码泄露了

**现在**：必须通过环境变量配置
```go
RedisAddr:    getEnv("REDIS_ADDR", ""),
RedisPassword: getEnv("REDIS_PASSWORD", ""),
```
**好处**：安全，灵活

---

## 环境变量列表

### 必须配置的（否则无法启动）

| 变量 | 说明 | 示例 |
|------|------|------|
| `REDIS_ADDR` | Redis地址 | `192.144.142.2:26739` |
| `REDIS_PASSWORD` | Redis密码 | `填写密码` |

### 可选配置（有默认值）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `7066` | HTTP端口 |
| `WEB_DIST_DIR` | `./web-dist` | 前端目录 |
| `TOP1000_API_URL` | `https://api.iyuu.cn/top1000.php` | 数据源API |
| `CACHE_DURATION` | `24h` | 静态文件缓存 |
| `DATA_EXPIRE_DURATION` | `24h` | 数据过期检测 |
| `REDIS_ENABLED` | `true` | 是否启用Redis |
| `REDIS_DB` | `0` | Redis数据库编号 |
| `REDIS_KEY_PREFIX` | `top1000:` | 键前缀 |
| `CORS_ORIGINS` | `*` | CORS允许的来源 |

---

## 常见问题

### Q: 为何Redis配置是必须的？

**A**: 此版本依赖Redis存储数据，没有Redis无法运行。因此必须配置。

### Q: 默认值是什么？

**A**: 参考上表，除了Redis地址和密码，其他都有默认值。

### Q: 如何配置？

**A**: 创建`.env`文件：
```bash
REDIS_ADDR=192.144.142.2:26739
REDIS_PASSWORD=填写密码
```

### Q: 配置错误会怎样？

**A**: 程序启动时检查，错误会直接退出：
```
❌ 配置验证失败: REDIS_ADDR 环境变量未设置
```

### Q: 能否修改默认值？

**A**: 可以，修改`config.go`里的Load函数，或设置环境变量覆盖。

---

## 代码优化

### 安全性提升

| 方面 | 优化前 | 优化后 |
|------|--------|--------|
| Redis地址 | 硬编码 | 必须环境变量 |
| Redis密码 | 硬编码 | 必须环境变量 |
| 配置验证 | 无 | 启动时检查 |

### 新增功能

- ✅ `Validate()`函数
- ✅ 友好的错误提示
- ✅ 强制环境变量配置

---

## 相关文件

- `config.go` - 配置管理代码（115行）
- `.env.example` - 环境变量模板
- `../server/server.go` - 启动时调用Validate()

---

**总结**：配置管理应优先考虑安全性，不要将密码硬编码在代码中，使用环境变量！

**更新**: 2026-01-10
**代码质量**: A级
**优化**: 配置验证 + 移除硬编码
