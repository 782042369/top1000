# 数据模型

> 定义数据结构并提供验证

---

## 模块功能

**定义Top1000数据的格式，提供验证功能**

核心内容：
1. `SiteItem` - 一条站点数据的结构
2. `ProcessedData` - 完整的Top1000数据
3. `Validate()` - 验证数据有效性

---

## 数据结构

### SiteItem - 单条数据

```go
type SiteItem struct {
    SiteName    string  // 站点名字（如"朋友"）
    SiteID      string  // 站点ID（数字字符串）
    Duplication string  // 重复度（如"95%"）
    Size        string  // 文件大小（如"1.5GB"）
    ID          int     // 序号（1,2,3...）
}
```

### ProcessedData - 完整数据

```go
type ProcessedData struct {
    Time  string     // 时间（如"2025-12-11 07:52:33"）
    Items []SiteItem // 1000条数据
}
```

---

## 验证方法

### SiteItem.Validate()

**检查内容**：
```go
1. 站点名称不能为空
2. 站点ID必须是数字
3. 重复度格式：数字（如"95"、"87.5"）
4. 文件大小格式：数字 + 单位（KB/MB/GB/TB）
5. ID必须大于0
```

**示例**：
```go
item := model.SiteItem{
    SiteName: "朋友",
    SiteID:   "123",
    Duplication: "95%",
    Size: "1.5GB",
    ID: 1,
}

if err := item.Validate(); err != nil {
    log.Printf("数据错误: %v", err)
}
```

### ProcessedData.Validate()

**检查内容**：
```go
1. 时间不能为空
2. 至少有一条数据
3. 每条数据都要通过SiteItem验证
```

**示例**：
```go
data := model.ProcessedData{
    Time: "2025-12-11 07:52:33",
    Items: []model.SiteItem{...},
}

if err := data.Validate(); err != nil {
    log.Printf("数据错误: %v", err)
}
```

---

## 验证规则详解

### 站点名称

```go
if strings.TrimSpace(s.SiteName) == "" {
    return fmt.Errorf("站点名称不能为空")
}
```

### 站点ID

```go
if s.SiteID == "" {
    return fmt.Errorf("站点ID不能为空")
}
if _, err := strconv.ParseInt(s.SiteID, 10, 64); err != nil {
    return fmt.Errorf("站点ID必须是数字")
}
```

### 重复度

```go
// 重复度必须是数字
if s.Duplication != "" {
    if _, err := strconv.ParseFloat(s.Duplication, 64); err != nil {
        return fmt.Errorf("重复度必须为数字")
    }
}
```

**注意**：允许为空，如果有值，必须是数字（如 "95"、"87.5"）

### 文件大小

```go
if s.Size != "" {
    sizePattern := regexp.MustCompile(`^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`)
    if !sizePattern.MatchString(s.Size) {
        return fmt.Errorf("文件大小格式错误")
    }
}
```

**注意**：允许为空，但如果有值必须符合格式

### ID

```go
if s.ID <= 0 {
    return fmt.Errorf("ID必须大于0")
}
```

---

## JSON示例

```json
{
  "time": "2025-12-11 07:52:33",
  "items": [
    {
      "siteName": "朋友",
      "siteid": "123456",
      "duplication": "95%",
      "size": "1.5GB",
      "id": 1
    },
    {
      "siteName": "馒头",
      "siteid": "789012",
      "duplication": "87%",
      "size": "2.3GB",
      "id": 2
    }
  ]
}
```

---

## 使用场景

### 爬虫解析数据

```go
// crawler/scheduler.go
items = append(items, model.SiteItem{
    SiteName:    siteName,
    SiteID:      siteID,
    Duplication: duplication,
    Size:        size,
    ID:          len(items) + 1,
})

result := model.ProcessedData{
    Time:  parseTime(timeLine),
    Items: items,
}

// 验证一下
if err := result.Validate(); err != nil {
    log.Printf("⚠️ 数据验证失败: %v", err)
}
```

### 存Redis前验证

```go
// storage/redis.go
func SaveData(data model.ProcessedData) error {
    // 先验证
    if err := data.Validate(); err != nil {
        log.Printf("❌ 数据验证失败，拒绝保存: %v", err)
        return fmt.Errorf("数据验证失败: %v", err)
    }

    // 序列化保存...
}
```

---

## 常见问题

### Q: 为何需要验证？

**A**: 防止存储无效数据。爬虫解析错误、API返回错误都会被拦截。

### Q: 是否允许空值？

**A**: 部分字段允许：
- `Duplication`: 允许为空
- `Size`: 允许为空
- 其他：不允许

### Q: 验证失败会怎样？

**A**:
- 爬虫：记录警告日志，但仍然返回（容错）
- 存储：直接拒绝，不会存入Redis

### Q: 为何ID是int不是string？

**A**:
- `ID`: 序号（1,2,3...），方便排序和显示
- `SiteID`: 资源ID（字符串，可能包含字母数字）

---

## 相关文件

- `types.go` - 数据结构定义
- `../crawler/scheduler.go` - 数据解析
- `../storage/redis.go` - 数据存储（调用Validate）
- `../../web/src/types.d.ts` - 前端TypeScript定义
