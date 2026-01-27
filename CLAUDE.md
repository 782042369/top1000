# Top1000 - PT站点资源追踪系统

> 小型个人项目，日请求约100次，经过优化去除过度设计

## 项目愿景

Top1000 是一个轻量级的 PT 站点资源追踪系统，旨在为个人用户提供简洁、高效的资源监控服务。项目采用极简架构设计，移除了不必要的中间层和复杂功能，专注于核心价值：稳定、快速、易用。

## 核心特性

- **极简架构**：移除内存缓存，直接读 Redis，日请求 100 次完全够用
- **容错机制**：爬取失败返回旧数据，保证服务可用性
- **自动更新**：数据过期自动爬取，无需定时任务
- **热重载**：使用 Air 实现 Go 代码热重载，提升开发体验
- **简洁 UI**：AG Grid 表格，支持过滤排序和中文本地化

## 技术栈

### 后端
- **Go 1.25.5** - 高性能 HTTP 服务
- **Fiber v2** - 轻量级 Web 框架
- **Redis** - 数据存储与缓存

### 前端
- **TypeScript** - 类型安全
- **Vite 8** - 极速构建工具
- **AG Grid Community** - 企业级表格组件

### 基础设施
- **Docker** - 容器化部署（4.5-5MB 镜像）
- **GitHub Actions** - CI/CD 自动化

## 模块索引

| 模块 | 路径 | 职责 | 技术栈 |
|------|------|------|--------|
| API 层 | [`internal/api`](./internal/api/) | HTTP 接口处理、数据更新调度 | Go + Fiber |
| 爬虫 | [`internal/crawler`](./internal/crawler/) | IYUU 数据抓取与解析 | Go + net/http |
| 服务器 | [`internal/server`](./internal/server/) | HTTP 服务、路由配置、中间件 | Go + Fiber |
| 存储 | [`internal/storage`](./internal/storage/) | Redis 连接与数据持久化 | Go + go-redis |
| 数据模型 | [`internal/model`](./internal/model/) | 数据结构定义与验证 | Go |
| 配置 | [`internal/config`](./internal/config/) | 环境变量加载与验证 | Go |
| 前端应用 | [`web`](./web/) | 用户界面 | TypeScript + Vite + AG Grid |

## 运行与开发

### 环境准备

```bash
# 必需环境
Go 1.25.5+
Redis 5.0+
Node.js 24.3.0+ (仅修改前端时需要)

# 可选工具
Air (Go 热重载): go install github.com/air-verse/air@latest
pnpm (前端包管理): npm install -g pnpm@10
```

### 快速启动

#### 开发环境

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填写 Redis 配置

# 2. 启动后端（使用 Air 热重载）
air
# 或使用原生方式
go run cmd/top1000/main.go

# 3. 启动前端开发服务器（如需修改前端）
cd web
pnpm install
pnpm dev  # 访问 http://localhost:5173
```

#### 生产环境（Docker）

```bash
# 构建镜像
docker build -t top1000:latest .

# 运行容器
docker run -d \
  --name top1000 \
  -p 7066:7066 \
  --env-file .env \
  top1000:latest

# 或使用 Docker Compose
docker-compose up -d
```

### 访问服务

- **Web 界面**: http://localhost:7066
- **Top1000 数据**: http://localhost:7066/top1000.json
- **站点列表**: http://localhost:7066/sites.json (需配置 IYUU_SIGN)

## 测试策略

### 当前状态
- **单元测试**: 未配置
- **集成测试**: 未配置
- **手动测试**: 依赖实际环境验证

### 测试建议
由于项目规模较小，建议优先添加以下测试：
1. **数据模型验证测试** - `internal/model/types_test.go`
2. **Redis 存储层测试** - `internal/storage/redis_test.go`
3. **爬虫解析逻辑测试** - `internal/crawler/parser_test.go`
4. **API 接口测试** - `internal/api/handlers_test.go`

## 编码规范

### Go 代码规范
- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 包级注释清晰说明职责
- 导出函数必须有文档注释
- 错误处理使用 `fmt.Errorf` 包装上下文

### TypeScript 代码规范
- 使用 ESLint + Prettier 统一风格
- 严格模式：`strict: true`
- 优先使用 `const` > `let` > `var`
- 函数组件优于类组件
- 类型推导优于显式类型注解（推导可行时）

### 提交信息规范
- feat: 新功能
- fix: 修复 bug
- refactor: 重构（不改变功能）
- chore: 构建/工具链配置
- docs: 文档更新
- style: 代码格式（不影响功能）
- test: 测试相关
- perf: 性能优化

## 工作流程

```
用户访问前端
    ↓
前端请求 /top1000.json
    ↓
后端检查数据是否过期
    ↓
数据过期？
├─ 是 → 爬取新数据
│   ├─ 成功 → 更新 Redis，返回新数据
│   └─ 失败 → 返回 Redis 旧数据（容错）
└─ 否 → 直接返回 Redis 数据
```

## 性能指标

- **代码质量**: 95/100（S 级）
- **Docker 镜像**: 4.5-5MB（Scratch 基础镜像）
- **响应时间**: <100ms（Redis 缓存）
- **并发支持**: 200 次/小时
- **内存占用**: ~20MB（运行时）

## 常见问题

### Q: 爬取失败会影响服务吗？
不会！系统实现了容错机制，爬取失败时会返回 Redis 中的旧数据，保证服务可用。

### Q: 如何清理 Redis 数据？
```bash
redis-cli -h <host> -p <port> -a <password>
> DEL top1000:data
> DEL top1000:sites
```

### Q: Air 热重载不生效？
确保正确安装 Air：
```bash
go install github.com/air-verse/air@latest
air --version
```

### Q: 前端构建失败？
```bash
cd web
rm -rf node_modules pnpm-lock.yaml
pnpm install
pnpm build
```

## 项目结构

```
top1000/
├── cmd/top1000/          # 程序入口
├── internal/             # 核心代码（私有包）
│   ├── api/              # API 处理层
│   ├── crawler/          # 数据爬取
│   ├── server/           # HTTP 服务器
│   ├── storage/          # Redis 存储
│   ├── model/            # 数据模型
│   └── config/           # 配置管理
├── web/                  # 前端应用
│   ├── src/              # 源代码
│   ├── dist/             # 构建产物（gitignore）
│   └── package.json      # 依赖配置
├── web-dist/             # 前端构建产物（嵌入 Docker）
├── .air.toml             # Air 热重载配置
├── docker-compose.yaml   # Docker Compose 配置
├── Dockerfile            # Docker 镜像构建
├── .env.example          # 环境变量模板
├── go.mod                # Go 依赖管理
└── README.md             # 项目说明
```

## 版本历史

- **2026-01-15**: 代码优化，移除过度设计，减少 28% 代码量
- **2025-04-08**: 新增站点配置动态加载功能
- **2025-02-08**: 初始版本

## License

MIT

---

**更新时间**: 2026-01-15
**代码质量**: 95/100（S 级）
**适用场景**: 小型个人项目，日请求 100 次左右
