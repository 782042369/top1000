package server

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"top1000/internal/api"
	"top1000/internal/config"
	"top1000/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	appName            = "Top1000 Service"
	requestBodyLimit   = 4 * 1024 * 1024
	maxRequestsPerHour = 60
	corsMaxAge         = 86400 // 24小时
)

var corsOrigins = func() string {
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		return origins
	}
	return "*"
}()

// StartWatcher 启动Web服务器
func StartWatcher() {
	cfg := config.Get()

	if err := config.Validate(); err != nil {
		log.Fatalf("❌ 配置验证失败: %v", err)
	}

	printStartupBanner()

	app := fiber.New(fiber.Config{
		AppName:      appName,
		StrictRouting: true,
		BodyLimit:    requestBodyLimit,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	setupMiddleware(app)
	setupRoutes(app, cfg)
	initStorage()

	printStartupInfo(cfg)

	// 确保程序退出时关闭Redis连接
	defer func() {
		log.Println("🔌 正在关闭Redis连接...")
		if err := storage.CloseRedis(); err != nil {
			log.Printf("❌ 关闭Redis连接失败: %v", err)
		} else {
			log.Println("✅ Redis连接已关闭")
		}
	}()

	log.Fatal(app.Listen(":" + cfg.Port))
}

// setupMiddleware 配置中间件
func setupMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Shanghai",
	}))
	app.Use(corsMiddleware())
	app.Use(securityHeaders())
	app.Use(rateLimiter())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
}

// corsMiddleware CORS中间件
func corsMiddleware() fiber.Handler {
	allowCredentials := corsOrigins != "*"
	return cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length,ETag,Cache-Control",
		MaxAge:           corsMaxAge,
		AllowCredentials: allowCredentials,
	})
}

// securityHeaders 安全响应头
func securityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://log.939593.xyz; "+
				"img-src 'self' data: https: https://lsky.939593.xyz:11111; "+
				"style-src 'self' 'unsafe-inline'; "+
				"connect-src 'self' https://log.939593.xyz;")
		return c.Next()
	}
}

// rateLimiter 速率限制
func rateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequestsPerHour,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "请求过于频繁，请稍后再试",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// setupRoutes 配置路由
func setupRoutes(app *fiber.App, cfg *config.Config) {
	app.Get("/top1000.json", api.GetTop1000Data)
	app.Static("/", cfg.WebDistDir, fiber.Static{
		CacheDuration:  cfg.CacheDuration,
		Browse:         true,
		MaxAge:         0,
		ModifyResponse: staticFileCacheHeaders,
	})
}

// staticFileCacheHeaders 设置静态文件缓存头
func staticFileCacheHeaders(c *fiber.Ctx) error {
	const (
		oneYearMaxAge = "public, max-age=31536000"
		noCache       = "no-cache, no-store, must-revalidate"
	)

	path := c.Path()
	isHTML := filepath.Ext(path) == ".html" || path == "/"

	if !isHTML && c.Response().StatusCode() == fiber.StatusOK {
		c.Response().Header.Set("Cache-Control", oneYearMaxAge)
	} else {
		c.Response().Header.Set("Cache-Control", noCache)
		c.Response().Header.Set("Pragma", "no-cache")
		c.Response().Header.Set("Expires", "0")
	}
	return nil
}

// initStorage 初始化存储
func initStorage() {
	log.Println("正在初始化 Redis 连接...")
	if err := storage.InitRedis(); err != nil {
		log.Fatalf("❌ Redis 初始化失败: %v", err)
	}
	log.Println("✅ Redis 初始化成功")
}

// printStartupBanner 打印启动横幅
func printStartupBanner() {
	log.Println("========================================")
	log.Println("   Top1000 服务正在启动...")
	log.Println("========================================")
}

// printStartupInfo 打印启动信息
func printStartupInfo(cfg *config.Config) {
	log.Println("========================================")
	log.Printf("✅ 服务已启动，监听端口: %s", cfg.Port)
	log.Printf("📦 存储方式: Redis (%s)", cfg.RedisAddr)
	log.Println("🔄 数据更新策略: 实时更新（过期时自动获取）")
	log.Println("🔒 安全措施: 速率限制、安全响应头、CORS 保护")
	log.Println("========================================")
}
