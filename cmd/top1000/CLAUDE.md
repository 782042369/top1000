# 程序入口

> [根目录](../CLAUDE.md) > **cmd/top1000**

---

## 模块职责

**程序的入口点，负责初始化配置和启动服务器**

核心功能：
1. 加载 `.env` 环境变量文件
2. 验证必需的环境变量（Redis配置）
3. 启动 HTTP 服务器

---

## 入口文件

### main.go

```go
package main

func main() {
    // 1. 加载 .env 文件
    godotenv.Load()

    // 2. 检查必需环境变量
    requiredEnvs := []string{"REDIS_ADDR", "REDIS_PASSWORD"}

    // 3. 启动服务器
    server.StartWatcher()
}
```

---

## 启动流程

```
main.go 启动
    ↓
1. godotenv.Load()
   ├─ 成功 → 加载 .env 文件
   └─ 失败 → 使用系统环境变量
    ↓
2. 检查必需环境变量
   ├─ REDIS_ADDR（Redis地址）
   └─ REDIS_PASSWORD（Redis密码）
    ↓
3. 缺少必需变量？
   ├─ 是 → fatal 退出，提示配置错误
   └─ 否 → 继续
    ↓
4. server.StartWatcher()
   ├─ 验证配置
   ├─ 初始化 Redis
   ├─ 启动 HTTP 服务器
   └─ 监听端口 7066
```

---

## 环境变量检查

### 必需的环境变量

```go
requiredEnvs := []string{
    "REDIS_ADDR",      // Redis地址
    "REDIS_PASSWORD",  // Redis密码
}
```

**为什么必须？**
- 此版本完全依赖 Redis 存储数据
- 没有配置则无法运行
- 因此启动时强制检查

### 检查逻辑

```go
missingEnvs := []string{}
for _, env := range requiredEnvs {
    if os.Getenv(env) == "" {
        missingEnvs = append(missingEnvs, env)
    }
}

if len(missingEnvs) > 0 {
    log.Fatalf("❌ 缺少必需的环境变量: %v", missingEnvs)
}
```

**特点**：
- 收集所有缺失的变量
- 一次性提示所有错误
- 友好的错误提示

---

## 错误处理

### .env 文件加载失败

```
⚠️ 警告: 无法加载 .env 文件: open .env: no such file or directory
🔧 将使用系统环境变量
```

**处理方式**：
- 不中断程序
- 记录警告日志
- 尝试使用系统环境变量

### 缺少必需环境变量

```
❌ 缺少必需的环境变量: [REDIS_ADDR REDIS_PASSWORD]
请检查 .env 文件或系统环境变量配置
```

**处理方式**：
- 直接 fatal 退出
- 列出所有缺失的变量
- 提示配置方法

---

## 相关文件

### 源代码

- `main.go` - 入口文件（18行）

### 配置文件

- `.env` - 环境变量（不提交到 Git）
- `.env.example` - 环境变量模板

### 依赖模块

- `../internal/server` - HTTP 服务器模块
- `../internal/config` - 配置管理模块

---

## 常见问题

### Q: 为何使用 godotenv 而不是直接读取环境变量？

**A**: godotenv 提供：
- 方便的开发体验（`.env` 文件）
- 灵活的生产配置（系统环境变量）
- 不影响容器化部署（覆盖机制）

### Q: .env 文件不存在会怎样？

**A**: 程序不会退出，会：
- 记录警告日志
- 尝试使用系统环境变量
- 如果系统环境变量也没有，才会在后续检查中退出

### Q: 为何不在这里验证配置？

**A**: 职责分离：
- `main.go` 只检查**必需变量是否存在**
- `config.Validate()` 检查**配置值的有效性**
- 避免入口文件过于复杂

---

**更新时间**: 2026-01-20
**代码行数**: 18行
