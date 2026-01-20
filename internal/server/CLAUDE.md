# HTTP 服务器

> 启动HTTP服务器并配置路由

---

## 模块功能

**启动HTTP服务器，配置路由和中间件**

核心功能：
1. 创建Fiber应用
2. 配置中间件（日志、CORS、安全头、限流）
3. 注册路由（API、静态文件）
4. 初始化Redis
5. 启动时预加载数据 ⭐
6. 启动服务

---

## 启动流程

```go
func Start() {
    cfg := config.Get()

    // 1. 验证配置
    if err := config.Validate(); err != nil {
        log.Fatalf("❌ 配置验证失败: %v", err)
    }

    // 2. 打印启动横幅
    printStartupBanner()

    // 3. 创建Fiber应用
    app := createApp(cfg)

    // 4. 初始化Redis
    initStorage()

    // 5. 启动时预加载数据 ⭐
    preloadData()

    // 6. 打印启动信息
    printStartupInfo(cfg)

    // 7. 确保程序退出时关闭Redis连接
    defer closeRedis()

    // 8. 启动服务
    log.Fatal(app.Listen(":" + cfg.Port))
}
```

**预加载功能**：
- 在Redis初始化之后，服务启动之前执行
- 检查Redis中是否已有数据
- 如果没有数据或数据过期，自动从API获取并存储
- 预加载失败不影响服务启动（容错机制）
- **避免首次访问超时问题**

---

## 中间件配置

### 错误恢复

```go
app.Use(recover.New())
```

**作用**：panic不会导致崩溃，会恢复并记录日志

### 日志

```go
app.Use(logger.New(logger.Config{
    Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
    TimeFormat: "2006-01-02 15:04:05",
    TimeZone:   "Asia/Shanghai",
}))
```

**格式**：`[2025-12-11 07:52:33] 200 - GET /top1000.json - 10ms`

### CORS

```go
corsOrigins := os.Getenv("CORS_ORIGINS")
if corsOrigins == "" {
    corsOrigins = "*"
}

// 当使用通配符时，不能启用 AllowCredentials
allowCredentials := corsOrigins != "*"

app.Use(cors.New(cors.Config{
    AllowOrigins:     corsOrigins,
    AllowMethods:     "GET,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    ExposeHeaders:    "Content-Length,ETag,Cache-Control",
    MaxAge:           86400,
    AllowCredentials: allowCredentials,
}))
```

**特点**：
- 通配符（*）+ 携带凭证存在安全风险
- 因此通配符时禁用AllowCredentials
- 指定域名时才允许携带凭证

### 安全头

手动配置安全头（不使用Helmet，因为COEP无法禁用）：

```go
app.Use(func(c *fiber.Ctx) error {
    // XSS保护
    c.Set("X-XSS-Protection", "1; mode=block")
    // 禁止MIME类型嗅探
    c.Set("X-Content-Type-Options", "nosniff")
    // 防止点击劫持
    c.Set("X-Frame-Options", "DENY")
    // CSP：允许外部监控脚本、图片、数据上报
    c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; img-src 'self' data: https: https://lsky.939593.xyz:11111; style-src 'self' 'unsafe-inline'; connect-src 'self' https://log.939593.xyz;")
    // 不设置COEP和COOP，允许跨域资源加载
    return c.Next()
})
```

**作用**：
- **防XSS攻击**：`X-XSS-Protection`
- **防止MIME类型嗅探**：`X-Content-Type-Options`
- **防止点击劫持**：`X-Frame-Options`
- **CSP白名单**：允许监控脚本和图片加载
- **禁用COEP/COOP**：让跨域能正常加载

### 速率限制

```go
app.Use(limiter.New(limiter.Config{
    Max:        200,  // 每小时最多200次（小项目）
    Expiration: 1 * time.Hour,
    KeyGenerator: func(c *fiber.Ctx) string {
        return c.IP()  // 基于IP限流
    },
    LimitReached: func(c *fiber.Ctx) error {
        return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
            "error": "请求过于频繁，请稍后再试",
        })
    },
}))
```

**作用**：防止DDoS，每个IP每小时最多200次请求

### 响应压缩

```go
app.Use(compress.New(compress.Config{
    Level: compress.LevelBestSpeed,
}))
```

**作用**：压缩响应体，节省带宽

---

## 路由配置

### API路由

```go
app.Get("/top1000.json", api.GetTop1000Data)
```

**返回**：Top1000的JSON数据

### 静态文件

```go
app.Static("/", cfg.WebDistDir, fiber.Static{
    CacheDuration: 0, // Fiber内部缓存禁用，完全由ModifyResponse自定义
    Browse:        true,
    MaxAge:        0,
    ModifyResponse: func(c *fiber.Ctx) error {
        path := c.Path()
        // 非HTML文件：长期缓存（1年）
        if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, "/") {
            c.Response().Header.Set("Cache-Control", "public, max-age=31536000")
        } else {
            // HTML文件：不缓存
            c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
        }
        return nil
    },
})
```

**缓存策略**：
- HTML文件：不缓存（每次都请求最新的）
- 其他文件：缓存1年（JS、CSS等）

---

## Fiber配置

```go
app := fiber.New(fiber.Config{
    AppName:      "Top1000 Service",
    StrictRouting: true,        // 启用严格路由
    BodyLimit:    4 * 1024 * 1024, // 限制请求体4MB
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
})
```

**说明**：
- StrictRouting：`/api`和`/api/`不同（严格匹配）
- BodyLimit：防止大文件攻击
- 超时：10秒足够

---

## 启动日志

```
========================================
   Top1000 服务正在启动...
========================================
正在初始化 Redis 连接...
正在连接 Redis: 192.144.142.2:26739 (DB: 0)
✅ Redis 连接成功
✅ Redis 初始化成功
========================================
[🔍 爬虫] 检查是否需要预加载数据...
[🔍 爬虫] ✅ 预加载成功，已存入Redis（共 1000 条记录）
========================================
✅ 服务已启动，监听端口: 7066
📦 存储方式: Redis (192.144.142.2:26739)
🔄 数据更新策略: 实时更新（过期时自动获取）
🔒 安全措施: 速率限制、安全响应头、CORS 保护
========================================
```

---

## 常见问题

### Q: 为何Redis失败就fatal？

**A**: 此版本依赖Redis存储数据，没有Redis无法运行。因此直接退出。

### Q: 速率限制能否调整？

**A**: 可以，修改`Max`和`Expiration`：
```go
Max:        200,  // 每小时200次
Expiration: 1 * time.Hour,
```

### Q: CORS配置错误会怎样？

**A**: 程序会panic退出。现已动态判断，不会崩溃。

### Q: 能否修改端口？

**A**: 可以，修改`.env`：
```bash
PORT=8080
```

---

## 相关文件

- `server.go` - 服务器代码
- `../api/handlers.go` - API处理器
- `../config/config.go` - 配置管理
- `../storage/redis.go` - Redis初始化
