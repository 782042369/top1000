# Cmd/Top1000 模块

[根目录](../../CLAUDE.md) > [cmd](../) > **top1000**

## 模块职责

Cmd/Top1000 模块是应用程序的入口点，负责环境变量加载和应用启动。它是整个程序的启动引导器。

## 入口与启动

- **入口文件**: `main.go`
- **入口函数**: `main()`
- **调用链**: `main()` → `server.Start()`

## 对外接口

### 导出函数

```go
// 无（main 包不导出函数）
```

### 内部函数

```go
// 主函数
func main()
```

## 关键依赖与配置

### 依赖模块

- `github.com/joho/godotenv` - .env 文件加载（非强制）
- `top1000/internal/server` - HTTP 服务器启动
- `log` - 日志输出

### 环境变量

| 变量 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `.env` 文件 | 否 | 无 | 环境变量文件（不强制） |

## 核心逻辑

### 启动流程

```
main()
    ↓
加载 .env 文件（非必需）
    ├─ 成功 → 环境变量已加载
    └─ 失败 → 使用系统环境变量（容错）
    ↓
记录日志 "✅ 环境变量已加载"
    ↓
启动服务器 (server.Start())
    ├─ 验证配置 (config.Validate)
    ├─ 初始化 Redis (storage.InitRedis)
    ├─ 预加载数据 (crawler.PreloadData)
    └─ 监听端口 :7066（阻塞）
```

### 代码实现

```go
package main

import (
    "log"

    "github.com/joho/godotenv"
    "top1000/internal/server"
)

func main() {
    // 加载 .env 文件（非必需，失败时使用系统环境变量）
    _ = godotenv.Load()

    log.Println("✅ 环境变量已加载")

    server.Start()
}
```

### 容错设计

**godotenv.Load() 错误处理**：
- 使用 `_` 忽略错误
- .env 文件不存在时，自动使用系统环境变量
- 不影响程序启动

**优点**：
- Docker 部署时不需要 .env 文件（使用环境变量）
- 本地开发时可以使用 .env 文件（方便）
- 提高部署灵活性

## 测试与质量

### 当前状态
- 无单元测试（main 包通常不测试）
- 依赖实际环境验证

### 测试建议

由于 `main()` 函数通常不进行单元测试，建议：

1. **集成测试** - 测试完整启动流程
2. **端到端测试** - 启动程序后验证服务可用

示例集成测试：
```go
// cmd/top1000/integration_test.go
func TestAppStartup(t *testing.T) {
    // 设置测试环境变量
    os.Setenv("REDIS_ADDR", "localhost:6379")
    os.Setenv("REDIS_PASSWORD", "testpassword")

    // 启动服务器（在 goroutine 中）
    go server.Start()

    // 等待服务器启动
    time.Sleep(2 * time.Second)

    // 测试 HTTP 请求
    resp, err := http.Get("http://localhost:7066/top1000.json")
    if err != nil {
        t.Fatalf("请求失败: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("期望状态码 200，得到 %d", resp.StatusCode)
    }
}
```

## 相关文件清单

### 核心文件
- `main.go` - 程序入口（18 行）
  - `main()` - 主函数

### 测试文件（待创建）
- `integration_test.go` - 集成测试

### 依赖文件
- `../../internal/server/server.go` - 服务器启动
- `../../internal/config/config.go` - 配置管理

## 最佳实践

### 1. 环境变量管理

**开发环境**（使用 .env 文件）：
```bash
# .env
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=localpassword
```

**生产环境**（使用环境变量）：
```bash
# Docker
docker run -e REDIS_ADDR=redis:6379 -e REDIS_PASSWORD=prodpassword ...

# Kubernetes
env:
  - name: REDIS_ADDR
    value: "redis-service:6379"
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: redis-secret
        key: password
```

### 2. 日志配置

考虑在 main.go 中配置日志：
```go
import (
    "log"
    "os"
)

func main() {
    // 配置日志输出
    log.SetOutput(os.Stdout)
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    _ = godotenv.Load()
    log.Println("✅ 环境变量已加载")

    server.Start()
}
```

### 3. 优雅关闭

实现优雅关闭机制：
```go
func main() {
    _ = godotenv.Load()

    // 创建服务器
    app := server.CreateApp()

    // 监听信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    // 启动服务器（在 goroutine 中）
    go func() {
        if err := app.Listen(":7066"); err != nil {
            log.Fatalf("服务器启动失败: %v", err)
        }
    }()

    // 等待信号
    <-sigCh
    log.Println("⏸️ 收到关闭信号，正在优雅关闭...")

    // 关闭服务器
    app.Shutdown()
    log.Println("✅ 服务器已关闭")
}
```

## 常见问题

### Q: 为什么不检查 godotenv.Load() 错误？
- .env 文件是可选的，不存在时使用系统环境变量
- Docker 部署通常不使用 .env 文件
- 忽略错误提高部署灵活性

### Q: 如何传递命令行参数？
当前版本不支持命令行参数，全部使用环境变量。如需支持：
```go
import "flag"

func main() {
    port := flag.String("port", "7066", "HTTP 服务端口")
    flag.Parse()

    os.Setenv("PORT", *port)
    // ...
}
```

### Q: 如何调试启动问题？
1. 检查环境变量是否正确加载
2. 查看 `server.Start()` 的日志输出
3. 验证 Redis 连接是否成功
4. 检查端口 7066 是否被占用

### Q: 如何修改日志级别？
当前版本使用标准 `log` 包，不支持日志级别。建议：
- 使用 `logrus` 或 `zap` 替换
- 或保持简单，依赖日志内容区分

### Q: 如何禁用 .env 文件加载？
注释 `godotenv.Load()` 调用：
```go
// _ = godotenv.Load()
```

## 扩展建议

### 可能的扩展

1. **命令行参数** - 支持端口、配置文件路径等
2. **版本信息** - 添加 `--version` 和 `--help`
3. **健康检查** - 添加 `/health` 端点
4. **指标暴露** - 添加 `/metrics` 端点（Prometheus）

示例（版本信息）：
```go
var (
    version   = "dev"
    buildTime = "unknown"
)

func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "--version":
            fmt.Printf("Top1000 version %s (build %s)\n", version, buildTime)
            return
        case "--help":
            fmt.Println("Usage: top1000 [--version] [--help]")
            return
        }
    }

    _ = godotenv.Load()
    server.Start()
}
```

---

**最后更新**: 2026-01-27
**代码行数**: ~18 行
**维护状态**: 稳定
