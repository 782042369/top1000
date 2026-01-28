# internal/model - 数据模型

[根目录](../../CLAUDE.md) > [internal](../) > **model**

## 模块快照

**职责**：数据结构定义、验证逻辑

**关键文件**：`types.go`

## 数据结构

### SiteItem - 站点数据

```go
type SiteItem struct {
    SiteName    string  // 站点名称
    SiteID      string  // 站点 ID
    Duplication string  // 重复度
    Size        string  // 文件大小
    ID          int     // 序号
}
```

### ProcessedData - 完整数据

```go
type ProcessedData struct {
    Time  string     // 数据时间
    Items []SiteItem // 站点列表
}
```

## 验证规则

### SiteItem 验证

- `SiteName`：不能为空
- `SiteID`：必须为数字字符串
- `Duplication`：必须为数字（可选）
- `Size`：格式必须匹配 `^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`
- `ID`：必须大于 0

### ProcessedData 验证

- `Time`：不能为空
- `Items`：列表不能为空，且每项需通过 SiteItem 验证

## 正则表达式

```go
sizePattern = regexp.MustCompile(`^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`)
sitePattern = regexp.MustCompile(`站名：(.*?) 【ID：(\d+)】`)
```

## 测试

无测试文件。

**建议**：添加数据验证测试、边界条件测试。

---

*文档生成时间：2026-01-28 13:08:52*
