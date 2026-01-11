# Web 前端应用

> 展示Top1000数据的表格界面

---

## 模块功能

**使用AG Grid将Top1000数据展示为表格，支持查看、搜索、点击操作**

核心功能：

1. 从后端API获取JSON数据
2. 使用AG Grid展示表格（企业版，功能完整）
3. 支持列过滤、排序
4. 点击操作按钮跳转站点详情或下载种子

前端就一个页面，简洁！

---

## 入口文件

```
index.html (入口HTML)
    ↓
main.ts (入口TS)
    ↓
AG Grid渲染表格
```

**开发命令**：

```bash
pnpm dev      # 启动开发服务器（端口自动分配）
pnpm build    # 构建到 ../web-dist/
```

---

## 启动流程

```typescript
// main.ts 的功能
1. 注册AG Grid模块（企业版）
2. 配置列定义（名字、ID、重复度、大小、操作）
3. 设置中文本地化
4. 加载数据（fetch /top1000.json）
5. 渲染表格
```

**流程图**：

```
index.html
    <div id="root">
        ↓
    main.ts
        ├── 注册 AG Grid
        ├── 配置列定义
        └── 加载数据
            ↓
        AG Grid 渲染
            ├── 数据表格
            ├── 过滤器
            ├── 排序
            └── 操作按钮
```

---

## AG Grid配置

### 基本配置

```typescript
localeText: AG_GRID_LOCALE_CN       // 中文界面
getRowId: params => params.data.id  // 行ID用数据里的id字段
defaultColDef.flex: 1               // 列宽度自适应
defaultColDef.sortable: false       // 默认不可排序（部分列开启）
```

### 列定义

| 列名     | 字段          | 能过滤？ | 能排序？ | 说明                          |
| -------- | ------------- | -------- | -------- | ----------------------------- |
| 名字     | `siteName`    | ✅       | ❌       | 站点名称（如"朋友"）          |
| 资源ID   | `siteid`      | ❌       | ❌       | 站点内资源ID                  |
| 重复度   | `duplication` | ❌       | ✅       | 重复度百分比（如"95%"）       |
| 文件大小 | `size`        | ❌       | ✅       | 自定义比较器（1.5GB > 512MB） |
| 操作     | -             | ❌       | ❌       | 自定义渲染器（跳转链接）      |

---

## 核心依赖

### AG Grid

```json
{
  "dependencies": {
    "@ag-grid-community/locale": "^35.0.0", // 中文包
    "ag-grid-community": "^35.0.0", // 社区版
    "ag-grid-enterprise": "^35.0.0" // 企业版（功能全）
  }
}
```

**企业版功能**：

- 高级过滤
- 范围选择
- Excel导出
- 更多主题

### 开发工具

```json
{
  "devDependencies": {
    "@antfu/eslint-config": "^6.7.3", // ESLint配置
    "typescript": "^5.9.3", // TypeScript
    "vite": "8.0.0-beta.5", // 构建工具
    "vite-plugin-html": "^3.2.2" // HTML插件
  }
}
```

### 环境要求

- **Node.js**: >=24.3.0（新版）
- **pnpm**: >=10.12.4（快速包管理器）

---

## 构建配置

### Vite配置

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [
    splitChunks(), // 代码分割
    createHtmlPlugin({ minify: true }), // HTML压缩
  ],
  build: {
    rollupOptions: {
      output: {
        chunkFileNames: 'js/[name]-[hash].js', // chunk文件命名
        entryFileNames: 'js/[name]-[hash].js', // 入口文件命名
        assetFileNames: '[ext]/[name]-[hash].[ext]', // 资源文件命名
      },
    },
    outDir: resolve(__dirname, '../web-dist'), // 输出到项目根目录的web-dist
  },
})
```

**输出结构**：

```
../web-dist/
    ├── index.html        # 压缩后的HTML
    ├── js/               # JS文件（带hash）
    │   ├── main-abc123.js
    │   └── vendor-def456.js
    └── css/              # CSS文件（带hash）
        └── style-ghi789.css
```

---

## 数据模型

### TypeScript类型

```typescript
// types.d.ts
export interface DataType {
  siteName: string // 站点名称（如"朋友"）
  siteid: string // 资源ID（数字字符串）
  duplication: string // 重复度（如"95%"）
  size: string // 文件大小（如"1.5GB"）
  id: number // 序号（1,2,3...）
}

export interface ResDataType {
  items: DataType[] // 种子列表（1000条）
  time: string // 更新时间（如"2025-12-11 07:52:33"）
}
```

### 数据来源

```typescript
// src/utils/index.ts
const response = await fetch('http://localhost:7066/top1000.json')
const json: ResDataType = await response.json()
```

**后端API**：

- 地址：`http://localhost:7066/top1000.json`
- 返回：JSON格式
- 更新：按需更新（TTL < 24小时就自动拉新的）

---

## 站点配置

### 站点元数据

```typescript
// src/utils/iyuuSites.ts
const siteData = [
  {
    id: 1,
    site: 'keepfrds', // 站点标识
    nickname: '朋友', // 站点昵称（显示用）
    base_url: 'pt.keepfrds.com', // 站点域名
    download_page: 'download.php?id={}&passkey={passkey}', // 下载页面
    details_page: 'details.php?id={}', // 详情页面
    is_https: 2, // 是否HTTPS
    cookie_required: 0, // 是否需要Cookie
  },
  // ... 共118个站点配置
]
```

**用途**：

- 操作列渲染器根据站点ID查询配置
- 生成正确的跳转链接（详情、下载）

---

## 操作列渲染

### 自定义渲染器

```typescript
// src/utils/operationRender.ts
export function operationRender(params: ICellRendererParams) {
  const data = params.data as DataType
  const site = iyuuSites.find(s => s.nickname === data.siteName)

  if (!site) {
    return `<span style="color: #999">暂不支持</span>`
  }

  return `
    <a href="${detailsUrl}" target="_blank">详情</a>
    <a href="${downloadUrl}">下载种子</a>
  `
}
```

**逻辑**：

1. 根据站点名称查找配置
2. 生成详情页链接（`details.php?id=xxx`）
3. 生成下载链接（`download.php?id=xxx&passkey=xxx`）
4. 不支持的站点显示"暂不支持"

---

## 工具函数

### 大小转换

```typescript
// src/utils/index.ts
export function convertSizeToKb(size: string): number {
  const match = size.match(/^(\d+(?:\.\d+)?)\s*(KB|MB|GB|TB)$/i)
  if (!match)
    return 0

  const value = Number.parseFloat(match[1])
  const unit = match[2].toUpperCase()

  const multipliers = {
    KB: 1,
    MB: 1024,
    GB: 1024 * 1024,
    TB: 1024 * 1024 * 1024,
  }

  return value * multipliers[unit]
}
```

**用途**：AG Grid自定义排序比较器

```typescript
comparator: (valueA, valueB) => {
  const kbA = convertSizeToKb(valueA)
  const kbB = convertSizeToKb(valueB)
  return kbA - kbB
}
```

---

## 常见问题

### Q: 如何修改数据源地址？

**A**: 修改`src/utils/index.ts`：

```typescript
const response = await fetch('http://localhost:7066/top1000.json')
// 修改为实际地址
const response = await fetch('https://your-domain.com/top1000.json')
```

### Q: 如何添加新列？

**A**: 修改`src/main.ts`的`columnDefs`：

```typescript
columnDefs: [
  {
    headerName: '名字',
    field: 'siteName',
    filter: true,
  },
  {
    headerName: '新列', // 新增列
    field: 'newField',
    sortable: true,
  },
  // ...
]
```

### Q: 如何固定左侧列？

**A**: 设置`pinned: 'left'`：

```typescript
{
  headerName: '名字',
  field: 'siteName',
  pinned: 'left',  // 固定到左侧
}
```

### Q: 如何启用行选择？

**A**:

```typescript
const gridOptions: GridOptions<DataType> = {
  rowSelection: 'multiple', // 多选
  // rowSelection: 'single', // 单选
}

// 获取选中的行
const selectedRows = gridApi.getSelectedRows()
```

### Q: 如何导出Excel？

**A**: AG Grid企业版内置支持：

```typescript
import { ExcelExportModule, ModuleRegistry } from 'ag-grid-enterprise'

ModuleRegistry.registerModules([
  // ... 其他模块
  ExcelExportModule,
])

// 导出
gridApi.exportDataAsExcel({
  fileName: 'top1000.xlsx',
})
```

### Q: 如何自定义主题？

**A**: 修改`src/index.css`：

```css
.ag-theme-alpine {
  --ag-header-background-color: #0d47a1; /* 表头背景 */
  --ag-odd-row-background-color: #f5f5f5; /* 奇数行背景 */
  --ag-font-size: 14px; /* 字体大小 */
  --ag-border-color: #ddd; /* 边框颜色 */
}

```

### Q: 如何添加搜索框？

**A**: AG Grid内置快速搜索：

```typescript
const gridOptions: GridOptions<DataType> = {
  quickFilterText: '',  // 绑定到输入框
}

// HTML中加个输入框
<input
  type="text"
  placeholder="搜索..."
  oninput="gridOptions.api.setGridOption('quickFilterText', this.value)"
/>
```

### Q: 大数据性能如何？

**A**: AG Grid虚拟滚动，1000条数据轻松处理：

- 虚拟滚动：只渲染可见行
- 需要更快？启用分页：
  ```typescript
  pagination: true,
  paginationPageSize: 50,
  ```

---

## 代码质量

### 当前状态

- ❌ 没有单元测试
- ❌ 没有E2E测试
- ❌ 没有组件测试

前端确实未编写测试。就一个页面，值得吗？

### 建议补充

如果要添加测试，可以这样做：

#### 1. 单元测试（Vitest）

```typescript
// utils/index.test.ts
import { describe, expect, it } from 'vitest'

import { convertSizeToKb } from './index'

describe('convertSizeToKb', () => {
  it('should convert GB to KB', () => {
    expect(convertSizeToKb('1.5GB')).toBe(1.5 * 1024 * 1024)
  })

  it('should handle invalid format', () => {
    expect(convertSizeToKb('invalid')).toBe(0)
  })
})
```

#### 2. E2E测试（Playwright）

```typescript
// e2e/basic.spec.ts
import { expect, test } from '@playwright/test'

test('data table renders correctly', async ({ page }) => {
  await page.goto('http://localhost:7066')
  const table = page.locator('.ag-root-wrapper')
  await expect(table).toBeVisible()

  // 检查数据行数
  const rowCount = await page.locator('.ag-row').count()
  expect(rowCount).toBe(1000)
})
```

#### 3. 组件测试（Testing Library）

```typescript
// operationRender.test.ts
import { render } from '@testing-library/dom'

import { operationRender } from './operationRender'

test('renders download link for supported sites', () => {
  const params = {
    data: { siteName: 'keepfrds', siteid: '123456' }
  }
  const container = render(operationRender(params))
  expect(container.innerHTML).toContain('下载种子')
})
```

---

## 扩展建议

### 1. 添加状态管理（Zustand）

```typescript
// store.ts
import { create } from 'zustand'

const useStore = create(set => ({
  data: [],
  loading: false,
  error: null,

  fetchData: async () => {
    set({ loading: true, error: null })
    try {
      const response = await fetch('/top1000.json')
      const json = await response.json()
      set({ data: json.items, loading: false })
    }
    catch (error) {
      set({ error: error.message, loading: false })
    }
  },
}))
```

### 2. 添加错误处理

```typescript
// src/main.ts
async function onGridReady(event: GridReadyEvent<DataType>) {
  try {
    const response = await fetch('/top1000.json')
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const json: ResDataType = await response.json()
    event.api?.setGridOption('rowData', json.items)
  }
  catch (error) {
    console.error('加载数据失败:', error)
    alert('数据加载失败，请刷新重试')
  }
}
```

### 3. 添加加载动画

```typescript
const gridOptions: GridOptions<DataType> = {
  loadingOverlayComponent: CustomLoadingOverlay,
  loadingOverlayComponentParams: {
    loadingMessage: '数据加载中...',
  },
}

// 自定义加载组件
function CustomLoadingOverlay() {
  return '<div class="ag-overlay-loading-center">数据加载中...</div>'
}
```

### 4. 添加刷新按钮

```typescript
// 定时刷新（5分钟）
setInterval(() => {
  gridApi.purgeServerSideCache() // 清除缓存
  fetchData() // 重新加载
}, 5 * 60 * 1000)

// 手动刷新
function onRefresh() {
  fetchData()
}
```

### 5. 保存用户偏好

```typescript
// 保存列状态
function saveColumnState() {
  const state = gridApi.getColumnState()
  localStorage.setItem('columnState', JSON.stringify(state))
}

// 恢复列状态
function restoreColumnState() {
  const state = JSON.parse(localStorage.getItem('columnState') || '[]')
  gridApi.applyColumnState({ state, applyOrder: true })
}

// 窗口关闭前保存
window.addEventListener('beforeunload', saveColumnState)
```

---

## 相关文件

### 源代码

- `src/main.ts` - 入口文件（AG Grid配置）
- `src/types.d.ts` - TypeScript类型定义
- `src/utils/index.ts` - 工具函数（fetch、大小转换）
- `src/utils/operationRender.ts` - 操作列渲染器
- `src/utils/config.ts` - 站点URL配置
- `src/utils/iyuuSites.ts` - 站点元数据（118个）
- `src/index.css` - 全局样式

### 配置文件

- `index.html` - HTML模板
- `package.json` - NPM依赖
- `pnpm-lock.yaml` - PNPM锁定文件
- `vite.config.ts` - Vite构建配置
- `eslint.config.js` - ESLint配置

### 构建产物

- `../web-dist/` - 构建输出目录
  - `index.html` - 压缩后的HTML
  - `js/` - JS文件（带hash）
  - `css/` - CSS文件（带hash）
  - `assets/` - 其他资源

---

## 性能优化

### 当前优化

1. **虚拟滚动**：只渲染可见行
2. **代码分割**：vendor和main分开
3. **资源压缩**：HTML/CSS/JS都压缩
4. **长期缓存**：文件名带hash，CDN友好

### 可进一步优化

1. **CDN加速**：将web-dist部署到CDN
2. **Gzip压缩**：服务器启用Gzip
3. **图片懒加载**：如有图片资源
4. **预加载**：使用`<link rel="preload">`

---

**总结**：前端就一个页面，使用AG Grid企业版，功能完整。测试确实未编写，但就这点功能，写测试有些过度设计。

**更新**: 2026-01-11
**代码质量**: B+ 级（没测试，但代码已优化）
**技术栈**: Vite + AG Grid企业版
**最近优化**: 移除过时注释 + 添加空值检查 + 优化导入顺序 + 改进错误处理
