# Server 模块

[根目录](../../CLAUDE.md) > [internal](../) > **server**

## 模块职责

Server 模块负责 HTTP 服务器的创建、配置和启动，包括路由注册、中间件配置、静态文件服务和日志记录。

## 入口与启动

- **入口文件**: `server.go`
- **入口函数**: `Start()`
- **调用位置**: `cmd/top1000/main.go`

## 对外接口

### 导出函数

```go
// 启动 Web 服务器（阻塞）
func Start()
```

### 内部函数

```go
// 创建 Fiber 应用
func createApp() *fiber.App

// 配置中间件
func setupMiddleware(app *fiber.App)

// 日志中间件
func loggerMiddleware() fiber.Handler

// 安全响应头中间件
func securityHeadersMiddleware() fiber.Handler

// 配置路由
func setupRoutes(app *fiber.App)

// 设置静态文件缓存头
func setCacheHeaders(c *fiber.Ctx) error

// 初始化 Redis
func initStorage()

// 关闭 Redis
func closeRedis()

// 打印启动横幅
func printStartupBanner()

// 打印启动信息
func printStartupInfo(cfg *config.Config)

// 预加载数据
func preloadData()
```

## 关键依赖与配置

### 依赖模块

- `internal/api` - API 处理器
- `internal/config` - 配置管理
- `internal/crawler` - 数据爬取
- `internal/storage` - Redis 存储
- `github.com/gofiber/fiber/v2` - Web 框架

### 常量配置

```go
const (
    appName          = "Top1000"
    requestBodyLimit = 4 * 1024 * 1024  // 4MB 请求体限制
    oneYearMaxAge    = "public, max-age=31536000"  // 1 年缓存
    noCache          = "no-cache, no-store, must-revalidate"
    cspHeader        = "default-src 'self'; ..."  // CSP 策略
)
```

### 默认端口

| 变量 | 默认值 | 环境变量 |
|------|--------|----------|
| `PORT` | `7066` | `PORT` |

## 路由配置

### API 路由

| 路径 | 方法 | 处理器 | 描述 |
|------|------|--------|------|
| `/top1000.json` | GET | `api.GetTop1000Data` | 获取 Top1000 数据 |
| `/sites.json` | GET | `api.GetSitesData` | 获取站点列表 |

### 静态文件

| 路径 | 目录 | 缓存策略 |
|------|------|----------|
| `/` | `./web-dist` | HTML: no-cache<br>其他资源: 1 年 |

## 中间件栈

```
请求 → Recover → Logger → Security Headers → Compress → 路由处理
         ↓         ↓           ↓              ↓           ↓
      恢复异常   记录日志    安全响应头      响应压缩    业务逻辑
```

### 中间件详情

#### 1. Recover（异常恢复）
- 捕获 panic，返回 500 错误
- 防止服务器崩溃

#### 2. Logger（请求日志）
- 记录每个请求的：
  - 时间戳
  - HTTP 方法
  - 请求路径
  - 响应状态码
  - 处理耗时

日志格式：
```
[2026-01-19 07:50:56] GET /top1000.json - 200 - 15ms
```

#### 3. Security Headers（安全响应头）
- `X-XSS-Protection: 1; mode=block` - XSS 保护
- `X-Content-Type-Options: nosniff` - 禁止 MIME 嗅探
- `X-Frame-Options: DENY` - 禁止 iframe 嵌入
- `Content-Security-Policy: ...` - CSP 策略

#### 4. Compress（响应压缩）
- 自动压缩文本响应（JSON、HTML、CSS、JS）
- 减少传输体积

## 启动流程

```
main.go
    ↓
server.Start()
    ↓
验证配置 (config.Validate)
    ↓
打印启动横幅
    ↓
创建 Fiber 应用 (createApp)
    ├─ 配置中间件 (setupMiddleware)
    └─ 配置路由 (setupRoutes)
    ↓
初始化 Redis (initStorage)
    ↓
预加载数据 (preloadData)
    ↓
打印启动信息
    ↓
监听端口 :7066（阻塞）
    ↓
程序退出 → 关闭 Redis (defer closeRedis)
```

## 缓存策略

### 静态文件缓存

| 文件类型 | Cache-Control | 原因 |
|----------|---------------|------|
| `.html` | `no-cache` | 频繁更新，确保获取最新版本 |
| 其他资源 | `max-age=31536000` | 带文件名哈希，内容变化时 URL 变化 |

### 实现逻辑

```go
func setCacheHeaders(c *fiber.Ctx) error {
    path := c.Path()
    isHTML := filepath.Ext(path) == ".html" || path == "/"

    if !isHTML && c.Response().StatusCode() == fiber.StatusOK {
        c.Response().Header.Set("Cache-Control", oneYearMaxAge)
        return nil
    }

    // HTML 文件或错误状态：禁止缓存
    c.Response().Header.Set("Cache-Control", noCache)
    c.Response().Header.Set("Pragma", "no-cache")
    c.Response().Header.Set("Expires", "0")
    return nil
}
```

## 安全配置

### CSP 策略

```go
const cspHeader = "default-src 'self'; " +
    "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; " +
    "img-src 'self' data: https: https://lsky.939593.xyz:11111; " +
    "style-src 'self' 'unsafe-inline'; " +
    "connect-src 'self' https://log.939593.xyz;"
```

### 安全措施

1. **请求体限制** - 4MB 防止大文件攻击
2. **超时控制** - 读写超时 10 秒
3. **XSS 保护** - 禁用内联脚本（除必要情况）
4. **点击劫持防护** - 禁止 iframe 嵌入
5. **MIME 嗅探防护** - 防止内容类型混淆

## 测试与质量

### 当前状态
- 无单元测试
- 无集成测试
- 依赖手动测试

### 测试建议

**单元测试文件**: `server_test.go`

```go
func TestCreateApp(t *testing.T)
func TestSetupMiddleware(t *testing.T)
func TestSetupRoutes(t *testing.T)
func TestSecurityHeadersMiddleware(t *testing.T)
func TestLoggerMiddleware(t *testing.T)
func TestSetCacheHeaders_HTML(t *testing.T)
func TestSetCacheHeaders_Static(t *testing.T)
func TestSetCacheHeaders_Error(t *testing.T)
```

### 测试要点

1. **中间件顺序** - 验证中间件执行顺序正确
2. **路由注册** - 验证所有路由正确注册
3. **缓存策略** - 验证不同文件类型的缓存头
4. **安全头** - 验证安全响应头正确设置
5. **错误处理** - 验证 panic 被正确捕获

## 相关文件清单

### 核心文件
- `server.go` - 服务器配置（176 行）
  - `Start()` - 启动服务器
  - `createApp()` - 创建应用
  - `setupMiddleware()` - 配置中间件
  - `setupRoutes()` - 配置路由
  - `loggerMiddleware()` - 日志中间件
  - `securityHeadersMiddleware()` - 安全头中间件
  - `setCacheHeaders()` - 缓存头设置
  - `initStorage()` - 初始化存储
  - `closeRedis()` - 关闭存储
  - `printStartupBanner()` - 启动横幅
  - `printStartupInfo()` - 启动信息
  - `preloadData()` - 预加载数据

### 测试文件（待创建）
- `server_test.go` - 单元测试

### 依赖文件
- `../api/handlers.go` - API 处理器
- `../config/config.go` - 配置管理
- `../crawler/scheduler.go` - 数据爬取
- `../storage/redis.go` - Redis 存储

## 性能优化

### 已实现优化
1. **响应压缩** - 自动压缩文本响应
2. **静态文件缓存** - 1 年长缓存，减少请求
3. **读写超时** - 防止慢请求占用资源
4. **连接复用** - Fiber 自动管理连接池

### 可优化项
1. **HTTP/2** - 启用 HTTP/2 提升性能
2. **限流中间件** - 防止 DDoS 攻击
3. **监控指标** - 添加 Prometheus metrics
4. **优雅关闭** - 实现优雅关闭机制

## 常见问题

### Q: 如何修改端口？
设置环境变量 `PORT` 或修改 `config.DefaultPort`。

### Q: 如何禁用压缩？
注释 `setupMiddleware()` 中的 `app.Use(compress.New())`。

### Q: 静态文件 404？
检查 `web-dist` 目录是否存在，运行 `cd web && pnpm build`。

### Q: 如何添加新路由？
在 `setupRoutes()` 中添加：
```go
app.Get("/new-route", api.NewHandler)
```

### Q: 如何自定义 CSP？
修改 `cspHeader` 常量，根据需求调整策略。

---

**最后更新**: 2026-01-27
**代码行数**: ~176 行
**维护状态**: 活跃
