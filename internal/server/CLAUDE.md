# internal/server - HTTP 服务器

[根目录](../../CLAUDE.md) > [internal](../) > **server**

## 模块快照

**职责**：Fiber 应用配置、路由设置、中间件管理

**关键文件**：`server.go`

**核心函数**：
- `Start()` - 启动服务器
- `createApp()` - 创建 Fiber 实例
- `setupRoutes()` - 配置路由
- `setupMiddleware()` - 配置中间件

## 路由配置

| 路由 | 方法 | 处理器 | 说明 |
|------|------|--------|------|
| `/top1000.json` | GET | `api.GetTop1000Data` | Top1000 数据 |
| `/sites.json` | GET | `api.GetSitesData` | 站点列表 |
| `/` | GET | 静态文件 | `web-dist/` 目录 |

## 中间件链

```
请求 → recover.New() → loggerMiddleware → securityHeadersMiddleware → compress.New() → 路由处理
```

## 安全响应头

- `X-XSS-Protection: 1; mode=block`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy`（自定义 CSP）

## 缓存策略

- **静态资源**（JS/CSS/图片）：`max-age=31536000`（1年）
- **HTML 文件**：`no-cache`（禁止缓存）

## 启动流程

```go
1. 验证配置（config.Validate()）
2. 打印启动横幅
3. 创建 Fiber 应用
4. 初始化 Redis 连接
5. 预加载数据（crawler.PreloadData()）
6. 打印启动信息
7. 监听端口（默认 7066）
```

## 依赖关系

- `github.com/gofiber/fiber/v2` - Web 框架
- `internal/api` - 路由处理器
- `internal/config` - 配置管理
- `internal/crawler` - 数据预加载
- `internal/storage` - Redis 初始化

## 测试

无测试文件。

**建议**：添加路由测试、中间件测试、缓存策略测试。

---

*文档生成时间：2026-01-28 13:08:52*
