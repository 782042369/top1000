# Redis 存储层

> 管理Redis连接和数据存储

---

## 模块功能

**管理Redis连接，存储数据，读取数据，检查过期**

核心功能：
1. 连接Redis（程序启动时）
2. 存储数据到Redis（永久存储，不设置TTL）
3. 从Redis读取数据
4. 检查数据是否过期（基于数据time字段）
5. 防止并发更新（使用bool标记）

---

## 使用方法

### 初始化

```go
if err := storage.InitRedis(); err != nil {
    log.Fatalf("❌ Redis 连接失败: %v", err)
}
```

**注意**：Redis连接失败时会fatal退出。

### 存储数据

```go
err := storage.SaveData(data)
```

**内部处理**：
1. 先调用`data.Validate()`检查数据有效性
2. 序列化成JSON
3. 存入Redis，**不设置TTL**（永久存储）
4. 记录日志

**Redis Key**: `top1000:data`（可以通过环境变量改前缀）

### 读取数据

```go
data, err := storage.LoadData()
```

**返回**：
- `*ProcessedData`: 数据
- `error`: 不存在或解析失败

### 检查过期

```go
isExpired, err := storage.IsDataExpired()
```

**逻辑**：
- 从Redis读取数据
- 解析数据中的`time`字段（格式：`2025-12-11 07:52:33`）
- 计算`time`与当前时间的差值
- 超过24小时 → 返回true（过期了）
- 未超过24小时 → 返回false（还新鲜）
- 数据不存在或解析失败 → 返回true（强制更新）

### 防止并发更新

```go
// 标记正在更新
storage.SetUpdating(true)
defer storage.SetUpdating(false)

// 执行更新...
```

**检查状态**：
```go
if storage.IsUpdating() {
    // 正在更新中，跳过
    return
}
```

---

## Redis连接配置

```go
redisClient = redis.NewClient(&redis.Options{
    Addr:         cfg.RedisAddr,      // 从环境变量读
    Password:     cfg.RedisPassword,  // 从环境变量读
    DB:           cfg.RedisDB,        // 默认0
    DialTimeout:  5 * time.Second,    // 连接超时5秒
    ReadTimeout:  3 * time.Second,    // 读超时3秒
    WriteTimeout: 3 * time.Second,    // 写超时3秒
    PoolSize:     3,                  // 连接池3个（小项目）
    MinIdleConns: 1,                  // 至少保持1个空闲
})
```

**配置原因**：
- 超时不宜过长，否则影响用户体验
- 连接池3个足够（小访问量项目）
- 保持1个空闲连接，避免频繁建立连接

---

## 数据验证

存Redis之前会自动验证：

```go
// SiteItem验证
- 站点名称不能为空
- 站点ID必须是数字
- 重复度格式：数字
- 文件大小格式：数字 + 单位（KB/MB/GB/TB）
- ID必须大于0

// ProcessedData验证
- 时间不能为空
- 至少有一条数据
- 每条数据都要通过SiteItem验证
```

**验证失败**：
- 记录日志：`❌ 数据验证失败，拒绝保存`
- 返回错误，不会存到Redis

---

## TTL管理

### 不设置TTL

| 操作 | TTL | 说明 |
|------|-----|------|
| **存数据** | **0（永久）** | 不设置TTL，数据永久存储 |

### 过期逻辑

**基于数据time字段判断**：
```
1. 从Redis读取数据
2. 解析time字段（如：2025-12-11 07:52:33）
3. 计算时间差 = 当前时间 - time
4. 时间差 > 24h？
   ├─ 是 → 过期了，触发更新
   └─ 否 → 还新鲜，不更新
```

**为何不设置TTL**？
- 数据完全基于time字段判断是否过期
- 更新失败时可以返回旧数据（容错）
- 避免Redis自动删除导致数据丢失
- 手动清理即可（Redis key: `top1000:data`）

---

## Context使用优化

### 新增带context的函数

- `SaveDataWithContext(ctx, data)` - 支持外部传入context
- `LoadDataWithContext(ctx)` - 支持外部传入context
- `DataExistsWithContext(ctx)` - 支持外部传入context
- `IsDataExpiredWithContext(ctx)` - 支持外部传入context

**向后兼容**：
- 旧的函数（如`SaveData()`）保持不变，内部创建默认context
- 新函数支持外部传入context，实现更好的超时控制和取消机制

### 使用示例

**方式一：使用默认超时**
```go
// 使用默认5秒超时
err := storage.SaveData(data)
data, err := storage.LoadData()
```

**方式二：使用自定义context**
```go
// API层从Fiber的context提取
ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
defer cancel()

// 传递给storage层
data, err := storage.LoadDataWithContext(ctx)
```

---

## 环境变量

| 变量 | 示例 | 必需？ |
|------|------|--------|
| `REDIS_ADDR` | `192.144.142.2:26739` | **必须** |
| `REDIS_PASSWORD` | `填写密码` | **必须** |
| `REDIS_DB` | `0` | 可选，默认0 |
| `REDIS_KEY_PREFIX` | `top1000:` | 可选，默认top1000: |

**注意**：
- `REDIS_ADDR`和`REDIS_PASSWORD`**必须**在`.env`里配置
- 程序启动时会检查，没配置会直接报错退出

---

## 常见问题

### Q: Redis连接失败怎么办？

**A**: 检查以下几点：
1. `.env`文件中的`REDIS_ADDR`是否正确
2. 密码是否正确
3. Redis是否启动
4. 防火墙是否拦截

### Q: 数据存储多久？

**A**: 数据永久存储（不设置TTL），过期判断基于time字段：
- 数据time字段超过24小时 → 触发更新
- 更新成功 → 返回最新数据
- 更新失败 → 返回旧数据（容错机制）

### Q: 如何查看Redis中的数据？

**A**: 使用redis-cli：
```bash
redis-cli -h <host> -p <port> -a <password>
> GET top1000:data
```

### Q: 如何清理Redis中的数据？

**A**: 由于数据永久存储，需要手动清理：
```bash
redis-cli -h <host> -p <port> -a <password>
> DEL top1000:data
```

**注意**：删除后，下次访问会自动触发更新获取新数据。

---

## 相关文件

- `redis.go` - 所有Redis操作
- `../config/config.go` - 读配置
- `../model/types.go` - 数据结构（含验证）
