# Model 模块

[根目录](../../CLAUDE.md) > [internal](../) > **model**

## 模块职责

Model 模块定义系统的核心数据结构，包括站点数据、完整响应格式，以及数据验证逻辑。它是类型安全的基础。

## 入口与启动

- **入口文件**: `types.go`
- **无初始化逻辑** - 纯数据结构定义

## 对外接口

### 数据结构

```go
// SiteItem 一条站点数据
type SiteItem struct {
    SiteName    string `json:"siteName"`    // 站点名称
    SiteID      string `json:"siteid"`      // 站点 ID
    Duplication string `json:"duplication"` // 重复度
    Size        string `json:"size"`        // 文件大小
    ID          int    `json:"id"`          // 序号（自动递增）
}

// ProcessedData 完整的 Top1000 数据
type ProcessedData struct {
    Time  string     `json:"time"`  // 数据时间（北京时间）
    Items []SiteItem `json:"items"` // 站点列表
}
```

### 验证方法

```go
// Validate 验证单条数据正确性
func (s *SiteItem) Validate() error

// Validate 验证完整数据
func (p *ProcessedData) Validate() error
```

## 关键依赖与配置

### 依赖模块
- 无外部依赖（仅使用标准库）

### 常量配置

```go
var (
    sizePattern = regexp.MustCompile(`^\d+(\.\d+)?\s*(KB|MB|GB|TB)$`)
)
```

## 数据模型详解

### SiteItem

#### 字段说明

| 字段 | 类型 | JSON Tag | 必需 | 验证规则 | 示例 |
|------|------|----------|------|----------|------|
| `SiteName` | string | `siteName` | 是 | 非空 | `"站点名称"` |
| `SiteID` | string | `siteid` | 是 | 数字字符串 | `"123"` |
| `Duplication` | string | `duplication` | 否 | 浮点数字符串 | `"1.5"` |
| `Size` | string | `size` | 否 | 大小格式 | `"1.5 GB"` |
| `ID` | int | `id` | 是 | > 0 | `1` |

#### 验证规则

```go
func (s *SiteItem) Validate() error {
    // 1. 站点名称非空
    if strings.TrimSpace(s.SiteName) == "" {
        return fmt.Errorf("站点名称不能为空")
    }

    // 2. 站点 ID 非空且为数字
    if s.SiteID == "" {
        return fmt.Errorf("站点ID不能为空")
    }
    if _, err := strconv.ParseInt(s.SiteID, 10, 64); err != nil {
        return fmt.Errorf("站点ID必须是数字: %s", s.SiteID)
    }

    // 3. 重复度为浮点数（如果存在）
    if s.Duplication != "" {
        if _, err := strconv.ParseFloat(s.Duplication, 64); err != nil {
            return fmt.Errorf("重复度必须为数字: %s", s.Duplication)
        }
    }

    // 4. 文件大小格式正确（如果存在）
    if s.Size != "" && !sizePattern.MatchString(s.Size) {
        return fmt.Errorf("文件大小格式错误: %s", s.Size)
    }

    // 5. ID > 0
    if s.ID <= 0 {
        return fmt.Errorf("ID必须大于0: %d", s.ID)
    }

    return nil
}
```

### ProcessedData

#### 字段说明

| 字段 | 类型 | JSON Tag | 必需 | 验证规则 | 示例 |
|------|------|----------|------|----------|------|
| `Time` | string | `time` | 是 | 非空时间字符串 | `"2026-01-19 07:50:56"` |
| `Items` | []SiteItem | `items` | 是 | 至少 1 条且全部有效 | `[...]` |

#### 验证规则

```go
func (p *ProcessedData) Validate() error {
    // 1. 时间非空
    if strings.TrimSpace(p.Time) == "" {
        return fmt.Errorf("时间不能为空")
    }

    // 2. 数据条目非空
    if len(p.Items) == 0 {
        return fmt.Errorf("数据条目不能为空")
    }

    // 3. 每条数据都有效
    for i, item := range p.Items {
        if err := item.Validate(); err != nil {
            return fmt.Errorf("第%d条数据验证失败: %w", i+1, err)
        }
    }

    return nil
}
```

## 数据流转

### 输入（IYUU API 返回）

```
站名：站点名称 【ID：123】
重复度：1.5
文件大小：1.5 GB
```

### 解析（Crawler 模块）

```go
item := model.SiteItem{
    SiteName:    "站点名称",
    SiteID:      "123",
    Duplication: "1.5",
    Size:        "1.5 GB",
    ID:          1,  // 自动赋值
}
```

### 验证

```go
if err := item.Validate(); err != nil {
    // 处理验证失败
}
```

### 存储（Storage 模块）

```go
data := model.ProcessedData{
    Time:  "2026-01-19 07:50:56",
    Items: []model.SiteItem{item},
}
if err := data.Validate(); err != nil {
    // 处理验证失败
}
storage.SaveDataWithContext(ctx, data)
```

### 响应（API 模块）

```json
{
  "time": "2026-01-19 07:50:56",
  "items": [
    {
      "id": 1,
      "siteName": "站点名称",
      "siteid": "123",
      "duplication": "1.5",
      "size": "1.5 GB"
    }
  ]
}
```

## 测试与质量

### 当前状态
- 无单元测试
- 依赖其他模块的集成测试验证

### 测试建议

**单元测试文件**: `types_test.go`

```go
func TestSiteItem_Valid(t *testing.T)
func TestSiteItem_EmptySiteName(t *testing.T)
func TestSiteItem_InvalidSiteID(t *testing.T)
func TestSiteItem_InvalidDuplication(t *testing.T)
func TestSiteItem_InvalidSize(t *testing.T)
func TestSiteItem_ZeroID(t *testing.T)

func TestProcessedData_Valid(t *testing.T)
func TestProcessedData_EmptyTime(t *testing.T)
func TestProcessedData_EmptyItems(t *testing.T)
func TestProcessedData_InvalidItem(t *testing.T)
```

### 测试用例示例

```go
func TestSiteItem_Valid(t *testing.T) {
    item := model.SiteItem{
        SiteName:    "测试站点",
        SiteID:      "123",
        Duplication: "1.5",
        Size:        "1.5 GB",
        ID:          1,
    }
    if err := item.Validate(); err != nil {
        t.Errorf("验证失败: %v", err)
    }
}

func TestSiteItem_InvalidSize(t *testing.T) {
    item := model.SiteItem{
        SiteName: "测试站点",
        SiteID:   "123",
        Size:     "invalid",  // 错误格式
        ID:       1,
    }
    if err := item.Validate(); err == nil {
        t.Error("应该返回错误")
    }
}
```

## 相关文件清单

### 核心文件
- `types.go` - 数据模型定义（78 行）
  - `SiteItem` 结构体
  - `SiteItem.Validate()` 方法
  - `ProcessedData` 结构体
  - `ProcessedData.Validate()` 方法
  - `sizePattern` 正则表达式

### 测试文件（待创建）
- `types_test.go` - 单元测试

### 依赖文件
- 无（纯数据结构）

## 设计考虑

### 为什么使用 string 而不是 float64？

**SiteID** 使用 `string`:
- IYUU API 返回的是字符串格式
- 避免大整数精度问题（JavaScript Number 最大安全整数是 2^53-1）
- 便于正则匹配和解析

**Duplication** 使用 `string`:
- 不参与数值计算，仅用于展示
- 保留原始格式（如 "1.50"）
- 避免浮点数精度问题

**Size** 使用 `string`:
- 带单位（KB/MB/GB/TB），解析成本高
- 仅用于展示，不参与计算
- 保留原始格式

### 为什么 ID 使用 int？

- 自动递增序号，不会超过 2^31-1
- JSON 反序列化后可以安全排序
- 便于前端实现序号列

### JSON Tag 命名

遵循 Go 命名规范：
- 结构体字段：`PascalCase`（导出）
- JSON Tag：`camelCase`（前端约定）

## 扩展建议

### 可能的扩展字段

```go
type SiteItem struct {
    // 现有字段...
    SiteName    string `json:"siteName"`
    SiteID      string `json:"siteid"`
    Duplication string `json:"duplication"`
    Size        string `json:"size"`
    ID          int    `json:"id"`

    // 可能的扩展字段
    URL         string `json:"url,omitempty"`          // 站点 URL
    Category    string `json:"category,omitempty"`     // 分类
    Tags        []string `json:"tags,omitempty"`       // 标签
    Seeders     int    `json:"seeders,omitempty"`      // 做种数
    Leechers    int    `json:"leechers,omitempty"`     // 下载数
    Snatched    int    `json:"snatched,omitempty"`     // 完成数
    PublishTime string `json:"publishTime,omitempty"`  // 发布时间
}
```

### 验证增强

```go
func (s *SiteItem) ValidateAdvanced() error {
    // 基础验证
    if err := s.Validate(); err != nil {
        return err
    }

    // URL 格式验证
    if s.URL != nil {
        if _, err := url.Parse(s.URL); err != nil {
            return fmt.Errorf("URL 格式错误: %w", err)
        }
    }

    // 数值范围验证
    if s.Seeders < 0 {
        return fmt.Errorf("做种数不能为负数")
    }

    return nil
}
```

## 常见问题

### Q: 如何添加新字段？
在 `SiteItem` 结构体中添加字段，并更新 `Validate()` 方法。

### Q: 如何修改验证规则？
修改 `Validate()` 方法中的验证逻辑，或添加新的验证方法。

### Q: JSON Tag 可以省略吗？
不可以。省略后字段名将使用结构体字段名（PascalCase），不符合前端约定。

### Q: 为什么 omitempty？
对于可选字段，使用 `omitempty` 可以在字段为零值时不序列化到 JSON，减少数据体积。

---

**最后更新**: 2026-01-27
**代码行数**: ~78 行
**维护状态**: 稳定
