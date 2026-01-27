# API 模块

[根目录](../../CLAUDE.md) > [internal](../) > **api**

## 模块职责

API 模块负责处理所有 HTTP 请求，包括数据获取、更新调度和响应返回。它是前端与后端核心逻辑的桥梁。

## 入口与启动

- **入口文件**: `handlers.go`
- **注册位置**: `internal/server/server.go` 中的 `setupRoutes()`
- **主要函数**:
  - `GetTop1000Data()` - 处理 Top1000 数据请求
  - `GetSitesData()` - 处理 IYUU 站点列表请求

## 对外接口

### HTTP 端点

| 端点 | 方法 | 处理函数 | 描述 |
|------|------|---------|------|
| `/top1000.json` | GET | `GetTop1000Data` | 获取 Top1000 资源数据 |
| `/sites.json` | GET | `GetSitesData` | 获取 IYUU 站点列表（需配置 IYUU_SIGN） |

### 响应格式

**Top1000 数据响应**:
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

**站点数据响应**:
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

**错误响应**:
```json
{
  "error": "错误描述信息"
}
```

## 关键依赖与配置

### 依赖模块

- `internal/config` - 配置管理（获取 IYUU_SIGN）
- `internal/crawler` - 数据爬取
- `internal/storage` - Redis 存储
- `github.com/gofiber/fiber/v2` - Web 框架

### 环境变量

| 变量 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `IYUU_SIGN` | 否 | 空 | IYUU API 签名，用于访问站点列表 |

### 常量配置

```go
const (
    defaultAPITimeout        = 15 * time.Second // API 默认超时时间
    defaultHTTPClientTimeout = 5 * time.Second  // HTTP 客户端超时时间
)
```

## 数据模型

使用 `internal/model` 包中定义的数据结构：
- `model.ProcessedData` - Top1000 完整数据
- `model.SiteItem` - 单条站点数据

## 核心逻辑

### 数据更新流程

```
请求 /top1000.json
    ↓
检查数据是否需要更新 (shouldUpdateData)
    ↓
需要更新？
├─ 是 → 检查是否正在更新 (IsUpdating)
│   ├─ 否 → 设置更新标记，开始刷新 (refreshData)
│   │   ├─ 保存旧数据（容错）
│   │   ├─ 爬取新数据 (crawler.FetchTop1000WithContext)
│   │   ├─ 成功 → 保存到 Redis，返回新数据
│   │   └─ 失败 → 使用旧数据，记录日志
│   └─ 是 → 跳过，等待当前更新完成
└─ 否 → 直接返回 Redis 数据
```

### 并发控制

使用 `storage.IsUpdating()` 和 `storage.SetUpdating()` 实现简单的互斥锁，防止并发更新。

### 容错机制

1. **爬取失败**: 使用 Redis 旧数据
2. **超时保护**: 所有操作都有 context 超时控制
3. **错误日志**: 记录详细的错误信息用于排查

## 测试与质量

### 当前状态
- 无单元测试
- 无集成测试
- 依赖手动测试和实际环境验证

### 测试建议

**单元测试文件**: `handlers_test.go`

```go
func TestGetTop1000Data_Success(t *testing.T)
func TestGetTop1000Data_StorageError(t *testing.T)
func TestGetSitesData_MissingSign(t *testing.T)
func TestRefreshData_FallbackToOld(t *testing.T)
func TestShouldUpdateData_NoData(t *testing.T)
func TestShouldUpdateData_Expired(t *testing.T)
func TestShouldUpdateData_Fresh(t *testing.T)
```

### 测试要点

1. **并发更新测试** - 模拟多个请求同时触发更新
2. **容错机制测试** - 验证爬取失败时返回旧数据
3. **超时控制测试** - 验证 context 超时正确触发
4. **边界条件测试** - 数据不存在、格式错误等

## 相关文件清单

### 核心文件
- `handlers.go` - API 处理器（202 行）
  - `GetTop1000Data()` - Top1000 数据接口
  - `GetSitesData()` - 站点列表接口
  - `shouldUpdateData()` - 更新检查
  - `refreshData()` - 数据刷新
  - `shouldUpdateSitesData()` - 站点数据更新检查
  - `refreshSitesData()` - 站点数据刷新

### 测试文件（待创建）
- `handlers_test.go` - 单元测试

### 依赖文件
- `../config/config.go` - 配置管理
- `../crawler/scheduler.go` - 数据爬取
- `../storage/redis.go` - Redis 存储
- `../model/types.go` - 数据模型

## 性能优化

### 已实现优化
1. **Context 超时** - 防止请求挂起
2. **并发控制** - 避免重复爬取
3. **容错机制** - 失败时返回旧数据，减少重试

### 可优化项
1. **连接池复用** - HTTP 客户端可改为全局单例
2. **缓存头优化** - 可添加更精细的 Cache-Control
3. **监控指标** - 添加 Prometheus metrics

## 常见问题

### Q: 为什么站点接口返回 502？
检查是否配置了 `IYUU_SIGN` 环境变量。

### Q: 数据更新很慢？
可能是 IYUU API 响应慢，检查日志中的时间戳。

### Q: 如何禁用某个接口？
在 `internal/server/server.go` 的 `setupRoutes()` 中注释掉对应路由。

---

**最后更新**: 2026-01-27
**代码行数**: ~202 行
**维护状态**: 活跃
