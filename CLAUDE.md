# Top1000 项目文档

> PT站点资源追踪系统（小型个人项目，日请求约100次）
>
> Docker镜像：4.5-5MB（Scratch基础镜像）

---

## 项目简介

**Top1000** 是一个 PT（Private Tracker）站点资源追踪系统，专为小型个人项目设计，具有以下特点：

1. **极简架构**：去除过度设计，直接读Redis
2. **容错机制**：爬取失败时返回Redis旧数据，保证服务可用
3. **自动更新**：数据过期时自动爬取，无需定时任务
4. **热重载开发**：使用Air实现Go代码热重载，提升开发体验
5. **极简部署**：单Docker容器即可运行，镜像仅4.5-5MB

---

## 技术架构

### 系统架构图

```
┌─────────────────────────────────────┐
│     Docker容器（端口7066）            │
│                                     │
│  ┌───────────────────────────────┐  │
│  │   Go后端（Fiber框架）         │  │
│  │   • /top1000.json - 数据接口  │  │
│  │   • 静态文件服务              │  │
│  └───────────────────────────────┘  │
│             ↓                        │
│  ┌───────────────────────────────┐  │
│  │   前端（AG Grid表格）         │  │
│  │   • 显示1000个资源            │  │
│  │   • 序号列、过滤、排序        │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│     Redis（数据存储）                │
│  • 数据永久存储（不设置TTL）          │
│  • 24小时内算新鲜                   │
│  • 过期时自动拉新数据               │
└─────────────────────────────────────┘
             ↓
┌─────────────────────────────────────┐
│     IYUU API（数据源）               │
│  api.iyuu.cn/top1000.php            │
└─────────────────────────────────────┘
```

### 数据流

```
用户访问前端页面
    ↓
前端调用 /top1000.json
    ↓
后端检查数据是否过期（基于time字段）
    ↓
数据过期？
├─ 是 → 爬取新数据
│   ├─ 成功 → 更新Redis，返回新数据
│   └─ 失败 → 返回Redis旧数据（容错）
└─ 否 → 直接返回Redis数据
```

---

## 模块结构

```mermaid
graph TD
    Root["(根) Top1000 项目<br/>小型个人项目，日请求100次"] --> CMD["cmd/top1000<br/>程序入口"]
    Root --> Internal["internal<br/>核心业务逻辑"]
    Root --> Web["web<br/>前端应用"]

    Internal --> Config["config<br/>配置管理"]
    Internal --> Server["server<br/>HTTP服务器"]
    Internal --> API["api<br/>API处理层"]
    Internal --> Crawler["crawler<br/>数据爬取"]
    Internal --> Storage["storage<br/>Redis存储"]
    Internal --> Model["model<br/>数据模型"]

    CMD -->|启动| Server
    Server -->|使用| Config
    Server -->|注册路由| API
    API -->|调用| Crawler
    API -->|读取| Storage
    Crawler -->|解析| Model
    Crawler -->|存储| Storage
    Storage -->|验证| Model

    Web -->|调用| API

    click CMD "cmd/top1000/CLAUDE.md" "查看程序入口文档"
    click Config "internal/config/CLAUDE.md" "查看配置管理文档"
    click Server "internal/server/CLAUDE.md" "查看服务器文档"
    click API "internal/api/CLAUDE.md" "查看API层文档"
    click Crawler "internal/crawler/CLAUDE.md" "查看爬虫文档"
    click Storage "internal/storage/CLAUDE.md" "查看存储层文档"
    click Model "internal/model/CLAUDE.md" "查看数据模型文档"
    click Web "web/CLAUDE.md" "查看前端文档"
```

---

## 模块索引

| 模块路径 | 模块名称 | 语言 | 代码行数 | 职责 | 文档链接 |
|---------|---------|------|---------|------|---------|
| `cmd/top1000` | 程序入口 | Go | 18行 | 加载环境变量、启动服务器 | [查看](cmd/top1000/CLAUDE.md) |
| `internal/config` | 配置管理 | Go | 120行 | 从环境变量读取配置、启动时验证 | [查看](internal/config/CLAUDE.md) |
| `internal/model` | 数据模型 | Go | 75行 | 定义数据结构、提供数据验证 | [查看](internal/model/CLAUDE.md) |
| `internal/api` | API处理层 | Go | 103行 | 处理HTTP请求、容错机制 | [查看](internal/api/CLAUDE.md) |
| `internal/storage` | Redis存储 | Go | 183行 | 管理Redis连接、TTL管理 | [查看](internal/storage/CLAUDE.md) |
| `internal/crawler` | 数据爬取 | Go | 199行 | 从IYUU API获取数据、解析文本 | [查看](internal/crawler/CLAUDE.md) |
| `internal/server` | HTTP服务器 | Go | 142行 | 配置Fiber应用、中间件和路由 | [查看](internal/server/CLAUDE.md) |
| `web` | 前端应用 | TypeScript | - | AG Grid表格展示、用户交互 | [查看](web/CLAUDE.md) |

---

## 目录结构

```
top1000/
├── cmd/top1000/          # 程序入口（18行）
│   ├── main.go           # 启动服务器
│   └── CLAUDE.md         # 模块文档
│
├── internal/             # 核心代码（Go）
│   ├── api/              # API处理（103行）
│   │   ├── handlers.go
│   │   └── CLAUDE.md
│   ├── config/           # 配置管理（120行）
│   │   ├── config.go
│   │   └── CLAUDE.md
│   ├── crawler/          # 爬虫（199行）
│   │   ├── scheduler.go
│   │   └── CLAUDE.md
│   ├── model/            # 数据结构（75行）
│   │   ├── types.go
│   │   └── CLAUDE.md
│   ├── server/           # HTTP服务器（142行）
│   │   ├── server.go
│   │   └── CLAUDE.md
│   └── storage/          # Redis存储（183行）
│       ├── redis.go
│       └── CLAUDE.md
│
├── web/                  # 前端（TypeScript + Vite）
│   ├── src/              # 源码
│   │   ├── main.ts       # 入口文件
│   │   ├── gridConfig.ts # 表格配置
│   │   ├── types.d.ts    # 类型定义
│   │   └── utils/        # 工具函数
│   ├── package.json
│   ├── vite.config.ts
│   └── CLAUDE.md
│
├── web-dist/             # 前端构建产物（Docker中使用）
├── .env                  # 环境变量（Redis密码等）
├── .env.example          # 环境变量模板
├── .air.toml             # Air热重载配置
├── Dockerfile            # Docker打包文件（Scratch版，4.5-5MB）
├── docker-compose.yaml   # Docker Compose配置
├── go.mod               # Go依赖
├── CLAUDE.md            # 本文档（根级文档）
└── .claude/
    └── index.json       # 项目索引文件
```

---

## 快速开始

### 环境要求

- Go 1.25.5+
- Node.js 24.3.0+（如果自己改前端的话）
- Redis 5.0+（**这个必须有，没Redis跑不起来**）
- Docker（可选，建议使用）
- Air（推荐，用于Go热重载）

### 配置环境变量

创建`.env`文件（参考`.env.example`）：

```bash
# Redis配置（必填，否则无法运行）
REDIS_ADDR=127.0.0.1:26739
REDIS_PASSWORD=填写Redis密码
```

### 本地开发

**方式一：使用Air热重载（推荐）**

```bash
# 1. 安装Air
go install github.com/cosmtrek/air@latest

# 2. 启动服务（代码变更自动重启）
air

# 3. 打开浏览器
open http://localhost:7066
```

**方式二：直接运行**
```bash
# 设置环境变量（Linux/Mac）
export $(cat .env | grep -v '^#' | xargs)

# 运行程序
go run cmd/top1000/main.go
```

### Docker部署（生产环境）

**方式一：使用 docker-compose（推荐）**

```bash
# 1. 配置环境变量（必须配置外部 Redis）
cp .env.example .env
# 编辑 .env 文件，修改 REDIS_ADDR 和 REDIS_PASSWORD

# 2. 启动服务（使用 Scratch 镜像，4.5-5MB）
docker-compose up -d

# 3. 查看日志
docker-compose logs -f top1000

# 4. 停止服务
docker-compose down
```

**方式二：使用Docker命令**

```bash
# 1. 构建镜像（Scratch 极简版）
docker build -t top1000:scratch .

# 2. 跑容器（需要外部 Redis）
docker run -d \
  --name top1000 \
  -p 7066:7066 \
  --env-file .env \
  top1000:scratch

# 3. 查看日志
docker logs -f top1000
```

---

## 技术栈

### 后端

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.25.5 |
| 框架 | Fiber | v2.52.10 |
| 数据库 | Redis | 5.0+ |
| 依赖管理 | go.mod | - |

### 前端

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | TypeScript | 5.9.3 |
| 框架 | Vite | 8.0.0-beta.5 |
| UI库 | AG Grid Enterprise | 35.0.0 |
| 包管理器 | pnpm | 10.12.4+ |

### 部署

| 组件 | 技术 | 版本 |
|------|------|------|
| 容器 | Docker | - |
| 基础镜像 | Scratch | 4.5-5MB（UPX压缩） |
| 端口 | - | 7066 |
| CI/CD | GitHub Actions | - |

### 开发工具

| 工具 | 用途 |
|------|------|
| Air | Go热重载 |
| pnpm | 前端包管理 |
| ESLint | 代码检查 |
| Prettier | 代码格式化 |

---

## 常见问题

### Q: 程序启动失败，报Redis连接错误？

**A**: 检查`.env`文件，确认`REDIS_ADDR`和`REDIS_PASSWORD`是否正确。

### Q: 数据多久更新一次？

**A**: 根据数据time字段判断，24小时内算新鲜数据，过期后自动获取新数据。

### Q: Docker镜像有多大？

**A**: Scratch版：4.5-5MB（UPX压缩，极简版）

### Q: 能否不使用Redis？

**A**: 不能。此版本专为Redis设计，不使用Redis需要修改代码。

### Q: 为什么没有测试？

**A**: 小型个人项目（日访问100次），不追求高测试覆盖。核心逻辑清晰，易于理解和维护。

### Q: Air热重载不生效？

**A**: 确保安装了Air：
```bash
go install github.com/cosmtrek/air@latest
```

### Q: 爬取失败会影响服务吗？

**A**: 不会！爬取失败时会返回Redis旧数据，保证服务可用（容错机制）。

### Q: 如何清理Redis数据？

**A**:
```bash
redis-cli -h <host> -p <port> -a <password>
> DEL top1000:data
```
删除后，下次访问会自动触发更新获取新数据。

---

## 外部依赖

- **Redis**（必须有）: 数据存储 + 过期检测
  - 连接池：3个连接
  - 存储策略：永久存储（不设置TTL）
  - 更新检测：基于数据time字段，24小时阈值

- **IYUU API**: `https://api.iyuu.cn/top1000.php`
  - 超时：30秒
  - 重试：1次（小项目简化）
  - 更新策略：按需更新（过期才拉）

---

## 核心配置文件

- `go.mod` / `go.sum` - Go依赖管理
- `.env.example` - 环境变量模板（复制这个改成`.env`）
- `docker-compose.yaml` - Docker Compose配置
- `.air.toml` - Air热重载配置
- `Dockerfile` - Scratch 极简版（4.5-5MB）
- `web/package.json` - npm依赖
- `web/vite.config.ts` - Vite构建配置
- `web/index.html` - HTML入口
- `web/src/gridConfig.ts` - 表格配置

---

**更新时间**: 2026-01-20
**Docker镜像**: 4.5-5MB
**文档覆盖率**: 100%
