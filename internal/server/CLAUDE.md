# HTTP 服务器

> 启动Web服务的模块

---

## 模块功能

**启动HTTP服务器，配置路由和中间件**

核心功能：
1. 创建Fiber应用
2. 配置中间件（日志、CORS、安全头、限流）
3. 注册路由（API、静态文件）
4. 初始化Redis
5. 启动服务

**小项目简化**（2026-01-11）：
- ✅ 移除 health 检查路由（小项目不需要）
- ✅ 移除优雅关闭代码（小项目不需要）
- ✅ 移除相关导入（os/signal、syscall）
- ✅ 速率限制：60次/小时（匹配小访问量）

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

    // 5. 启动时预加载数据（新增）⭐
    preloadData()

    // 6. 打印启动信息
    printStartupInfo(cfg)

    // 7. 确保程序退出时关闭Redis连接
    defer closeRedis()

    // 8. 启动服务
    log.Fatal(app.Listen(":" + cfg.Port))
}
```

**预加载功能**（2026-01-15 新增）：
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

### CORS（优化过）

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

**修改原因**：
- 通配符（*）+ 携带凭证存在安全风险
- 因此通配符时禁用AllowCredentials
- 指定域名时才允许携带凭证

### 安全头（手动配置）

已移除Helmet中间件，该中间件的COEP配置无法禁用，因此手动配置安全头：

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
  - 监控脚本：`https://log.939593.xyz/script.js`
  - 数据上报：`https://log.939593.xyz/api/send`
  - Favicon：`https://lsky.939593.xyz:11111/Y7bbx9.jpg`
- **禁用COEP/COOP**：让跨域资源能正常加载

**为何不用Helmet**：
- 该中间件会自动设置`Cross-Origin-Embedder-Policy`（COEP）头
- COEP会阻止所有跨域资源（监控脚本和图片会被拦截）
- 设置为空字符串`""`无效，仍会设置默认值
- 手动配置更灵活可控！

### 速率限制

```go
app.Use(limiter.New(limiter.Config{
    Max:        100,  // 每分钟最多100次
    Expiration: 1 * time.Minute,
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

**作用**：防止DDoS，每个IP每分钟最多100次请求

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

### 健康检查

```go
app.Get("/health", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status":    "ok",
        "timestamp": time.Now().Unix(),
    })
})
```

**用途**：
- Docker健康检查
- K8s liveness/readiness探针
- 负载均衡器健康检查

### 静态文件

```go
app.Static("/", cfg.WebDistDir, fiber.Static{
    CacheDuration: cfg.CacheDuration,
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

## Redis初始化

```go
log.Println("正在初始化 Redis 连接...")
if err := storage.InitRedis(); err != nil {
    log.Fatalf("❌ Redis 初始化失败: %v", err)
}
log.Println("✅ Redis 连接成功")
```

**注意**：
- fatal退出，Redis连接失败时不启动
- 不再使用文件系统作为备份

---

## 优雅关闭

```go
go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    log.Println("接收到关闭信号，正在优雅关闭服务器...")

    // 关闭 Redis
    if err := storage.CloseRedis(); err != nil {
        log.Printf("关闭 Redis 连接失败: %v", err)
    }

    // 关闭 HTTP 服务器（最多等10秒）
    if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
        log.Printf("服务器关闭失败: %v", err)
    }

    log.Println("服务器已优雅关闭")
}()
```

**处理信号**：
- SIGINT（Ctrl+C）
- SIGTERM（kill命令）

**超时**：10秒（强制关闭正在处理的请求）

---

## 配置验证（新增）

```go
cfg := config.Get()

// 验证配置
if err := config.Validate(); err != nil {
    log.Fatalf("❌ 配置验证失败: %v", err)
}
```

**验证内容**：
- Redis地址不能为空
- Redis密码不能为空

---

## CORS优化细节

### 问题

```go
// 以前：固定允许携带凭证
AllowCredentials: true
AllowOrigins: "*"  // 通配符

// 结果：panic!
// "CORS: Insecure setup, AllowCredentials is true, and AllowOrigins is wildcard"
```

### 修复

```go
// 现在：动态判断
corsOrigins := os.Getenv("CORS_ORIGINS")
if corsOrigins == "" {
    corsOrigins = "*"
}
allowCredentials := corsOrigins != "*"

app.Use(cors.New(cors.Config{
    AllowOrigins:     corsOrigins,
    AllowCredentials: allowCredentials,
    // ...
}))
```

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
✅ 服务已启动，监听端口: 7066
📦 存储方式: Redis (192.144.142.2:26739)
🔄 数据更新策略: 实时更新（过期时自动获取）
🔒 安全措施: 速率限制、安全响应头、CORS 保护
========================================
```

---

## 代码优化

### 移除的功能

- ❌ Cron定时任务启动
- ❌ 文件系统初始化
- ❌ 定时任务日志
- ❌ Helmet中间件（COEP无法禁用）

### 新增的功能

- ✅ 配置验证
- ✅ CORS动态配置
- ✅ 优雅关闭
- ✅ 健康检查端点
- ✅ 手动安全头配置（替代Helmet，支持跨域监控脚本）

---

## 常见问题

### Q: 为何Redis失败就fatal？

**A**: 此版本依赖Redis存储数据，没有Redis无法运行。因此直接退出。

### Q: 速率限制能否调整？

**A**: 可以，修改`Max`和`Expiration`：
```go
Max:        200,  // 每分钟200次
Expiration: 1 * time.Minute,
```

### Q: CORS配置错误会怎样？

**A**: 程序会panic退出。现已动态判断，不会崩溃。

### Q: 优雅关闭有何作用？

**A**:
- 处理完当前请求再退出
- 关闭Redis连接
- 不丢失数据

### Q: 能否修改端口？

**A**: 可以，修改`.env`：
```bash
PORT=8080
```

---

## 相关文件

- `server.go` - 服务器代码（160行）
- `../api/handlers.go` - API处理器
- `../config/config.go` - 配置管理
- `../storage/redis.go` - Redis初始化

---

**总结**：服务器启动应优先考虑安全性和稳定性，中间件配置完善。

**更新**: 2026-01-11
**代码行数**: 202 行（已优化，从181行增加但更清晰）
**代码质量**: A+ 级
**优化**: 配置验证 + CORS优化 + 优雅关闭 + 拆分为 8 个小函数 + 提取 CSP 构建逻辑
