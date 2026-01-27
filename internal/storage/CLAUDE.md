# Storage 模块

[根目录](../../CLAUDE.md) > [internal](../) > **storage**

## 模块职责

Storage 模块负责 Redis 连接管理、数据持久化、数据过期检查和并发更新控制。它是系统数据存储的唯一接口。

## 入口与启动

- **入口文件**: `redis.go`
- **初始化函数**: `InitRedis()`
- **调用位置**: `internal/server/server.go` 中的 `initStorage()`

## 对外接口

### 连接管理

```go
// 初始化 Redis 连接
func InitRedis() error

// 关闭 Redis 连接
func CloseRedis() error

// 测试连接
func Ping() error
```

### Top1000 数据操作

```go
// 保存数据（向后兼容）
func SaveData(data model.ProcessedData) error

// 保存数据（支持 context）
func SaveDataWithContext(ctx context.Context, data model.ProcessedData) error

// 加载数据（向后兼容）
func LoadData() (*model.ProcessedData, error)

// 加载数据（支持 context）
func LoadDataWithContext(ctx context.Context) (*model.ProcessedData, error

// 检查数据是否存在（向后兼容）
func DataExists() (bool, error)

// 检查数据是否存在（支持 context）
func DataExistsWithContext(ctx context.Context) (bool, error)

// 检查数据是否过期（向后兼容）
func IsDataExpired() (bool, error)

// 检查数据是否过期（支持 context）
func IsDataExpiredWithContext(ctx context.Context) (bool, error)
```

### 站点数据操作

```go
// 保存站点数据（带 24 小时 TTL）
func SaveSitesData(data interface{}) error

// 保存站点数据（支持 context）
func SaveSitesDataWithContext(ctx context.Context, data interface{}) error

// 加载站点数据
func LoadSitesData() (interface{}, error)

// 加载站点数据（支持 context）
func LoadSitesDataWithContext(ctx context.Context) (interface{}, error)

// 检查站点数据是否存在
func SitesDataExists() (bool, error)

// 检查站点数据是否存在（支持 context）
func SitesDataExistsWithContext(ctx context.Context) (bool, error)
```

### 并发控制

```go
// 检查是否正在更新 Top1000 数据
func IsUpdating() bool

// 设置更新标记
func SetUpdating(updating bool)

// 检查是否正在更新站点数据
func IsSitesUpdating() bool

// 设置站点数据更新标记
func SetSitesUpdating(updating bool)
```

## 关键依赖与配置

### 依赖模块

- `internal/config` - 配置管理
- `internal/model` - 数据模型
- `github.com/redis/go-redis/v9` - Redis 客户端

### 环境变量

| 变量 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `REDIS_ADDR` | 是 | 无 | Redis 地址（`host:port`） |
| `REDIS_PASSWORD` | 是 | 无 | Redis 密码 |
| `REDIS_DB` | 否 | `0` | Redis 数据库编号 |

### 常量配置

```go
const (
    dialTimeout  = 10 * time.Second  // 连接超时
    readTimeout  = 5 * time.Second   // 读取超时
    writeTimeout = 5 * time.Second   // 写入超时
    poolSize     = 3                 // 连接池大小
    minIdleConns = 1                 // 最小空闲连接
    timeFormat   = "2006-01-02 15:04:05"  // 时间格式
)

// Redis Key
const (
    DefaultRedisKey    = "top1000:data"   // Top1000 数据 key
    DefaultSitesKey    = "top1000:sites"  // 站点数据 key
)

// 过期时间
const (
    DefaultDataExpire  = 24 * time.Hour   // Top1000 数据过期阈值
    DefaultSitesExpire = 24 * time.Hour   // 站点数据 TTL
)
```

## 数据模型

### Redis 存储格式

**Top1000 数据**（永久存储，基于 time 字段判断过期）:
```json
{
  "time": "2026-01-19 07:50:56",
  "items": [
    {
      "id": 1,
      "siteName": "站点名称",
      "siteid": "123",
      "duplication": "1.5",
      "size": "1.5 GB"
    }
  ]
}
```

**站点数据**（24 小时 TTL）:
```json
{
  "sites": [
    {
      "id": "1",
      "name": "站点名称",
      "url": "https://example.com"
    }
  ]
}
```

## 核心逻辑

### 连接管理

```
InitRedis()
    ↓
创建 Redis 客户端
    ├─ Addr: REDIS_ADDR
    ├─ Password: REDIS_PASSWORD
    ├─ DB: REDIS_DB
    ├─ DialTimeout: 10s
    ├─ ReadTimeout: 5s
    ├─ WriteTimeout: 5s
    ├─ PoolSize: 3
    └─ MinIdleConns: 1
    ↓
发送 Ping 命令测试连接
    ↓
成功 → 记录日志，返回 nil
失败 → 记录错误，返回 error
```

### 数据过期检查

```
IsDataExpiredWithContext(ctx)
    ↓
LoadDataWithContext(ctx)
    ├─ 从 Redis 读取 JSON
    ├─ 反序列化为 ProcessedData
    └─ 返回数据或错误
    ↓
解析 data.Time（北京时间 UTC+8）
    ├─ 解析失败 → 返回 true（认为过期）
    └─ 解析成功 → 减 8 小时转换为 UTC
    ↓
计算时间差 time.Since(dataTime)
    ↓
比较 age > DefaultDataExpire (24h)
    ├─ true → 数据过期
    └─ false → 数据新鲜
    ↓
记录日志并返回结果
```

### 并发更新控制

使用 `sync.Mutex` 实现简单的互斥锁：

```go
var (
    isUpdating   bool
    updateMutex sync.Mutex
)

func IsUpdating() bool {
    updateMutex.Lock()
    defer updateMutex.Unlock()
    return isUpdating
}

func SetUpdating(updating bool) {
    updateMutex.Lock()
    defer updateMutex.Unlock()
    isUpdating = updating
}
```

使用流程：
```
API 请求检查更新
    ↓
IsUpdating()？
├─ true → 跳过，正在更新中
└─ false → SetUpdating(true)
    ↓
执行更新逻辑
    ↓
SetUpdating(false)
```

### Context 超时控制

所有 Redis 操作都支持 context 超时：

```go
func SaveDataWithContext(ctx context.Context, data model.ProcessedData) error {
    // ctx 由调用者传入，带有超时控制
    return redisClient.Set(ctx, key, jsonData, 0).Err()
}

// 调用示例
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
SaveDataWithContext(ctx, data)
```

## 测试与质量

### 当前状态
- 无单元测试
- 无集成测试
- 依赖实际 Redis 验证

### 测试建议

**单元测试文件**: `redis_test.go`

```go
func TestInitRedis_Success(t *testing.T)
func TestInitRedis_Failure(t *testing.T)
func TestSaveDataWithContext(t *testing.T)
func TestLoadDataWithContext(t *testing.T)
func TestDataExistsWithContext(t *testing.T)
func TestIsDataExpiredWithContext_Fresh(t *testing.T)
func TestIsDataExpiredWithContext_Expired(t *testing.T)
func TestSaveSitesDataWithContext(t *testing.T)
func TestLoadSitesDataWithContext(t *testing.T)
func TestIsUpdating_Concurrent(t *testing.T)
func TestPing(t *testing.T)
```

### 测试要点

1. **连接测试** - 验证 Redis 连接成功/失败
2. **CRUD 测试** - 验证数据的增删改查
3. **过期检查** - 验证时间计算正确性
4. **并发测试** - 验证互斥锁正确工作
5. **Context 超时** - 验证超时正确触发

### Mock 建议

使用 `miniredis` 进行单元测试：
```go
func setupMockRedis() (*miniredis.Miniredis, *redis.Client) {
    m := miniredis.RunT(t)
    client := redis.NewClient(&redis.Options{
        Addr: m.Addr(),
    })
    return m, client
}
```

## 相关文件清单

### 核心文件
- `redis.go` - Redis 存储层（311 行）
  - `InitRedis()` - 初始化连接
  - `CloseRedis()` - 关闭连接
  - `SaveDataWithContext()` - 保存 Top1000 数据
  - `LoadDataWithContext()` - 加载 Top1000 数据
  - `DataExistsWithContext()` - 检查存在性
  - `IsDataExpiredWithContext()` - 过期检查
  - `SaveSitesDataWithContext()` - 保存站点数据
  - `LoadSitesDataWithContext()` - 加载站点数据
  - `SitesDataExistsWithContext()` - 检查站点数据存在性
  - `IsUpdating()` / `SetUpdating()` - Top1000 数据更新标记
  - `IsSitesUpdating()` / `SetSitesUpdating()` - 站点数据更新标记

### 测试文件（待创建）
- `redis_test.go` - 单元测试

### 依赖文件
- `../config/config.go` - 配置管理
- `../model/types.go` - 数据模型

## 性能优化

### 已实现优化
1. **连接池复用** - PoolSize=3，MinIdleConns=1
2. **超时控制** - 读写超时 5 秒，防止阻塞
3. **Context 传播** - 所有操作支持 context 超时
4. **数据压缩** - JSON 自动压缩（Redis 内部）

### 可优化项
1. **Pipeline** - 批量操作使用 pipeline 提升性能
2. **连接池调优** - 根据实际负载调整 PoolSize
3. **监控指标** - 添加 Redis 指标监控
4. **重试机制** - 添加网络错误重试

## 常见问题

### Q: Redis 连接失败？
检查环境变量 `REDIS_ADDR` 和 `REDIS_PASSWORD` 是否正确。

### Q: 数据过期检查不准确？
确保系统时区设置正确，代码会将北京时间（UTC+8）转换为 UTC。

### Q: 如何清空数据？
```bash
redis-cli -h <host> -p <port> -a <password>
> DEL top1000:data
> DEL top1000:sites
```

### Q: 如何查看过期时间？
```bash
redis-cli
> TTL top1000:data    # 返回 -1（永不过期）
> TTL top1000:sites   # 返回剩余秒数
```

### Q: 并发更新冲突？
使用 `IsUpdating()` 和 `SetUpdating()` 实现互斥，已内置在 API 层。

---

**最后更新**: 2026-01-27
**代码行数**: ~311 行
**维护状态**: 活跃
