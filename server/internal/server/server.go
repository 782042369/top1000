package server

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"top1000/internal/api"
	"top1000/internal/config"
	"top1000/internal/crawler"
	"top1000/internal/storage"

	docs "top1000/docs" // Swagger docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	appName          = "Top1000"
	requestBodyLimit = 4 * 1024 * 1024
	oneYearMaxAge    = "public, max-age=31536000"
	noCache          = "no-cache, no-store, must-revalidate"
	separatorLength  = 40
)

// Start 启动Web服务器
func Start() {
	cfg := config.Get()

	if err := config.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	printSeparator()
	log.Println("   Top1000 服务正在启动...")
	printSeparator()

	initStorage()
	app := createApp()
	preloadData()
	printStartupInfo(cfg)

	if err := app.Listen(":" + config.DefaultPort); err != nil {
		log.Printf("服务启动失败: %v", err)
		closeRedis()
		return
	}

	closeRedis()
}

// createApp 创建Fiber应用并配置中间件和路由
func createApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      appName,
		StrictRouting: true,
		BodyLimit:    requestBodyLimit,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	setupMiddleware(app)
	setupRoutes(app)

	return app
}

// setupMiddleware 配置中间件
func setupMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(loggerMiddleware())
	app.Use(securityHeadersMiddleware())
	app.Use(compress.New())
}

// loggerMiddleware 日志中间件
func loggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Printf("[%s] %s %s - %d - %v",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			time.Since(start),
		)
		return err
	}
}

// securityHeadersMiddleware 安全响应头中间件
func securityHeadersMiddleware() fiber.Handler {
	cspHeader := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; " +
		"img-src 'self' data: https: https://lsky.939593.xyz:11111; " +
		"style-src 'self' 'unsafe-inline'; " +
		"connect-src 'self' https://log.939593.xyz;"

	return func(c *fiber.Ctx) error {
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Content-Security-Policy", cspHeader)
		return c.Next()
	}
}

// setupRoutes 配置路由
func setupRoutes(app *fiber.App) {
	// 创建 Handler 实例（依赖注入）
	// RedisStore 实现了所有三个接口（DataStore、SitesStore、UpdateLock）
	handler := api.NewHandler(
		storage.GetDefaultStore(),
		storage.GetDefaultSitesStore(),
		storage.GetDefaultLock(),
	)

	// 注册 Handler 的路由
	handler.RegisterRoutes(app)

	// Swagger UI
	app.Get("/swagger/*", swaggerUI)
	app.Get("/swagger/doc.json", swaggerJSON)

	// 静态文件
	app.Static("/", config.DefaultWebDistDir, fiber.Static{
		CacheDuration:  0, // Fiber内部缓存禁用，完全由ModifyResponse自定义
		Browse:         true,
		MaxAge:         0,
		ModifyResponse: setCacheHeaders,
	})
}

// swaggerUI 返回 Swagger UI HTML
func swaggerUI(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>Top1000 API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; padding: 0; font-family: "Helvetica Neue", Helvetica, Arial, sans-serif; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js" charset="UTF-8"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            })
            window.ui = ui
        }
    </script>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send([]byte(html))
}

// swaggerJSON 返回 Swagger JSON 文档
func swaggerJSON(c *fiber.Ctx) error {
	return c.JSON(docs.SwaggerInfo)
}

// setCacheHeaders 设置静态文件缓存头
func setCacheHeaders(c *fiber.Ctx) error {
	path := c.Path()
	isHTML := filepath.Ext(path) == ".html" || path == "/"

	if !isHTML && c.Response().StatusCode() == fiber.StatusOK {
		c.Response().Header.Set("Cache-Control", oneYearMaxAge)
		return nil
	}

	// HTML文件或错误状态:禁止缓存
	c.Response().Header.Set("Cache-Control", noCache)
	c.Response().Header.Set("Pragma", "no-cache")
	c.Response().Header.Set("Expires", "0")
	return nil
}

// initStorage 初始化Redis连接
func initStorage() {
	log.Println("正在初始化Redis连接...")
	if err := storage.InitRedis(); err != nil {
		log.Fatalf("Redis初始化失败: %v", err)
	}
	log.Println("Redis初始化成功")
}

// closeRedis 关闭Redis连接
func closeRedis() {
	log.Println("正在关闭Redis连接...")
	if err := storage.CloseRedis(); err != nil {
		log.Printf("关闭Redis连接失败: %v", err)
	} else {
		log.Println("Redis连接已关闭")
	}
}

// printSeparator 打印分隔线
func printSeparator() {
	log.Println(strings.Repeat("=", separatorLength))
}

// printStartupInfo 打印启动信息
func printStartupInfo(cfg *config.Config) {
	printSeparator()
	log.Printf("服务已启动，监听端口: %s", config.DefaultPort)
	log.Printf("存储方式: Redis (%s)", cfg.RedisAddr)
	log.Println("数据更新策略: 过期自动更新（容错机制）")
	log.Println("安全措施: 速率限制、安全响应头")
	printSeparator()
}

// preloadData 启动时预加载数据
func preloadData() {
	printSeparator()
	crawler.PreloadData()
	printSeparator()
}
