# 数据爬取

> 从IYUU API获取数据并解析

---

## 模块功能

**调用IYUU API，获取Top1000数据，解析后存入Redis**

核心功能：
1. 发送HTTP请求到IYUU API
2. 解析返回的文本数据
3. 转换成结构化数据
4. 存入Redis（永久存储，不设置TTL）
5. **启动时预加载数据**

---

## 核心函数

### FetchTop1000 - 获取数据

```go
func FetchTop1000() (*model.ProcessedData, error)
```

**功能**：
1. 加锁（防止并发更新）
2. 调用doFetch()执行
3. 解锁

**特点**：
- TryLock（不会阻塞）
- 失败记录详细日志

### PreloadData - 启动时预加载 ⭐

```go
func PreloadData()
```

**功能**：
1. 检查Redis中是否已有数据
2. 检查数据是否过期（基于time字段）
3. 如果有数据且未过期，跳过预加载
4. 如果没有数据或已过期，调用API获取新数据
5. 存入Redis

**优点**：
- **避免首次访问超时**：服务启动时数据就准备好了
- **提升用户体验**：用户访问时数据已经在Redis中
- **容错机制**：预加载失败不影响服务启动，首次访问时自动重试

**使用场景**：
- 程序启动时自动调用
- 首次部署时自动获取数据
- 数据过期时自动更新

### doFetch - 执行请求

```go
func doFetch() error
```

**流程**：
```go
1. 创建HTTP请求（30秒超时）
2. 发送到IYUU API
3. 检查状态码（200 OK）
4. 读响应体
5. 调用processData()解析
6. 存到Redis
```

### processData - 解析数据

```go
func processData(rawData string) model.ProcessedData
```

**数据格式**：
```
create time 2025-12-11 07:52:33 by https://api.iyuu.cn/

站名：朋友 【ID：123456】
重复度：95%
文件大小：1.5GB

站名：馒头 【ID：789012】
重复度：87%
文件大小：2.3GB

...
```

**解析规则**：
- 第1行：时间
- 第3行开始：每3行一条数据
- 正则提取：`站名：(.*?) 【ID：(\d+)】`
- 分割提取：重复度、文件大小

---

## 常量定义

```go
const (
    httpTimeout     = 30 * time.Second  // HTTP超时
    linesPerItem    = 3                 // 每3行一条数据
    timeLineIndex   = 0                 // 时间行索引
    dataStartLineIndex = 2             // 数据起始行
)
```

**设置原因**：
- 30秒超时：API响应慢时不继续等待
- 3行一组：数据格式固定

---

## 并发控制

### 防止并发更新

```go
var taskMutex sync.Mutex

func FetchData() error {
    if !taskMutex.TryLock() {
        return fmt.Errorf("任务正在执行中")
    }
    defer taskMutex.Unlock()
    // ...
}
```

**特点**：
- TryLock：不阻塞，直接返回
- 友好提示："任务正在执行中"
- API层也会检查IsUpdating()

---

## 数据验证

解析完成后验证：

```go
result := model.ProcessedData{
    Time:  parseTime(timeLine),
    Items: items,
}

// 验证一下
if err := result.Validate(); err != nil {
    log.Printf("⚠️ 数据验证失败: %v", err)
    // 爬虫容错，还是会返回
}
```

**注意**：
- 爬虫：验证失败记录警告，还是返回（容错）
- 存储：验证失败直接拒绝（不保存）

---

## 正则表达式

```go
siteRegex = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)
```

**匹配示例**：
```
输入: "站名：朋友 【ID：123456】"
匹配:
  - match[0]: "站名：朋友 【ID：123456】"
  - match[1]: "朋友"
  - match[2]: "123456"
```

---

## Context使用优化

### 新增带context的函数

- `FetchTop1000WithContext(ctx)` - 支持外部传入context
- `doFetchWithContext(ctx)` - 内部执行HTTP请求，使用传入的context
- `checkDataLoadRequired(ctx)` - 检查数据加载需求，支持context

**向后兼容**：
- 旧的函数（如`FetchTop1000()`）保持不变，内部创建默认context
- 新函数支持外部传入context，实现更好的超时控制和取消机制

### 使用示例

**方式一：使用默认超时**
```go
// 使用默认30秒超时
data, err := crawler.FetchTop1000()
```

**方式二：使用自定义context**
```go
// API层从Fiber的context提取
ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
defer cancel()

// 传递给crawler层
data, err := crawler.FetchTop1000WithContext(ctx)
```

---

## 常见问题

### Q: 为何没有定时任务？

**A**: 已改为按需更新。数据time字段超过24小时时自动获取新数据。

### Q: 数据解析失败会怎样？

**A**: 跳过该条数据，记录警告。只要有一条成功就算成功。

### Q: 为何使用TryLock？

**A**: 防止并发更新。如果正在更新，直接返回错误，不阻塞。

---

## 依赖配置

### 环境变量

```go
cfg := config.Get()
url := cfg.Top1000APIURL  // https://api.iyuu.cn/top1000.php
```

### 超时配置

```go
ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
defer cancel()
```

---

## 相关文件

- `scheduler.go` - 爬虫代码
- `../storage/redis.go` - 数据存储
- `../config/config.go` - 配置管理
- `../model/types.go` - 数据结构
