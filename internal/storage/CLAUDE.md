# internal/storage - Redis 存储层

[根目录](../../CLAUDE.md) > [internal](../) > **storage**

## 模块快照

**职责**：Redis 连接管理、数据持久化、过期检查、并发控制

**关键文件**：`redis.go`

**核心函数**：
- `InitRedis()` - 初始化连接
- `LoadDataWithContext()` - 加载 Top1000 数据
- `SaveDataWithContext()` - 保存 Top1000 数据
- `IsDataExpiredWithContext()` - 过期检查
- `LoadSitesDataWithContext()` - 加载站点数据
- `SaveSitesDataWithContext()` - 保存站点数据

## 连接配置

```go
DialTimeout:  10 * time.Second
ReadTimeout:   5 * time.Second
WriteTimeout:  5 * time.Second
PoolSize:      3
MinIdleConns:  1
```

## Redis Key 设计

| Key | 说明 | TTL |
|-----|------|-----|
| `top1000:data` | Top1000 数据 | 永久（基于时间字段判断过期） |
| `top1000:sites` | 站点列表数据 | 24 小时 |

## 过期检查逻辑

```go
1. 从 Redis 读取数据
2. 解析时间字段（北京时间 UTC+8）
3. 转换为 UTC（减 8 小时）
4. 计算时间差
5. 判断是否超过 24 小时
```

## 并发控制

```go
// Top1000 数据更新锁
IsUpdating() / SetUpdating(bool)

// 站点数据更新锁（独立）
IsSitesUpdating() / SetSitesUpdating(bool)
```

## Context 支持

所有数据操作都支持 `context.Context`：
- 超时控制
- 取消操作
- 传播请求追踪

## 错误处理

- 数据不存在：返回 `redis.Nil`
- JSON 解析失败：返回解析错误
- Redis 连接失败：返回连接错误

## 测试

无测试文件。

**建议**：添加 CRUD 测试、过期检查测试、并发锁测试、Context 超时测试（使用 miniredis）。

---

*文档生成时间：2026-01-28 13:08:52*
