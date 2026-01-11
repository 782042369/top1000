package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"top1000/internal/server"
)

func main() {
	// 加载 .env 文件（非必需，失败时使用系统环境变量）
	_ = godotenv.Load()

	// 验证必需的环境变量
	if missing := checkRequiredEnvVars(); len(missing) > 0 {
		log.Fatalf("❌ 缺少必需的环境变量: %v\n请检查 .env 文件或系统环境变量配置", missing)
	}

	server.StartWatcher()
}

// checkRequiredEnvVars 检查必需的环境变量是否已设置
func checkRequiredEnvVars() []string {
	var missing []string

	for _, env := range []string{"REDIS_ADDR", "REDIS_PASSWORD"} {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	// 如果全部存在，记录日志提示来源
	if len(missing) == 0 {
		log.Println("✅ 必需环境变量检查通过")
	}

	return missing
}
