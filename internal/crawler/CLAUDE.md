# Crawler 模块

[根目录](../../CLAUDE.md) > [internal](../) > **crawler**

## 模块职责

Crawler 模块负责从 IYUU API 爬取 Top1000 数据，并进行解析和验证。它是系统数据源的唯一入口点。

## 入口与启动

- **入口文件**: `scheduler.go`
- **主要函数**:
  - `FetchTop1000()` - 向后兼容的默认超时版本
  - `FetchTop1000WithContext(ctx)` - 支持外部 context 的版本
  - `PreloadData()` - 启动时预加载数据
- **调用位置**:
  - `internal/api/handlers.go` - API 层调用
  - `internal/server/server.go` - 启动预加载

## 对外接口

### 导出函数

```go
// 向后兼容（使用默认超时）
func FetchTop1000() (*model.ProcessedData, error)

// 支持外部传入 context（推荐）
func FetchTop1000WithContext(ctx context.Context) (*model.ProcessedData, error)

// 启动时预加载数据
func PreloadData()
```

### 内部函数

```go
// 执行 HTTP 请求
func doFetchWithContext(ctx context.Context) (*model.ProcessedData, error)

// 解析原始文本
func parseResponse(rawData string) model.ProcessedData

// 解析数据行
func parseDataLines(dataLines []string) ([]model.SiteItem, int)

// 解析单组数据（3 行）
func parseItemGroup(group []string) (model.SiteItem, bool)
```

## 关键依赖与配置

### 依赖模块

- `internal/config` - 配置管理（API URL）
- `internal/model` - 数据模型
- `internal/storage` - Redis 存储
- `net/http` - HTTP 客户端
- `context` - 超时控制

### 常量配置

```go
const (
    logPrefix       = "🔍 爬虫"
    httpTimeout     = 10 * time.Second  // HTTP 超时
    maxRetries      = 1                 // 最大重试次数
    retryInterval   = 1 * time.Second   // 重试间隔
    linesPerItem    = 3                 // 每条数据占 3 行
    timeLineIndex   = 0                 // 时间行索引
    dataStartLine   = 2                 // 数据开始行
    timePrefix      = "create time "
    timeSuffix      = " by "
    fieldSeparator  = "："
    sitePattern     = `站名：(.*?) 【ID：(\d+)】`
)
```

### 环境变量

| 变量 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `API_URL` | 否 | `https://api.iyuu.cn/top1000.php` | IYUU API 地址 |

## 数据模型

### 输入格式（API 返回的纯文本）

```
create time 2026-01-19 07:50:56 by IYUU

站名：站点名称 【ID：123】
重复度：1.5
文件大小：1.5 GB

站名：站点名称2 【ID：456】
重复度：2.0
文件大小：2.3 GB
...
```

### 输出格式（解析后）

```go
type ProcessedData struct {
    Time  string     // "2026-01-19 07:50:56"
    Items []SiteItem // 解析后的站点列表
}

type SiteItem struct {
    SiteName    string // "站点名称"
    SiteID      string // "123"
    Duplication string // "1.5"
    Size        string // "1.5 GB"
    ID          int    // 自动递增
}
```

## 核心逻辑

### 数据爬取流程

```
FetchTop1000WithContext(ctx)
    ↓
获取任务锁 (tryLock)
    ↓
循环重试（最多 maxRetries 次）
    ↓
doFetchWithContext(ctx)
    ├─ 创建 HTTP 请求（带 ctx 超时）
    ├─ 发送 GET 请求到 IYUU API
    ├─ 读取响应体
    └─ parseResponse(body)
        ├─ 提取时间行
        ├─ 分割数据行（每 3 行一条）
        ├─ parseItemGroup() - 正则提取
        ├─ 数据验证
        └─ 返回 ProcessedData
    ↓
释放任务锁
```

### 解析逻辑

1. **标准化换行符**: 将 `\r\n` 统一为 `\n`
2. **提取时间**: 从第一行提取 `create time 2026-01-19 07:50:56 by IYUU`
3. **分组解析**: 每 3 行为一组（站名行、重复度行、大小行）
4. **正则提取**: 使用 `站名：(.*?) 【ID：(\d+)】` 提取站名和 ID
5. **字段分割**: 使用 `：` 分割字段名和值
6. **ID 赋值**: 按顺序自动递增

### 并发控制

使用 `sync.Mutex` 实现简单的任务锁：
- `taskMutex.TryLock()` - 尝试获取锁，失败表示任务正在进行
- `defer taskMutex.Unlock()` - 确保锁被释放

### 预加载机制

启动时检查 Redis 中是否有数据：
```
PreloadData()
    ↓
checkDataLoadRequired(ctx)
    ├─ 检查数据是否存在 (storage.DataExistsWithContext)
    └─ 检查数据是否过期 (storage.IsDataExpiredWithContext)
    ↓
需要加载？
├─ 是 → FetchTop1000WithContext(ctx)
│   └─ 保存到 Redis (storage.SaveDataWithContext)
└─ 否 → 跳过
```

## 测试与质量

### 当前状态
- 无单元测试
- 无集成测试
- 依赖实际 API 验证

### 测试建议

**单元测试文件**: `scheduler_test.go`

```go
func TestParseResponse_Success(t *testing.T)
func TestParseResponse_EmptyData(t *testing.T)
func TestParseItemGroup_ValidFormat(t *testing.T)
func TestParseItemGroup_InvalidFormat(t *testing.T)
func TestExtractTime_Valid(t *testing.T)
func TestExtractTime_Invalid(t *testing.T)
func TestParseDataLines_SkippedCount(t *testing.T)
func TestFetchTop1000WithContext_Timeout(t *testing.T)
func TestFetchTop1000WithContext_Retry(t *testing.T)
```

### 测试要点

1. **解析逻辑测试** - 使用模拟数据验证各种格式
2. **边界条件** - 空数据、格式错误、不完整数据
3. **并发测试** - 验证任务锁正确工作
4. **超时测试** - 验证 context 超时正确触发
5. **重试逻辑** - 验证失败重试机制

### Mock 建议

使用 `httptest.Server` 模拟 IYUU API：
```go
func setupMockServer(response string) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(response))
    }))
}
```

## 相关文件清单

### 核心文件
- `scheduler.go` - 爬虫调度器（254 行）
  - `FetchTop1000()` - 向后兼容入口
  - `FetchTop1000WithContext()` - 核心爬取逻辑
  - `doFetchWithContext()` - HTTP 请求执行
  - `parseResponse()` - 响应解析
  - `parseDataLines()` - 数据行解析
  - `parseItemGroup()` - 单组解析
  - `extractTime()` - 时间提取
  - `PreloadData()` - 启动预加载
  - `checkDataLoadRequired()` - 加载检查

### 测试文件（待创建）
- `scheduler_test.go` - 单元测试

### 依赖文件
- `../config/config.go` - 配置管理
- `../model/types.go` - 数据模型
- `../storage/redis.go` - Redis 存储

## 性能优化

### 已实现优化
1. **并发控制** - 避免重复爬取
2. **重试机制** - 提高成功率（最多 1 次重试）
3. **Context 超时** - 防止请求挂起（10 秒）
4. **正则预编译** - `siteRegex` 在包初始化时编译

### 可优化项
1. **连接复用** - 使用全局 `http.Client` 连接池
2. **压缩传输** - 启用 gzip 压缩（如果 API 支持）
3. **并发爬取** - 如果需要爬取多个 API
4. **缓存策略** - 失败时缓存响应，便于调试

## 常见问题

### Q: 解析失败数据丢失？
解析失败的行会被跳过，日志会记录跳过数量。不会影响其他数据的解析。

### Q: 如何查看解析日志？
查看日志输出，搜索 `🔍 爬虫` 前缀：
```
[🔍 爬虫] 数据获取成功（12345 字节）
[🔍 爬虫] 数据解析完成（1000 条）
[🔍 爬虫] 警告：跳过 5 条格式错误的数据
```

### Q: 爬取超时怎么办？
调整 `httpTimeout` 常量，或检查网络连接和 IYUU API 状态。

### Q: 如何禁用启动预加载？
注释 `internal/server/server.go` 中的 `preloadData()` 调用。

---

**最后更新**: 2026-01-27
**代码行数**: ~254 行
**维护状态**: 活跃
