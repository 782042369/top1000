# Config 模块

[根目录](../../CLAUDE.md) > [internal](../) > **config**

## 模块职责

Config 模块负责从环境变量加载应用配置，验证配置有效性，并提供配置访问接口。采用单例模式确保全局唯一。

## 入口与启动

- **入口文件**: `config.go`
- **初始化函数**: `Load()`
- **调用位置**: 其他模块通过 `Get()` 获取配置实例

## 对外接口

### 配置加载

```go
// 加载配置（单例模式，首次调用时初始化）
func Load() *Config

// 获取配置实例（如果未初始化则自动调用 Load）
func Get() *Config
```

### 配置验证

```go
// 验证配置的有效性（返回所有错误）
func Validate() error
```

### 辅助函数

```go
// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string

// 获取整数环境变量，如果不存在或解析失败则返回默认值
func getEnvInt(key string, defaultValue int) int
```

## 关键依赖与配置

### 依赖模块
- 无（仅使用标准库）

### 环境变量

| 变量 | 必需 | 默认值 | 描述 | 示例 |
|------|------|--------|------|------|
| `REDIS_ADDR` | 是 | 无 | Redis 地址（`host:port`） | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | 是 | 无 | Redis 密码 | `yourpassword` |
| `REDIS_DB` | 否 | `0` | Redis 数据库编号 | `0` |
| `IYUU_SIGN` | 否 | 空 | IYUU API 签名 | `your_sign` |
| `PORT` | 否 | `7066` | HTTP 服务端口 | `7066` |

### 常量配置

```go
const (
    DefaultPort         = "7066"
    DefaultWebDistDir   = "./web-dist"
    DefaultAPIURL       = "https://api.iyuu.cn/top1000.php"
    DefaultDataExpire   = 24 * time.Hour  // 数据过期阈值
    DefaultRedisDB      = 0               // Redis 数据库编号
    DefaultRedisKey     = "top1000:data"  // Top1000 数据 key
    DefaultSitesKey     = "top1000:sites" // 站点数据 key
    DefaultSitesExpire  = 24 * time.Hour  // 站点数据 TTL
)
```

## 数据模型

### Config 结构

```go
type Config struct {
    RedisAddr     string // Redis 地址（必须配置）
    RedisPassword string // Redis 密码（必须配置）
    RedisDB       int    // Redis 数据库编号（可选，默认 0）
    IYYUSign      string // IYUU 签名（可选，用于调用站点 API）
}
```

### ValidationError 结构

```go
type ValidationError struct {
    errors []string  // 收集所有验证错误
}

func (e *ValidationError) Error() string
func (e *ValidationError) Add(field string)
func (e *ValidationError) IsValid() bool
```

## 核心逻辑

### 配置加载流程

```
Load()
    ↓
检查单例是否已存在
    ├─ 存在 → 直接返回
    └─ 不存在 → 继续加载
        ↓
    从环境变量读取配置
        ├─ REDIS_ADDR     (必需)
        ├─ REDIS_PASSWORD (必需)
        ├─ REDIS_DB       (可选，默认 0)
        └─ IYUU_SIGN      (可选，默认空)
        ↓
    保存到单例 appConfig
        ↓
    返回配置实例
```

### 配置验证流程

```
Validate()
    ↓
创建 ValidationError 收集器
    ↓
检查必需字段
    ├─ REDIS_ADDR == "" ？
    │   └─ 是 → Add("REDIS_ADDR")
    ├─ REDIS_PASSWORD == "" ？
    │   └─ 是 → Add("REDIS_PASSWORD")
    └─ ...
    ↓
检查是否有错误
    ├─ 有错误 → 返回 ValidationError
    └─ 无错误 → 返回 nil
```

### 单例模式实现

```go
var appConfig *Config  // 全局单例

func Load() *Config {
    if appConfig != nil {
        return appConfig  // 已初始化，直接返回
    }

    // 首次初始化
    appConfig = &Config{
        RedisAddr:     getEnv("REDIS_ADDR", ""),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       getEnvInt("REDIS_DB", DefaultRedisDB),
        IYYUSign:      getEnv("IYUU_SIGN", ""),
    }

    return appConfig
}

func Get() *Config {
    if appConfig == nil {
        return Load()  // 懒加载
    }
    return appConfig
}
```

## 环境变量优先级

1. **.env 文件** - 通过 `godotenv.Load()` 加载（不强制）
2. **系统环境变量** - 覆盖 .env 文件
3. **Docker 环境变量** - 覆盖系统环境变量

示例：
```bash
# .env 文件
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=defaultpassword

# 系统环境变量（优先级更高）
export REDIS_PASSWORD=systempassword

# 最终使用
# RedisAddr: 127.0.0.1:6379
# RedisPassword: systempassword
```

## 测试与质量

### 当前状态
- 无单元测试
- 依赖实际环境验证

### 测试建议

**单元测试文件**: `config_test.go`

```go
func TestLoad_Success(t *testing.T)
func TestLoad_Singleton(t *testing.T)
func TestGet_NotInitialized(t *testing.T)
func TestGet_Initialized(t *testing.T)
func TestValidate_AllValid(t *testing.T)
func TestValidate_MissingAddr(t *testing.T)
func TestValidate_MissingPassword(t *testing.T)
func TestValidate_MissingBoth(t *testing.T)
func TestGetEnv_Exists(t *testing.T)
func TestGetEnv_NotExists(t *testing.T)
func TestGetEnvInt_Valid(t *testing.T)
func TestGetEnvInt_Invalid(t *testing.T)
```

### 测试用例示例

```go
func TestValidate_MissingAddr(t *testing.T) {
    // 设置环境变量
    os.Setenv("REDIS_ADDR", "")
    os.Setenv("REDIS_PASSWORD", "password")

    // 重新加载配置
    appConfig = nil  // 清除单例
    cfg := Load()

    // 验证失败
    if err := Validate(); err == nil {
        t.Error("应该返回验证错误")
    } else {
        fmt.Println(err)  // "配置验证失败: REDIS_ADDR"
    }
}

func TestGetEnvInt_Valid(t *testing.T) {
    os.Setenv("TEST_DB", "5")
    result := getEnvInt("TEST_DB", 0)
    if result != 5 {
        t.Errorf("期望 5，得到 %d", result)
    }
}
```

### Mock 建议

使用 `os.Setenv()` 和 `os.Unsetenv()` 进行测试：
```go
func setupTestEnv() func() {
    // 保存原始环境变量
    originalAddr := os.Getenv("REDIS_ADDR")
    originalPwd := os.Getenv("REDIS_PASSWORD")

    // 设置测试环境变量
    os.Setenv("REDIS_ADDR", "localhost:6379")
    os.Setenv("REDIS_PASSWORD", "testpassword")

    // 返回清理函数
    return func() {
        if originalAddr == "" {
            os.Unsetenv("REDIS_ADDR")
        } else {
            os.Setenv("REDIS_ADDR", originalAddr)
        }
        if originalPwd == "" {
            os.Unsetenv("REDIS_PASSWORD")
        } else {
            os.Setenv("REDIS_PASSWORD", originalPwd)
        }
    }
}
```

## 相关文件清单

### 核心文件
- `config.go` - 配置管理（115 行）
  - `Config` 结构体
  - `ValidationError` 结构体及方法
  - `Load()` - 加载配置
  - `Get()` - 获取配置
  - `Validate()` - 验证配置
  - `getEnv()` - 获取字符串环境变量
  - `getEnvInt()` - 获取整数环境变量

### 测试文件（待创建）
- `config_test.go` - 单元测试

### 依赖文件
- `.env.example` - 环境变量模板

## 最佳实践

### 1. 敏感信息管理

**不要**将 `.env` 文件提交到 Git：
```bash
# .gitignore
.env
.env.local
.env.production
```

提供 `.env.example` 作为模板：
```bash
# .env.example
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=填写Redis密码
REDIS_DB=0
IYUU_SIGN=填写IYUU签名
```

### 2. 配置验证

在应用启动时验证配置：
```go
// cmd/top1000/main.go
func main() {
    _ = godotenv.Load()

    // 验证配置
    if err := config.Validate(); err != nil {
        log.Fatalf("❌ 配置验证失败: %v", err)
    }

    server.Start()
}
```

### 3. 配置访问

始终通过 `config.Get()` 访问配置：
```go
cfg := config.Get()
fmt.Println(cfg.RedisAddr)
```

### 4. 默认值使用

为可选字段提供合理的默认值：
```go
RedisDB: getEnvInt("REDIS_DB", DefaultRedisDB),  // 默认 0
IYYUSign: getEnv("IYUU_SIGN", ""),               // 默认空
```

## 扩展建议

### 可能的扩展配置

```go
type Config struct {
    // 现有字段...
    RedisAddr     string
    RedisPassword string
    RedisDB       int
    IYYUSign      string

    // 可能的扩展字段
    HTTPPort      int           `env:"HTTP_PORT" default:"7066"`
    LogLevel      string        `env:"LOG_LEVEL" default:"info"`
    DataExpire    time.Duration `env:"DATA_EXPIRE" default:"24h"`
    EnableCache   bool          `env:"ENABLE_CACHE" default:"true"`
    MaxRetries    int           `env:"MAX_RETRIES" default:"3"`
}
```

### 配置库集成

考虑使用成熟的配置库：
- [viper](https://github.com/spf13/viper) - 支持多种格式（JSON、YAML、TOML）
- [envconfig](https://github.com/kelseyhightower/envconfig) - 纯环境变量，结构体标签

示例（使用 viper）：
```go
func Load() *Config {
    viper.SetDefault("redis.addr", "localhost:6379")
    viper.SetDefault("redis.db", 0)
    viper.AutomaticEnv()
    viper.BindEnv("redis.addr", "REDIS_ADDR")

    cfg := &Config{}
    viper.Unmarshal(cfg)
    return cfg
}
```

## 常见问题

### Q: 为什么不使用配置文件？
- 项目规模小，环境变量足够
- Docker 部署更友好
- 避免配置文件泄漏风险

### Q: 如何在生产环境使用？
1. 使用 Docker secrets 或 Kubernetes secrets
2. 设置系统环境变量
3. 使用 `.env` 文件（不提交到 Git）

### Q: 配置热更新支持吗？
不支持。配置在启动时加载，修改后需要重启。

### Q: 如何添加新的环境变量？
1. 在 `.env.example` 中添加模板
2. 在 `Config` 结构体中添加字段
3. 在 `Load()` 中调用 `getEnv()` 或 `getEnvInt()`

### Q: 为什么不用结构体标签？
- 项目简单，硬编码更清晰
- 避免引入额外依赖
- 验证逻辑灵活

---

**最后更新**: 2026-01-27
**代码行数**: ~115 行
**维护状态**: 稳定
