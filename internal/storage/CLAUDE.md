# Redis 存储层

> 数据存储模块

---

## 模块功能

**管理Redis连接，存储数据，读取数据，检查过期**

核心功能：
1. 连接Redis（程序启动时）
2. 存储数据到Redis（爬虫获取的数据）
3. 从Redis读取数据（API需要返回时）
4. 检查数据是否过期（TTL < 24小时即为过期）
5. 防止并发更新（使用bool标记）

---

## 使用方法

### 初始化（程序启动时调用）

```go
if err := storage.InitRedis(); err != nil {
    log.Fatalf("❌ Redis 连接失败: %v", err)
}
```

**注意**：此处会**直接fatal**，Redis连接失败时退出，不使用备选方案。

---

### 存储数据

```go
err := storage.SaveData(data)
```

**内部处理**：
1. 先调用`data.Validate()`检查数据有效性
2. 序列化成JSON
3. 存入Redis，TTL设为48小时
4. 记录日志

**Redis Key**: `top1000:data`（可以通过环境变量改前缀）

---

### 读取数据

```go
data, err := storage.LoadData()
```

**返回**：
- `*ProcessedData`: 数据
- `error`: 不存在或解析失败

**错误类型**：
- `redis.Nil`: 数据不存在（正常情况）
- 其他：JSON解析失败（数据损坏）

---

### 检查过期

```go
isExpired, err := storage.IsDataExpired()
```

**逻辑**：
- 获取Redis key的TTL
- TTL < 24小时 → 返回true（过期了，该更新了）
- TTL >= 24小时 → 返回false（还新鲜）
- Key不存在 → 返回true

**使用场景**：
```go
if !exists || isExpired {
    // 触发更新
    go crawler.FetchData()
}
```

---

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

连接池配置：

```go
redisClient = redis.NewClient(&redis.Options{
    Addr:         cfg.RedisAddr,      // 从环境变量读
    Password:     cfg.RedisPassword,  // 从环境变量读
    DB:           cfg.RedisDB,        // 默认0
    DialTimeout:  5 * time.Second,    // 连接超时5秒
    ReadTimeout:  3 * time.Second,    // 读超时3秒
    WriteTimeout: 3 * time.Second,    // 写超时3秒
    PoolSize:     10,                 // 连接池10个
    MinIdleConns: 5,                  // 至少保持5个空闲
})
```

**配置原因**：
- 超时不宜过长，否则影响用户体验
- 连接池10个足够，非高并发系统
- 保持5个空闲连接，避免频繁建立连接

---

## 数据验证

存Redis之前会自动验证：

```go
// SiteItem验证
- 站点名称不能为空
- 站点ID必须是数字
- 重复度格式：XX%
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

## TTL管理（过期时间）

### TTL设置

| 操作 | TTL | 说明 |
|------|-----|------|
| **存数据** | 48小时 | 2 * DATA_EXPIRE_DURATION |
| **检测阈值** | 24小时 | DATA_EXPIRE_DURATION |

### 过期逻辑

```
当前时间 ←←←← 数据存入（TTL=48h）
    ↓
    24小时过去了
    ↓
TTL < 24h？
  ├─ 是 → 过期了，触发更新
  └─ 否 → 还新鲜，不管
```

**为何48小时TTL但24小时就算过期**？
- 预留24小时缓冲，避免等到最后一刻才更新
- 即使更新失败，仍有24小时数据可用

---

## 环境变量（必须配置）

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

## 错误处理

### 错误分类

**1. 连接错误**（最严重）
```
错误：Redis 连接失败
处理：程序fatal退出，不启动
原因：地址不对、密码错、Redis没起
```

**2. 数据验证错误**
```
错误：数据验证失败，拒绝保存
处理：返回错误，不保存
原因：数据格式不对、字段缺失
```

**3. 序列化错误**
```
错误：序列化/反序列化数据失败
处理：返回错误
原因：JSON格式错误、类型不匹配
```

**4. 数据不存在**
```
错误：数据不存在（redis.Nil）
处理：返回空，不报错
原因：第一次运行、数据过期被删了
```

---

## 常见问题

### Q: Redis连接失败怎么办？

**A**: 检查以下几点：
1. `.env`文件中的`REDIS_ADDR`是否正确
2. 密码是否正确
3. Redis是否启动
4. 防火墙是否拦截

### Q: 数据存储多久？

**A**: 48小时TTL，但24小时后会触发更新：
- 0-24小时：新鲜，不更新
- 24-48小时：已过期，会更新
- 48小时后：Redis删除，重新获取

### Q: 如何查看Redis中的数据？

**A**: 使用redis-cli：
```bash
redis-cli -h <host> -p <port> -a <password>
> GET top1000:data
> TTL top1000:data
```

### Q: 能否更换数据库？

**A**: 可以，修改`.env`：
```bash
REDIS_DB=1  # 使用1号数据库
```

### Q: Key前缀能否修改？

**A**: 可以，修改`.env`：
```bash
REDIS_KEY_PREFIX=myservice:  # 改为myservice:data
```

---

## 性能优化建议

### 调整连接池

高并发场景可以调大点：

```go
PoolSize:     20,  # 从10改成20
MinIdleConns: 10,  # 从5改成10
```

### 添加重试

临时性错误可以重试：

```go
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    err := storage.SaveData(data)
    if err == nil {
        break
    }
    time.Sleep(time.Second * time.Duration(i+1))
}
```

### 监控指标

建议监控：
- Redis连接数（别超过连接池）
- 命令执行耗时（别超过100ms）
- 数据命中率（别总是Miss）
- TTL分布（别都集中过期）

---

## 测试建议

### 单元测试示例

```go
func TestSaveAndLoad(t *testing.T) {
    // 准备数据
    data := model.ProcessedData{
        Time: "2025-12-11 07:52:33",
        Items: []model.SiteItem{...},
    }

    // 存
    err := storage.SaveData(data)
    assert.NoError(t, err)

    // 取
    loaded, err := storage.LoadData()
    assert.NoError(t, err)
    assert.Equal(t, data.Time, loaded.Time)
}
```

---

## 相关文件

- `redis.go` - 所有Redis操作都在这
- `../config/config.go` - 读配置
- `../model/types.go` - 数据结构（含验证）

---

**总结**：此模块负责数据存储和读取，需要Redis才能运行。

**更新**: 2026-01-10
**代码质量**: A级
