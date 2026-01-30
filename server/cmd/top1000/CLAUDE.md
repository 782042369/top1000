# cmd/top1000 - 应用入口

[根目录](../../CLAUDE.md) > **cmd/top1000**

## 模块快照

**职责**：应用程序入口，环境变量加载与启动引导

**入口文件**：`main.go`

**关键操作**：
- 加载 `.env` 环境变量文件（非必需）
- 启动 Fiber 服务器

## 代码结构

```go
func main() {
    _ = godotenv.Load()  // 加载 .env（可选）
    server.Start()        // 启动服务器
}
```

## 依赖关系

- `github.com/joho/godotenv` - 环境变量加载
- `internal/server` - 服务器启动逻辑

## 测试

无测试文件。

**建议**：添加启动流程测试。

---

*文档生成时间：2026-01-28 13:08:52*
