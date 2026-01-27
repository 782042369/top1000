# Web 模块

[根目录](../CLAUDE.md) > **web**

## 模块职责

Web 模块是 Top1000 系统的前端应用，提供用户界面和数据展示。采用 TypeScript + Vite + AG Grid 技术栈，实现简洁高效的数据浏览体验。

## 入口与启动

- **入口文件**: `src/main.ts`
- **HTML 模板**: `index.html`
- **构建输出**: `dist/`（开发时）→ `../web-dist/`（生产构建）

### 开发服务器

```bash
cd web
pnpm install
pnpm dev  # 访问 http://localhost:5173
```

### 生产构建

```bash
cd web
pnpm build  # 输出到 ../web-dist/
```

## 技术栈

### 核心框架
- **TypeScript 5.9** - 类型安全
- **Vite 8** - 构建工具和开发服务器
- **AG Grid Community 35** - 表格组件

### 开发工具
- **ESLint** - 代码检查（@antfu/eslint-config）
- **Prettier** - 代码格式化
- **pnpm 10** - 包管理器

### 构建优化
- **@xiaowaibuzheng/rolldown-vite-split-chunks** - 代码分割
- **vite-plugin-html** - HTML 模板处理

## 项目结构

```
web/
├── src/
│   ├── main.ts              # 应用入口
│   ├── types.d.ts           # TypeScript 类型定义
│   ├── gridConfig.ts        # AG Grid 配置
│   ├── utils/
│   │   ├── index.ts         # 工具函数导出
│   │   ├── config.ts        # 站点配置加载
│   │   └── operationRender.ts  # 操作列渲染器
│   ├── index.css            # 全局样式
│   └── vite-env.d.ts        # Vite 类型声明
├── index.html               # HTML 模板
├── vite.config.ts           # Vite 配置
├── tsconfig.json            # TypeScript 配置
├── .prettierrc              # Prettier 配置
├── package.json             # 依赖配置
└── pnpm-lock.yaml           # 锁定文件
```

## 核心功能

### 1. 表格展示

使用 AG Grid 实现高性能数据表格：
- **序号列** - 固定在左侧，显示行号
- **名字列** - 站点名称，支持文本过滤
- **资源 ID 列** - 站点 ID
- **重复度列** - 数值排序
- **文件大小列** - 大小排序（智能转换为 KB）
- **操作列** - 固定在右侧，提供快捷操作

### 2. 站点配置动态加载

启动时预加载站点配置：
```typescript
async function initApp() {
  await loadSitesConfig()  // 预加载站点配置
  initGrid()               // 初始化表格
}
```

### 3. 数据获取

```typescript
export async function fetchData() {
  const response = await fetch('/top1000.json')
  const data: ResDataType = await response.json()
  // 更新表格数据...
}
```

### 4. 中文本地化

```typescript
import { AG_GRID_LOCALE_CN } from '@ag-grid-community/locale'

const gridOptions: GridOptions<DataType> = {
  localeText: AG_GRID_LOCALE_CN,
  // ...
}
```

## 关键依赖与配置

### 依赖包

```json
{
  "dependencies": {
    "@ag-grid-community/locale": "^35.0.0",
    "ag-grid-community": "^35.0.0"
  },
  "devDependencies": {
    "@antfu/eslint-config": "^6.7.3",
    "typescript": "^5.9.3",
    "vite": "8.0.0-beta.5",
    "vite-plugin-html": "^3.2.2"
  }
}
```

### 环境要求

| 工具 | 版本 | 必需 |
|------|------|------|
| Node.js | >=24.3.0 | 是 |
| pnpm | >=10.12.4 | 是 |

### Vite 配置

```typescript
export default defineConfig({
  plugins: [
    splitChunks(),  // 代码分割
    createHtmlPlugin({ minify: true }),
  ],
  build: {
    rollupOptions: {
      output: {
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: '[ext]/[name]-[hash].[ext]',
      },
    },
    emptyOutDir: true,
    chunkSizeWarningLimit: 1000,
  },
  server: {
    open: true,
    host: '0.0.0.0',
    proxy: {
      '/top1000.json': 'http://127.0.0.1:7066',
      '/sites.json': 'http://127.0.0.1:7066',
    },
  },
})
```

## 数据模型

### DataType（站点数据）

```typescript
interface DataType {
  siteName: string      // 站点名称
  siteid: string        // 资源 ID
  duplication: string   // 重复度
  size: string          // 文件大小
  id: number            // 序号
}
```

### ResDataType（API 响应）

```typescript
interface ResDataType {
  items: DataType[]     // 种子列表
  time: string          // 更新时间
}
```

### SitesConfig（站点配置）

```typescript
interface SitesConfig {
  [key: string]: {
    url: string         // 站点 URL
    name?: string       // 站点名称（可选）
  }
}
```

## 开发指南

### 本地开发

```bash
# 1. 安装依赖
cd web
pnpm install

# 2. 启动开发服务器
pnpm dev

# 3. 访问 http://localhost:5173
#    - 前端运行在 5173 端口
#    - API 请求代理到后端 7066 端口
```

### 代码检查

```bash
# ESLint 检查并修复
pnpm lint

# TypeScript 类型检查
pnpm build  # 构建时自动检查
```

### 代码格式化

```bash
# Prettier 格式化（集成在 ESLint 中）
pnpm lint
```

### 生产构建

```bash
# 构建到 ../web-dist/
pnpm build

# 构建产物
# ../web-dist/
# ├── index.html
# ├── js/
# │   ├── main-xxx.js
# │   └── ...
# └── assets/
#     └── ...
```

## 测试与质量

### 当前状态
- 无单元测试
- 无集成测试
- 依赖手动测试

### 测试建议

**单元测试文件**: `src/utils/index.test.ts`

```typescript
describe('convertSizeToKb', () => {
  it('should convert GB to KB', () => {
    expect(convertSizeToKb('1.5 GB')).toBe(1.5 * 1024 * 1024)
  })

  it('should handle MB', () => {
    expect(convertSizeToKb('500 MB')).toBe(500 * 1024)
  })

  it('should handle invalid format', () => {
    expect(convertSizeToKb('invalid')).toBe(0)
  })
})

describe('loadSitesConfig', () => {
  it('should load sites config', async () => {
    await loadSitesConfig()
    expect(getSitesConfig()).toBeDefined()
  })
})
```

### E2E 测试建议

使用 Playwright 或 Cypress：
```typescript
test('display table data', async ({ page }) => {
  await page.goto('http://localhost:5173')
  await expect(page.locator('.ag-root-wrapper')).toBeVisible()
  await expect(page.locator('.ag-row')).toHaveCount(1000)
})
```

## 性能优化

### 已实现优化

1. **按需导入** - 仅导入 AG Grid 必要模块
2. **代码分割** - 使用 rolldown-vite-split-chunks
3. **虚拟滚动** - AG Grid 内置虚拟滚动（rowBuffer: 10）
4. **文件哈希** - 内容变化时 URL 变化，利用长缓存
5. **压缩传输** - Vite 自动压缩，Gzip/Brotli

### 可优化项

1. **预加载关键资源** - 使用 `<link rel="modulepreload">`
2. **CDN 加速** - 静态资源上传到 CDN
3. **图片优化** - 使用 WebP 格式（如果有图片）
4. **Service Worker** - 缓存静态资源，离线访问
5. **懒加载** - 操作列渲染器可以懒加载

## 相关文件清单

### 核心文件
- `src/main.ts` - 应用入口（90 行）
- `src/types.d.ts` - 类型定义（29 行）
- `src/gridConfig.ts` - 表格配置（90 行）
- `src/utils/index.ts` - 工具函数
- `src/utils/config.ts` - 站点配置加载
- `src/utils/operationRender.ts` - 操作列渲染器
- `src/index.css` - 全局样式
- `index.html` - HTML 模板
- `vite.config.ts` - Vite 配置
- `package.json` - 依赖配置

### 测试文件（待创建）
- `src/utils/index.test.ts` - 工具函数测试
- `src/e2e/` - E2E 测试

## 常见问题

### Q: 前端开发时 API 请求失败？
确保后端服务运行在 7066 端口，检查 Vite proxy 配置。

### Q: AG Grid 不显示数据？
1. 检查 `/top1000.json` 是否返回数据
2. 打开浏览器控制台查看错误
3. 验证 `columnDefs` 配置是否正确

### Q: 如何添加新列？
编辑 `src/gridConfig.ts`：
```typescript
export const columnDefs = [
  // 现有列...
  {
    headerName: '新列',
    field: 'newField',
    width: 100,
  },
]
```

### Q: 如何自定义样式？
编辑 `src/index.css` 或使用 AG Grid 主题：
```typescript
import { themeAlpine, themeBalham, themeQuartz } from 'ag-grid-community'

const gridOptions = {
  theme: themeAlpine,  // 切换主题
}
```

### Q: 构建产物过大？
检查 `chunkSizeWarningLimit`，考虑代码分割：
```typescript
// vite.config.ts
build: {
  chunkSizeWarningLimit: 1000,
  rollupOptions: {
    output: {
      manualChunks: {
        'ag-grid': ['ag-grid-community'],
      },
    },
  },
}
```

## 扩展建议

### 功能扩展

1. **数据导出** - 导出为 CSV/Excel
2. **高级过滤** - 自定义过滤条件
3. **列管理** - 显示/隐藏列，列顺序调整
4. **数据刷新** - 手动刷新按钮，自动刷新间隔
5. **主题切换** - 亮色/暗色主题

### 技术升级

1. **Vue/React 集成** - 如果需要更复杂的交互
2. **状态管理** - 使用 Pinia/Zustand 管理状态
3. **组件库** - 添加 Naive UI/Ant Design Vue
4. **国际化** - 使用 vue-i18n 支持多语言

---

**最后更新**: 2026-01-27
**代码行数**: ~300 行（src/）
**维护状态**: 活跃
