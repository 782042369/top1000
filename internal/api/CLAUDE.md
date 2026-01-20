# API 处理层

> 处理HTTP请求并返回JSON数据

---

## 模块功能

**处理HTTP请求，返回Top1000的JSON数据**

核心功能：
1. 从Fiber的context提取context.Context
2. 检查数据是否需要更新
3. 需要更新时触发刷新
4. 从Redis加载数据并返回

---

## 核心函数

### GetTop1000Data - 主函数

API入口，处理`/top1000.json`请求。

**流程**：
```go
1. 从Fiber提取context（15秒超时）
2. 检查数据是否需要更新
3. 需要更新就刷新数据
4. 从Redis加载数据
5. 返回JSON数据
```

---

## Context传递

### 提取context

```go
ctx, cancel := context.WithTimeout(c.Context(), defaultAPITimeout)
defer cancel()
```

**配置**：默认15秒超时

### 传递给下层

- `shouldUpdateData(ctx)` - 检查是否需要更新
- `refreshData(ctx)` - 刷新数据
- `storage.LoadDataWithContext(ctx)` - 从Redis加载数据
- `storage.DataExistsWithContext(ctx)` - 检查数据存在性
- `storage.IsDataExpiredWithContext(ctx)` - 检查数据过期
- `crawler.FetchTop1000WithContext(ctx)` - 获取新数据
- `storage.SaveDataWithContext(ctx, *newData)` - 保存数据

---

## 常量定义

```go
const (
    defaultAPITimeout = 15 * time.Second  // API超时15秒
)
```

---

## 容错机制

### 刷新失败时返回旧数据

```go
// 需要更新
if shouldUpdate {
    err := refreshData(ctx)
    if err != nil {
        // 刷新失败，尝试返回旧数据
        if data, err := storage.LoadDataWithContext(ctx); err == nil {
            return c.JSON(data)  // 返回旧数据
        }
    }
}
```

**优点**：
- 即使API故障，用户也能看到旧数据
- 不会因为更新失败导致服务不可用
- 提高系统的容错性和可用性

---

## 并发控制

### 读写锁

```go
cacheMutex sync.RWMutex
```

**使用场景**：
- **读锁（RLock）**：读取缓存、检查状态
- **写锁（Lock）**：更新缓存、设置loadingFlag

### 更新标记

```go
loadingFlag bool      // 是否正在加载
loadDone   chan struct{}  // 加载完成通知
```

**流程**：
```go
1. 设置loadingFlag=true
2. 创建loadDone通道
3. 执行加载
4. 关闭loadDone通道
5. 清除loadingFlag
```

---

## Context优点

1. **超时控制**：API请求超时时间可配置（默认15秒）
2. **取消机制**：客户端断开连接时，可以取消正在执行的操作
3. **资源节约**：避免无用的后台操作
4. **调用链追踪**：为未来的分布式追踪预留了基础

---

## 常见问题

### Q: 数据会持续刷新吗？

**A**: 不会。检查数据time字段，24小时内算新鲜数据，不会更新。

### Q: 多个请求同时到达会怎样？

**A**:
- 第一个请求触发更新
- 其他请求等待更新完成
- 都等待同一个loadDone通道

### Q: 更新失败会怎样？

**A**:
- 返回Redis中的旧数据（容错机制）
- 保证服务可用，用户无感知
- 记录错误日志，下次请求再重试

---

## 相关文件

- `handlers.go` - API处理代码
- `../storage/redis.go` - Redis存储
- `../crawler/scheduler.go` - 数据更新
- `../model/types.go` - 数据结构
