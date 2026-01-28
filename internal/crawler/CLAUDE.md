# internal/crawler - 数据爬取模块

[根目录](../../CLAUDE.md) > [internal](../) > **crawler**

## 模块快照

**职责**：从 IYUU API 爬取数据并解析

**关键文件**：`scheduler.go`

**核心函数**：
- `FetchTop1000WithContext(ctx)` - 带超时的数据爬取
- `PreloadData()` - 启动时预加载
- `parseResponse()` - 文本解析

## 数据源

- **API URL**：`https://api.iyuu.cn/top1000.php`
- **响应格式**：纯文本（自定义格式）
- **超时设置**：10 秒

## 解析逻辑

```
原始文本格式：
create time 2026-01-19 07:50:56 by xxx
站名：XXX 【ID：123】
重复度：：85.5%
文件大小：：1.2TB

解析为：
SiteItem {
    SiteName: "XXX",
    SiteID: "123",
    Duplication: "85.5%",
    Size: "1.2TB"
}
```

## 重试机制

- 最大重试次数：1
- 重试间隔：1 秒
- 并发锁：防止重复爬取

## 依赖关系

- `net/http` - HTTP 请求
- `internal/config` - API URL 配置
- `internal/model` - 数据模型

## 测试

无测试文件。

**建议**：添加解析测试、边界条件测试、并发锁测试。

---

*文档生成时间：2026-01-28 13:08:52*
