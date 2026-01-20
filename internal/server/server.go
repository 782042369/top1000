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

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	appName          = "Top1000"
	requestBodyLimit = 4 * 1024 * 1024 // 4MB
	oneYearMaxAge    = "public, max-age=31536000"
	noCache          = "no-cache, no-store, must-revalidate"
	cspHeader        = "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; " +
		"img-src 'self' data: https: https://lsky.939593.xyz:11111; " +
		"style-src 'self' 'unsafe-inline'; " +
		"connect-src 'self' https://log.939593.xyz;"
)

// Start 启动Web服务器
func Start() {
	cfg := config.Get()

	// 验证配置
	if err := config.Validate(); err != nil {
		log.Fatalf("❌ 配置验证失败: %v", err)
	}

	printStartupBanner()

	app := createApp()
	initStorage()
	preloadData() // 启动时预加载数据
	printStartupInfo(cfg)

	// 确保程序退出时关闭Redis连接
	defer closeRedis()

	log.Fatal(app.Listen(":" + config.DefaultPort))
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
	app.Get("/top1000.json", api.GetTop1000Data)

	app.Static("/", config.DefaultWebDistDir, fiber.Static{
		CacheDuration:  0, // Fiber内部缓存禁用，完全由ModifyResponse自定义
		Browse:         true,
		MaxAge:         0,
		ModifyResponse: setCacheHeaders,
	})
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
	log.Println("🔌 正在初始化Redis连接...")
	if err := storage.InitRedis(); err != nil {
		log.Fatalf("❌ Redis初始化失败: %v", err)
	}
	log.Println("✅ Redis初始化成功")
}

// closeRedis 关闭Redis连接
func closeRedis() {
	log.Println("🔌 正在关闭Redis连接...")
	if err := storage.CloseRedis(); err != nil {
		log.Printf("❌ 关闭Redis连接失败: %v", err)
	} else {
		log.Println("✅ Redis连接已关闭")
	}
}

// printStartupBanner 打印启动横幅
func printStartupBanner() {
	log.Println(strings.Repeat("=", 40))
	log.Println("   Top1000 服务正在启动...")
	log.Println(strings.Repeat("=", 40))
}

// printStartupInfo 打印启动信息
func printStartupInfo(cfg *config.Config) {
	log.Println(strings.Repeat("=", 40))
	log.Printf("✅ 服务已启动，监听端口: %s", config.DefaultPort)
	log.Printf("📦 存储方式: Redis (%s)", cfg.RedisAddr)
	log.Println("🔄 数据更新策略: 过期自动更新（容错机制）")
	log.Println("🔒 安全措施: 速率限制、安全响应头")
	log.Println(strings.Repeat("=", 40))
}

// preloadData 启动时预加载数据
func preloadData() {
	log.Println(strings.Repeat("=", 40))
	crawler.PreloadData()
	log.Println(strings.Repeat("=", 40))
}
