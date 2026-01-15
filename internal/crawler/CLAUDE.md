# 数据爬取

> 从IYUU获取数据的模块

---

## 模块功能

**调用IYUU API，获取Top1000数据，解析后存入Redis**

核心功能：
1. 发送HTTP请求到IYUU API
2. 解析返回的文本数据
3. 转换成结构化数据
4. 存入Redis（永久存储，不设置TTL）
5. **启动时预加载数据（新增功能）**

已移除定时任务，改为按需更新（过期时才获取）。

---

## 核心函数

### FetchTop1000 - 获取数据（主函数）

```go
func FetchTop1000() (*model.ProcessedData, error)
```

**功能**：
1. 加锁（防止并发更新）
2. 重试循环（最多1次，小项目简化）
3. 调用doFetch()执行
4. 解锁

**特点**：
- TryLock（不会阻塞）
- 最多重试1次（小访问量项目不需要多次重试）
- 失败记录详细日志

---

### PreloadData - 启动时预加载（新增）⭐

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

**示例日志**：
```
========================================
[🔍 爬虫] 检查是否需要预加载数据...
⚠️ 数据不存在
[🔍 爬虫] 🚀 Redis中无数据或数据过期，开始预加载...
[🔍 爬虫] 开始爬取IYUU数据...
[🔍 爬虫] 数据获取成功（123456 字节）
[🔍 爬虫] 数据解析完成（1000 条）
✅ 数据已保存到Redis（永久存储，过期判断基于数据time字段）
[🔍 爬虫] ✅ 预加载成功，已存入Redis（共 1000 条记录）
========================================
```

---

### doFetch - 执行请求

```go
func FetchData() error
```

**功能**：
1. 加锁（防止并发更新）
2. 设置updating标记
3. 重试循环（最多3次，间隔5秒）
4. 调用doFetch()执行
5. 清理标记

**特点**：
- TryLock（不会阻塞）
- 最多重试3次
- 每次重试等5秒
- 失败记录详细日志

---

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

---

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

已提取常量，消除了魔法数字：

```go
const (
    httpTimeout     = 30 * time.Second  // HTTP超时
    maxRetries      = 3                 // 最多重试3次
    retryInterval   = 5 * time.Second   // 重试间隔5秒
    linesPerItem    = 3                 // 每3行一条数据
    timeLineIndex   = 0                 // 时间行索引
    dataStartLineIndex = 2             // 数据起始行
)
```

**设置原因**：
- 30秒超时：API响应慢时不继续等待
- 3次重试：失败后重试，但避免无限重试
- 5秒间隔：给API恢复时间
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
- 存储：验证失败直接拒绝（存）

---

## 重试机制

### 重试循环

```go
for attempt := 0; attempt < maxRetries; attempt++ {
    if attempt > 0 {
        log.Printf("第 %d 次重试中...", attempt)
        time.Sleep(retryInterval)
    }

    err := doFetch()
    if err == nil {
        return nil  // 成功，退出
    }

    lastErr = err
    log.Printf("第 %d 次尝试失败: %v", attempt+1, err)
}

log.Printf("重试 %d 次后仍失败，放弃", maxRetries)
return lastErr
```

**重试策略**：
- 最多3次
- 每次间隔5秒
- 第1次失败 → 等5秒 → 第2次 → 等5秒 → 第3次 → 放弃

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

## 代码优化

### 代码行数

| 版本 | 行数 | 说明 |
|------|------|------|
| 以前 | 426行 | 有定时任务、文件操作 |
| 现在 | 209行 | 纯爬虫，-51% |

### 移除的功能

- ❌ InitializeData() - 不需要初始化文件
- ❌ ScheduleJob() - 不需要定时任务
- ❌ checkExpired() - 不需要检查文件过期
- ❌ 文件读写 - 全用Redis

### 保留的功能

- ✅ FetchData() - 按需获取数据
- ✅ processData() - 数据解析
- ✅ 重试机制
- ✅ 并发控制

### 新增功能

- ✅ 数据验证
- ✅ 常量提取
- ✅ 更新状态标记

---

## 错误处理

### HTTP错误

```go
if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
}
```

### 解析警告

```go
if skippedCount > 0 {
    log.Printf("警告：跳过了 %d 条格式不正确的数据", skippedCount)
}
```

### 验证警告

```go
if err := result.Validate(); err != nil {
    log.Printf("⚠️ 数据验证失败: %v", err)
}
```

---

## 常见问题

### Q: 为何没有定时任务？

**A**: 已改为按需更新。TTL < 24小时时自动获取新数据。

### Q: 重试3次都失败会怎样？

**A**: 返回最后一次的错误，记录日志。下次请求再试。

### Q: 数据解析失败会怎样？

**A**: 跳过该条数据，记录警告。只要有一条成功就算成功。

### Q: 为何使用TryLock？

**A**: 防止并发更新。如果正在更新，直接返回错误，不阻塞。

### Q: 能否调整重试次数？

**A**: 可以，修改`maxRetries`常量。建议不要设置过多，3次足够。

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

## Context使用优化（已完成 2026-01-15）

### ✅ 已完成的优化

**新增带context的函数**：
- `FetchTop1000WithContext(ctx)` - 支持外部传入context
- `doFetchWithContext(ctx)` - 内部执行HTTP请求，使用传入的context
- `checkDataLoadRequired(ctx)` - 检查数据加载需求，支持context

**向后兼容**：
- 旧的函数（如`FetchTop1000()`）保持不变，内部创建默认context
- 新函数支持外部传入context，实现更好的超时控制和取消机制

### 使用示例

**方式一：使用默认超时（向后兼容）**
```go
// 使用默认10秒超时
data, err := crawler.FetchTop1000()
```

**方式二：使用自定义context（推荐）**
```go
// API层从Fiber的context提取
ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
defer cancel()

// 传递给crawler层
data, err := crawler.FetchTop1000WithContext(ctx)
```

### 优点

1. **超时控制**：API请求可以根据客户端需求设置超时时间
2. **取消机制**：客户端断开连接时，可以取消正在执行的HTTP请求
3. **调用链追踪**：为未来的分布式追踪预留了基础
4. **向后兼容**：旧的调用方式不受影响

---

## 相关文件

- `scheduler.go` - 爬虫代码（255行）
- `../storage/redis.go` - 数据存储
- `../config/config.go` - 配置管理
- `../model/types.go` - 数据结构

---

## 性能数据

### HTTP请求

- 超时：10秒
- 重试：最多1次（小项目简化）
- 间隔：1秒

### 数据解析

- 平均1000条数据
- 解析速度：<10ms
- 验证速度：<5ms

### 存储时间

- JSON序列化：<5ms
- Redis存储：<10ms

---

**总结**：爬虫稳定性第一，重试、超时、验证缺一不可！

**更新**: 2026-01-15
**代码行数**: 255 行
**代码质量**: S 级
**优化**: Context使用优化 + 移除定时任务 + 拆分函数 + 使用 `%w` 包装错误 + 提取常量

**小项目简化**（2026-01-11）：
- ✅ 重试次数：3次 → 1次（小项目不需要太多重试）
- ✅ 重试间隔：5秒 → 1秒（缩短重试间隔）

**Context优化**（2026-01-15）：
- ✅ API层从Fiber提取context，传递给下层
- ✅ Storage层支持外部传入context
- ✅ Crawler层支持外部传入context
- ✅ 向后兼容，旧的调用方式不受影响
